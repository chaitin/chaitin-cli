package tanswer

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type alarmRelatedOptions struct {
	id       string
	window   string
	relation string
	limit    int
}

func newAlarmRelatedCommand(opts *RootOptions) *cobra.Command {
	var alarmOpts alarmRelatedOptions
	cmd := &cobra.Command{
		Use:   "related",
		Short: "查看相关告警",
		Long: "查看相关告警，用于从一条告警出发，查询原告警前后时间窗口内同攻击源或同受害对象的其他告警。默认窗口为前后 30m，默认关联口径为 both。\n\n" +
			"输出：原告警摘要、时间窗口、关联口径、相关告警数量、最早/最新告警时间、相关告警摘要。",
		Example: "  chaitin-cli tanswer alarm related --id '<doc_id>'\n" +
			"  chaitin-cli tanswer alarm related --id '<doc_id>' --window 1h --relation attacker",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlarmRelatedCommand(cmd, opts, alarmOpts)
		},
	}
	cmd.Flags().StringVar(&alarmOpts.id, "id", "", "alarm doc_id from alarm list")
	cmd.Flags().StringVar(&alarmOpts.window, "window", "30m", "time window before and after the source alarm, for example 10m, 30m, 1h")
	cmd.Flags().StringVar(&alarmOpts.relation, "relation", "both", "relation scope: both, attacker, victim")
	cmd.Flags().IntVar(&alarmOpts.limit, "limit", 20, "maximum related alarms to return after de-duplication, 1-100")
	return cmd
}

func runAlarmRelatedCommand(cmd *cobra.Command, opts *RootOptions, alarmOpts alarmRelatedOptions) error {
	const task = "查看相关告警"
	const command = "chaitin-cli tanswer alarm related"

	if strings.TrimSpace(alarmOpts.id) == "" {
		return writeAlarmListError(cmd, task, command, "MISSING_ALARM_ID", "missing alarm doc_id: set --id", false)
	}
	window, err := parseRelatedWindowMillis(alarmOpts.window)
	if err != nil {
		return writeAlarmListError(cmd, task, command, "INVALID_WINDOW", err.Error(), false)
	}
	relation := strings.ToLower(strings.TrimSpace(alarmOpts.relation))
	if relation == "" {
		relation = "both"
	}
	if relation != "both" && relation != "attacker" && relation != "victim" {
		return writeAlarmListError(cmd, task, command, "INVALID_RELATION", "relation must be one of both, attacker, victim", false)
	}
	if alarmOpts.limit < 1 || alarmOpts.limit > 100 {
		return writeAlarmListError(cmd, task, command, "INVALID_LIMIT", "limit must be between 1 and 100", false)
	}

	cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
	if err != nil {
		return err
	}
	client := NewClient(cfg)
	var detail alarmDetailRPCResult
	if err := client.CallRPC(cmd.Context(), "AlarmService.GetAlarm", map[string]any{"doc_id": strings.TrimSpace(alarmOpts.id)}, &detail); err != nil {
		return writeAlarmListError(cmd, task, command, "ALARM_DETAIL_FAILED", err.Error(), true)
	}
	if len(detail.Data) == 0 {
		return writeAlarmListError(cmd, task, command, "ALARM_NOT_FOUND", "alarm detail is empty", false)
	}

	var merged []map[string]any
	var scanned int64
	if relation == "both" || relation == "attacker" {
		req, err := buildAlarmRelatedListRequest(detail.Data, "attacker", window, alarmOpts.limit)
		if err != nil {
			return writeAlarmListError(cmd, task, command, "INVALID_SOURCE_ALARM", err.Error(), false)
		}
		var result alarmListRPCResult
		if err := client.CallRPC(cmd.Context(), "AlarmService.SearchAlarmList", req, &result); err != nil {
			return writeAlarmListError(cmd, task, command, "ALARM_RELATED_FAILED", err.Error(), true)
		}
		scanned += result.Total
		merged = append(merged, markRelatedScope(result.Data, "same_attacker")...)
	}
	if relation == "both" || relation == "victim" {
		req, err := buildAlarmRelatedListRequest(detail.Data, "victim", window, alarmOpts.limit)
		if err != nil {
			return writeAlarmListError(cmd, task, command, "INVALID_SOURCE_ALARM", err.Error(), false)
		}
		var result alarmListRPCResult
		if err := client.CallRPC(cmd.Context(), "AlarmService.SearchAlarmList", req, &result); err != nil {
			return writeAlarmListError(cmd, task, command, "ALARM_RELATED_FAILED", err.Error(), true)
		}
		scanned += result.Total
		merged = append(merged, markRelatedScope(result.Data, "same_victim")...)
	}

	sourceID := strings.TrimSpace(alarmOpts.id)
	sourceIDs := map[string]struct{}{sourceID: {}}
	if detailDocID := strings.TrimSpace(fmt.Sprint(detail.Data["doc_id"])); detailDocID != "" && detailDocID != "<nil>" {
		sourceIDs[detailDocID] = struct{}{}
	}
	related := dedupeRelatedAlarms(merged, sourceIDs, alarmOpts.limit)
	data := map[string]any{
		"source_alarm":  summarizeSourceAlarm(detail.Data),
		"window":        buildRelatedWindow(detail.Data, window),
		"relation":      relation,
		"scanned_total": scanned,
		"related_total": len(related),
		"current_count": len(related),
		"earliest_time": minAlarmTimestamp(related),
		"latest_time":   maxAlarmTimestamp(related),
		"alarms":        related,
	}
	raw, err := RenderJSON(SuccessEnvelope{
		Success: true,
		Task:    task,
		Command: command,
		Query: map[string]any{
			"id":       sourceID,
			"window":   firstNonEmpty(alarmOpts.window, "30m"),
			"relation": relation,
			"limit":    alarmOpts.limit,
		},
		Data: data,
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func buildAlarmRelatedListRequest(original map[string]any, relation string, windowMillis int64, limit int) (map[string]any, error) {
	timestamp, err := alarmTimestamp(original)
	if err != nil {
		return nil, err
	}
	if limit < 1 {
		limit = 20
	}
	req := map[string]any{
		"time_range_start": timestamp - windowMillis,
		"time_range_end":   timestamp + windowMillis,
		"offset":           int64(0),
		"count":            int64(limit),
		"save_history":     false,
	}
	switch relation {
	case "attacker":
		attacker := strings.TrimSpace(fmt.Sprint(original["attacker"]))
		if attacker == "" || attacker == "<nil>" {
			return nil, fmt.Errorf("source alarm has no attacker")
		}
		req["attacker"] = stringQuery([]string{attacker}, strings.TrimSpace)
	case "victim":
		victim := strings.TrimSpace(fmt.Sprint(original["victim"]))
		if victim == "" || victim == "<nil>" {
			return nil, fmt.Errorf("source alarm has no victim")
		}
		req["victim"] = stringQuery([]string{victim}, strings.TrimSpace)
	default:
		return nil, fmt.Errorf("unsupported relation %q", relation)
	}
	return req, nil
}

func parseRelatedWindowMillis(value string) (int64, error) {
	if strings.TrimSpace(value) == "" {
		value = "30m"
	}
	d, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, err
	}
	if d <= 0 {
		return 0, fmt.Errorf("window must be greater than 0")
	}
	if d > 24*time.Hour {
		return 0, fmt.Errorf("window must be no more than 24h")
	}
	return d.Milliseconds(), nil
}

func alarmTimestamp(item map[string]any) (int64, error) {
	for _, key := range []string{"timestamp", "alert_time", "time"} {
		if value, ok := item[key]; ok {
			if ts, ok := anyToInt64(value); ok {
				return ts, nil
			}
		}
	}
	return 0, fmt.Errorf("source alarm has no timestamp")
}

func anyToInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case jsonNumber:
		n, err := strconv.ParseInt(string(v), 10, 64)
		return n, err == nil
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return n, err == nil
	default:
		return 0, false
	}
}

type jsonNumber string

func markRelatedScope(items []map[string]any, scope string) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		copyItem := copyMap(item)
		copyItem["matched_by"] = scope
		out = append(out, copyItem)
	}
	return out
}

func dedupeRelatedAlarms(items []map[string]any, sourceIDs map[string]struct{}, limit int) []map[string]any {
	seen := map[string]int{}
	var merged []map[string]any
	for _, item := range items {
		docID := strings.TrimSpace(fmt.Sprint(item["doc_id"]))
		if docID == "" {
			continue
		}
		if _, isSource := sourceIDs[docID]; isSource {
			continue
		}
		if idx, ok := seen[docID]; ok {
			existing := strings.TrimSpace(fmt.Sprint(merged[idx]["matched_by"]))
			next := strings.TrimSpace(fmt.Sprint(item["matched_by"]))
			if existing != next && !strings.Contains(existing, next) {
				merged[idx]["matched_by"] = existing + "," + next
			}
			continue
		}
		seen[docID] = len(merged)
		merged = append(merged, item)
	}
	sort.Slice(merged, func(i, j int) bool {
		left, _ := alarmTimestamp(merged[i])
		right, _ := alarmTimestamp(merged[j])
		return left < right
	})
	if limit > 0 && len(merged) > limit {
		merged = merged[:limit]
	}
	return summarizeRelatedAlarms(merged)
}

func summarizeRelatedAlarms(items []map[string]any) []map[string]any {
	out := summarizeAlarmList(items)
	for i := range out {
		if matchedBy, ok := items[i]["matched_by"]; ok {
			out[i]["matched_by"] = matchedBy
		}
	}
	return out
}

func summarizeSourceAlarm(item map[string]any) map[string]any {
	summary := map[string]any{}
	for _, key := range []string{"doc_id", "name", "severity", "timestamp", "attacker", "victim", "result", "phase", "tag"} {
		if value, ok := item[key]; ok {
			summary[key] = value
		}
	}
	return summary
}

func buildRelatedWindow(item map[string]any, windowMillis int64) map[string]any {
	timestamp, _ := alarmTimestamp(item)
	return map[string]any{
		"center": timestamp,
		"before": windowMillis,
		"after":  windowMillis,
		"start":  timestamp - windowMillis,
		"end":    timestamp + windowMillis,
	}
}

func minAlarmTimestamp(items []map[string]any) any {
	var min int64
	for _, item := range items {
		ts, err := alarmTimestamp(item)
		if err != nil {
			continue
		}
		if min == 0 || ts < min {
			min = ts
		}
	}
	if min == 0 {
		return nil
	}
	return min
}

func maxAlarmTimestamp(items []map[string]any) any {
	var max int64
	for _, item := range items {
		ts, err := alarmTimestamp(item)
		if err != nil {
			continue
		}
		if ts > max {
			max = ts
		}
	}
	if max == 0 {
		return nil
	}
	return max
}
