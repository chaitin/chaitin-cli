package monkeyscan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func printDryRun(cmd *cobra.Command, snapshot *diffSnapshot) {
	fmt.Fprintln(cmd.OutOrStdout(), "MonkeyScan Review Dry Run")
	fmt.Fprintf(cmd.OutOrStdout(), "服务地址: %s\n", normalizedURL(runtimeCfg.URL))
	fmt.Fprintf(cmd.OutOrStdout(), "仓库: %s\n", snapshot.RemoteURL)
	fmt.Fprintf(cmd.OutOrStdout(), "分支: %s\n", firstNonEmpty(snapshot.CurrentBranch, "(detached HEAD)"))
	fmt.Fprintf(cmd.OutOrStdout(), "HEAD: %s\n", snapshot.HeadSHA)
	fmt.Fprintf(cmd.OutOrStdout(), "检查范围: %s\n", snapshot.Scope)
	fmt.Fprintf(cmd.OutOrStdout(), "Diff 字节数: %d\n", len(snapshot.Diff))
	fmt.Fprintf(cmd.OutOrStdout(), "涉及文件数: %d\n", len(snapshot.Files))
	for _, file := range snapshot.Files {
		fmt.Fprintf(cmd.OutOrStdout(), "- %s %s (+%d/-%d)\n", file.Status, file.Path, file.Additions, file.Deletions)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "dry-run 未上传代码、未创建任务、未生成 Review 结果。")
}

func printReviewSummary(cmd *cobra.Command, detail *reviewDetail, reviewPath string) {
	fmt.Fprintln(cmd.OutOrStdout(), "MonkeyScan Review 结果")
	fmt.Fprintf(cmd.OutOrStdout(), "Run ID: %s\n", detail.Run.ID)
	fmt.Fprintf(cmd.OutOrStdout(), "状态: %s\n", firstNonEmpty(detail.Run.Status, detail.Run.TaskGroupStatus))
	fmt.Fprintf(cmd.OutOrStdout(), "仓库: %s\n", detail.Run.RepositoryName)
	fmt.Fprintf(cmd.OutOrStdout(), "发现问题: %d\n", len(detail.Findings))
	fmt.Fprintf(cmd.OutOrStdout(), "结果文件: %s\n", reviewPath)
	fmt.Fprintln(cmd.OutOrStdout(), "提示: .monkeyscan 可能包含代码 diff 和修复建议，请不要提交到仓库。")
}

func printScanCreated(cmd *cobra.Command, resp *scanCreateResponse) {
	fmt.Fprintln(cmd.OutOrStdout(), "MonkeyScan 扫描任务已创建")
	fmt.Fprintf(cmd.OutOrStdout(), "Task Group ID: %s\n", resp.TaskGroupID)
	fmt.Fprintf(cmd.OutOrStdout(), "状态: %s\n", resp.Status)
	fmt.Fprintf(cmd.OutOrStdout(), "扫描对象: %s\n", scanTargetLabel(resp.ScanTarget))
	for _, next := range resp.NextCommands {
		fmt.Fprintf(cmd.OutOrStdout(), "后续查询: %s\n", next)
	}
}

func printScanList(cmd *cobra.Command, resp *scanListResponse) {
	fmt.Fprintln(cmd.OutOrStdout(), "最近 7 天全量扫描任务")
	if len(resp.Items) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "暂无扫描任务。")
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%-36s  %-14s  %-24s  %-8s  %-20s\n", "TASK_GROUP_ID", "STATUS", "TARGET", "DEFECTS", "CREATED_AT")
	for _, item := range resp.Items {
		fmt.Fprintf(
			cmd.OutOrStdout(),
			"%-36s  %-14s  %-24s  %-8d  %-20s\n",
			item.TaskGroupID,
			item.Status,
			truncate(scanTargetLabel(item.ScanTarget), 24),
			item.DefectCount,
			formatTime(item.CreatedAt),
		)
	}
}

func outputScanResult(cmd *cobra.Command, result *scanResultResponse, output string) error {
	content := renderScanResultMarkdown(result)
	if strings.TrimSpace(output) != "" {
		if err := os.WriteFile(output, []byte(content), 0o600); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "扫描结果已写入: %s\n", output)
		return nil
	}
	fmt.Fprint(cmd.OutOrStdout(), content)
	return nil
}

func renderScanResultMarkdown(result *scanResultResponse) string {
	var b strings.Builder
	b.WriteString("# MonkeyScan 扫描结果\n\n")
	b.WriteString("## 基础信息\n\n")
	fmt.Fprintf(&b, "- Task Group ID: %s\n", result.Task.TaskGroupID)
	fmt.Fprintf(&b, "- 状态: %s\n", result.Task.Status)
	fmt.Fprintf(&b, "- 扫描对象: %s\n", scanTargetLabel(result.Task.ScanTarget))
	fmt.Fprintf(&b, "- 创建时间: %s\n", formatTime(result.Task.CreatedAt))
	fmt.Fprintf(&b, "- 更新时间: %s\n\n", formatTime(result.Task.UpdatedAt))
	b.WriteString("## 统计摘要\n\n")
	fmt.Fprintf(&b, "- 漏洞总数: %d\n", result.Total)
	fmt.Fprintf(&b, "- Critical: %d\n", result.Severity.Critical)
	fmt.Fprintf(&b, "- High: %d\n", result.Severity.High)
	fmt.Fprintf(&b, "- Medium: %d\n", result.Severity.Medium)
	fmt.Fprintf(&b, "- Low: %d\n\n", result.Severity.Low)
	if result.Truncated {
		fmt.Fprintf(&b, "> 结果数量较多，本次仅展示前 %d 条。请在 MonkeyScan 页面查看完整结果。\n\n", len(result.Items))
	}
	b.WriteString("## 漏洞列表\n\n")
	if len(result.Items) == 0 {
		b.WriteString("未发现漏洞。\n")
		return b.String()
	}
	for i, item := range result.Items {
		title := firstNonEmpty(item.DefectNameZh, item.DefectName, item.RuleName, item.ID)
		fmt.Fprintf(&b, "### %d. %s\n\n", i+1, title)
		fmt.Fprintf(&b, "- 严重性: %s\n", item.Severity)
		if item.FilePath != "" {
			fmt.Fprintf(&b, "- 位置: %s:%d\n", item.FilePath, item.Line)
		}
		if item.RuleName != "" {
			fmt.Fprintf(&b, "- 规则: %s\n", item.RuleName)
		}
		fmt.Fprintf(&b, "- 验证状态: %d\n", item.AiVerifyStatus)
		if !result.Full {
			b.WriteString("\n")
			continue
		}
		if text := firstNonEmpty(item.MessageZh, item.Message, item.CheckerMessage); text != "" {
			fmt.Fprintf(&b, "\n**描述**\n\n%s\n", text)
		}
		if item.CodeSnippet != "" {
			fmt.Fprintf(&b, "\n**代码片段**\n\n```text\n%s\n```\n", item.CodeSnippet)
		}
		if text := firstNonEmpty(item.Recommendation, item.AiSuggestion); text != "" {
			fmt.Fprintf(&b, "\n**修复建议**\n\n%s\n", text)
		}
		if item.AiVerification != "" {
			fmt.Fprintf(&b, "\n**验证/降噪结论**\n\n%s\n", item.AiVerification)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func scanTargetLabel(target scanTarget) string {
	switch target.Type {
	case "git":
		if target.Ref != "" {
			return fmt.Sprintf("%s@%s", firstNonEmpty(target.URL, target.Name), target.Ref)
		}
		return firstNonEmpty(target.URL, target.Name)
	default:
		return firstNonEmpty(target.Name, target.URL, target.Type)
	}
}

func truncate(value string, width int) string {
	if len(value) <= width {
		return value
	}
	if width <= 3 {
		return value[:width]
	}
	return value[:width-3] + "..."
}

func writeReviewMarkdown(path string, detail *reviewDetail) error {
	var b strings.Builder
	b.WriteString("# MonkeyScan Review 结果\n\n")
	b.WriteString("## 基础信息\n\n")
	fmt.Fprintf(&b, "- Run ID: %s\n", detail.Run.ID)
	fmt.Fprintf(&b, "- 仓库: %s\n", detail.Run.RepositoryName)
	fmt.Fprintf(&b, "- 状态: %s\n", firstNonEmpty(detail.Run.Status, detail.Run.TaskGroupStatus))
	fmt.Fprintf(&b, "- Base: %s\n", detail.Run.BaseBranch)
	fmt.Fprintf(&b, "- Head: %s %s\n", detail.Run.HeadBranch, detail.Run.HeadSHA)
	fmt.Fprintf(&b, "- 发布时间: %s\n\n", formatTime(detail.Run.UpdatedAt))
	b.WriteString("## 总结\n\n")
	fmt.Fprintf(&b, "- 问题数量: %d\n", len(detail.Findings))
	fmt.Fprintf(&b, "- 评论数量: %d\n", len(detail.Comments))
	if detail.Run.ErrorMessage != "" {
		fmt.Fprintf(&b, "- 错误信息: %s\n", sanitizeErrorMessage(detail.Run.ErrorMessage))
	}
	b.WriteString("\n## 问题列表\n\n")
	if len(detail.Findings) == 0 {
		b.WriteString("未发现问题。\n\n")
	}
	for i, finding := range detail.Findings {
		fmt.Fprintf(&b, "### %d. %s\n\n", i+1, firstNonEmpty(finding.Title, "未命名问题"))
		fmt.Fprintf(&b, "- 严重性: %s\n", finding.Severity)
		fmt.Fprintf(&b, "- 类别: %s\n", finding.Category)
		fmt.Fprintf(&b, "- 置信度: %s\n", finding.Confidence)
		if finding.Location.Path != "" {
			fmt.Fprintf(&b, "- 位置: %s:%d-%d\n", finding.Location.Path, finding.Location.StartLine, finding.Location.EndLine)
		}
		b.WriteString("\n")
		if finding.Description != "" {
			b.WriteString("**描述**\n\n")
			b.WriteString(finding.Description)
			b.WriteString("\n\n")
		}
		if finding.ProblemCode != "" {
			b.WriteString("**问题代码**\n\n```text\n")
			b.WriteString(finding.ProblemCode)
			b.WriteString("\n```\n\n")
		}
		if finding.Recommendation != "" {
			b.WriteString("**修复建议**\n\n")
			b.WriteString(finding.Recommendation)
			b.WriteString("\n\n")
		}
		b.WriteString("**推荐修复 Diff**\n\n```diff\n")
		b.WriteString(firstNonEmpty(finding.SuggestedDiff, finding.RecommendedDiff, finding.FixDiff, finding.SuggestedPatch, "无"))
		b.WriteString("\n```\n\n")
	}
	b.WriteString("## Comments\n\n")
	if len(detail.Comments) == 0 {
		b.WriteString("无。\n")
	}
	for i, comment := range detail.Comments {
		fmt.Fprintf(&b, "### Comment %d\n\n", i+1)
		fmt.Fprintf(&b, "- 类型: %s\n", comment.CommentType)
		fmt.Fprintf(&b, "- 发布状态: %s\n", comment.PublishStatus)
		if comment.ErrorMessage != "" {
			fmt.Fprintf(&b, "- 说明: %s\n", comment.ErrorMessage)
		}
		if comment.Body != "" {
			b.WriteString("\n")
			b.WriteString(comment.Body)
			b.WriteString("\n\n")
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func writeLocalStatus(runDir string, status localStatus) error {
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(runDir, "status.json"), append(data, '\n'), 0o600)
}

func findLocalStatus(repoRoot, run string) (localStatus, string, error) {
	base := filepath.Join(repoRoot, reviewsDir)
	if run != "" {
		candidate := run
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(base, run)
		}
		if st, err := readLocalStatus(filepath.Join(candidate, "status.json")); err == nil {
			return st, candidate, nil
		}
		if st, dir, err := findStatusByServerRunID(base, run); err == nil {
			return st, dir, nil
		}
		return localStatus{RunID: run, ClientRunID: run}, filepath.Join(base, run), nil
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		return localStatus{}, "", fmt.Errorf("未找到本地 Review 状态目录")
	}
	type candidate struct {
		status localStatus
		dir    string
	}
	var candidates []candidate
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(base, entry.Name())
		st, err := readLocalStatus(filepath.Join(dir, "status.json"))
		if err != nil {
			continue
		}
		if !isTerminalReviewStatus(st.Status, "") {
			candidates = append(candidates, candidate{status: st, dir: dir})
		}
	}
	if len(candidates) == 0 {
		return localStatus{}, "", fmt.Errorf("未找到未完成的 Review，本地可用 --run 指定任务")
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].status.UpdatedAt.After(candidates[j].status.UpdatedAt)
	})
	return candidates[0].status, candidates[0].dir, nil
}

func findStatusByServerRunID(base, runID string) (localStatus, string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return localStatus{}, "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(base, entry.Name())
		st, err := readLocalStatus(filepath.Join(dir, "status.json"))
		if err == nil && st.RunID == runID {
			return st, dir, nil
		}
	}
	return localStatus{}, "", fmt.Errorf("not found")
}

func readLocalStatus(path string) (localStatus, error) {
	var st localStatus
	data, err := os.ReadFile(path)
	if err != nil {
		return st, err
	}
	err = json.Unmarshal(data, &st)
	return st, err
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "未知"
	}
	return t.Format(time.RFC3339)
}
