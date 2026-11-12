package tanswer

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type alarmOverviewOptions struct {
	time     string
	start    string
	end      string
	severity string
	result   string
	phase    string
}

func newAlarmCommand(opts *RootOptions) *cobra.Command {
	alarm := &cobra.Command{
		Use:   "alarm",
		Short: "Threat alarm semantic shortcuts",
		Long:  "Threat alarm semantic shortcuts. Use overview for duty inspection summaries; use Open API fallback for raw interface calls.",
	}
	alarm.AddCommand(newAlarmOverviewCommand(opts))
	alarm.AddCommand(newAlarmListCommand(opts))
	alarm.AddCommand(newAlarmHighPriorityCommand(opts))
	alarm.AddCommand(newAlarmDetailCommand(opts))
	alarm.AddCommand(newAlarmTimelineCommand(opts))
	alarm.AddCommand(newAlarmByAttackerCommand(opts))
	alarm.AddCommand(newAlarmByVictimCommand(opts))
	alarm.AddCommand(newAlarmByThreatCommand(opts))
	alarm.AddCommand(newAlarmImportantAssetsCommand(opts))
	alarm.AddCommand(newAlarmRankCommand(opts, alarmRankAttacker))
	alarm.AddCommand(newAlarmRankCommand(opts, alarmRankVictim))
	alarm.AddCommand(newAlarmRankCommand(opts, alarmRankPhase))
	alarm.AddCommand(newAlarmRelatedCommand(opts))
	return alarm
}

func newAlarmOverviewCommand(opts *RootOptions) *cobra.Command {
	var alarmOpts alarmOverviewOptions
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "查看威胁告警概览",
		Long:  "查看威胁告警概览，用于值班巡检时快速判断当前威胁态势和优先处理方向。该命令返回聚合概览、分布和 Top 排行，不返回原始告警列表。\n\n输出：查询时间范围、实际筛选条件、告警总数、等级分布、攻击结果分布、攻击阶段分布、威胁类型 Top、攻击源 Top、受害对象 Top。",
		Example: "  chaitin-cli tanswer alarm overview --time today\n" +
			"  chaitin-cli tanswer alarm overview --time 24h --severity critical,high --result success,control",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			rng, err := ParseTimeRange(TimeRangeOptions{Time: alarmOpts.time, Start: alarmOpts.start, End: alarmOpts.end})
			if err != nil {
				raw, renderErr := RenderJSON(ErrorEnvelope{
					Success: false,
					Task:    "查看威胁告警概览",
					Command: "chaitin-cli tanswer alarm overview",
					Error:   CLIError{Code: "INVALID_TIME_RANGE", Message: err.Error(), Retryable: false},
				})
				if renderErr != nil {
					return renderErr
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			client := NewClient(cfg)
			filters := buildAlarmFilters(alarmOpts)
			data, err := fetchAlarmOverview(cmd.Context(), client, rng, filters)
			if err != nil {
				raw, renderErr := RenderJSON(ErrorEnvelope{
					Success: false,
					Task:    "查看威胁告警概览",
					Command: "chaitin-cli tanswer alarm overview",
					Error:   CLIError{Code: "ALARM_OVERVIEW_FAILED", Message: err.Error(), Retryable: true},
				})
				if renderErr != nil {
					return renderErr
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查看威胁告警概览",
				Command: "chaitin-cli tanswer alarm overview",
				Query: map[string]any{
					"time_range": rng,
					"filters":    filters,
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
	cmd.Flags().StringVar(&alarmOpts.time, "time", "today", "time range: today, 24h, 7d")
	cmd.Flags().StringVar(&alarmOpts.start, "start", "", "custom start time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&alarmOpts.end, "end", "", "custom end time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&alarmOpts.severity, "severity", "", "severity filter: critical,high,medium,low or 1,2,3,4")
	cmd.Flags().StringVar(&alarmOpts.result, "result", "", "attack result filter: success,control,failed,unknown")
	cmd.Flags().StringVar(&alarmOpts.phase, "phase", "", "attack phase filter: recon,intrustion,delivery,control,lateral,goal,other")
	return cmd
}

func fetchAlarmOverview(ctx context.Context, client *Client, rng TimeRange, filters map[string]any) (map[string]any, error) {
	base := map[string]any{
		"time_range_start": rng.Start,
		"time_range_end":   rng.End,
	}
	for key, value := range filters {
		base[key] = value
	}

	data := map[string]any{}
	count := map[string]any{}
	if err := client.CallRPC(ctx, "AlarmService.SearchAlarmCount", base, &count); err != nil {
		if isAuthRequiredRPCError(err) {
			return fetchAlarmOverviewFromList(ctx, client, rng, filters)
		}
		return nil, err
	}
	data["summary"] = map[string]any{"alarm_total": count["total"]}

	aggs := []struct {
		agg      string
		outField string
	}{
		{agg: "severity", outField: "severity_distribution"},
		{agg: "result", outField: "result_distribution"},
		{agg: "phase", outField: "phase_distribution"},
		{agg: "tag", outField: "threat_type_top"},
		{agg: "attacker", outField: "attacker_top"},
		{agg: "victim", outField: "victim_top"},
	}
	for _, item := range aggs {
		req := copyMap(base)
		req["agg"] = item.agg
		req["top"] = 10
		aggResp := map[string]any{}
		if err := client.CallRPC(ctx, "AlarmService.SearchAlarmAggTop", req, &aggResp); err != nil {
			if isAuthRequiredRPCError(err) {
				return fetchAlarmOverviewFromList(ctx, client, rng, filters)
			}
			return nil, err
		}
		data[item.outField] = aggResp["data"]
	}
	return data, nil
}

func isAuthRequiredRPCError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "需要登录") || strings.Contains(strings.ToLower(msg), "unauthorized")
}

func fetchAlarmOverviewFromList(ctx context.Context, client *Client, rng TimeRange, filters map[string]any) (map[string]any, error) {
	req := map[string]any{
		"time_range_start": rng.Start,
		"time_range_end":   rng.End,
		"offset":           int64(0),
		"count":            int64(100),
		"save_history":     false,
	}
	for key, value := range filters {
		req[key] = value
	}
	var result alarmListRPCResult
	if err := client.CallRPC(ctx, "AlarmService.SearchAlarmList", req, &result); err != nil {
		return nil, err
	}
	data := map[string]any{
		"summary": map[string]any{
			"alarm_total":   result.Total,
			"source":        "list_fallback",
			"current_count": len(result.Data),
			"partial":       int64(len(result.Data)) < result.Total,
		},
		"severity_distribution": topCount(result.Data, "severity", 10),
		"result_distribution":   topCount(result.Data, "result", 10),
		"phase_distribution":    topCount(result.Data, "phase", 10),
		"threat_type_top":       topCount(result.Data, "tag", 10),
		"attacker_top":          topCount(result.Data, "attacker", 10),
		"victim_top":            topCount(result.Data, "victim", 10),
	}
	return data, nil
}

func topCount(items []map[string]any, field string, limit int) []map[string]any {
	counts := map[string]int{}
	for _, item := range items {
		value, ok := item[field]
		if !ok || value == nil {
			continue
		}
		key := strings.TrimSpace(fmt.Sprint(value))
		if key == "" {
			continue
		}
		counts[key]++
	}
	rows := make([]map[string]any, 0, len(counts))
	for key, count := range counts {
		rows = append(rows, map[string]any{"key": key, "count": count})
	}
	sort.Slice(rows, func(i, j int) bool {
		left := rows[i]["count"].(int)
		right := rows[j]["count"].(int)
		if left == right {
			return rows[i]["key"].(string) < rows[j]["key"].(string)
		}
		return left > right
	})
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	return rows
}

func buildAlarmFilters(opts alarmOverviewOptions) map[string]any {
	filters := map[string]any{}
	if values := parseCSV(opts.severity); len(values) > 0 {
		filters["severity"] = intQuery(values, severityValue)
	}
	if values := parseCSV(opts.result); len(values) > 0 {
		filters["result"] = stringQuery(values, strings.ToLower)
	}
	if values := parseCSV(opts.phase); len(values) > 0 {
		filters["phase"] = stringQuery(values, strings.ToLower)
	}
	return filters
}

func parseCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var out []string
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func stringQuery(values []string, normalize func(string) string) []map[string]any {
	out := make([]map[string]any, 0, len(values))
	for _, value := range values {
		out = append(out, map[string]any{"oper": "=", "target": normalize(value)})
	}
	return out
}

func intQuery(values []string, normalize func(string) int) []map[string]any {
	out := make([]map[string]any, 0, len(values))
	for _, value := range values {
		out = append(out, map[string]any{"oper": "=", "target": normalize(value)})
	}
	return out
}

func uintEqualityQuery(values []uint) []map[string]any {
	out := make([]map[string]any, 0, len(values))
	for _, value := range values {
		out = append(out, map[string]any{"oper": "=", "target": value})
	}
	return out
}

func severityValue(value string) int {
	switch strings.ToLower(value) {
	case "critical", "超危":
		return 1
	case "high", "高危":
		return 2
	case "medium", "中危":
		return 3
	case "low", "低危":
		return 4
	default:
		n, _ := strconv.Atoi(value)
		return n
	}
}

func copyMap(in map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range in {
		out[key] = value
	}
	return out
}
