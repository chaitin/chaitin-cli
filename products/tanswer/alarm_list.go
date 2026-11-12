package tanswer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type alarmListOptions struct {
	time      string
	start     string
	end       string
	page      int
	pageSize  int
	severity  string
	result    string
	phase     string
	attacker  string
	victim    string
	assetIP   string
	keyword   string
	name      string
	tag       string
	direction string
	appProto  string
	url       string
	host      string
	xff       string
	srcPort   string
	destPort  string
}

func newAlarmListCommand(opts *RootOptions) *cobra.Command {
	var alarmOpts alarmListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "查询威胁告警列表",
		Long: "查询威胁告警列表，用于从告警概览下钻到原始告警列表。该命令返回原始告警列表字段，并补充总数、分页和当前返回数量。\n\n" +
			"输出：查询时间范围、实际筛选条件、total、page、page_size、current_count、has_more、alarms。列表项使用产品已有告警列表字段，如 doc_id、name、severity、timestamp、attacker、victim、result、phase、tag、app_proto、url/host 摘要等。",
		Example: "  chaitin-cli tanswer alarm list --time today --page-size 10\n" +
			"  chaitin-cli tanswer alarm list --time 24h --severity critical,high --result success,control\n" +
			"  chaitin-cli tanswer alarm list --asset-ip 192.0.2.10 --attacker 198.51.100.10",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlarmListCommand(cmd, opts, alarmOpts, "查询威胁告警列表", "chaitin-cli tanswer alarm list")
		},
	}
	addAlarmListFlags(cmd, &alarmOpts)
	return cmd
}

func newAlarmHighPriorityCommand(opts *RootOptions) *cobra.Command {
	var alarmOpts alarmListOptions
	cmd := &cobra.Command{
		Use:   "high-priority",
		Short: "查询高优先级威胁告警",
		Long: "查询高优先级威胁告警，用于值班优先处置。该命令是告警列表的高频预设，默认筛选 severity=critical,high 且 result=success,control，返回原始告警列表字段和稳定分页信息。\n\n" +
			"输出：查询时间范围、实际筛选条件、total、page、page_size、current_count、has_more、alarms。",
		Example: "  chaitin-cli tanswer alarm high-priority --time today\n" +
			"  chaitin-cli tanswer alarm high-priority --time 24h --asset-ip 192.0.2.10",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(alarmOpts.severity) == "" {
				alarmOpts.severity = "critical,high"
			}
			if strings.TrimSpace(alarmOpts.result) == "" {
				alarmOpts.result = "success,control"
			}
			return runAlarmListCommand(cmd, opts, alarmOpts, "查询高优先级威胁告警", "chaitin-cli tanswer alarm high-priority")
		},
	}
	addAlarmListFlags(cmd, &alarmOpts)
	return cmd
}

func runAlarmListCommand(cmd *cobra.Command, opts *RootOptions, alarmOpts alarmListOptions, task string, command string) error {
	if alarmOpts.page < 1 {
		return writeAlarmListError(cmd, task, command, "INVALID_PAGE", "page must be greater than or equal to 1", false)
	}
	if alarmOpts.pageSize < 1 || alarmOpts.pageSize > 100 {
		return writeAlarmListError(cmd, task, command, "INVALID_PAGE_SIZE", "page-size must be between 1 and 100", false)
	}
	cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
	if err != nil {
		return err
	}
	rng, err := ParseTimeRange(TimeRangeOptions{Time: alarmOpts.time, Start: alarmOpts.start, End: alarmOpts.end})
	if err != nil {
		return writeAlarmListError(cmd, task, command, "INVALID_TIME_RANGE", err.Error(), false)
	}

	req := buildAlarmListRequest(rng, alarmOpts)
	client := NewClient(cfg)
	var result alarmListRPCResult
	if err := client.CallRPC(cmd.Context(), "AlarmService.SearchAlarmList", req, &result); err != nil {
		return writeAlarmListError(cmd, task, command, "ALARM_LIST_FAILED", err.Error(), true)
	}

	data := map[string]any{
		"total":         result.Total,
		"page_total":    result.PageTotal,
		"page":          alarmOpts.page,
		"page_size":     alarmOpts.pageSize,
		"current_count": len(result.Data),
		"has_more":      int64(alarmOpts.page*alarmOpts.pageSize) < result.Total,
		"alarms":        summarizeAlarmList(result.Data),
	}
	raw, err := RenderJSON(SuccessEnvelope{
		Success: true,
		Task:    task,
		Command: command,
		Query: map[string]any{
			"time_range": rng,
			"filters":    alarmListFilters(alarmOpts),
			"page":       alarmOpts.page,
			"page_size":  alarmOpts.pageSize,
		},
		Data: data,
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func summarizeAlarmList(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		summary := map[string]any{}
		for _, key := range []string{
			"doc_id",
			"name",
			"severity",
			"timestamp",
			"attacker",
			"victim",
			"attacker_port",
			"victim_port",
			"src_ip",
			"src_port",
			"dest_ip",
			"dest_port",
			"proto",
			"app_proto",
			"result",
			"phase",
			"tag",
			"type",
			"tool",
			"direction",
			"xff",
			"msg",
			"geo_attacker",
			"geo_victim",
			"attacker_asset",
			"victim_asset",
			"read",
			"ignore",
			"mark",
			"collapse_info",
		} {
			if value, ok := item[key]; ok {
				summary[key] = value
			}
		}
		if appbrief, ok := item["appbrief"].(map[string]any); ok {
			addAppBriefSummary(summary, appbrief)
		}
		out = append(out, summary)
	}
	return out
}

func addAppBriefSummary(summary map[string]any, appbrief map[string]any) {
	if httpBrief, ok := appbrief["http"].(map[string]any); ok {
		copyIfPresent(summary, httpBrief, "url")
		copyIfPresentAs(summary, httpBrief, "hostname", "host")
		copyIfPresentAs(summary, httpBrief, "method", "http_method")
		copyIfPresentAs(summary, httpBrief, "status", "http_status")
		copyIfPresentAs(summary, httpBrief, "xff_raw", "xff_raw")
	}
	if tiInfo, ok := appbrief["ti_info"].(map[string]any); ok {
		copyIfPresentAs(summary, tiInfo, "ti_type", "ti_type")
		copyIfPresentAs(summary, tiInfo, "ti_value", "ti_value")
		copyIfPresentAs(summary, tiInfo, "ti_desc", "ti_desc")
	}
}

func copyIfPresent(dst map[string]any, src map[string]any, key string) {
	copyIfPresentAs(dst, src, key, key)
}

func copyIfPresentAs(dst map[string]any, src map[string]any, sourceKey string, targetKey string) {
	if value, ok := src[sourceKey]; ok {
		dst[targetKey] = value
	}
}

func addAlarmListFlags(cmd *cobra.Command, opts *alarmListOptions) {
	cmd.Flags().StringVar(&opts.time, "time", "today", "time range: today, 24h, 7d")
	cmd.Flags().StringVar(&opts.start, "start", "", "custom start time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&opts.end, "end", "", "custom end time, format: 2006-01-02 15:04:05")
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().StringVar(&opts.severity, "severity", "", "severity filter: critical,high,medium,low or 1,2,3,4")
	cmd.Flags().StringVar(&opts.result, "result", "", "attack result filter: success,control,failed,unknown")
	cmd.Flags().StringVar(&opts.phase, "phase", "", "attack phase filter: recon,intrustion,delivery,control,lateral,goal,other")
	cmd.Flags().StringVar(&opts.attacker, "attacker", "", "attacker IP filter, comma separated")
	cmd.Flags().StringVar(&opts.victim, "victim", "", "victim IP filter, comma separated")
	cmd.Flags().StringVar(&opts.assetIP, "asset-ip", "", "asset IP filter matching attacker or victim, comma separated")
	cmd.Flags().StringVar(&opts.keyword, "keyword", "", "keyword filter, comma separated")
	cmd.Flags().StringVar(&opts.name, "name", "", "threat name or attack summary filter, comma separated")
	cmd.Flags().StringVar(&opts.tag, "tag", "", "threat type/tag filter, comma separated")
	cmd.Flags().StringVar(&opts.direction, "direction", "", "traffic direction filter: in,lateral,out,other")
	cmd.Flags().StringVar(&opts.appProto, "app-proto", "", "application protocol filter, comma separated")
	cmd.Flags().StringVar(&opts.url, "http-url", "", "HTTP URL filter, comma separated")
	cmd.Flags().StringVar(&opts.host, "host", "", "Host/hostname filter, comma separated")
	cmd.Flags().StringVar(&opts.xff, "xff", "", "XFF filter, comma separated")
	cmd.Flags().StringVar(&opts.srcPort, "src-port", "", "source port filter, comma separated")
	cmd.Flags().StringVar(&opts.destPort, "dest-port", "", "destination port filter, comma separated")
}

type alarmListRPCResult struct {
	Data      []map[string]any `json:"data"`
	Total     int64            `json:"total"`
	PageTotal int64            `json:"page_total"`
}

func buildAlarmListRequest(rng TimeRange, opts alarmListOptions) map[string]any {
	if opts.page < 1 {
		opts.page = 1
	}
	if opts.pageSize < 1 {
		opts.pageSize = 10
	}
	offset := int64((opts.page - 1) * opts.pageSize)
	count := int64(opts.pageSize)
	req := map[string]any{
		"time_range_start": rng.Start,
		"time_range_end":   rng.End,
		"offset":           offset,
		"count":            count,
		"save_history":     false,
	}
	for key, value := range alarmListFilters(opts) {
		req[key] = value
	}
	return req
}

func alarmListFilters(opts alarmListOptions) map[string]any {
	filters := map[string]any{}
	addIntFilter(filters, "severity", opts.severity, severityValue)
	addStringFilter(filters, "result", opts.result, strings.ToLower)
	addStringFilter(filters, "phase", opts.phase, strings.ToLower)
	addStringFilter(filters, "attacker", opts.attacker, strings.TrimSpace)
	addStringFilter(filters, "victim", opts.victim, strings.TrimSpace)
	addStringFilter(filters, "name", opts.name, strings.TrimSpace)
	addStringFilter(filters, "tag", opts.tag, strings.TrimSpace)
	addStringFilter(filters, "direction", opts.direction, strings.ToLower)
	addStringFilter(filters, "app_proto", opts.appProto, strings.ToLower)
	addStringFilter(filters, "url", opts.url, strings.TrimSpace)
	addStringFilter(filters, "hostname", opts.host, strings.TrimSpace)
	addStringFilter(filters, "xff", opts.xff, strings.TrimSpace)
	addIntFilter(filters, "src_port", opts.srcPort, portValue)
	addIntFilter(filters, "dest_port", opts.destPort, portValue)
	if values := parseCSV(opts.assetIP); len(values) > 0 {
		filters["asset_ip"] = values
	}
	if values := parseCSV(opts.keyword); len(values) > 0 {
		filters["keyword"] = values
	}
	return filters
}

func addStringFilter(filters map[string]any, key string, raw string, normalize func(string) string) {
	if values := parseCSV(raw); len(values) > 0 {
		filters[key] = stringQuery(values, normalize)
	}
}

func addIntFilter(filters map[string]any, key string, raw string, normalize func(string) int) {
	if values := parseCSV(raw); len(values) > 0 {
		filters[key] = intQuery(values, normalize)
	}
}

func portValue(value string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(value))
	return n
}

func writeAlarmListError(cmd *cobra.Command, task string, command string, code string, message string, retryable bool) error {
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
