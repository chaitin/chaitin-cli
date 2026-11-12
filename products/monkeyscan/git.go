package monkeyscan

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func collectDiff(ctx context.Context, opts reviewScope) (*diffSnapshot, error) {
	repoRoot, err := gitOutput(ctx, "", "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("当前目录不是 Git 仓库或无法读取仓库根目录: %w", err)
	}
	repoRoot = strings.TrimSpace(repoRoot)
	remoteURL := resolveRemoteURL(ctx, repoRoot)
	branch, _ := gitOutput(ctx, repoRoot, "branch", "--show-current")
	headSHA, _ := gitOutput(ctx, repoRoot, "rev-parse", "HEAD")
	s := &diffSnapshot{
		RepoRoot:      repoRoot,
		RemoteURL:     remoteURL,
		CurrentBranch: strings.TrimSpace(branch),
		HeadSHA:       strings.TrimSpace(headSHA),
		Scope:         scopeName(opts),
	}
	if opts.All {
		staged, stagedFiles, err := diffForMode(ctx, repoRoot, "staged", "", "")
		if err != nil {
			return nil, err
		}
		unstaged, unstagedFiles, err := diffForMode(ctx, repoRoot, "uncommitted", "", "")
		if err != nil {
			return nil, err
		}
		s.Diff = joinDiffs(staged, unstaged)
		s.Files = mergeFiles(stagedFiles, unstagedFiles)
		return s, nil
	}
	if opts.Type == "committed" {
		baseSHA := opts.BaseCommit
		baseBranch := opts.Base
		if opts.Base != "" {
			out, err := gitOutput(ctx, repoRoot, "merge-base", "HEAD", opts.Base)
			if err != nil {
				return nil, fmt.Errorf("计算 merge-base 失败: %w", err)
			}
			baseSHA = strings.TrimSpace(out)
		}
		s.BaseSHA = baseSHA
		s.BaseBranch = baseBranch
		diff, files, err := diffForMode(ctx, repoRoot, "committed", baseSHA, "")
		if err != nil {
			return nil, err
		}
		s.Diff = diff
		s.Files = files
		return s, nil
	}
	diff, files, err := diffForMode(ctx, repoRoot, opts.Type, "", "")
	if err != nil {
		return nil, err
	}
	s.Diff = diff
	s.Files = files
	return s, nil
}

func resolveRemoteURL(ctx context.Context, repoRoot string) string {
	if remoteURL, err := gitOutput(ctx, repoRoot, "remote", "get-url", "origin"); err == nil {
		return strings.TrimSpace(remoteURL)
	}
	remotes, err := gitOutput(ctx, repoRoot, "remote")
	if err != nil {
		return ""
	}
	for _, remote := range strings.Split(remotes, "\n") {
		remote = strings.TrimSpace(remote)
		if remote == "" {
			continue
		}
		if remoteURL, err := gitOutput(ctx, repoRoot, "remote", "get-url", remote); err == nil {
			return strings.TrimSpace(remoteURL)
		}
	}
	return ""
}

func diffForMode(ctx context.Context, repoRoot, mode, base, head string) (string, []changedFile, error) {
	diffArgs := []string{"diff", "--no-ext-diff", "--src-prefix=a/", "--dst-prefix=b/"}
	statArgs := []string{"diff", "--numstat"}
	nameArgs := []string{"diff", "--name-status"}
	switch mode {
	case "staged":
		diffArgs = append(diffArgs, "--cached", "--", ".")
		statArgs = append(statArgs, "--cached", "--", ".")
		nameArgs = append(nameArgs, "--cached", "--", ".")
	case "uncommitted":
		diffArgs = append(diffArgs, "--", ".")
		statArgs = append(statArgs, "--", ".")
		nameArgs = append(nameArgs, "--", ".")
	case "committed":
		rev := base + "..HEAD"
		if head != "" {
			rev = base + ".." + head
		}
		diffArgs = append(diffArgs, rev, "--", ".")
		statArgs = append(statArgs, rev, "--", ".")
		nameArgs = append(nameArgs, rev, "--", ".")
	default:
		return "", nil, fmt.Errorf("unsupported diff mode %s", mode)
	}
	diff, err := gitOutput(ctx, repoRoot, diffArgs...)
	if err != nil {
		return "", nil, fmt.Errorf("采集 %s diff 失败: %w", mode, err)
	}
	numstat, err := gitOutput(ctx, repoRoot, statArgs...)
	if err != nil {
		return "", nil, fmt.Errorf("采集 %s 文件统计失败: %w", mode, err)
	}
	nameStatus, err := gitOutput(ctx, repoRoot, nameArgs...)
	if err != nil {
		return "", nil, fmt.Errorf("采集 %s 文件列表失败: %w", mode, err)
	}
	files := parseChangedFiles(nameStatus, numstat, splitPatches(diff))
	if mode == "uncommitted" {
		untrackedDiff, untrackedFiles, err := collectUntrackedFiles(ctx, repoRoot)
		if err != nil {
			return "", nil, err
		}
		diff = joinDiffs(diff, untrackedDiff)
		files = mergeFiles(files, untrackedFiles)
	}
	return diff, files, nil
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		return "", err
	}
	return string(out), nil
}

func gitOutputAllowExit(ctx context.Context, dir string, allowed map[int]bool, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && allowed[exitErr.ExitCode()] {
			return string(out), nil
		}
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		return "", err
	}
	return string(out), nil
}

func collectUntrackedFiles(ctx context.Context, repoRoot string) (string, []changedFile, error) {
	raw, err := gitOutput(ctx, repoRoot, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return "", nil, fmt.Errorf("采集未跟踪文件列表失败: %w", err)
	}
	if raw == "" {
		return "", nil, nil
	}
	var patches []string
	var files []changedFile
	for _, path := range strings.Split(raw, "\x00") {
		if path == "" {
			continue
		}
		cleanPath := filepath.ToSlash(path)
		if isIgnoredLocalReviewPath(cleanPath) {
			continue
		}
		patch, err := gitOutputAllowExit(ctx, repoRoot, map[int]bool{1: true}, "diff", "--no-ext-diff", "--src-prefix=a/", "--dst-prefix=b/", "--no-index", "--", "/dev/null", cleanPath)
		if err != nil {
			return "", nil, fmt.Errorf("采集未跟踪文件 %s diff 失败: %w", cleanPath, err)
		}
		patches = append(patches, patch)
		additions, deletions := countPatchChanges(patch)
		files = append(files, changedFile{
			Path:      cleanPath,
			Status:    "added",
			Additions: additions,
			Deletions: deletions,
			Changes:   additions + deletions,
			Patch:     patch,
		})
	}
	return joinDiffs(patches...), files, nil
}

func isIgnoredLocalReviewPath(path string) bool {
	return path == ".monkeyscan" || strings.HasPrefix(path, ".monkeyscan/")
}

func parseChangedFiles(nameStatus, numstat string, patches map[string]string) []changedFile {
	stats := parseNumstat(numstat)
	var files []changedFile
	for _, line := range strings.Split(strings.TrimSpace(nameStatus), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		statusToken := parts[0]
		status := statusName(statusToken)
		path := ""
		var oldPath *string
		if strings.HasPrefix(statusToken, "R") || strings.HasPrefix(statusToken, "C") {
			if len(parts) >= 3 {
				old := parts[1]
				oldPath = &old
				path = parts[2]
			}
		} else if len(parts) >= 2 {
			path = parts[1]
		}
		if path == "" {
			continue
		}
		file := changedFile{Path: path, OldPath: oldPath, Status: status, Patch: patches[path]}
		if st, ok := stats[path]; ok {
			file.Additions = st.additions
			file.Deletions = st.deletions
			file.Changes = st.additions + st.deletions
		}
		files = append(files, file)
	}
	return files
}

type fileStat struct {
	additions int
	deletions int
}

func parseNumstat(raw string) map[string]fileStat {
	out := map[string]fileStat{}
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		additions, _ := strconv.Atoi(parts[0])
		deletions, _ := strconv.Atoi(parts[1])
		path := parts[len(parts)-1]
		out[path] = fileStat{additions: additions, deletions: deletions}
	}
	return out
}

func splitPatches(diff string) map[string]string {
	patches := map[string]string{}
	const marker = "diff --git "
	chunks := strings.Split(diff, marker)
	for _, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}
		patch := marker + chunk
		firstLine, _, _ := strings.Cut(patch, "\n")
		path := diffNewPath(firstLine)
		if path == "" {
			continue
		}
		patches[path] = patch
	}
	return patches
}

func diffNewPath(firstLine string) string {
	const marker = " b/"
	idx := strings.LastIndex(firstLine, marker)
	if idx < 0 {
		return ""
	}
	path := strings.TrimSpace(firstLine[idx+len(marker):])
	return strings.Trim(path, "\"")
}

func countPatchChanges(patch string) (int, int) {
	additions, deletions := 0, 0
	for _, line := range strings.Split(patch, "\n") {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			continue
		case strings.HasPrefix(line, "+"):
			additions++
		case strings.HasPrefix(line, "-"):
			deletions++
		}
	}
	return additions, deletions
}

func statusName(token string) string {
	switch {
	case strings.HasPrefix(token, "A"):
		return "added"
	case strings.HasPrefix(token, "D"):
		return "deleted"
	case strings.HasPrefix(token, "R"):
		return "renamed"
	case strings.HasPrefix(token, "C"):
		return "copied"
	default:
		return "modified"
	}
}

func scopeName(opts reviewScope) string {
	if opts.All {
		return "all"
	}
	if opts.Type == "committed" && opts.Base != "" {
		return "committed:" + opts.Base
	}
	if opts.Type == "committed" && opts.BaseCommit != "" {
		return "committed:" + opts.BaseCommit
	}
	return opts.Type
}

func joinDiffs(parts ...string) string {
	var b strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(part)
		b.WriteString("\n")
	}
	return b.String()
}

func mergeFiles(groups ...[]changedFile) []changedFile {
	byPath := map[string]changedFile{}
	order := []string{}
	for _, group := range groups {
		for _, file := range group {
			if _, ok := byPath[file.Path]; !ok {
				order = append(order, file.Path)
				byPath[file.Path] = file
				continue
			}
			existing := byPath[file.Path]
			existing.Additions += file.Additions
			existing.Deletions += file.Deletions
			existing.Changes += file.Changes
			if file.Patch != "" {
				existing.Patch = joinDiffs(existing.Patch, file.Patch)
			}
			byPath[file.Path] = existing
		}
	}
	files := make([]changedFile, 0, len(order))
	for _, path := range order {
		files = append(files, byPath[path])
	}
	return files
}
