package tanswer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const metadataConfigUpdateConfirmToken = "CONFIRM_METADATA_CONFIG_UPDATE"

type metadataListOptions struct {
	time          string
	start         string
	end           string
	page          int
	pageSize      int
	protocol      string
	srcIP         string
	destIP        string
	srcPort       string
	destPort      string
	httpURL       string
	dnsRRName     string
	advancedQuery string
}

type metadataDetailOptions struct {
	id        string
	timestamp string
	protocol  string
}

type metadataConfigOptions struct {
	nodeID string
}

type metadataConfigWriteOptions struct {
	nodeID  string
	enable  string
	disable string
	preview bool
	confirm string
}

type metadataNearAlarmOptions struct {
	id       string
	window   string
	protocol string
	pageSize int
}

type metadataListRPCResult struct {
	Logs  []map[string]any `json:"logs"`
	Total int64            `json:"total"`
}

type metadataDetailRPCResult struct {
	Log any `json:"log"`
}

type metadataConfigRPCResult struct {
	Configurations []map[string]any `json:"configurations"`
}

func newMetadataCommand(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metadata",
		Short: "流量元数据语义命令",
		Long:  "流量元数据语义命令。用于检索 HTTP、DNS、TCP、UDP 和其他协议元数据，查看详情、查询告警附近上下文和读取元数据数据配置；配置调整使用受保护的 metadata config-update 命令。元数据不应直接判定为攻击证据。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newMetadataProtocolCommand(opts))
	cmd.AddCommand(newMetadataSearchCommand(opts))
	cmd.AddCommand(newMetadataDetailCommand(opts))
	cmd.AddCommand(newMetadataNearAlarmCommand(opts))
	cmd.AddCommand(newMetadataConfigCommand(opts))
	cmd.AddCommand(newMetadataConfigUpdateCommand(opts))
	return cmd
}

func newMetadataProtocolCommand(opts *RootOptions) *cobra.Command {
	var metaOpts metadataListOptions
	cmd := &cobra.Command{
		Use:   "protocol",
		Short: "按协议检索元数据",
		Long:  "按协议检索元数据，用于按 HTTP、DNS、TCP、UDP 或其他协议查看已有流量元数据摘要。该命令只读取元数据索引，返回稳定摘要字段和分页信息。\n\n输出：查询时间范围、实际筛选条件、total、page、page_size、current_count、has_more、metadata。",
		Example: "  chaitin-cli tanswer metadata protocol --protocol http --time today --page-size 10\n" +
			"  chaitin-cli tanswer metadata protocol --protocol DNS --src-ip 198.51.100.10 --page-size 20",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMetadataListCommand(cmd, opts, metaOpts, "按协议检索元数据", "chaitin-cli tanswer metadata protocol")
		},
	}
	addMetadataListFlags(cmd, &metaOpts)
	return cmd
}

func newMetadataSearchCommand(opts *RootOptions) *cobra.Command {
	var metaOpts metadataListOptions
	cmd := &cobra.Command{
		Use:   "search",
		Short: "按高级条件检索元数据",
		Long:  "按高级条件检索元数据，用于在指定协议内结合高级查询语句和基础五元组筛选检索已有流量元数据。该命令不保存查询历史。\n\n输出：查询时间范围、实际筛选条件、advanced_query、total、page、page_size、current_count、has_more、metadata。",
		Example: "  chaitin-cli tanswer metadata search --protocol dns --advanced-query \"dns_rrname = 'example.com'\"\n" +
			"  chaitin-cli tanswer metadata search --protocol http --advanced-query \"http_method = 'GET'\" --time 24h",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(metaOpts.advancedQuery) == "" {
				return writeMetadataError(cmd, "按高级条件检索元数据", "chaitin-cli tanswer metadata search", "MISSING_ADVANCED_QUERY", "missing advanced query: set --advanced-query", false)
			}
			return runMetadataListCommand(cmd, opts, metaOpts, "按高级条件检索元数据", "chaitin-cli tanswer metadata search")
		},
	}
	addMetadataListFlags(cmd, &metaOpts)
	return cmd
}

func newMetadataDetailCommand(opts *RootOptions) *cobra.Command {
	var metaOpts metadataDetailOptions
	cmd := &cobra.Command{
		Use:     "detail",
		Short:   "查看元数据详情",
		Long:    "查看元数据详情，用于从元数据列表返回的 id、timestamp 和 protocol 下钻到单条原始元数据详情。该命令只读取已有详情，不下载文件，不修改配置。\n\n输出：id、timestamp、protocol、detail。",
		Example: "  chaitin-cli tanswer metadata detail --id '<metadata_id>' --timestamp 1784282400000 --protocol http",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(metaOpts.id) == "" {
				return writeMetadataError(cmd, "查看元数据详情", "chaitin-cli tanswer metadata detail", "MISSING_METADATA_ID", "missing metadata id: set --id", false)
			}
			timestamp, err := strconv.ParseInt(strings.TrimSpace(metaOpts.timestamp), 10, 64)
			if err != nil || timestamp <= 0 {
				return writeMetadataError(cmd, "查看元数据详情", "chaitin-cli tanswer metadata detail", "INVALID_METADATA_TIMESTAMP", "timestamp must be a positive millisecond value", false)
			}
			info := metadataProtocolInfo(metaOpts.protocol)
			if info.eventType == "" {
				return writeMetadataError(cmd, "查看元数据详情", "chaitin-cli tanswer metadata detail", "MISSING_METADATA_PROTOCOL", "missing protocol: set --protocol", false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var result metadataDetailRPCResult
			req := map[string]any{
				"id":               metaOpts.id,
				"timestamp":        timestamp,
				"event_type":       info.eventType,
				"event_type_group": info.group,
			}
			if err := client.CallRPC(cmd.Context(), "LogSearchService.GetOrigDataLogDetail", req, &result); err != nil {
				return writeMetadataError(cmd, "查看元数据详情", "chaitin-cli tanswer metadata detail", "METADATA_DETAIL_FAILED", err.Error(), true)
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查看元数据详情",
				Command: "chaitin-cli tanswer metadata detail",
				Query:   req,
				Data: map[string]any{
					"id":        metaOpts.id,
					"timestamp": timestamp,
					"protocol":  info.eventType,
					"detail":    result.Log,
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&metaOpts.id, "id", "", "metadata id from metadata list commands")
	cmd.Flags().StringVar(&metaOpts.timestamp, "timestamp", "", "metadata timestamp in milliseconds")
	cmd.Flags().StringVar(&metaOpts.protocol, "protocol", "", "metadata protocol, for example http, dns, tcp, udp")
	return cmd
}

func newMetadataConfigCommand(opts *RootOptions) *cobra.Command {
	var metaOpts metadataConfigOptions
	cmd := &cobra.Command{
		Use:   "config",
		Short: "查询元数据数据配置",
		Long:  "查询元数据数据配置，用于只读查看当前节点或指定节点的元数据协议采集配置。该命令不调整配置。\n\n输出：node_id、configurations。",
		Example: "  chaitin-cli tanswer metadata config\n" +
			"  chaitin-cli tanswer metadata config --node-id '<node_id>'",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			req := map[string]any{}
			if strings.TrimSpace(metaOpts.nodeID) != "" {
				req["node_id"] = strings.TrimSpace(metaOpts.nodeID)
			}
			client := NewClient(cfg)
			var result metadataConfigRPCResult
			if err := client.CallRPC(cmd.Context(), "LogSearchService.GetOrigDataLogProtocolList", req, &result); err != nil {
				return writeMetadataError(cmd, "查询元数据数据配置", "chaitin-cli tanswer metadata config", "METADATA_CONFIG_FAILED", err.Error(), true)
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查询元数据数据配置",
				Command: "chaitin-cli tanswer metadata config",
				Query:   req,
				Data: map[string]any{
					"node_id":        metaOpts.nodeID,
					"configurations": result.Configurations,
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&metaOpts.nodeID, "node-id", "", "node id; empty means current/global node")
	return cmd
}

func newMetadataConfigUpdateCommand(opts *RootOptions) *cobra.Command {
	var metaOpts metadataConfigWriteOptions
	commandText := "chaitin-cli tanswer metadata config-update"
	cmd := &cobra.Command{
		Use:   "config-update",
		Short: "调整元数据数据配置",
		Long: "调整元数据数据配置，用于修改指定节点的元数据协议存储范围。该命令是高影响写操作：预览阶段会读取当前配置并返回 before/after；必须使用 --confirm CONFIRM_METADATA_CONFIG_UPDATE 才会调用后端更新接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer metadata config-update --node-id '<node_id>' --enable http,dns --preview\n" +
			"  chaitin-cli tanswer metadata config-update --node-id '<node_id>' --disable tcp,udp --confirm CONFIRM_METADATA_CONFIG_UPDATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(metaOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(metaOpts.confirm, metadataConfigUpdateConfirmToken); err != nil {
					return writeMetadataError(cmd, "调整元数据数据配置", commandText, "METADATA_CONFIG_UPDATE_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			readReq := map[string]any{"node_id": strings.TrimSpace(metaOpts.nodeID)}
			var current metadataConfigRPCResult
			if err := client.CallRPC(cmd.Context(), "LogSearchService.GetOrigDataLogProtocolList", readReq, &current); err != nil {
				return writeMetadataError(cmd, "调整元数据数据配置", commandText, "METADATA_CONFIG_UPDATE_READ_FAILED", err.Error(), true)
			}
			req, err := buildMetadataConfigUpdateRequest(current.Configurations, metaOpts)
			if err != nil {
				return writeMetadataError(cmd, "调整元数据数据配置", commandText, "INVALID_METADATA_CONFIG_UPDATE_REQUEST", err.Error(), false)
			}
			changes := metadataConfigChanges(current.Configurations, req["configurations"].([]map[string]any))
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "调整元数据数据配置预览",
				Command:       commandText,
				OperationType: "metadata_config_update",
				RiskLevel:     "write_high",
				Target:        map[string]any{"node_id": req["node_id"], "enable": parseCSV(metaOpts.enable), "disable": parseCSV(metaOpts.disable)},
				ChangeSummary: map[string]any{"before": current.Configurations, "after": req["configurations"], "changes": changes},
				Impact:        map[string]any{"node_id": req["node_id"], "changed_protocol_count": len(changes), "configuration_count": len(req["configurations"].([]map[string]any))},
				RiskWarnings: []string{
					"将修改元数据协议存储范围，停用协议后对应流量元数据可能不再保存。",
					"执行前应确认节点和协议范围，避免影响后续检索、研判和取证。",
				},
				ConfirmToken: metadataConfigUpdateConfirmToken,
			})
			if metaOpts.preview || strings.TrimSpace(metaOpts.confirm) == "" {
				return writeMetadataSuccess(cmd, "调整元数据数据配置预览", commandText, req, preview)
			}
			var result struct{}
			if err := client.CallRPC(cmd.Context(), "LogSearchService.ConfigureOrigDataLogProtocol", req, &result); err != nil {
				return writeMetadataError(cmd, "调整元数据数据配置", commandText, "METADATA_CONFIG_UPDATE_FAILED", err.Error(), true)
			}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "metadata_config_update",
				Object:        map[string]any{"node_id": req["node_id"], "changes": changes},
				Action:        "update",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   map[string]any{"before": current.Configurations, "after": req["configurations"]},
				Result:        "success",
			})
			return writeMetadataSuccess(cmd, "调整元数据数据配置", commandText, req, data)
		},
	}
	cmd.Flags().StringVar(&metaOpts.nodeID, "node-id", "", "node id to update, required")
	cmd.Flags().StringVar(&metaOpts.enable, "enable", "", "protocol event_type list to enable, comma separated")
	cmd.Flags().StringVar(&metaOpts.disable, "disable", "", "protocol event_type list to disable, comma separated")
	cmd.Flags().BoolVar(&metaOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&metaOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newMetadataNearAlarmCommand(opts *RootOptions) *cobra.Command {
	var metaOpts metadataNearAlarmOptions
	cmd := &cobra.Command{
		Use:   "near-alarm",
		Short: "查询告警附近元数据",
		Long:  "查询告警附近元数据，用于围绕一条威胁告警的时间点和五元组检索附近流量元数据，辅助还原上下文。返回结果仅代表同时间窗内的上下文，不直接判定为攻击证据。\n\n输出：alarm、window、metadata_query、total、page_size、current_count、metadata。",
		Example: "  chaitin-cli tanswer metadata near-alarm --id '<alarm_doc_id>' --window 30m --page-size 10\n" +
			"  chaitin-cli tanswer metadata near-alarm --id '<alarm_doc_id>' --protocol http --window 10m",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(metaOpts.id) == "" {
				return writeMetadataError(cmd, "查询告警附近元数据", "chaitin-cli tanswer metadata near-alarm", "MISSING_ALARM_ID", "missing alarm doc_id: set --id", false)
			}
			if metaOpts.pageSize < 1 || metaOpts.pageSize > 100 {
				return writeMetadataError(cmd, "查询告警附近元数据", "chaitin-cli tanswer metadata near-alarm", "INVALID_PAGE_SIZE", "page-size must be between 1 and 100", false)
			}
			window, err := time.ParseDuration(strings.TrimSpace(metaOpts.window))
			if err != nil || window <= 0 {
				return writeMetadataError(cmd, "查询告警附近元数据", "chaitin-cli tanswer metadata near-alarm", "INVALID_WINDOW", "window must be a positive duration, for example 10m or 1h", false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var alarmResult alarmDetailRPCResult
			if err := client.CallRPC(cmd.Context(), "AlarmService.GetAlarm", map[string]any{"doc_id": metaOpts.id}, &alarmResult); err != nil {
				return writeMetadataError(cmd, "查询告警附近元数据", "chaitin-cli tanswer metadata near-alarm", "ALARM_DETAIL_FAILED", err.Error(), true)
			}
			protocol := firstString(firstPresent(alarmResult.Data, "app_proto", "proto", "event_type"))
			if strings.TrimSpace(metaOpts.protocol) != "" {
				protocol = metaOpts.protocol
			}
			info := metadataProtocolInfo(protocol)
			if info.eventType == "" {
				info = metadataProtocolInfo("http")
			}
			timestamp := int64Value(firstPresent(alarmResult.Data, "timestamp", "time"))
			rng := TimeRange{
				Type:  "near_alarm",
				Start: timestamp - window.Milliseconds(),
				End:   timestamp + window.Milliseconds(),
			}
			listOpts := metadataListOptions{
				page:     1,
				pageSize: metaOpts.pageSize,
				protocol: info.eventType,
				srcIP:    firstString(firstPresent(alarmResult.Data, "src_ip", "attacker")),
				destIP:   firstString(firstPresent(alarmResult.Data, "dest_ip", "victim")),
				srcPort:  firstString(firstPresent(alarmResult.Data, "src_port")),
				destPort: firstString(firstPresent(alarmResult.Data, "dest_port")),
			}
			req := buildMetadataListRequest(rng, listOpts)
			var result metadataListRPCResult
			if err := client.CallRPC(cmd.Context(), info.searchMethod, req, &result); err != nil {
				return writeMetadataError(cmd, "查询告警附近元数据", "chaitin-cli tanswer metadata near-alarm", "METADATA_NEAR_ALARM_FAILED", err.Error(), true)
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查询告警附近元数据",
				Command: "chaitin-cli tanswer metadata near-alarm",
				Query: map[string]any{
					"alarm_id":       metaOpts.id,
					"window":         metaOpts.window,
					"metadata_query": req,
				},
				Data: map[string]any{
					"alarm":         summarizeMetadataAlarm(alarmResult.Data),
					"window":        metaOpts.window,
					"total":         result.Total,
					"page_size":     metaOpts.pageSize,
					"current_count": len(result.Logs),
					"metadata":      summarizeMetadataLogs(result.Logs),
					"note":          "metadata near an alarm is context only, not direct attack evidence",
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&metaOpts.id, "id", "", "alarm doc_id from alarm list/detail commands")
	cmd.Flags().StringVar(&metaOpts.window, "window", "30m", "time window around alarm timestamp, for example 10m or 1h")
	cmd.Flags().StringVar(&metaOpts.protocol, "protocol", "", "override metadata protocol; default uses alarm app_proto/proto")
	cmd.Flags().IntVar(&metaOpts.pageSize, "page-size", 10, "page size, 1-100")
	return cmd
}

func runMetadataListCommand(cmd *cobra.Command, opts *RootOptions, metaOpts metadataListOptions, task string, command string) error {
	if metaOpts.page < 1 {
		return writeMetadataError(cmd, task, command, "INVALID_PAGE", "page must be greater than or equal to 1", false)
	}
	if metaOpts.pageSize < 1 || metaOpts.pageSize > 100 {
		return writeMetadataError(cmd, task, command, "INVALID_PAGE_SIZE", "page-size must be between 1 and 100", false)
	}
	info := metadataProtocolInfo(metaOpts.protocol)
	if info.eventType == "" {
		return writeMetadataError(cmd, task, command, "MISSING_METADATA_PROTOCOL", "missing protocol: set --protocol", false)
	}
	cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
	if err != nil {
		return err
	}
	rng, err := ParseTimeRange(TimeRangeOptions{Time: metaOpts.time, Start: metaOpts.start, End: metaOpts.end})
	if err != nil {
		return writeMetadataError(cmd, task, command, "INVALID_TIME_RANGE", err.Error(), false)
	}
	req := buildMetadataListRequest(rng, metaOpts)
	client := NewClient(cfg)
	var result metadataListRPCResult
	if err := client.CallRPC(cmd.Context(), info.searchMethod, req, &result); err != nil {
		return writeMetadataError(cmd, task, command, "METADATA_LIST_FAILED", err.Error(), true)
	}
	data := map[string]any{
		"total":         result.Total,
		"page":          metaOpts.page,
		"page_size":     metaOpts.pageSize,
		"current_count": len(result.Logs),
		"has_more":      int64(metaOpts.page*metaOpts.pageSize) < result.Total,
		"metadata":      summarizeMetadataLogs(result.Logs),
	}
	raw, err := RenderJSON(SuccessEnvelope{
		Success: true,
		Task:    task,
		Command: command,
		Query: map[string]any{
			"time_range": rng,
			"filters":    metadataListFilters(metaOpts),
			"page":       metaOpts.page,
			"page_size":  metaOpts.pageSize,
		},
		Data: data,
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func addMetadataListFlags(cmd *cobra.Command, opts *metadataListOptions) {
	cmd.Flags().StringVar(&opts.time, "time", "today", "time range: today, 24h, 7d")
	cmd.Flags().StringVar(&opts.start, "start", "", "custom start time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&opts.end, "end", "", "custom end time, format: 2006-01-02 15:04:05")
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().StringVar(&opts.protocol, "protocol", "", "metadata protocol: http, http2, dns, tcp, udp or other event_type")
	cmd.Flags().StringVar(&opts.srcIP, "src-ip", "", "source IP filter, comma separated")
	cmd.Flags().StringVar(&opts.destIP, "dest-ip", "", "destination IP filter, comma separated")
	cmd.Flags().StringVar(&opts.srcPort, "src-port", "", "source port filter, comma separated")
	cmd.Flags().StringVar(&opts.destPort, "dest-port", "", "destination port filter, comma separated")
	cmd.Flags().StringVar(&opts.httpURL, "http-url", "", "HTTP URL filter, comma separated")
	cmd.Flags().StringVar(&opts.dnsRRName, "dns-rrname", "", "DNS rrname filter, comma separated")
	cmd.Flags().StringVar(&opts.advancedQuery, "advanced-query", "", "advanced query expression")
}

func buildMetadataListRequest(rng TimeRange, opts metadataListOptions) map[string]any {
	if opts.page < 1 {
		opts.page = 1
	}
	if opts.pageSize < 1 {
		opts.pageSize = 10
	}
	info := metadataProtocolInfo(opts.protocol)
	req := map[string]any{
		"time_range_start": rng.Start,
		"time_range_end":   rng.End,
		"offset":           int64((opts.page - 1) * opts.pageSize),
		"count":            int64(opts.pageSize),
		"save_history":     false,
	}
	if info.eventType != "" {
		req["event_type"] = info.eventType
	}
	if value := strings.TrimSpace(opts.advancedQuery); value != "" {
		req["advanced_query"] = value
	}
	for key, value := range metadataListFilters(opts) {
		req[key] = value
	}
	return req
}

func metadataListFilters(opts metadataListOptions) map[string]any {
	filters := map[string]any{}
	addStringFilter(filters, "src_ip", opts.srcIP, strings.TrimSpace)
	addStringFilter(filters, "dest_ip", opts.destIP, strings.TrimSpace)
	addIntFilter(filters, "src_port", opts.srcPort, portValue)
	addIntFilter(filters, "dest_port", opts.destPort, portValue)
	addStringFilter(filters, "http_url", opts.httpURL, strings.TrimSpace)
	addStringFilter(filters, "dns_rrname", opts.dnsRRName, strings.TrimSpace)
	info := metadataProtocolInfo(opts.protocol)
	if info.group == "tcp/udp" {
		addStringFilter(filters, "proto", info.eventType, strings.ToLower)
	}
	return filters
}

type metadataProtocol struct {
	eventType    string
	group        string
	searchMethod string
}

func metadataProtocolInfo(protocol string) metadataProtocol {
	value := strings.ToLower(strings.TrimSpace(protocol))
	switch value {
	case "":
		return metadataProtocol{}
	case "http", "http2":
		return metadataProtocol{eventType: value, group: "http", searchMethod: "LogSearchService.SearchOrigDataHTTPLog"}
	case "dns":
		return metadataProtocol{eventType: "dns", group: "dns", searchMethod: "LogSearchService.SearchOrigDataDNSLog"}
	case "tcp", "udp":
		return metadataProtocol{eventType: value, group: "tcp/udp", searchMethod: "LogSearchService.SearchOrigDataTCPUDPLog"}
	default:
		return metadataProtocol{eventType: value, group: "other", searchMethod: "LogSearchService.SearchOtherOrigDataLog"}
	}
}

func buildMetadataConfigUpdateRequest(current []map[string]any, opts metadataConfigWriteOptions) (map[string]any, error) {
	nodeID := strings.TrimSpace(opts.nodeID)
	if nodeID == "" {
		return nil, fmt.Errorf("missing node id: set --node-id")
	}
	enable := metadataProtocolSet(opts.enable)
	disable := metadataProtocolSet(opts.disable)
	if len(enable) == 0 && len(disable) == 0 {
		return nil, fmt.Errorf("missing protocol changes: set --enable or --disable")
	}
	for protocol := range enable {
		if disable[protocol] {
			return nil, fmt.Errorf("protocol %q appears in both --enable and --disable", protocol)
		}
	}
	configs := make([]map[string]any, 0, len(current))
	known := map[string]bool{}
	for _, item := range current {
		eventType := strings.ToLower(strings.TrimSpace(firstString(item["event_type"])))
		if eventType == "" {
			continue
		}
		known[eventType] = true
		selected := metadataBoolValue(item["selected"])
		if enable[eventType] {
			selected = true
		}
		if disable[eventType] {
			selected = false
		}
		configs = append(configs, map[string]any{
			"event_type": eventType,
			"selected":   selected,
		})
	}
	for protocol := range enable {
		if !known[protocol] {
			return nil, fmt.Errorf("protocol %q is not present in current metadata config", protocol)
		}
	}
	for protocol := range disable {
		if !known[protocol] {
			return nil, fmt.Errorf("protocol %q is not present in current metadata config", protocol)
		}
	}
	return map[string]any{
		"node_id":        nodeID,
		"configurations": configs,
	}, nil
}

func metadataProtocolSet(value string) map[string]bool {
	out := map[string]bool{}
	for _, protocol := range parseCSV(value) {
		normalized := strings.ToLower(strings.TrimSpace(protocol))
		if normalized != "" {
			out[normalized] = true
		}
	}
	return out
}

func metadataBoolValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		parsed, _ := strconv.ParseBool(strings.TrimSpace(typed))
		return parsed
	default:
		return false
	}
}

func metadataConfigChanges(before []map[string]any, after []map[string]any) []map[string]any {
	beforeMap := map[string]bool{}
	for _, item := range before {
		eventType := strings.ToLower(strings.TrimSpace(firstString(item["event_type"])))
		if eventType != "" {
			beforeMap[eventType] = metadataBoolValue(item["selected"])
		}
	}
	changes := []map[string]any{}
	for _, item := range after {
		eventType := strings.ToLower(strings.TrimSpace(firstString(item["event_type"])))
		if eventType == "" {
			continue
		}
		afterSelected := metadataBoolValue(item["selected"])
		beforeSelected, ok := beforeMap[eventType]
		if ok && beforeSelected != afterSelected {
			changes = append(changes, map[string]any{
				"event_type": eventType,
				"before":     beforeSelected,
				"after":      afterSelected,
			})
		}
	}
	return changes
}

func summarizeMetadataLogs(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		summary := map[string]any{}
		for _, key := range []string{
			"id",
			"timestamp",
			"event_type",
			"src_ip",
			"src_port",
			"dest_ip",
			"dest_port",
			"proto",
			"app_proto",
			"src_mac",
			"dest_mac",
			"node_id",
			"node_name",
		} {
			copyIfPresent(summary, item, key)
		}
		addMetadataNestedSummary(summary, item)
		out = append(out, summary)
	}
	return out
}

func addMetadataNestedSummary(summary map[string]any, item map[string]any) {
	if httpInfo, ok := item["http"].(map[string]any); ok {
		for _, key := range []string{"hostname", "url", "req_line", "resp_line", "method", "status", "user_agent"} {
			copyIfPresent(summary, httpInfo, key)
		}
	}
	if dnsInfo, ok := item["dns"].(map[string]any); ok {
		for _, key := range []string{"rrname", "rrtype", "rcode", "answers"} {
			copyIfPresent(summary, dnsInfo, key)
		}
	}
	if flowInfo, ok := item["flow"].(map[string]any); ok {
		summary["flow"] = flowInfo
	}
	if tcpInfo, ok := item["tcp"].(map[string]any); ok {
		summary["tcp"] = tcpInfo
	}
	if udpInfo, ok := item["udp"].(map[string]any); ok {
		summary["udp"] = udpInfo
	}
}

func summarizeMetadataAlarm(in map[string]any) map[string]any {
	out := map[string]any{}
	for _, key := range []string{"doc_id", "id", "timestamp", "name", "severity", "src_ip", "src_port", "dest_ip", "dest_port", "attacker", "victim", "app_proto", "proto"} {
		copyIfPresent(out, in, key)
	}
	return out
}

func firstString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case jsonNumber:
		return string(v)
	default:
		return ""
	}
}

func int64Value(value any) int64 {
	if n, ok := anyToInt64(value); ok {
		return n
	}
	return 0
}

func writeMetadataSuccess(cmd *cobra.Command, task string, command string, query any, data map[string]any) error {
	raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: task, Command: command, Query: query, Data: data})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func writeMetadataError(cmd *cobra.Command, task string, command string, code string, message string, retryable bool) error {
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
