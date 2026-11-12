package tanswer

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type alarmSubjectKind string

const (
	alarmSubjectAttacker  alarmSubjectKind = "attacker"
	alarmSubjectVictim    alarmSubjectKind = "victim"
	alarmSubjectThreat    alarmSubjectKind = "threat"
	alarmSubjectImportant alarmSubjectKind = "important_assets"
)

type alarmSubjectOptions struct {
	time      string
	start     string
	end       string
	pageSize  int
	subject   string
	name      string
	tag       string
	phase     string
	severity  string
	result    string
	direction string
}

type alarmRankKind string

const (
	alarmRankAttacker alarmRankKind = "attacker"
	alarmRankVictim   alarmRankKind = "victim"
	alarmRankPhase    alarmRankKind = "phase"
)

type alarmRankOptions struct {
	time     string
	start    string
	end      string
	top      int
	severity string
	result   string
	phase    string
}

type alarmAggTopRPCResult struct {
	Data []map[string]any `json:"data"`
}

func newAlarmByAttackerCommand(opts *RootOptions) *cobra.Command {
	var alarmOpts alarmSubjectOptions
	cmd := &cobra.Command{
		Use:   "by-attacker",
		Short: "查询指定攻击源相关告警",
		Long: "查询指定攻击源相关告警，用于已知攻击源后快速查看攻击范围、威胁类型和最近活动。\n\n" +
			"输出：攻击源、查询时间范围、相关告警数、最高等级、成功/失陷数量、受害对象 Top、主要威胁类型、告警摘要。",
		Example: "  chaitin-cli tanswer alarm by-attacker --attacker 198.51.100.10 --time today",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlarmSubjectCommand(cmd, opts, alarmOpts, alarmSubjectAttacker)
		},
	}
	addAlarmSubjectCommonFlags(cmd, &alarmOpts)
	cmd.Flags().StringVar(&alarmOpts.subject, "attacker", "", "attacker IP")
	return cmd
}

func newAlarmByVictimCommand(opts *RootOptions) *cobra.Command {
	var alarmOpts alarmSubjectOptions
	cmd := &cobra.Command{
		Use:   "by-victim",
		Short: "查询指定受害对象相关告警",
		Long: "查询指定受害对象相关告警，用于已知资产或受害对象后快速判断其被攻击情况。\n\n" +
			"输出：受害对象、查询时间范围、相关告警数、最高等级、成功/失陷数量、攻击源 Top、主要威胁类型、告警摘要。",
		Example: "  chaitin-cli tanswer alarm by-victim --victim 203.0.113.20 --time today",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlarmSubjectCommand(cmd, opts, alarmOpts, alarmSubjectVictim)
		},
	}
	addAlarmSubjectCommonFlags(cmd, &alarmOpts)
	cmd.Flags().StringVar(&alarmOpts.subject, "victim", "", "victim IP or object")
	return cmd
}

func newAlarmByThreatCommand(opts *RootOptions) *cobra.Command {
	var alarmOpts alarmSubjectOptions
	cmd := &cobra.Command{
		Use:   "by-threat",
		Short: "查询指定威胁相关告警",
		Long: "查询指定威胁相关告警，用于围绕威胁名称、威胁类型或攻击阶段查看当前影响范围。必须至少提供 --name、--tag 或 --phase 中的一项。\n\n" +
			"输出：威胁筛选条件、查询时间范围、相关告警数、最高等级、成功/失陷数量、攻击源 Top、受害对象 Top、告警摘要。",
		Example: "  chaitin-cli tanswer alarm by-threat --name SQL注入 --time today\n" +
			"  chaitin-cli tanswer alarm by-threat --tag Webshell --phase intrustion",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlarmSubjectCommand(cmd, opts, alarmOpts, alarmSubjectThreat)
		},
	}
	addAlarmSubjectCommonFlags(cmd, &alarmOpts)
	cmd.Flags().StringVar(&alarmOpts.name, "name", "", "threat name filter")
	cmd.Flags().StringVar(&alarmOpts.tag, "tag", "", "threat type/tag filter")
	cmd.Flags().StringVar(&alarmOpts.phase, "phase", "", "attack phase filter")
	return cmd
}

func newAlarmImportantAssetsCommand(opts *RootOptions) *cobra.Command {
	var alarmOpts alarmSubjectOptions
	cmd := &cobra.Command{
		Use:   "important-assets",
		Short: "查询重点资产相关告警",
		Long: "查询重点资产相关告警，用于优先确认重点资产是否被攻击、是否存在成功或失陷风险。\n\n" +
			"输出：重点资产相关告警数、最高等级、成功/失陷数量、受害对象 Top、攻击源 Top、告警摘要。",
		Example: "  chaitin-cli tanswer alarm important-assets --time today",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlarmSubjectCommand(cmd, opts, alarmOpts, alarmSubjectImportant)
		},
	}
	addAlarmSubjectCommonFlags(cmd, &alarmOpts)
	return cmd
}

func runAlarmSubjectCommand(cmd *cobra.Command, opts *RootOptions, alarmOpts alarmSubjectOptions, kind alarmSubjectKind) error {
	task, command := alarmSubjectTaskAndCommand(kind)
	if alarmOpts.pageSize < 1 || alarmOpts.pageSize > 100 {
		return writeAlarmListError(cmd, task, command, "INVALID_PAGE_SIZE", "page-size must be between 1 and 100", false)
	}
	if err := validateAlarmSubjectOptions(alarmOpts, kind); err != nil {
		return writeAlarmListError(cmd, task, command, "INVALID_ALARM_SUBJECT", err.Error(), false)
	}
	cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
	if err != nil {
		return err
	}
	rng, err := ParseTimeRange(TimeRangeOptions{Time: alarmOpts.time, Start: alarmOpts.start, End: alarmOpts.end})
	if err != nil {
		return writeAlarmListError(cmd, task, command, "INVALID_TIME_RANGE", err.Error(), false)
	}
	req := buildAlarmSubjectListRequest(rng, alarmOpts, kind)
	client := NewClient(cfg)
	var result alarmListRPCResult
	if err := client.CallRPC(cmd.Context(), "AlarmService.SearchAlarmList", req, &result); err != nil {
		return writeAlarmListError(cmd, task, command, "ALARM_SUBJECT_FAILED", err.Error(), true)
	}
	data := summarizeAlarmSubject(kind, alarmOpts, result)
	raw, err := RenderJSON(SuccessEnvelope{
		Success: true,
		Task:    task,
		Command: command,
		Query: map[string]any{
			"time_range": rng,
			"filters":    alarmSubjectFilters(alarmOpts, kind),
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

func buildAlarmSubjectListRequest(rng TimeRange, opts alarmSubjectOptions, kind alarmSubjectKind) map[string]any {
	pageSize := opts.pageSize
	if pageSize < 1 {
		pageSize = 10
	}
	req := map[string]any{
		"time_range_start": rng.Start,
		"time_range_end":   rng.End,
		"offset":           int64(0),
		"count":            int64(pageSize),
		"save_history":     false,
	}
	for key, value := range alarmSubjectFilters(opts, kind) {
		req[key] = value
	}
	return req
}

func alarmSubjectFilters(opts alarmSubjectOptions, kind alarmSubjectKind) map[string]any {
	filters := alarmListFilters(alarmListOptions{
		severity:  opts.severity,
		result:    opts.result,
		phase:     opts.phase,
		name:      opts.name,
		tag:       opts.tag,
		direction: opts.direction,
	})
	switch kind {
	case alarmSubjectAttacker:
		addStringFilter(filters, "attacker", opts.subject, strings.TrimSpace)
	case alarmSubjectVictim:
		addStringFilter(filters, "victim", opts.subject, strings.TrimSpace)
	case alarmSubjectImportant:
		filters["asset_importance"] = intQuery([]string{"1"}, severityValue)
	}
	return filters
}

func validateAlarmSubjectOptions(opts alarmSubjectOptions, kind alarmSubjectKind) error {
	switch kind {
	case alarmSubjectAttacker:
		if strings.TrimSpace(opts.subject) == "" {
			return fmt.Errorf("missing attacker: set --attacker")
		}
	case alarmSubjectVictim:
		if strings.TrimSpace(opts.subject) == "" {
			return fmt.Errorf("missing victim: set --victim")
		}
	case alarmSubjectThreat:
		if strings.TrimSpace(opts.name) == "" && strings.TrimSpace(opts.tag) == "" && strings.TrimSpace(opts.phase) == "" {
			return fmt.Errorf("set at least one of --name, --tag, or --phase")
		}
	}
	return nil
}

func summarizeAlarmSubject(kind alarmSubjectKind, opts alarmSubjectOptions, result alarmListRPCResult) map[string]any {
	data := map[string]any{
		"related_total":         result.Total,
		"current_count":         len(result.Data),
		"highest_severity":      highestSeverity(result.Data),
		"success_control_count": successControlCount(result.Data),
		"alarms":                summarizeAlarmList(result.Data),
	}
	switch kind {
	case alarmSubjectAttacker:
		data["attacker"] = strings.TrimSpace(opts.subject)
		data["victim_top"] = topCount(result.Data, "victim", 10)
		data["threat_type_top"] = topCount(result.Data, "tag", 10)
	case alarmSubjectVictim:
		data["victim"] = strings.TrimSpace(opts.subject)
		data["attacker_top"] = topCount(result.Data, "attacker", 10)
		data["threat_type_top"] = topCount(result.Data, "tag", 10)
	case alarmSubjectThreat:
		data["threat"] = map[string]any{"name": opts.name, "tag": opts.tag, "phase": opts.phase}
		data["attacker_top"] = topCount(result.Data, "attacker", 10)
		data["victim_top"] = topCount(result.Data, "victim", 10)
	case alarmSubjectImportant:
		data["asset_importance"] = "important"
		data["attacker_top"] = topCount(result.Data, "attacker", 10)
		data["victim_top"] = topCount(result.Data, "victim", 10)
	}
	return data
}

func highestSeverity(items []map[string]any) any {
	best := 0
	for _, item := range items {
		value := severityNumber(item["severity"])
		if value == 0 {
			continue
		}
		if best == 0 || value < best {
			best = value
		}
	}
	if best == 0 {
		return nil
	}
	return best
}

func severityNumber(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		return severityValue(v)
	default:
		return 0
	}
}

func successControlCount(items []map[string]any) int {
	count := 0
	for _, item := range items {
		result := strings.ToLower(strings.TrimSpace(fmt.Sprint(item["result"])))
		if result == "success" || result == "control" {
			count++
		}
	}
	return count
}

func addAlarmSubjectCommonFlags(cmd *cobra.Command, opts *alarmSubjectOptions) {
	cmd.Flags().StringVar(&opts.time, "time", "today", "time range: today, 24h, 7d")
	cmd.Flags().StringVar(&opts.start, "start", "", "custom start time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&opts.end, "end", "", "custom end time, format: 2006-01-02 15:04:05")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "number of alarm summaries to return, 1-100")
	cmd.Flags().StringVar(&opts.severity, "severity", "", "severity filter: critical,high,medium,low or 1,2,3,4")
	cmd.Flags().StringVar(&opts.result, "result", "", "attack result filter: success,control,failed,unknown")
	cmd.Flags().StringVar(&opts.direction, "direction", "", "traffic direction filter: in,lateral,out,other")
}

func newAlarmRankCommand(opts *RootOptions, kind alarmRankKind) *cobra.Command {
	var alarmOpts alarmRankOptions
	use, task, summary, example := alarmRankMetadata(kind)
	cmd := &cobra.Command{
		Use:     use,
		Short:   task,
		Long:    summary + "\n\n输出：查询时间范围、排行类型、排行数量、排行列表。",
		Example: example,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlarmRankCommand(cmd, opts, alarmOpts, kind)
		},
	}
	cmd.Flags().StringVar(&alarmOpts.time, "time", "today", "time range: today, 24h, 7d")
	cmd.Flags().StringVar(&alarmOpts.start, "start", "", "custom start time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&alarmOpts.end, "end", "", "custom end time, format: 2006-01-02 15:04:05")
	cmd.Flags().IntVar(&alarmOpts.top, "top", 10, "ranking size, 1-100")
	cmd.Flags().StringVar(&alarmOpts.severity, "severity", "", "severity filter: critical,high,medium,low or 1,2,3,4")
	cmd.Flags().StringVar(&alarmOpts.result, "result", "", "attack result filter: success,control,failed,unknown")
	cmd.Flags().StringVar(&alarmOpts.phase, "phase", "", "attack phase filter")
	return cmd
}

func runAlarmRankCommand(cmd *cobra.Command, opts *RootOptions, alarmOpts alarmRankOptions, kind alarmRankKind) error {
	_, task, _, _ := alarmRankMetadata(kind)
	command := "chaitin-cli tanswer alarm " + string(alarmRankUse(kind))
	if alarmOpts.top < 1 || alarmOpts.top > 100 {
		return writeAlarmListError(cmd, task, command, "INVALID_TOP", "top must be between 1 and 100", false)
	}
	cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
	if err != nil {
		return err
	}
	rng, err := ParseTimeRange(TimeRangeOptions{Time: alarmOpts.time, Start: alarmOpts.start, End: alarmOpts.end})
	if err != nil {
		return writeAlarmListError(cmd, task, command, "INVALID_TIME_RANGE", err.Error(), false)
	}
	req := buildAlarmRankRequest(rng, alarmOpts, kind)
	client := NewClient(cfg)
	var result alarmAggTopRPCResult
	if err := client.CallRPC(cmd.Context(), "AlarmService.SearchAlarmAggTop", req, &result); err != nil {
		return writeAlarmListError(cmd, task, command, "ALARM_RANK_FAILED", err.Error(), true)
	}
	raw, err := RenderJSON(SuccessEnvelope{
		Success: true,
		Task:    task,
		Command: command,
		Query: map[string]any{
			"time_range": rng,
			"filters":    alarmRankFilters(alarmOpts),
			"top":        alarmOpts.top,
		},
		Data: map[string]any{
			"rank_type":  string(kind),
			"rank_count": len(result.Data),
			"rank":       result.Data,
		},
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func buildAlarmRankRequest(rng TimeRange, opts alarmRankOptions, kind alarmRankKind) map[string]any {
	req := map[string]any{
		"time_range_start": rng.Start,
		"time_range_end":   rng.End,
		"agg":              string(kind),
		"top":              opts.top,
	}
	for key, value := range alarmRankFilters(opts) {
		req[key] = value
	}
	return req
}

func alarmRankFilters(opts alarmRankOptions) map[string]any {
	return alarmListFilters(alarmListOptions{
		severity: opts.severity,
		result:   opts.result,
		phase:    opts.phase,
	})
}

func alarmSubjectTaskAndCommand(kind alarmSubjectKind) (string, string) {
	switch kind {
	case alarmSubjectAttacker:
		return "查询指定攻击源相关告警", "chaitin-cli tanswer alarm by-attacker"
	case alarmSubjectVictim:
		return "查询指定受害对象相关告警", "chaitin-cli tanswer alarm by-victim"
	case alarmSubjectThreat:
		return "查询指定威胁相关告警", "chaitin-cli tanswer alarm by-threat"
	case alarmSubjectImportant:
		return "查询重点资产相关告警", "chaitin-cli tanswer alarm important-assets"
	default:
		return "查询威胁告警", "chaitin-cli tanswer alarm"
	}
}

func alarmRankUse(kind alarmRankKind) string {
	switch kind {
	case alarmRankAttacker:
		return "attacker-rank"
	case alarmRankVictim:
		return "victim-rank"
	case alarmRankPhase:
		return "phase-distribution"
	default:
		return "rank"
	}
}

func alarmRankMetadata(kind alarmRankKind) (use string, task string, summary string, example string) {
	switch kind {
	case alarmRankAttacker:
		return "attacker-rank", "查看攻击源排行", "查看攻击源排行，用于快速定位攻击最集中的来源对象。", "  chaitin-cli tanswer alarm attacker-rank --time today --top 10"
	case alarmRankVictim:
		return "victim-rank", "查看受害对象排行", "查看受害对象排行，用于快速定位被攻击最多或风险最高的对象。", "  chaitin-cli tanswer alarm victim-rank --time today --top 10"
	case alarmRankPhase:
		return "phase-distribution", "查看攻击链阶段分布", "查看攻击链阶段分布，用于判断当前时间范围或筛选条件下攻击主要集中在哪些阶段。", "  chaitin-cli tanswer alarm phase-distribution --time today"
	default:
		return "rank", "查看告警排行", "查看告警排行。", "  chaitin-cli tanswer alarm rank"
	}
}
