package tanswer

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type alarmTimelineOptions struct {
	time      string
	start     string
	end       string
	interval  string
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

type alarmTimelineRPCResult struct {
	Data []map[string]any `json:"data"`
}

func newAlarmTimelineCommand(opts *RootOptions) *cobra.Command {
	var alarmOpts alarmTimelineOptions
	cmd := &cobra.Command{
		Use:   "timeline",
		Short: "查看威胁告警趋势",
		Long: "查看威胁告警趋势，用于按时间曲线判断告警爆发、回落和等级变化。该命令返回时间点、告警数量和等级分布，不返回原始告警列表。\n\n" +
			"输出：查询时间范围、实际筛选条件、interval、point_count、points。每个曲线点包含 key、doc_count 和 severity buckets。",
		Example: "  chaitin-cli tanswer alarm timeline --time today\n" +
			"  chaitin-cli tanswer alarm timeline --time 24h --interval 1h --severity critical,high",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlarmTimelineCommand(cmd, opts, alarmOpts)
		},
	}
	addAlarmTimelineFlags(cmd, &alarmOpts)
	return cmd
}

func runAlarmTimelineCommand(cmd *cobra.Command, opts *RootOptions, alarmOpts alarmTimelineOptions) error {
	const task = "查看威胁告警趋势"
	const command = "chaitin-cli tanswer alarm timeline"

	cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
	if err != nil {
		return err
	}
	rng, err := ParseTimeRange(TimeRangeOptions{Time: alarmOpts.time, Start: alarmOpts.start, End: alarmOpts.end})
	if err != nil {
		return writeAlarmListError(cmd, task, command, "INVALID_TIME_RANGE", err.Error(), false)
	}

	req := buildAlarmTimelineRequest(rng, alarmOpts)
	client := NewClient(cfg)
	var result alarmTimelineRPCResult
	if err := client.CallRPC(cmd.Context(), "AlarmService.SearchAlarmListChart", req, &result); err != nil {
		return writeAlarmListError(cmd, task, command, "ALARM_TIMELINE_FAILED", err.Error(), true)
	}

	interval := strings.TrimSpace(alarmOpts.interval)
	if interval == "" {
		interval = "auto"
	}
	data := map[string]any{
		"interval":    interval,
		"range_mode":  int64(0),
		"point_count": len(result.Data),
		"points":      summarizeAlarmTimelinePoints(result.Data),
	}
	raw, err := RenderJSON(SuccessEnvelope{
		Success: true,
		Task:    task,
		Command: command,
		Query: map[string]any{
			"time_range": rng,
			"filters":    alarmTimelineFilters(alarmOpts),
			"interval":   interval,
		},
		Data: data,
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func buildAlarmTimelineRequest(rng TimeRange, opts alarmTimelineOptions) map[string]any {
	req := map[string]any{
		"time_range_start": rng.Start,
		"time_range_end":   rng.End,
		"interval":         strings.TrimSpace(opts.interval),
		"range_mode":       int64(0),
	}
	for key, value := range alarmTimelineFilters(opts) {
		req[key] = value
	}
	return req
}

func alarmTimelineFilters(opts alarmTimelineOptions) map[string]any {
	return alarmListFilters(alarmListOptions{
		severity:  opts.severity,
		result:    opts.result,
		phase:     opts.phase,
		attacker:  opts.attacker,
		victim:    opts.victim,
		assetIP:   opts.assetIP,
		keyword:   opts.keyword,
		name:      opts.name,
		tag:       opts.tag,
		direction: opts.direction,
		appProto:  opts.appProto,
		url:       opts.url,
		host:      opts.host,
		xff:       opts.xff,
		srcPort:   opts.srcPort,
		destPort:  opts.destPort,
	})
}

func summarizeAlarmTimelinePoints(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		point := map[string]any{}
		for _, key := range []string{"key", "doc_count", "severity"} {
			if value, ok := item[key]; ok {
				point[key] = value
			}
		}
		out = append(out, point)
	}
	return out
}

func addAlarmTimelineFlags(cmd *cobra.Command, opts *alarmTimelineOptions) {
	cmd.Flags().StringVar(&opts.time, "time", "today", "time range: today, 24h, 7d")
	cmd.Flags().StringVar(&opts.start, "start", "", "custom start time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&opts.end, "end", "", "custom end time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&opts.interval, "interval", "", "chart interval, for example 5m, 1h, 24h; empty means backend auto")
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
