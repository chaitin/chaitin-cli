package tanswer

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type fileAlarmListOptions struct {
	time     string
	start    string
	end      string
	page     int
	pageSize int
	severity string
	tag      string
	fileType string
	srcIP    string
	destIP   string
	srcPort  string
	destPort string
	appProto string
	keyword  string
}

type fileAlarmDetailOptions struct {
	id string
}

type fileAlarmListRPCResult struct {
	Data      []map[string]any `json:"data"`
	Total     int64            `json:"total"`
	PageTotal int64            `json:"page_total"`
}

type fileAlarmDetailRPCResult struct {
	Data map[string]any `json:"data"`
}

func newFileAlarmCommand(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file-alarm",
		Short: "文件告警语义命令",
		Long:  "文件告警语义命令。用于只读查询恶意文件、Webshell、沙箱检测告警和单条详情；不触发样本下载或新的沙箱分析流程。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newFileAlarmOverviewCommand(opts))
	cmd.AddCommand(newFileAlarmListCommand(opts, "malicious", "查询恶意文件告警", "elf", "AlarmService.SearchAlarmFdetectList"))
	cmd.AddCommand(newFileAlarmListCommand(opts, "webshell", "查询 Webshell 告警", "webshell", "AlarmService.SearchAlarmFdetectList"))
	cmd.AddCommand(newFileAlarmListCommand(opts, "sandbox", "查询沙箱检测告警", "", "AlarmService.SearchSandboxAlarmFdetectList"))
	cmd.AddCommand(newFileAlarmDetailCommand(opts))
	return cmd
}

func newFileAlarmOverviewCommand(opts *RootOptions) *cobra.Command {
	var fileOpts fileAlarmListOptions
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "查看文件告警概览",
		Long:  "查看文件告警概览，用于值班巡检时快速判断当前是否存在恶意文件、Webshell 或沙箱检测风险。该命令返回文件告警总数和已有文件类型聚合，不返回原始告警摘要。\n\n输出：查询时间范围、实际筛选条件、文件告警总数、文件类型分布。",
		Example: "  chaitin-cli tanswer file-alarm overview --time today\n" +
			"  chaitin-cli tanswer file-alarm overview --time 24h --severity critical,high",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			rng, err := ParseTimeRange(TimeRangeOptions{Time: fileOpts.time, Start: fileOpts.start, End: fileOpts.end})
			if err != nil {
				return writeFileAlarmError(cmd, "查看文件告警概览", "chaitin-cli tanswer file-alarm overview", "INVALID_TIME_RANGE", err.Error(), false)
			}
			req := buildFileAlarmBaseRequest(rng, fileOpts)
			client := NewClient(cfg)
			count := map[string]any{}
			if err := client.CallRPC(cmd.Context(), "AlarmService.SearchAlarmFdetectCount", req, &count); err != nil {
				return writeFileAlarmError(cmd, "查看文件告警概览", "chaitin-cli tanswer file-alarm overview", "FILE_ALARM_OVERVIEW_FAILED", err.Error(), true)
			}
			aggReq := copyMap(req)
			aggReq["agg"] = "type"
			aggReq["top"] = 10
			agg := map[string]any{}
			if err := client.CallRPC(cmd.Context(), "AlarmService.SearchAlarmFdetectAggTop", aggReq, &agg); err != nil {
				return writeFileAlarmError(cmd, "查看文件告警概览", "chaitin-cli tanswer file-alarm overview", "FILE_ALARM_OVERVIEW_FAILED", err.Error(), true)
			}
			data := map[string]any{
				"summary": map[string]any{
					"file_alarm_total": count["total"],
				},
				"type_distribution": agg["data"],
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查看文件告警概览",
				Command: "chaitin-cli tanswer file-alarm overview",
				Query: map[string]any{
					"time_range": rng,
					"filters":    fileAlarmListFilters(fileOpts),
				},
				Data: data,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	addFileAlarmCommonFlags(cmd, &fileOpts)
	return cmd
}

func newFileAlarmListCommand(opts *RootOptions, use string, task string, defaultTag string, rpcMethod string) *cobra.Command {
	var fileOpts fileAlarmListOptions
	cmd := &cobra.Command{
		Use:   use,
		Short: task,
		Long:  fmt.Sprintf("%s，用于查询文件检测已有告警摘要。该命令返回发现时间、风险等级、文件名、文件类型、源/目的 IP、应用层协议、MD5、SHA256 和分页信息；字段无值时返回空。\n\n输出：查询时间范围、实际筛选条件、total、page、page_size、current_count、has_more、file_alarms。", task),
		Example: fmt.Sprintf("  chaitin-cli tanswer file-alarm %s --time today --page-size 10\n", use) +
			fmt.Sprintf("  chaitin-cli tanswer file-alarm %s --src-ip 198.51.100.10 --dest-ip 192.0.2.10", use),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(fileOpts.tag) == "" {
				fileOpts.tag = defaultTag
			}
			return runFileAlarmListCommand(cmd, opts, fileOpts, task, "chaitin-cli tanswer file-alarm "+use, rpcMethod)
		},
	}
	addFileAlarmListFlags(cmd, &fileOpts)
	return cmd
}

func newFileAlarmDetailCommand(opts *RootOptions) *cobra.Command {
	var fileOpts fileAlarmDetailOptions
	cmd := &cobra.Command{
		Use:     "detail",
		Short:   "查看文件告警详情",
		Long:    "查看文件告警详情，用于研判单条恶意文件、Webshell 或沙箱检测告警。该命令只读取已有详情，不下载原始样本，不触发新的深度分析。\n\n输出：doc_id、detail。detail 中保留后端文件告警详情字段，包括基础信息、检测结果、检测依据、文件内容片段或沙箱报告字段。",
		Example: "  chaitin-cli tanswer file-alarm detail --id '<doc_id>'",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(fileOpts.id) == "" {
				return writeFileAlarmError(cmd, "查看文件告警详情", "chaitin-cli tanswer file-alarm detail", "MISSING_FILE_ALARM_ID", "missing file alarm doc_id: set --id", false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var result fileAlarmDetailRPCResult
			if err := client.CallRPC(cmd.Context(), "AlarmService.GetAlarmFdetect", map[string]any{"doc_id": fileOpts.id}, &result); err != nil {
				return writeFileAlarmError(cmd, "查看文件告警详情", "chaitin-cli tanswer file-alarm detail", "FILE_ALARM_DETAIL_FAILED", err.Error(), true)
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查看文件告警详情",
				Command: "chaitin-cli tanswer file-alarm detail",
				Query:   map[string]any{"doc_id": fileOpts.id},
				Data: map[string]any{
					"doc_id": fileOpts.id,
					"detail": result.Data,
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&fileOpts.id, "id", "", "file alarm doc_id from file-alarm list commands")
	return cmd
}

func runFileAlarmListCommand(cmd *cobra.Command, opts *RootOptions, fileOpts fileAlarmListOptions, task string, command string, rpcMethod string) error {
	if fileOpts.page < 1 {
		return writeFileAlarmError(cmd, task, command, "INVALID_PAGE", "page must be greater than or equal to 1", false)
	}
	if fileOpts.pageSize < 1 || fileOpts.pageSize > 100 {
		return writeFileAlarmError(cmd, task, command, "INVALID_PAGE_SIZE", "page-size must be between 1 and 100", false)
	}
	cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
	if err != nil {
		return err
	}
	rng, err := ParseTimeRange(TimeRangeOptions{Time: fileOpts.time, Start: fileOpts.start, End: fileOpts.end})
	if err != nil {
		return writeFileAlarmError(cmd, task, command, "INVALID_TIME_RANGE", err.Error(), false)
	}
	req := buildFileAlarmListRequest(rng, fileOpts)
	client := NewClient(cfg)
	var result fileAlarmListRPCResult
	if err := client.CallRPC(cmd.Context(), rpcMethod, req, &result); err != nil {
		return writeFileAlarmError(cmd, task, command, "FILE_ALARM_LIST_FAILED", err.Error(), true)
	}
	data := map[string]any{
		"total":         result.Total,
		"page_total":    result.PageTotal,
		"page":          fileOpts.page,
		"page_size":     fileOpts.pageSize,
		"current_count": len(result.Data),
		"has_more":      int64(fileOpts.page*fileOpts.pageSize) < result.Total,
		"file_alarms":   summarizeFileAlarms(result.Data),
	}
	raw, err := RenderJSON(SuccessEnvelope{
		Success: true,
		Task:    task,
		Command: command,
		Query: map[string]any{
			"time_range": rng,
			"filters":    fileAlarmListFilters(fileOpts),
			"page":       fileOpts.page,
			"page_size":  fileOpts.pageSize,
		},
		Data: data,
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func addFileAlarmCommonFlags(cmd *cobra.Command, opts *fileAlarmListOptions) {
	cmd.Flags().StringVar(&opts.time, "time", "today", "time range: today, 24h, 7d")
	cmd.Flags().StringVar(&opts.start, "start", "", "custom start time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&opts.end, "end", "", "custom end time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&opts.severity, "severity", "", "severity filter: critical,high,medium,low or 1,2,3,4")
	cmd.Flags().StringVar(&opts.fileType, "file-type", "", "file alarm type filter, for example virus,backdoor,php_webshell")
	cmd.Flags().StringVar(&opts.keyword, "keyword", "", "keyword filter, comma separated")
}

func addFileAlarmListFlags(cmd *cobra.Command, opts *fileAlarmListOptions) {
	addFileAlarmCommonFlags(cmd, opts)
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().StringVar(&opts.tag, "tag", "", "file alarm tag filter: elf or webshell")
	cmd.Flags().StringVar(&opts.srcIP, "src-ip", "", "source IP filter, comma separated")
	cmd.Flags().StringVar(&opts.destIP, "dest-ip", "", "destination IP filter, comma separated")
	cmd.Flags().StringVar(&opts.srcPort, "src-port", "", "source port filter, comma separated")
	cmd.Flags().StringVar(&opts.destPort, "dest-port", "", "destination port filter, comma separated")
	cmd.Flags().StringVar(&opts.appProto, "app-proto", "", "application protocol filter, comma separated")
}

func buildFileAlarmListRequest(rng TimeRange, opts fileAlarmListOptions) map[string]any {
	if opts.page < 1 {
		opts.page = 1
	}
	if opts.pageSize < 1 {
		opts.pageSize = 10
	}
	req := buildFileAlarmBaseRequest(rng, opts)
	req["offset"] = int64((opts.page - 1) * opts.pageSize)
	req["count"] = int64(opts.pageSize)
	return req
}

func buildFileAlarmBaseRequest(rng TimeRange, opts fileAlarmListOptions) map[string]any {
	req := map[string]any{
		"time_range_start": rng.Start,
		"time_range_end":   rng.End,
	}
	for key, value := range fileAlarmListFilters(opts) {
		req[key] = value
	}
	return req
}

func fileAlarmListFilters(opts fileAlarmListOptions) map[string]any {
	filters := map[string]any{}
	addIntFilter(filters, "severity", opts.severity, severityValue)
	addStringFilter(filters, "tag", opts.tag, strings.ToLower)
	addStringFilter(filters, "type", opts.fileType, strings.ToLower)
	addStringFilter(filters, "src_ip", opts.srcIP, strings.TrimSpace)
	addStringFilter(filters, "dest_ip", opts.destIP, strings.TrimSpace)
	addStringFilter(filters, "app_proto", opts.appProto, strings.ToLower)
	addIntFilter(filters, "src_port", opts.srcPort, portValue)
	addIntFilter(filters, "dest_port", opts.destPort, portValue)
	if values := parseCSV(opts.keyword); len(values) > 0 {
		filters["keyword"] = values
	}
	return filters
}

func summarizeFileAlarms(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		summary := map[string]any{}
		for _, key := range []string{
			"doc_id",
			"id",
			"timestamp",
			"severity",
			"tag",
			"type",
			"filename",
			"file_type",
			"src_ip",
			"src_port",
			"dest_ip",
			"dest_port",
			"app_proto",
			"proto",
			"md5",
			"sha256",
			"node_id",
			"node_name",
			"sandbox_node_id",
			"sandbox_node_name",
			"score",
			"operating_env",
		} {
			copyIfPresent(summary, item, key)
		}
		addFileAlarmNestedSummary(summary, item)
		out = append(out, summary)
	}
	return out
}

func addFileAlarmNestedSummary(summary map[string]any, item map[string]any) {
	if mainResult, ok := item["main_result"].(map[string]any); ok {
		copyIfPresentAs(summary, mainResult, "file", "filename")
		copyIfPresentAs(summary, mainResult, "type", "file_type")
		copyIfPresentAs(summary, mainResult, "sub_type", "file_sub_type")
		copyIfPresentAs(summary, mainResult, "name", "detect_name")
		copyIfPresentAs(summary, mainResult, "engine", "detect_engine")
	}
	if metaConf, ok := item["meta_conf"].(map[string]any); ok {
		if metaData, ok := metaConf["meta_data"].(map[string]any); ok {
			copyIfPresent(summary, metaData, "src_ip")
			copyIfPresent(summary, metaData, "src_port")
			copyIfPresent(summary, metaData, "dest_ip")
			copyIfPresent(summary, metaData, "dest_port")
			copyIfPresent(summary, metaData, "app_proto")
			copyIfPresent(summary, metaData, "proto")
			if fileinfo, ok := metaData["fileinfo"].(map[string]any); ok {
				copyIfPresent(summary, fileinfo, "filename")
			}
		}
	}
	if sandbox, ok := item["sandbox"].(map[string]any); ok {
		copyIfPresent(summary, sandbox, "score")
		if info, ok := sandbox["info"].(map[string]any); ok {
			copyIfPresent(summary, info, "score")
			copyIfPresentAs(summary, info, "platform", "operating_env")
		}
		if target, ok := sandbox["target"].(map[string]any); ok {
			if file, ok := target["file"].(map[string]any); ok {
				copyIfPresentAs(summary, file, "name", "filename")
				copyIfPresent(summary, file, "md5")
				copyIfPresent(summary, file, "sha256")
			}
		}
	}
}

func writeFileAlarmError(cmd *cobra.Command, task string, command string, code string, message string, retryable bool) error {
	raw, renderErr := RenderJSON(ErrorEnvelope{
		Success: false,
		Task:    task,
		Command: command,
		Error:   CLIError{Code: code, Message: message, Retryable: retryable},
	})
	if renderErr != nil {
		return renderErr
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}
