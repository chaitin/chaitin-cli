package tanswer

import (
	"fmt"

	"github.com/spf13/cobra"
)

type CommandManifest struct {
	SchemaVersion string            `json:"schema_version"`
	Product       string            `json:"product"`
	Binary        string            `json:"binary"`
	Namespace     string            `json:"namespace"`
	Env           map[string]string `json:"env"`
	Bootstrap     BootstrapManifest `json:"bootstrap"`
	Commands      []ManifestCommand `json:"commands"`
}

type ConfigSource struct {
	Flag       string `json:"flag,omitempty"`
	Env        string `json:"env,omitempty"`
	ConfigPath string `json:"config_path,omitempty"`
	Secret     bool   `json:"secret,omitempty"`
}

type AuthenticationManifest struct {
	URL           ConfigSource `json:"url"`
	APIKey        ConfigSource `json:"api_key"`
	Timeout       ConfigSource `json:"timeout"`
	Insecure      ConfigSource `json:"insecure"`
	Precedence    []string     `json:"precedence"`
	StatusCommand string       `json:"status_command"`
	VerifyCommand string       `json:"verify_command"`
}

type BootstrapManifest struct {
	Authentication AuthenticationManifest `json:"authentication"`
	AgentProtocol  []string               `json:"agent_protocol"`
}

type ManifestCommand struct {
	Name                  string                    `json:"name"`
	FullCommand           string                    `json:"full_command"`
	Layer                 string                    `json:"layer"`
	Summary               string                    `json:"summary"`
	UseWhen               []string                  `json:"use_when"`
	DoNotUseWhen          []string                  `json:"do_not_use_when,omitempty"`
	Arguments             []ManifestArgument        `json:"arguments,omitempty"`
	Flags                 []ManifestFlag            `json:"flags,omitempty"`
	InputConstraints      []ManifestInputConstraint `json:"input_constraints,omitempty"`
	OutputType            string                    `json:"output_type"`
	OutputFields          []string                  `json:"output_fields"`
	PreviewOutputFields   []string                  `json:"preview_output_fields,omitempty"`
	RiskLevel             string                    `json:"risk_level"`
	RequiresConfirmation  bool                      `json:"requires_confirmation"`
	ConfirmationCondition string                    `json:"confirmation_condition,omitempty"`
	Examples              []ManifestExample         `json:"examples"`
	Backend               map[string]interface{}    `json:"backend,omitempty"`
}

type ManifestInputConstraint struct {
	Rule        string   `json:"rule"`
	Flags       []string `json:"flags"`
	Description string   `json:"description"`
}

type ManifestArgument struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type ManifestFlag struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Default     string   `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description"`
}

type ManifestExample struct {
	Description string `json:"description"`
	Command     string `json:"command"`
}

func newManifestCommand(opts *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "manifest",
		Short:   "查看 CLI 命令清单",
		Long:    "查看 CLI 命令清单，并输出全悉 CLI 的机器可读命令、配置和安全契约。面向 AI Agent 和集成方返回机器可读的命令元数据，包括命令层级、参数、默认值、枚举、output 类型、risk 等级、确认要求和示例。",
		Example: "  chaitin-cli tanswer manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查看 CLI 命令清单",
				Command: "chaitin-cli tanswer manifest",
				Data:    BuildCommandManifest(),
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
}

func BuildCommandManifest() CommandManifest {
	manifest := CommandManifest{
		SchemaVersion: "2026-07-29",
		Product:       "tanswer",
		Binary:        "chaitin-cli",
		Namespace:     "tanswer",
		Env: map[string]string{
			"TANSWER_URL":      "Quanxi console address, for example https://<全悉 Web 端 IP>",
			"TANSWER_API_KEY":  "OpenAPI Token used by Quanxi TokenAuth",
			"TANSWER_TIMEOUT":  "request timeout, default 30s",
			"TANSWER_INSECURE": "skip TLS certificate verification when certificate validation must be bypassed, default false",
		},
		Bootstrap: BootstrapManifest{
			Authentication: AuthenticationManifest{
				URL:           ConfigSource{Flag: "--url", Env: "TANSWER_URL", ConfigPath: "tanswer.url"},
				APIKey:        ConfigSource{Flag: "--api-key", Env: "TANSWER_API_KEY", ConfigPath: "tanswer.api_key", Secret: true},
				Timeout:       ConfigSource{Flag: "--timeout", Env: "TANSWER_TIMEOUT", ConfigPath: "tanswer.timeout"},
				Insecure:      ConfigSource{Flag: "--insecure", Env: "TANSWER_INSECURE", ConfigPath: "tanswer.insecure"},
				Precedence:    []string{"command flag", "environment variable", "config.yaml", "default"},
				StatusCommand: "chaitin-cli tanswer auth status",
				VerifyCommand: "chaitin-cli tanswer auth check",
			},
			AgentProtocol: []string{
				"discover commands and unknown flags with --help",
				"prefer semantic commands; use api only for known authorized endpoints not covered by semantic commands",
				"run preview before protected writes and require the documented confirmation token",
				"wait for explicit user confirmation after preview before executing a protected write",
				"for api fallback, execute non-GET/HEAD requests only after preview, explicit user confirmation, and CONFIRM_TANSWER_RAW_API_WRITE",
				"do not guess unsupported commands or direct API endpoints",
			},
		},
		Commands: []ManifestCommand{
			{
				Name:        "tanswer manifest",
				FullCommand: "chaitin-cli tanswer manifest",
				Layer:       "foundation",
				Summary:     "输出全悉 CLI 的机器可读命令、配置和安全契约。",
				UseWhen: []string{
					"需要枚举当前版本支持的命令、参数、输出字段、风险和确认要求。",
					"AI Agent 需要在读取 help 后结构化确认命令契约。",
				},
				OutputType:           "command_manifest",
				OutputFields:         []string{"schema_version", "bootstrap", "commands"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "输出当前版本命令清单", Command: "chaitin-cli tanswer manifest"},
				},
			},
			{
				Name:        "tanswer auth status",
				FullCommand: "chaitin-cli tanswer auth status",
				Layer:       "foundation",
				Summary:     "查看当前连接状态，不校验 Token 权限。",
				UseWhen: []string{
					"需要确认 CLI 当前指向哪个全悉环境。",
					"需要确认是否已经配置 OpenAPI Token。",
				},
				DoNotUseWhen: []string{
					"需要确认 Token 是否有效或具备接口权限时，使用 tanswer auth check。",
				},
				OutputType:           "connection_status",
				OutputFields:         []string{"environment", "token_set", "timeout", "insecure_skip_verify"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查看当前配置", Command: "chaitin-cli tanswer auth status --url 'https://<全悉 Web 端 IP>'"},
				},
			},
			{
				Name:        "tanswer auth check",
				FullCommand: "chaitin-cli tanswer auth check",
				Layer:       "foundation",
				Summary:     "校验 OpenAPI Token 是否能访问当前全悉环境。",
				UseWhen: []string{
					"执行语义任务或 Open API 兜底调用前，需要确认 Token 链路可用。",
					"排查地址、Token 或网络配置是否正确。",
				},
				OutputType:           "token_check_result",
				OutputFields:         []string{"token_available", "base_info"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "校验指定环境和 Token", Command: "chaitin-cli tanswer auth check --url 'https://<全悉 Web 端 IP>' --api-key \"<全悉 OpenAPI Token>\""},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"OpsService.GetBaseInfo", "AssetService.SearchTags"},
				},
			},
			{
				Name:        "tanswer api",
				FullCommand: "chaitin-cli tanswer api <METHOD> <PATH>",
				Layer:       "openapi_fallback",
				Summary:     "调用语义快捷层未覆盖、但用户已知且已授权的全悉 Open API。",
				UseWhen: []string{
					"目标能力已有 Open API，但没有对应语义命令。",
					"AI Agent 已经从 Open API 文档中确定接口路径、方法和参数。",
				},
				DoNotUseWhen: []string{
					"已有语义快捷命令可以满足目标任务时，优先使用语义快捷命令。",
					"需要 CLI 提供专属参数解释、字段摘要或业务口径时，不应依赖该兜底入口。",
					"尚未取得用户对具体请求的明确确认时，不得执行非 GET/HEAD 请求。",
				},
				Arguments: []ManifestArgument{
					{Name: "METHOD", Required: true, Description: "HTTP method, for example GET or POST"},
					{Name: "PATH", Required: true, Description: "Open API path, for example /rpc"},
				},
				Flags: []ManifestFlag{
					{Name: "--query", Type: "json_object", Required: false, Description: "query parameters as JSON object"},
					{Name: "--body", Type: "json_object_or_file", Required: false, Description: "request body as JSON object or @file path"},
					{Name: "--preview", Type: "boolean", Required: false, Description: "return the non-GET/HEAD request preview without sending it"},
					{Name: "--confirm", Type: "string", Required: false, Description: "required with CONFIRM_TANSWER_RAW_API_WRITE to execute a non-GET/HEAD request after explicit user confirmation"},
				},
				OutputType:            "raw_openapi_response",
				OutputFields:          []string{"status_code", "raw"},
				PreviewOutputFields:   []string{"requires_confirmation", "confirmed", "operation_type", "risk_level", "target", "change_summary", "impact", "risk_warnings", "confirmation_token", "confirmation_note"},
				RiskLevel:             "read_write",
				RequiresConfirmation:  true,
				ConfirmationCondition: "required for non-GET/HEAD requests",
				Examples: []ManifestExample{
					{Description: "预览已知但未封装的非 GET/HEAD Open API 请求；占位符必须按官方 OpenAPI 文档替换", Command: "chaitin-cli tanswer api POST '/<known-unwrapped-openapi-path>' --body '{\"<known_parameter>\":\"<value>\"}' --preview"},
					{Description: "查看 Open API 文档", Command: "chaitin-cli tanswer api GET /api/openapi/rpc/openapi.json"},
				},
			},
			{
				Name:        "tanswer alarm overview",
				FullCommand: "chaitin-cli tanswer alarm overview",
				Layer:       "semantic_shortcut",
				Summary:     "查看威胁告警概览，用于值班巡检和态势快速判断。",
				UseWhen: []string{
					"需要回答今天、近 24 小时或近 7 天的威胁告警整体情况。",
					"需要等级分布、攻击结果分布、攻击阶段分布、威胁类型 Top、攻击源 Top 或受害对象 Top。",
				},
				DoNotUseWhen: []string{
					"需要逐条查看原始告警列表或告警详情时，使用后续列表或详情语义任务。",
				},
				Flags: []ManifestFlag{
					{Name: "--time", Type: "string", Required: false, Default: "today", Enum: []string{"today", "24h", "7d"}, Description: "preset time range"},
					{Name: "--start", Type: "datetime", Required: false, Description: "custom start time, format 2006-01-02 15:04:05"},
					{Name: "--end", Type: "datetime", Required: false, Description: "custom end time, format 2006-01-02 15:04:05"},
					{Name: "--severity", Type: "csv_enum", Required: false, Enum: []string{"critical", "high", "medium", "low", "1", "2", "3", "4"}, Description: "severity filter"},
					{Name: "--result", Type: "csv_enum", Required: false, Enum: []string{"success", "control", "failed", "unknown"}, Description: "attack result filter"},
					{Name: "--phase", Type: "csv_enum", Required: false, Enum: []string{"recon", "intrustion", "delivery", "control", "lateral", "goal", "other"}, Description: "attack phase filter"},
				},
				OutputType: "alarm_overview",
				OutputFields: []string{
					"summary.alarm_total",
					"summary.source",
					"summary.current_count",
					"summary.partial",
					"severity_distribution",
					"result_distribution",
					"phase_distribution",
					"threat_type_top",
					"attacker_top",
					"victim_top",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查看今天告警概览", Command: "chaitin-cli tanswer alarm overview --time today"},
					{Description: "查看近 24 小时高优先级成功或失陷告警概览", Command: "chaitin-cli tanswer alarm overview --time 24h --severity critical,high --result success,control"},
				},
				Backend: map[string]interface{}{
					"rpc_methods":          []string{"AlarmService.SearchAlarmCount", "AlarmService.SearchAlarmAggTop"},
					"fallback_rpc_methods": []string{"AlarmService.SearchAlarmList"},
					"fallback_note":        "If aggregation methods require session login in the current environment, the command falls back to list sampling and marks summary.source=list_fallback.",
				},
			},
			{
				Name:        "tanswer alarm list",
				FullCommand: "chaitin-cli tanswer alarm list",
				Layer:       "semantic_shortcut",
				Summary:     "查询威胁告警列表，用于从概览下钻到原始告警列表。",
				UseWhen: []string{
					"需要按时间、等级、攻击结果、攻击源、受害对象、资产、威胁名称、威胁类型等条件查询原始告警列表。",
					"需要获取 doc_id 后继续查看单条告警详情。",
				},
				DoNotUseWhen: []string{
					"只需要态势概览、分布或 Top 排行时，使用 tanswer alarm overview。",
					"只需要高优先级成功或失陷告警时，优先使用 tanswer alarm high-priority。",
				},
				Flags:      alarmListManifestFlags(false),
				OutputType: "alarm_list",
				OutputFields: []string{
					"total",
					"page_total",
					"page",
					"page_size",
					"current_count",
					"has_more",
					"alarms",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询今天告警列表", Command: "chaitin-cli tanswer alarm list --time today --page-size 10"},
					{Description: "查询近 24 小时成功或失陷的高危告警列表", Command: "chaitin-cli tanswer alarm list --time 24h --severity critical,high --result success,control"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.SearchAlarmList"},
				},
			},
			{
				Name:        "tanswer alarm timeline",
				FullCommand: "chaitin-cli tanswer alarm timeline",
				Layer:       "semantic_shortcut",
				Summary:     "查看威胁告警趋势，用于按时间曲线判断告警爆发、回落和等级变化。",
				UseWhen: []string{
					"需要回答今天、近 24 小时或近 7 天告警数量随时间如何变化。",
					"需要按等级观察告警曲线、峰值或爆发时间段。",
				},
				DoNotUseWhen: []string{
					"需要整体态势、Top 排行或分布时，使用 tanswer alarm overview。",
					"需要逐条查看原始告警时，使用 tanswer alarm list。",
				},
				Flags:      alarmTimelineManifestFlags(),
				OutputType: "alarm_timeline",
				OutputFields: []string{
					"interval",
					"range_mode",
					"point_count",
					"points",
					"points[].key",
					"points[].doc_count",
					"points[].severity.buckets",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查看今天告警趋势", Command: "chaitin-cli tanswer alarm timeline --time today"},
					{Description: "查看近 24 小时高危告警小时级趋势", Command: "chaitin-cli tanswer alarm timeline --time 24h --interval 1h --severity critical,high"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.SearchAlarmListChart"},
					"range_mode":  "0 means backend/default automatic bucketing",
				},
			},
			{
				Name:        "tanswer alarm high-priority",
				FullCommand: "chaitin-cli tanswer alarm high-priority",
				Layer:       "semantic_shortcut",
				Summary:     "查询高优先级威胁告警，默认筛选超危/高危且攻击结果为成功/失陷。",
				UseWhen: []string{
					"需要值班优先处置列表。",
					"用户询问高优先级、高危、超危、成功或失陷告警。",
				},
				DoNotUseWhen: []string{
					"需要任意筛选条件的通用告警列表时，使用 tanswer alarm list。",
				},
				Flags:      alarmListManifestFlags(true),
				OutputType: "alarm_list",
				OutputFields: []string{
					"total",
					"page_total",
					"page",
					"page_size",
					"current_count",
					"has_more",
					"alarms",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询今天高优先级告警", Command: "chaitin-cli tanswer alarm high-priority --time today"},
					{Description: "查询重点资产相关高优先级告警", Command: "chaitin-cli tanswer alarm high-priority --asset-ip 192.0.2.10"},
				},
				Backend: map[string]interface{}{
					"rpc_methods":     []string{"AlarmService.SearchAlarmList"},
					"preset_filters":  map[string]string{"severity": "critical,high", "result": "success,control"},
					"equivalent_call": "chaitin-cli tanswer alarm list --severity critical,high --result success,control",
				},
			},
			{
				Name:        "tanswer alarm detail",
				FullCommand: "chaitin-cli tanswer alarm detail --id '<doc_id>'",
				Layer:       "semantic_shortcut",
				Summary:     "查看威胁告警详情，用于从列表 doc_id 下钻到单条告警研判信息。",
				UseWhen: []string{
					"已经从告警列表拿到 doc_id，需要查看单条告警详情。",
					"需要 payload、cve_list、alert_msg、alarm_description 等详情信息。",
				},
				DoNotUseWhen: []string{
					"没有 doc_id 且需要先查找告警时，使用 tanswer alarm list。",
				},
				Flags: []ManifestFlag{
					{Name: "--id", Type: "string", Required: true, Description: "alarm doc_id returned by alarm list"},
				},
				OutputType:           "alarm_detail",
				OutputFields:         []string{"doc_id", "detail"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查看单条告警详情", Command: "chaitin-cli tanswer alarm detail --id '<doc_id>'"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.GetAlarm"},
				},
			},
			{
				Name:        "tanswer alarm by-attacker",
				FullCommand: "chaitin-cli tanswer alarm by-attacker --attacker '<ip>'",
				Layer:       "semantic_shortcut",
				Summary:     "查询指定攻击源相关告警，用于查看攻击范围、威胁类型和最近活动。",
				UseWhen: []string{
					"已知攻击源 IP，需要查看其攻击了哪些受害对象。",
					"需要围绕一个攻击源统计相关告警数、最高等级、成功/失陷数量和受害对象 Top。",
				},
				DoNotUseWhen: []string{
					"需要通用告警列表时，使用 tanswer alarm list。",
					"需要攻击源整体排行时，使用 tanswer alarm attacker-rank。",
				},
				Flags:      alarmSubjectManifestFlags("--attacker"),
				OutputType: "alarm_subject_summary",
				OutputFields: []string{
					"attacker",
					"related_total",
					"highest_severity",
					"success_control_count",
					"victim_top",
					"threat_type_top",
					"alarms",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询攻击源相关告警", Command: "chaitin-cli tanswer alarm by-attacker --attacker 198.51.100.10 --time today"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.SearchAlarmList"},
				},
			},
			{
				Name:        "tanswer alarm by-victim",
				FullCommand: "chaitin-cli tanswer alarm by-victim --victim '<ip>'",
				Layer:       "semantic_shortcut",
				Summary:     "查询指定受害对象相关告警，用于判断资产或对象被攻击情况。",
				UseWhen: []string{
					"已知受害对象或资产 IP，需要查看被哪些攻击源攻击。",
					"需要围绕一个受害对象统计相关告警数、最高等级、成功/失陷数量和攻击源 Top。",
				},
				DoNotUseWhen: []string{
					"需要通用告警列表时，使用 tanswer alarm list。",
					"需要受害对象整体排行时，使用 tanswer alarm victim-rank。",
				},
				Flags:      alarmSubjectManifestFlags("--victim"),
				OutputType: "alarm_subject_summary",
				OutputFields: []string{
					"victim",
					"related_total",
					"highest_severity",
					"success_control_count",
					"attacker_top",
					"threat_type_top",
					"alarms",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询受害对象相关告警", Command: "chaitin-cli tanswer alarm by-victim --victim 203.0.113.20 --time today"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.SearchAlarmList"},
				},
			},
			{
				Name:        "tanswer alarm by-threat",
				FullCommand: "chaitin-cli tanswer alarm by-threat",
				Layer:       "semantic_shortcut",
				Summary:     "查询指定威胁相关告警，用于按威胁名称、威胁类型或攻击阶段查看影响范围。",
				UseWhen: []string{
					"需要查看某个威胁名称、威胁类型或攻击阶段当前影响了哪些对象。",
					"需要围绕威胁维度统计相关告警数、最高等级、攻击源 Top 和受害对象 Top。",
				},
				DoNotUseWhen: []string{
					"没有指定威胁名称、威胁类型或攻击阶段时，使用 tanswer alarm overview 或 list。",
				},
				Flags: alarmThreatManifestFlags(),
				InputConstraints: []ManifestInputConstraint{
					{
						Rule:        "at_least_one",
						Flags:       []string{"--name", "--tag", "--phase"},
						Description: "provide at least one threat selector",
					},
				},
				OutputType: "alarm_subject_summary",
				OutputFields: []string{
					"threat",
					"related_total",
					"highest_severity",
					"success_control_count",
					"attacker_top",
					"victim_top",
					"alarms",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "按威胁名称查询相关告警", Command: "chaitin-cli tanswer alarm by-threat --name SQL注入 --time today"},
					{Description: "按威胁类型和阶段查询相关告警", Command: "chaitin-cli tanswer alarm by-threat --tag Webshell --phase intrustion"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.SearchAlarmList"},
				},
			},
			{
				Name:        "tanswer alarm important-assets",
				FullCommand: "chaitin-cli tanswer alarm important-assets",
				Layer:       "semantic_shortcut",
				Summary:     "查询重点资产相关告警，用于优先确认重点资产是否存在成功或失陷风险。",
				UseWhen: []string{
					"需要查看重点资产是否被攻击。",
					"需要优先处理重点资产相关的成功或失陷告警。",
				},
				DoNotUseWhen: []string{
					"需要资产风险、漏洞风险或风险主机时，当前版本不包含资产风险能力。",
				},
				Flags:      alarmSubjectManifestFlags(""),
				OutputType: "alarm_subject_summary",
				OutputFields: []string{
					"asset_importance",
					"related_total",
					"highest_severity",
					"success_control_count",
					"attacker_top",
					"victim_top",
					"alarms",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询今天重点资产相关告警", Command: "chaitin-cli tanswer alarm important-assets --time today"},
				},
				Backend: map[string]interface{}{
					"rpc_methods":    []string{"AlarmService.SearchAlarmList"},
					"preset_filters": map[string]string{"asset_importance": "1"},
				},
			},
			{
				Name:        "tanswer alarm attacker-rank",
				FullCommand: "chaitin-cli tanswer alarm attacker-rank",
				Layer:       "semantic_shortcut",
				Summary:     "查看攻击源排行，用于快速定位攻击最集中的来源对象。",
				UseWhen: []string{
					"需要按告警数查看攻击源 Top。",
					"需要快速定位攻击最集中的来源对象。",
				},
				DoNotUseWhen: []string{
					"已知单个攻击源并需要相关告警详情时，使用 tanswer alarm by-attacker。",
				},
				Flags:                alarmRankManifestFlags(),
				OutputType:           "alarm_rank",
				OutputFields:         []string{"rank_type", "rank_count", "rank"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples:             []ManifestExample{{Description: "查看攻击源 Top 10", Command: "chaitin-cli tanswer alarm attacker-rank --time today --top 10"}},
				Backend:              map[string]interface{}{"rpc_methods": []string{"AlarmService.SearchAlarmAggTop"}, "agg": "attacker"},
			},
			{
				Name:        "tanswer alarm victim-rank",
				FullCommand: "chaitin-cli tanswer alarm victim-rank",
				Layer:       "semantic_shortcut",
				Summary:     "查看受害对象排行，用于快速定位被攻击最多或风险最高的对象。",
				UseWhen: []string{
					"需要按告警数查看受害对象 Top。",
					"需要快速定位被攻击最多或风险最高的对象。",
				},
				DoNotUseWhen: []string{
					"已知单个受害对象并需要相关告警详情时，使用 tanswer alarm by-victim。",
				},
				Flags:                alarmRankManifestFlags(),
				OutputType:           "alarm_rank",
				OutputFields:         []string{"rank_type", "rank_count", "rank"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples:             []ManifestExample{{Description: "查看受害对象 Top 10", Command: "chaitin-cli tanswer alarm victim-rank --time today --top 10"}},
				Backend:              map[string]interface{}{"rpc_methods": []string{"AlarmService.SearchAlarmAggTop"}, "agg": "victim"},
			},
			{
				Name:        "tanswer alarm phase-distribution",
				FullCommand: "chaitin-cli tanswer alarm phase-distribution",
				Layer:       "semantic_shortcut",
				Summary:     "查看攻击链阶段分布，用于判断告警主要集中在哪些攻击阶段。",
				UseWhen: []string{
					"需要查看攻击阶段分布。",
					"需要判断当前告警集中在入侵、控制、横向移动还是目标达成等阶段。",
				},
				DoNotUseWhen: []string{
					"需要完整告警态势和多个维度分布时，使用 tanswer alarm overview。",
				},
				Flags:                alarmRankManifestFlags(),
				OutputType:           "alarm_rank",
				OutputFields:         []string{"rank_type", "rank_count", "rank"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples:             []ManifestExample{{Description: "查看今天攻击阶段分布", Command: "chaitin-cli tanswer alarm phase-distribution --time today"}},
				Backend:              map[string]interface{}{"rpc_methods": []string{"AlarmService.SearchAlarmAggTop"}, "agg": "phase"},
			},
			{
				Name:        "tanswer alarm related",
				FullCommand: "chaitin-cli tanswer alarm related --id '<doc_id>'",
				Layer:       "semantic_shortcut",
				Summary:     "查看相关告警，用于从一条告警出发查询前后时间窗口内同攻击源或同受害对象的其他告警。",
				UseWhen: []string{
					"已经有一条告警 doc_id，需要查看前后时间窗口内是否有同攻击源或同受害对象的其他告警。",
					"需要从单条告警出发做轻量上下文排查。",
				},
				DoNotUseWhen: []string{
					"需要同时间窗口内全部告警、完整事件链或因果判断时，当前命令不执行自动工作流。",
					"没有 doc_id 时，先使用 tanswer alarm list 或 high-priority 查询。",
				},
				Flags:      alarmRelatedManifestFlags(),
				OutputType: "alarm_related",
				OutputFields: []string{
					"source_alarm",
					"window",
					"relation",
					"scanned_total",
					"related_total",
					"earliest_time",
					"latest_time",
					"alarms",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查看原告警前后 30 分钟同攻击源或同受害对象告警", Command: "chaitin-cli tanswer alarm related --id '<doc_id>'"},
					{Description: "只查看同攻击源相关告警", Command: "chaitin-cli tanswer alarm related --id '<doc_id>' --window 1h --relation attacker"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.GetAlarm", "AlarmService.SearchAlarmList"},
				},
			},
			{
				Name:        "tanswer file-alarm overview",
				FullCommand: "chaitin-cli tanswer file-alarm overview",
				Layer:       "semantic_shortcut",
				Summary:     "查看文件告警概览，用于判断恶意文件、Webshell 或沙箱检测风险。",
				UseWhen: []string{
					"需要回答今天、近 24 小时或近 7 天文件告警整体数量。",
					"需要查看文件检测已有分类或类型分布，不需要逐条告警摘要。",
				},
				DoNotUseWhen: []string{
					"需要逐条查看恶意文件、Webshell 或沙箱检测告警时，使用对应列表命令。",
					"需要触发新的沙箱分析、样本下载或处置动作时，不使用该只读命令。",
				},
				Flags:      fileAlarmOverviewManifestFlags(),
				OutputType: "file_alarm_overview",
				OutputFields: []string{
					"summary.file_alarm_total",
					"type_distribution",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查看今天文件告警概览", Command: "chaitin-cli tanswer file-alarm overview --time today"},
					{Description: "查看近 24 小时高优先级文件告警概览", Command: "chaitin-cli tanswer file-alarm overview --time 24h --severity critical,high"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.SearchAlarmFdetectCount", "AlarmService.SearchAlarmFdetectAggTop"},
				},
			},
			{
				Name:        "tanswer file-alarm malicious",
				FullCommand: "chaitin-cli tanswer file-alarm malicious",
				Layer:       "semantic_shortcut",
				Summary:     "查询恶意文件告警，用于聚焦文件检测引擎检出的恶意文件。",
				UseWhen: []string{
					"需要查询恶意文件告警摘要列表和分页信息。",
					"需要获取文件告警 doc_id 后继续查看详情。",
				},
				DoNotUseWhen: []string{
					"只需要 Webshell 专项告警时，使用 tanswer file-alarm webshell。",
					"只需要动态沙箱结果时，使用 tanswer file-alarm sandbox。",
				},
				Flags:      fileAlarmListManifestFlags("elf"),
				OutputType: "file_alarm_list",
				OutputFields: []string{
					"total",
					"page_total",
					"page",
					"page_size",
					"current_count",
					"has_more",
					"file_alarms",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询今天恶意文件告警", Command: "chaitin-cli tanswer file-alarm malicious --time today --page-size 10"},
					{Description: "按源和目的 IP 查询恶意文件告警", Command: "chaitin-cli tanswer file-alarm malicious --src-ip 198.51.100.10 --dest-ip 192.0.2.10"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.SearchAlarmFdetectList"},
				},
			},
			{
				Name:        "tanswer file-alarm webshell",
				FullCommand: "chaitin-cli tanswer file-alarm webshell",
				Layer:       "semantic_shortcut",
				Summary:     "查询 Webshell 告警，用于聚焦 Webshell 专项检测结果。",
				UseWhen: []string{
					"需要查询 Webshell 告警摘要列表。",
					"需要按 Webshell 类型、文件名、IP、协议或风险等级筛选。",
				},
				DoNotUseWhen: []string{
					"只需要动态沙箱结果时，使用 tanswer file-alarm sandbox。",
					"需要普通威胁告警 Webshell 类型时，使用 tanswer alarm by-threat 或 alarm list。",
				},
				Flags:      fileAlarmListManifestFlags("webshell"),
				OutputType: "file_alarm_list",
				OutputFields: []string{
					"total",
					"page_total",
					"page",
					"page_size",
					"current_count",
					"has_more",
					"file_alarms",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询今天 Webshell 告警", Command: "chaitin-cli tanswer file-alarm webshell --time today --page-size 10"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.SearchAlarmFdetectList"},
				},
			},
			{
				Name:        "tanswer file-alarm sandbox",
				FullCommand: "chaitin-cli tanswer file-alarm sandbox",
				Layer:       "semantic_shortcut",
				Summary:     "查询沙箱检测告警，用于查看已有动态沙箱分析结果。",
				UseWhen: []string{
					"需要查看动态沙箱评分、运行环境和对应文件告警摘要。",
					"需要确认已有沙箱检测结果，不触发新的分析流程。",
				},
				DoNotUseWhen: []string{
					"需要提交样本或重新分析样本时，不使用该只读命令。",
					"需要恶意文件或 Webshell 分类列表时，使用对应列表命令。",
				},
				Flags:      fileAlarmListManifestFlags(""),
				OutputType: "file_alarm_sandbox_list",
				OutputFields: []string{
					"total",
					"page_total",
					"page",
					"page_size",
					"current_count",
					"has_more",
					"file_alarms",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询今天沙箱检测告警", Command: "chaitin-cli tanswer file-alarm sandbox --time today --page-size 10"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.SearchSandboxAlarmFdetectList"},
				},
			},
			{
				Name:        "tanswer file-alarm detail",
				FullCommand: "chaitin-cli tanswer file-alarm detail --id '<doc_id>'",
				Layer:       "semantic_shortcut",
				Summary:     "查看文件告警详情，用于研判单条文件告警风险。",
				UseWhen: []string{
					"已经从文件告警列表拿到 doc_id，需要查看基础信息、检测结果、检测依据或沙箱报告。",
					"需要按已有详情判断恶意文件、Webshell 或沙箱检测风险。",
				},
				DoNotUseWhen: []string{
					"没有 doc_id 且需要先查找文件告警时，使用列表命令。",
					"需要下载原始样本或触发新分析流程时，不使用该只读命令。",
				},
				Flags: []ManifestFlag{
					{Name: "--id", Type: "string", Required: true, Description: "file alarm doc_id returned by file-alarm list commands"},
				},
				OutputType:           "file_alarm_detail",
				OutputFields:         []string{"doc_id", "detail"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查看文件告警详情", Command: "chaitin-cli tanswer file-alarm detail --id '<doc_id>'"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.GetAlarmFdetect"},
				},
			},
			{
				Name:        "tanswer system status",
				FullCommand: "chaitin-cli tanswer system status",
				Layer:       "semantic_shortcut",
				Summary:     "查看系统基础状态，用于确认全悉环境版本、授权、节点和自检结果。",
				UseWhen: []string{
					"需要快速确认全悉环境是否健康。",
					"需要查看产品版本、部署版本、License 状态、节点列表、节点运行状态或系统自检摘要。",
				},
				DoNotUseWhen: []string{
					"只需要确认 CLI 本地配置时，使用 tanswer auth status。",
					"只需要校验 OpenAPI Token 时，使用 tanswer auth check。",
					"需要修改系统配置、上传 License 或执行破坏性系统操作时，不使用该只读命令。",
				},
				OutputType: "system_status",
				OutputFields: []string{
					"version.product_name",
					"version.product_version",
					"version.version",
					"license.valid_expired",
					"license.valid_expired_soon",
					"license.product_version",
					"health.status",
					"health.online_count",
					"health.total_count",
					"health.nodes",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查看系统基础状态", Command: "chaitin-cli tanswer system status"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"OpsService.GetBaseInfo", "HeraLicenseService.GetLicense", "OpsService.GetSystemStatusResult"},
				},
			},
			{
				Name:        "tanswer asset list",
				FullCommand: "chaitin-cli tanswer asset list",
				Layer:       "semantic_shortcut",
				Summary:     "查询资产列表，用于查看当前资产配置和筛选结果。",
				UseWhen: []string{
					"需要按资产名称、IP、MAC、资产类型、资产等级、资产标签或资产组筛选资产。",
					"需要获取资产 ID 后继续查看资产详情。",
				},
				DoNotUseWhen: []string{
					"需要资产风险、漏洞或端口风险时，当前版本不包含资产风险能力。",
					"需要创建、编辑、删除、导入或导出资产时，应使用后续带确认保护的资产维护命令。",
				},
				Flags:      assetListManifestFlags(),
				OutputType: "asset_list",
				OutputFields: []string{
					"total",
					"page",
					"page_size",
					"current_count",
					"has_more",
					"assets",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询资产列表", Command: "chaitin-cli tanswer asset list --page-size 10"},
					{Description: "按 IP 查询资产", Command: "chaitin-cli tanswer asset list --ip 192.0.2.10"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AssetService.GetAssetList"},
				},
			},
			{
				Name:        "tanswer asset detail",
				FullCommand: "chaitin-cli tanswer asset detail --id '<asset_id>'",
				Layer:       "semantic_shortcut",
				Summary:     "查看资产详情，用于确认资产归属和基础信息。",
				UseWhen: []string{
					"已经从资产列表拿到资产 ID，需要查看资产档案详情。",
					"需要资产名称、IP、MAC、资产类型、资产等级、资产组、标签、负责人、位置、备注、来源和更新时间。",
				},
				DoNotUseWhen: []string{
					"需要资产风险、漏洞或端口风险时，当前版本不包含资产风险能力。",
					"没有资产 ID 且需要先查找资产时，使用 tanswer asset list。",
				},
				Flags: []ManifestFlag{
					{Name: "--id", Type: "integer", Required: true, Description: "asset id returned by asset list"},
				},
				OutputType:           "asset_detail",
				OutputFields:         []string{"id", "detail"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查看资产详情", Command: "chaitin-cli tanswer asset detail --id '<asset_id>'"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AssetService.GetAssetInfo"},
				},
			},
			{
				Name:        "tanswer asset group-tree",
				FullCommand: "chaitin-cli tanswer asset group-tree",
				Layer:       "semantic_shortcut",
				Summary:     "查询资产组树，用于选择和引用资产组。",
				UseWhen: []string{
					"需要查看当前资产组层级、资产组 ID、层级路径或组内资产数量。",
					"需要在资产查询、资产导入或批量维护前确认目标资产组。",
				},
				DoNotUseWhen: []string{
					"需要创建、重命名、删除或移动资产组时，应使用后续带确认保护的资产组维护命令。",
					"需要资产风险、漏洞或风险主机信息时，当前版本不包含资产风险能力。",
				},
				Flags: []ManifestFlag{
					{Name: "--id", Type: "integer", Required: false, Default: "1", Description: "asset group id to use as tree root"},
					{Name: "--depth", Type: "integer", Required: false, Default: "3", Description: "tree depth, 1-100"},
					{Name: "--with-asset", Type: "bool", Required: false, Default: "false", Description: "include asset nodes under groups"},
				},
				OutputType: "asset_group_tree",
				OutputFields: []string{
					"root_id",
					"depth",
					"with_asset",
					"target",
					"groups",
					"groups[].id",
					"groups[].name",
					"groups[].type",
					"groups[].type_label",
					"groups[].count",
					"groups[].path",
					"groups[].children",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询根资产组树", Command: "chaitin-cli tanswer asset group-tree"},
					{Description: "查询指定资产组两层子树", Command: "chaitin-cli tanswer asset group-tree --id 3 --depth 2"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AssetService.SearchGroups"},
				},
			},
			{
				Name:        "tanswer asset download-template",
				FullCommand: "chaitin-cli tanswer asset download-template",
				Layer:       "semantic_shortcut",
				Summary:     "下载资产导入模板，用于按产品模板准备资产导入文件。",
				UseWhen: []string{
					"需要获取资产导入 Excel 模板。",
					"需要让 AI 或运维人员按产品模板字段填写资产后再导入。",
				},
				DoNotUseWhen: []string{
					"需要实际导入资产时，应使用后续带预览和确认保护的导入命令。",
					"需要导出当前资产数据时，使用 tanswer asset export。",
				},
				Flags: []ManifestFlag{
					{Name: "--output", Type: "path", Required: false, Description: "local output file path"},
					{Name: "--with-example", Type: "bool", Required: false, Default: "false", Description: "download template with example rows"},
				},
				OutputType: "asset_template_file",
				OutputFields: []string{
					"file_name",
					"file_path",
					"size_bytes",
					"status_code",
					"method",
					"download_query",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "下载资产导入模板", Command: "chaitin-cli tanswer asset download-template --output ./asset-template.xlsx"},
				},
				Backend: map[string]interface{}{
					"download_method": assetDownloadMethod,
				},
			},
			{
				Name:        "tanswer asset export",
				FullCommand: "chaitin-cli tanswer asset export",
				Layer:       "semantic_shortcut",
				Summary:     "导出资产配置文件，用于备份或离线编辑。",
				UseWhen: []string{
					"需要导出当前资产配置文件。",
					"需要按指定资产 ID 导出选中资产。",
				},
				DoNotUseWhen: []string{
					"需要资产风险、漏洞或风险主机信息时，当前版本不包含资产风险能力。",
					"需要创建、编辑、删除或导入资产时，应使用后续带预览和确认保护的资产维护命令。",
				},
				Flags: []ManifestFlag{
					{Name: "--output", Type: "path", Required: false, Description: "local output file path"},
					{Name: "--id-list", Type: "csv_integer", Required: false, Description: "asset IDs to export; empty exports all assets"},
				},
				OutputType: "asset_export_file",
				OutputFields: []string{
					"file_name",
					"file_path",
					"size_bytes",
					"status_code",
					"method",
					"download_query",
					"export_scope",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "导出全部资产", Command: "chaitin-cli tanswer asset export --output ./asset-export.xlsx"},
					{Description: "导出选中资产", Command: "chaitin-cli tanswer asset export --id-list 3,7 --output ./selected-assets.xlsx"},
				},
				Backend: map[string]interface{}{
					"download_method": assetDownloadMethod,
				},
			},
			{
				Name:        "tanswer asset create",
				FullCommand: "chaitin-cli tanswer asset create",
				Layer:       "semantic_shortcut",
				Summary:     "新增资产，将新资产纳入识别和运营范围。",
				UseWhen: []string{
					"需要新增一个资产配置，且用户已明确提供资产名称和 IP。",
					"需要先生成新增资产预览并等待用户或上层系统确认。",
				},
				DoNotUseWhen: []string{
					"只需要查询资产时，使用 tanswer asset list 或 tanswer asset detail。",
					"需要批量导入资产时，应使用后续带预览和确认保护的导入命令。",
					"需要资产风险、漏洞或风险主机信息时，当前版本不包含资产风险能力。",
				},
				Flags: []ManifestFlag{
					{Name: "--name", Type: "string", Required: true, Description: "asset name"},
					{Name: "--ip", Type: "string", Required: true, Description: "asset IP list, comma separated"},
					{Name: "--contact", Type: "string", Required: false, Description: "asset owner/contact"},
					{Name: "--importance", Type: "enum", Required: false, Default: "normal", Enum: []string{"important", "normal", "重点", "普通", "1", "2"}, Description: "asset importance"},
					{Name: "--remark", Type: "string", Required: false, Description: "asset remark"},
					{Name: "--asset-type", Type: "string", Required: false, Description: "asset type"},
					{Name: "--location", Type: "string", Required: false, Description: "asset location"},
					{Name: "--tag-id", Type: "csv_integer", Required: false, Description: "asset tag IDs"},
					{Name: "--group-id", Type: "integer", Required: false, Default: "1", Description: "asset group ID"},
					{Name: "--ip-mac", Type: "json_array", Required: false, Description: "asset IP/MAC bindings"},
					{Name: "--preview", Type: "bool", Required: false, Default: "false", Description: "return write preview without executing"},
					{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_ASSET_CREATE to execute"},
				},
				OutputType: "asset_create_result",
				OutputFields: []string{
					"preview.requires_confirmation",
					"preview.confirmation_token",
					"confirmed",
					"result",
					"object",
					"audit",
				},
				RiskLevel:            "write_high",
				RequiresConfirmation: true,
				Examples: []ManifestExample{
					{Description: "预览新增资产", Command: "chaitin-cli tanswer asset create --name core-db --ip 192.0.2.10 --preview"},
					{Description: "确认新增资产", Command: "chaitin-cli tanswer asset create --name core-db --ip 192.0.2.10 --confirm CONFIRM_ASSET_CREATE"},
				},
				Backend: map[string]interface{}{
					"rpc_methods":         []string{"AssetService.CreateAsset"},
					"confirmation_token":  assetCreateConfirmToken,
					"preview_is_required": true,
				},
			},
			{
				Name:        "tanswer asset update",
				FullCommand: "chaitin-cli tanswer asset update --id '<asset_id>'",
				Layer:       "semantic_shortcut",
				Summary:     "编辑资产，更新单个资产基础信息。",
				UseWhen: []string{
					"需要更新一个已有资产的基础信息。",
					"需要先读取当前资产详情并生成 before/after 预览。",
				},
				DoNotUseWhen: []string{
					"只需要查询资产时，使用 tanswer asset list 或 tanswer asset detail。",
					"需要批量维护资产字段时，应使用后续带预览和确认保护的批量维护命令。",
					"需要资产风险、漏洞或风险主机信息时，当前版本不包含资产风险能力。",
				},
				Flags: []ManifestFlag{
					{Name: "--id", Type: "integer", Required: true, Description: "asset id"},
					{Name: "--name", Type: "string", Required: true, Description: "asset name"},
					{Name: "--ip", Type: "string", Required: true, Description: "asset IP list, comma separated"},
					{Name: "--contact", Type: "string", Required: false, Description: "asset owner/contact"},
					{Name: "--importance", Type: "enum", Required: false, Default: "normal", Enum: []string{"important", "normal", "重点", "普通", "1", "2"}, Description: "asset importance"},
					{Name: "--remark", Type: "string", Required: false, Description: "asset remark"},
					{Name: "--asset-type", Type: "string", Required: false, Description: "asset type"},
					{Name: "--location", Type: "string", Required: false, Description: "asset location"},
					{Name: "--tag-id", Type: "csv_integer", Required: false, Description: "asset tag IDs"},
					{Name: "--group-id", Type: "integer", Required: false, Default: "1", Description: "asset group ID"},
					{Name: "--ip-mac", Type: "json_array", Required: false, Description: "asset IP/MAC bindings"},
					{Name: "--preview", Type: "bool", Required: false, Default: "false", Description: "return write preview without executing"},
					{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_ASSET_UPDATE to execute"},
				},
				OutputType: "asset_update_result",
				OutputFields: []string{
					"preview.requires_confirmation",
					"preview.confirmation_token",
					"preview.change_summary.before",
					"preview.change_summary.after",
					"confirmed",
					"result",
					"object",
					"audit",
				},
				RiskLevel:            "write_high",
				RequiresConfirmation: true,
				Examples: []ManifestExample{
					{Description: "预览编辑资产", Command: "chaitin-cli tanswer asset update --id 9 --name core-db-new --ip 192.0.2.11 --preview"},
					{Description: "确认编辑资产", Command: "chaitin-cli tanswer asset update --id 9 --name core-db-new --ip 192.0.2.11 --confirm CONFIRM_ASSET_UPDATE"},
				},
				Backend: map[string]interface{}{
					"read_before_rpc_methods": []string{"AssetService.GetAssetInfo"},
					"rpc_methods":             []string{"AssetService.UpdateAsset"},
					"confirmation_token":      assetUpdateConfirmToken,
					"preview_is_required":     true,
				},
			},
			{
				Name:        "tanswer asset delete",
				FullCommand: "chaitin-cli tanswer asset delete --id-list '<asset_ids>'",
				Layer:       "semantic_shortcut",
				Summary:     "删除资产，移除一个或多个已有资产配置。",
				UseWhen: []string{
					"需要删除一个或多个已有资产配置，且用户已明确提供资产 ID。",
					"需要先读取待删除资产详情并生成删除预览。",
				},
				DoNotUseWhen: []string{
					"只需要查询资产时，使用 tanswer asset list 或 tanswer asset detail。",
					"需要资产风险、漏洞或风险主机信息时，当前版本不包含资产风险能力。",
					"不确定资产 ID 或删除影响时，先使用 tanswer asset detail 或 tanswer asset group-tree 查询确认。",
				},
				Flags: []ManifestFlag{
					{Name: "--id-list", Type: "csv_integer", Required: true, Description: "asset IDs to delete"},
					{Name: "--preview", Type: "bool", Required: false, Default: "false", Description: "return write preview without executing"},
					{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_ASSET_DELETE to execute"},
				},
				OutputType: "asset_delete_result",
				OutputFields: []string{
					"preview.requires_confirmation",
					"preview.confirmation_token",
					"preview.change_summary.before",
					"preview.change_summary.after",
					"confirmed",
					"result",
					"object",
					"audit",
				},
				RiskLevel:            "write_high",
				RequiresConfirmation: true,
				Examples: []ManifestExample{
					{Description: "预览删除资产", Command: "chaitin-cli tanswer asset delete --id-list 9 --preview"},
					{Description: "确认删除资产", Command: "chaitin-cli tanswer asset delete --id-list 9,10 --confirm CONFIRM_ASSET_DELETE"},
				},
				Backend: map[string]interface{}{
					"read_before_rpc_methods": []string{"AssetService.GetAssetInfo"},
					"rpc_methods":             []string{"AssetService.DeleteAsset"},
					"confirmation_token":      assetDeleteConfirmToken,
					"preview_is_required":     true,
				},
			},
			{
				Name:        "tanswer asset batch-maintain",
				FullCommand: "chaitin-cli tanswer asset batch-maintain --id-list '<asset_ids>'",
				Layer:       "semantic_shortcut",
				Summary:     "批量维护资产，更新多个资产的资产组、负责人、地理位置或备注。",
				UseWhen: []string{
					"需要对多个已有资产批量维护负责人、地理位置、备注或资产组。",
					"需要先读取待维护资产详情并生成批量变更预览。",
				},
				DoNotUseWhen: []string{
					"只需要查询资产时，使用 tanswer asset list 或 tanswer asset detail。",
					"需要批量修改资产风险、漏洞或风险主机信息时，当前版本不包含资产风险能力。",
					"需要批量新增资产标签时，使用后续带确认保护的资产标签维护命令。",
				},
				Flags: []ManifestFlag{
					{Name: "--id-list", Type: "csv_integer", Required: true, Description: "asset IDs to maintain"},
					{Name: "--contact", Type: "string", Required: false, Description: "asset owner/contact to set"},
					{Name: "--remark", Type: "string", Required: false, Description: "asset remark to set"},
					{Name: "--location", Type: "string", Required: false, Description: "asset location to set"},
					{Name: "--group-id", Type: "integer", Required: false, Description: "asset group ID to move assets into"},
					{Name: "--preview", Type: "bool", Required: false, Default: "false", Description: "return write preview without executing"},
					{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_ASSET_BATCH_MAINTAIN to execute"},
				},
				OutputType: "asset_batch_maintain_result",
				OutputFields: []string{
					"preview.requires_confirmation",
					"preview.confirmation_token",
					"preview.change_summary.before",
					"preview.change_summary.after",
					"confirmed",
					"result",
					"object",
					"audit",
				},
				RiskLevel:            "write_high",
				RequiresConfirmation: true,
				Examples: []ManifestExample{
					{Description: "预览批量维护资产负责人", Command: "chaitin-cli tanswer asset batch-maintain --id-list 9,10 --contact secops --preview"},
					{Description: "确认批量移动资产到资产组", Command: "chaitin-cli tanswer asset batch-maintain --id-list 9,10 --group-id 2 --confirm CONFIRM_ASSET_BATCH_MAINTAIN"},
				},
				Backend: map[string]interface{}{
					"read_before_rpc_methods": []string{"AssetService.GetAssetInfo"},
					"rpc_methods":             []string{"AssetService.UpdateAssetBatch"},
					"confirmation_token":      assetBatchMaintainConfirmToken,
					"preview_is_required":     true,
				},
			},
			{
				Name:        "tanswer asset batch-tag",
				FullCommand: "chaitin-cli tanswer asset batch-tag --id-list '<asset_ids>' --tag-id '<tag_ids>'",
				Layer:       "semantic_shortcut",
				Summary:     "批量维护资产标签，给一个或多个资产设置标签。",
				UseWhen: []string{
					"需要对多个已有资产批量设置资产标签。",
					"需要先读取待维护资产详情并生成标签变更预览。",
				},
				DoNotUseWhen: []string{
					"需要创建新标签时，不使用该命令；该命令只使用已有 tag ID。",
					"只需要查询资产时，使用 tanswer asset list 或 tanswer asset detail。",
					"需要资产风险、漏洞或风险主机信息时，当前版本不包含资产风险能力。",
				},
				Flags: []ManifestFlag{
					{Name: "--id-list", Type: "csv_integer", Required: true, Description: "asset IDs to maintain"},
					{Name: "--tag-id", Type: "csv_integer", Required: true, Description: "asset tag IDs to set"},
					{Name: "--preview", Type: "bool", Required: false, Default: "false", Description: "return write preview without executing"},
					{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_ASSET_BATCH_TAG to execute"},
				},
				OutputType: "asset_batch_tag_result",
				OutputFields: []string{
					"preview.requires_confirmation",
					"preview.confirmation_token",
					"preview.change_summary.before",
					"preview.change_summary.after",
					"confirmed",
					"result",
					"object",
					"audit",
				},
				RiskLevel:            "write_high",
				RequiresConfirmation: true,
				Examples: []ManifestExample{
					{Description: "预览批量维护资产标签", Command: "chaitin-cli tanswer asset batch-tag --id-list 9,10 --tag-id 3,7 --preview"},
					{Description: "确认批量维护资产标签", Command: "chaitin-cli tanswer asset batch-tag --id-list 9,10 --tag-id 3,7 --confirm CONFIRM_ASSET_BATCH_TAG"},
				},
				Backend: map[string]interface{}{
					"read_before_rpc_methods": []string{"AssetService.GetAssetInfo"},
					"rpc_methods":             []string{"AssetService.UpdateAssetTagBatch"},
					"confirmation_token":      assetBatchTagConfirmToken,
					"preview_is_required":     true,
				},
			},
			{
				Name:        "tanswer asset group-create",
				FullCommand: "chaitin-cli tanswer asset group-create --name '<name>'",
				Layer:       "semantic_shortcut",
				Summary:     "创建资产组，在指定父资产组下新增一个资产组。",
				UseWhen: []string{
					"需要创建一个新的资产组。",
					"需要先确认父资产组并生成创建预览。",
				},
				DoNotUseWhen: []string{
					"只需要查看资产组层级时，使用 tanswer asset group-tree。",
					"需要自动分组、智能推荐分组或分组规则时，当前版本不提供。",
				},
				Flags: []ManifestFlag{
					{Name: "--name", Type: "string", Required: true, Description: "asset group name"},
					{Name: "--parent-id", Type: "integer", Required: false, Default: "1", Description: "parent asset group ID"},
					{Name: "--preview", Type: "bool", Required: false, Default: "false", Description: "return write preview without executing"},
					{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_ASSET_GROUP_CREATE to execute"},
				},
				OutputType:           "asset_group_create_result",
				OutputFields:         []string{"preview.requires_confirmation", "preview.confirmation_token", "preview.change_summary.before", "preview.change_summary.after", "confirmed", "result", "object", "audit"},
				RiskLevel:            "write_high",
				RequiresConfirmation: true,
				Examples: []ManifestExample{
					{Description: "预览创建资产组", Command: "chaitin-cli tanswer asset group-create --name 核心区 --parent-id 2 --preview"},
					{Description: "确认创建资产组", Command: "chaitin-cli tanswer asset group-create --name 核心区 --parent-id 2 --confirm CONFIRM_ASSET_GROUP_CREATE"},
				},
				Backend: map[string]interface{}{
					"read_before_rpc_methods": []string{"AssetService.SearchGroups"},
					"rpc_methods":             []string{"AssetService.CreateGroup"},
					"confirmation_token":      assetGroupCreateConfirmToken,
					"preview_is_required":     true,
				},
			},
			{
				Name:        "tanswer asset group-rename",
				FullCommand: "chaitin-cli tanswer asset group-rename --id '<group_id>' --name '<name>'",
				Layer:       "semantic_shortcut",
				Summary:     "重命名资产组，修改一个已有资产组名称。",
				UseWhen: []string{
					"需要修改一个已有资产组名称。",
					"需要先读取当前资产组并生成重命名前后预览。",
				},
				DoNotUseWhen: []string{
					"只需要查看资产组层级时，使用 tanswer asset group-tree。",
					"需要移动资产组或资产所在层级时，使用后续资产树移动命令。",
				},
				Flags: []ManifestFlag{
					{Name: "--id", Type: "integer", Required: true, Description: "asset group ID"},
					{Name: "--name", Type: "string", Required: true, Description: "new asset group name"},
					{Name: "--preview", Type: "bool", Required: false, Default: "false", Description: "return write preview without executing"},
					{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_ASSET_GROUP_RENAME to execute"},
				},
				OutputType:           "asset_group_rename_result",
				OutputFields:         []string{"preview.requires_confirmation", "preview.confirmation_token", "preview.change_summary.before", "preview.change_summary.after", "confirmed", "result", "object", "audit"},
				RiskLevel:            "write_high",
				RequiresConfirmation: true,
				Examples: []ManifestExample{
					{Description: "预览重命名资产组", Command: "chaitin-cli tanswer asset group-rename --id 3 --name 核心区 --preview"},
					{Description: "确认重命名资产组", Command: "chaitin-cli tanswer asset group-rename --id 3 --name 核心区 --confirm CONFIRM_ASSET_GROUP_RENAME"},
				},
				Backend: map[string]interface{}{
					"read_before_rpc_methods": []string{"AssetService.SearchGroups"},
					"rpc_methods":             []string{"AssetService.UpdateGroup"},
					"confirmation_token":      assetGroupRenameConfirmToken,
					"preview_is_required":     true,
				},
			},
			{
				Name:        "tanswer asset group-delete",
				FullCommand: "chaitin-cli tanswer asset group-delete --id-list '<group_ids>'",
				Layer:       "semantic_shortcut",
				Summary:     "删除资产组，删除一个或多个非根资产组。",
				UseWhen: []string{
					"需要删除一个或多个非根资产组。",
					"需要先读取待删除资产组并生成删除预览。",
				},
				DoNotUseWhen: []string{
					"需要删除根资产组时，根资产组不可删除。",
					"只需要查看资产组层级时，使用 tanswer asset group-tree。",
				},
				Flags: []ManifestFlag{
					{Name: "--id-list", Type: "csv_integer", Required: true, Description: "asset group IDs to delete"},
					{Name: "--preview", Type: "bool", Required: false, Default: "false", Description: "return write preview without executing"},
					{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_ASSET_GROUP_DELETE to execute"},
				},
				OutputType:           "asset_group_delete_result",
				OutputFields:         []string{"preview.requires_confirmation", "preview.confirmation_token", "preview.change_summary.before", "preview.change_summary.after", "confirmed", "result", "object", "audit"},
				RiskLevel:            "write_high",
				RequiresConfirmation: true,
				Examples: []ManifestExample{
					{Description: "预览删除资产组", Command: "chaitin-cli tanswer asset group-delete --id-list 3 --preview"},
					{Description: "确认删除资产组", Command: "chaitin-cli tanswer asset group-delete --id-list 3,4 --confirm CONFIRM_ASSET_GROUP_DELETE"},
				},
				Backend: map[string]interface{}{
					"read_before_rpc_methods": []string{"AssetService.SearchGroups"},
					"rpc_methods":             []string{"AssetService.DeleteGroup"},
					"confirmation_token":      assetGroupDeleteConfirmToken,
					"preview_is_required":     true,
				},
			},
			{
				Name:        "tanswer asset tree-move",
				FullCommand: "chaitin-cli tanswer asset tree-move --id '<id>' --type '<group|asset>' --prev-id '<id>' --prev-type '<group|asset>'",
				Layer:       "semantic_shortcut",
				Summary:     "调整资产树层级，移动资产或资产组在资产树中的位置。",
				UseWhen: []string{
					"需要移动资产或资产组在资产树中的层级位置。",
					"已知源节点、前置/目标节点和 top_layer 拖拽语义，需要执行产品同等移动操作。",
				},
				DoNotUseWhen: []string{
					"只需要查看资产组层级时，使用 tanswer asset group-tree。",
					"只需要批量把资产移动到某个资产组时，优先使用 tanswer asset batch-maintain --group-id。",
					"需要自动分组、智能推荐分组或分组规则时，当前版本不提供。",
				},
				Flags: []ManifestFlag{
					{Name: "--id", Type: "integer", Required: true, Description: "source node id"},
					{Name: "--type", Type: "enum", Required: true, Enum: []string{"group", "asset", "1", "2"}, Description: "source node type"},
					{Name: "--prev-id", Type: "integer", Required: true, Description: "previous/target node id"},
					{Name: "--prev-type", Type: "enum", Required: true, Enum: []string{"group", "asset", "1", "2"}, Description: "previous/target node type"},
					{Name: "--top-layer", Type: "bool", Required: false, Default: "false", Description: "backend top_layer flag from product tree drag"},
					{Name: "--preview", Type: "bool", Required: false, Default: "false", Description: "return write preview without executing"},
					{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_ASSET_TREE_MOVE to execute"},
				},
				OutputType:           "asset_tree_move_result",
				OutputFields:         []string{"preview.requires_confirmation", "preview.confirmation_token", "preview.change_summary.before", "preview.change_summary.after", "confirmed", "result", "object", "audit"},
				RiskLevel:            "write_high",
				RequiresConfirmation: true,
				Examples: []ManifestExample{
					{Description: "预览移动资产树节点", Command: "chaitin-cli tanswer asset tree-move --id 9 --type asset --prev-id 3 --prev-type group --top-layer --preview"},
					{Description: "确认移动资产树节点", Command: "chaitin-cli tanswer asset tree-move --id 9 --type asset --prev-id 3 --prev-type group --top-layer --confirm CONFIRM_ASSET_TREE_MOVE"},
				},
				Backend: map[string]interface{}{
					"read_before_rpc_methods": []string{"AssetService.GetAssetInfo", "AssetService.SearchGroups"},
					"rpc_methods":             []string{"AssetService.UpdateAssetTree"},
					"confirmation_token":      assetTreeMoveConfirmToken,
					"preview_is_required":     true,
				},
			},
			{
				Name:        "tanswer asset import",
				FullCommand: "chaitin-cli tanswer asset import --file '<xlsx>'",
				Layer:       "semantic_shortcut",
				Summary:     "导入资产，上传资产导入模板文件并批量初始化或更新资产。",
				UseWhen: []string{
					"需要按产品资产导入模板批量创建或更新资产。",
					"需要先确认本地导入文件元信息，再显式确认上传。",
				},
				DoNotUseWhen: []string{
					"需要获取导入模板时，使用 tanswer asset download-template。",
					"需要导出资产时，使用 tanswer asset export。",
					"需要资产风险、漏洞或风险主机信息时，当前版本不包含资产风险能力。",
				},
				Flags: []ManifestFlag{
					{Name: "--file", Type: "path", Required: true, Description: "asset import template file path"},
					{Name: "--preview", Type: "bool", Required: false, Default: "false", Description: "return write preview without uploading"},
					{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_ASSET_IMPORT to execute"},
				},
				OutputType: "asset_import_result",
				OutputFields: []string{
					"preview.requires_confirmation",
					"preview.confirmation_token",
					"confirmed",
					"result",
					"object",
					"audit",
					"import_result.record_total",
					"import_result.success_total",
					"import_result.fail_total",
					"import_result.cover_total",
					"import_result.failure_reason",
					"import_result.cover_reason",
				},
				RiskLevel:            "write_high",
				RequiresConfirmation: true,
				Examples: []ManifestExample{
					{Description: "预览资产导入文件", Command: "chaitin-cli tanswer asset import --file ./assets.xlsx --preview"},
					{Description: "确认导入资产", Command: "chaitin-cli tanswer asset import --file ./assets.xlsx --confirm CONFIRM_ASSET_IMPORT"},
				},
				Backend: map[string]interface{}{
					"upload_method":        assetImportMethod,
					"upload_path":          "/api/upload",
					"confirmation_token":   assetImportConfirmToken,
					"preview_is_required":  true,
					"multipart_file_field": "file",
				},
			},
			{
				Name:        "tanswer metadata protocol",
				FullCommand: "chaitin-cli tanswer metadata protocol --protocol '<protocol>'",
				Layer:       "semantic_shortcut",
				Summary:     "按协议检索流量元数据摘要。",
				UseWhen: []string{
					"需要按 HTTP、DNS、TCP、UDP 或其他协议查看已有流量元数据。",
					"需要按时间、源/目的 IP、源/目的端口、HTTP URL 或 DNS rrname 筛选元数据。",
				},
				DoNotUseWhen: []string{
					"需要查看单条完整元数据详情时，使用 tanswer metadata detail。",
					"需要调整元数据采集配置时，使用 tanswer metadata config-update。",
				},
				Flags:      metadataListManifestFlags(false),
				OutputType: "metadata_list",
				OutputFields: []string{
					"total",
					"page",
					"page_size",
					"current_count",
					"has_more",
					"metadata",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询 HTTP 元数据", Command: "chaitin-cli tanswer metadata protocol --protocol http --time today --page-size 10"},
					{Description: "按源 IP 查询 DNS 元数据", Command: "chaitin-cli tanswer metadata protocol --protocol dns --src-ip 198.51.100.10"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"LogSearchService.SearchOrigDataHTTPLog", "LogSearchService.SearchOrigDataDNSLog", "LogSearchService.SearchOrigDataTCPUDPLog", "LogSearchService.SearchOtherOrigDataLog"},
				},
			},
			{
				Name:        "tanswer metadata search",
				FullCommand: "chaitin-cli tanswer metadata search --protocol '<protocol>' --advanced-query '<query>'",
				Layer:       "semantic_shortcut",
				Summary:     "按高级条件检索流量元数据。",
				UseWhen: []string{
					"需要使用高级查询语句检索指定协议的已有流量元数据。",
					"需要在高级查询基础上叠加时间、五元组、HTTP URL 或 DNS rrname 条件。",
				},
				DoNotUseWhen: []string{
					"只需要按协议简单翻页时，使用 tanswer metadata protocol。",
					"需要保存高级查询历史时，不使用该只读 CLI；该命令固定 save_history=false。",
				},
				Flags:      metadataListManifestFlags(true),
				OutputType: "metadata_list",
				OutputFields: []string{
					"total",
					"page",
					"page_size",
					"current_count",
					"has_more",
					"metadata",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "按 DNS 高级查询检索元数据", Command: "chaitin-cli tanswer metadata search --protocol dns --advanced-query \"dns_rrname = 'example.com'\""},
					{Description: "按 HTTP 方法检索元数据", Command: "chaitin-cli tanswer metadata search --protocol http --advanced-query \"http_method = 'GET'\" --time 24h"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"LogSearchService.SearchOrigDataHTTPLog", "LogSearchService.SearchOrigDataDNSLog", "LogSearchService.SearchOrigDataTCPUDPLog", "LogSearchService.SearchOtherOrigDataLog"},
				},
			},
			{
				Name:        "tanswer metadata detail",
				FullCommand: "chaitin-cli tanswer metadata detail --id '<metadata_id>' --timestamp '<ms>' --protocol '<protocol>'",
				Layer:       "semantic_shortcut",
				Summary:     "查看单条流量元数据详情。",
				UseWhen: []string{
					"已经从 metadata 列表拿到 id、timestamp 和 protocol，需要查看单条元数据详情。",
					"需要保留后端返回的原始元数据详情字段用于研判。",
				},
				DoNotUseWhen: []string{
					"没有元数据 id 或 timestamp 且需要先查找时，使用 tanswer metadata protocol 或 search。",
					"需要下载文件或调整配置时，不使用该只读命令。",
				},
				Flags: []ManifestFlag{
					{Name: "--id", Type: "string", Required: true, Description: "metadata id returned by metadata list commands"},
					{Name: "--timestamp", Type: "integer", Required: true, Description: "metadata timestamp in milliseconds"},
					{Name: "--protocol", Type: "enum", Required: true, Enum: []string{"http", "http2", "dns", "tcp", "udp", "other"}, Description: "metadata protocol or event_type"},
				},
				OutputType:           "metadata_detail",
				OutputFields:         []string{"id", "timestamp", "protocol", "detail"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查看元数据详情", Command: "chaitin-cli tanswer metadata detail --id '<metadata_id>' --timestamp 1784282400000 --protocol http"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"LogSearchService.GetOrigDataLogDetail"},
				},
			},
			{
				Name:        "tanswer metadata near-alarm",
				FullCommand: "chaitin-cli tanswer metadata near-alarm --id '<alarm_doc_id>'",
				Layer:       "semantic_shortcut",
				Summary:     "查询告警附近流量元数据上下文。",
				UseWhen: []string{
					"需要围绕一条威胁告警的时间点和五元组检索附近流量元数据。",
					"需要为告警研判补充同时间窗上下文。",
				},
				DoNotUseWhen: []string{
					"需要证明元数据本身就是攻击证据时，不应直接使用该结果下结论；该命令只返回上下文。",
					"需要关联其他威胁告警时，使用 tanswer alarm related。",
				},
				Flags: []ManifestFlag{
					{Name: "--id", Type: "string", Required: true, Description: "alarm doc_id returned by alarm list/detail commands"},
					{Name: "--window", Type: "duration", Required: false, Default: "30m", Description: "time window before and after the alarm timestamp, for example 10m or 1h"},
					{Name: "--protocol", Type: "enum", Required: false, Enum: []string{"http", "http2", "dns", "tcp", "udp", "other"}, Description: "override metadata protocol; default uses alarm app_proto/proto"},
					{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "page size, 1-100"},
				},
				OutputType:           "metadata_near_alarm",
				OutputFields:         []string{"alarm", "window", "total", "page_size", "current_count", "metadata", "note"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询告警附近元数据", Command: "chaitin-cli tanswer metadata near-alarm --id '<alarm_doc_id>' --window 30m --page-size 10"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.GetAlarm", "LogSearchService.SearchOrigDataHTTPLog", "LogSearchService.SearchOrigDataDNSLog", "LogSearchService.SearchOrigDataTCPUDPLog", "LogSearchService.SearchOtherOrigDataLog"},
				},
			},
			{
				Name:        "tanswer metadata config",
				FullCommand: "chaitin-cli tanswer metadata config",
				Layer:       "semantic_shortcut",
				Summary:     "查询元数据数据配置。",
				UseWhen: []string{
					"需要只读查看当前节点或指定节点的元数据协议采集配置。",
					"需要确认 HTTP、DNS 等协议元数据是否已选择采集。",
				},
				DoNotUseWhen: []string{
					"需要修改元数据采集配置时，使用 tanswer metadata config-update。",
				},
				Flags: []ManifestFlag{
					{Name: "--node-id", Type: "string", Required: false, Description: "node id; empty means current/global node"},
				},
				OutputType:           "metadata_config",
				OutputFields:         []string{"node_id", "configurations"},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询当前节点元数据配置", Command: "chaitin-cli tanswer metadata config"},
					{Description: "查询指定节点元数据配置", Command: "chaitin-cli tanswer metadata config --node-id '<node_id>'"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"LogSearchService.GetOrigDataLogProtocolList"},
				},
			},
			{
				Name:        "tanswer metadata config-update",
				FullCommand: "chaitin-cli tanswer metadata config-update",
				Layer:       "semantic_shortcut",
				Summary:     "调整元数据数据配置。",
				UseWhen: []string{
					"需要修改指定节点的元数据协议存储范围。",
					"需要启用或停用 HTTP、DNS、TCP、UDP 等协议元数据存储。",
				},
				DoNotUseWhen: []string{
					"只需要查看当前配置时，使用 tanswer metadata config。",
					"没有明确 node-id 或没有确认 token 时，不执行写操作。",
				},
				Flags: []ManifestFlag{
					{Name: "--node-id", Type: "string", Required: true, Description: "node id to update"},
					{Name: "--enable", Type: "csv_string", Required: false, Description: "protocol event_type list to enable"},
					{Name: "--disable", Type: "csv_string", Required: false, Description: "protocol event_type list to disable"},
					{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
					{Name: "--confirm", Type: "string", Required: false, Description: "must equal " + metadataConfigUpdateConfirmToken + " to execute"},
				},
				OutputType: "metadata_config_update_result",
				OutputFields: []string{
					"preview.requires_confirmation",
					"preview.confirmation_token",
					"execution.confirmed",
					"execution.result",
					"execution.object",
					"execution.audit",
				},
				RiskLevel:            "write_high",
				RequiresConfirmation: true,
				Examples: []ManifestExample{
					{Description: "预览启用 HTTP/DNS 元数据存储", Command: "chaitin-cli tanswer metadata config-update --node-id '<node_id>' --enable http,dns --preview"},
					{Description: "确认停用 TCP/UDP 元数据存储", Command: "chaitin-cli tanswer metadata config-update --node-id '<node_id>' --disable tcp,udp --confirm CONFIRM_METADATA_CONFIG_UPDATE"},
				},
				Backend: map[string]interface{}{
					"rpc_methods":         []string{"LogSearchService.GetOrigDataLogProtocolList", "LogSearchService.ConfigureOrigDataLogProtocol"},
					"confirmation_token":  metadataConfigUpdateConfirmToken,
					"write_protection":    "preview_then_exact_confirm",
					"audit_actor_source":  "open_api_token",
					"requires_token_auth": true,
				},
			},
			{
				Name:        "tanswer policy detection-whitelist",
				FullCommand: "chaitin-cli tanswer policy detection-whitelist",
				Layer:       "semantic_shortcut",
				Summary:     "查询检测白名单，用于查看误报抑制策略。",
				UseWhen: []string{
					"需要查看当前检测白名单规则、状态、失效时间或匹配条件。",
					"需要按源/目的 IP、域名、URL 路径、User-Agent、XFF、响应码、威胁类型或检测规则筛选误报抑制规则。",
				},
				DoNotUseWhen: []string{
					"需要新增、编辑、启停、删除、导入或导出检测白名单时，后续使用带确认保护的写入命令。",
					"需要查询响应白名单时，不使用该检测白名单命令。",
				},
				Flags:      policyDetectionWhitelistManifestFlags(),
				OutputType: "policy_detection_whitelist_list",
				OutputFields: []string{
					"total",
					"page",
					"page_size",
					"current_count",
					"has_more",
					"detection_whitelists",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist --page-size 10"},
					{Description: "按源 IP 和启用状态查询检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist --src-ip 198.51.100.10 --status enabled"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.SearchWhiteList"},
				},
			},
			policyDetectionWhitelistWriteManifestCommand("tanswer policy detection-whitelist-create", "chaitin-cli tanswer policy detection-whitelist-create", "新增检测白名单。", []string{
				"需要新增检测白名单来抑制误报或重复告警。",
				"已经明确白名单名称、匹配条件、处置方式、状态和有效期。",
			}, policyDetectionWhitelistWriteManifestFlags(false, "CONFIRM_POLICY_DETECTION_WHITELIST_CREATE"), "policy_detection_whitelist_create_result", []ManifestExample{
				{Description: "预览新增检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-create --name 登录误报 --src-ip 198.51.100.10 --valid-time 3600 --preview"},
				{Description: "确认新增检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-create --name 登录误报 --src-ip 198.51.100.10 --valid-time 3600 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_CREATE"},
			}, []string{"AlarmService.CreateWhiteList"}, policyDetectionWhitelistCreateConfirmToken),
			policyDetectionWhitelistWriteManifestCommand("tanswer policy detection-whitelist-update", "chaitin-cli tanswer policy detection-whitelist-update", "编辑检测白名单。", []string{
				"需要更新单条检测白名单配置。",
				"需要执行前查看当前白名单与目标配置差异。",
			}, policyDetectionWhitelistWriteManifestFlags(true, "CONFIRM_POLICY_DETECTION_WHITELIST_UPDATE"), "policy_detection_whitelist_update_result", []ManifestExample{
				{Description: "预览编辑检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-update --id 21 --name 新白名单 --src-ip 198.51.100.11 --valid-time 3600 --preview"},
				{Description: "确认编辑检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-update --id 21 --name 新白名单 --src-ip 198.51.100.11 --valid-time 3600 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_UPDATE"},
			}, []string{"AlarmService.SearchWhiteList", "AlarmService.UpdateWhiteList"}, policyDetectionWhitelistUpdateConfirmToken),
			policyDetectionWhitelistWriteManifestCommand("tanswer policy detection-whitelist-enable", "chaitin-cli tanswer policy detection-whitelist-enable", "启用检测白名单。", []string{
				"需要启用一条或多条检测白名单。",
			}, policyDetectionWhitelistActionManifestFlags("CONFIRM_POLICY_DETECTION_WHITELIST_ENABLE"), "policy_detection_whitelist_enable_result", []ManifestExample{
				{Description: "确认启用检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-enable --id-list 21,22 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_ENABLE"},
			}, []string{"AlarmService.SearchWhiteList", "AlarmService.DeleteWhiteList"}, policyDetectionWhitelistEnableConfirmToken),
			policyDetectionWhitelistWriteManifestCommand("tanswer policy detection-whitelist-disable", "chaitin-cli tanswer policy detection-whitelist-disable", "禁用检测白名单。", []string{
				"需要禁用一条或多条检测白名单。",
			}, policyDetectionWhitelistActionManifestFlags("CONFIRM_POLICY_DETECTION_WHITELIST_DISABLE"), "policy_detection_whitelist_disable_result", []ManifestExample{
				{Description: "确认禁用检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-disable --id-list 21,22 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_DISABLE"},
			}, []string{"AlarmService.SearchWhiteList", "AlarmService.DeleteWhiteList"}, policyDetectionWhitelistDisableConfirmToken),
			policyDetectionWhitelistWriteManifestCommand("tanswer policy detection-whitelist-delete", "chaitin-cli tanswer policy detection-whitelist-delete", "删除检测白名单。", []string{
				"需要删除一条或多条检测白名单。",
			}, policyDetectionWhitelistActionManifestFlags("CONFIRM_POLICY_DETECTION_WHITELIST_DELETE"), "policy_detection_whitelist_delete_result", []ManifestExample{
				{Description: "确认删除检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-delete --id-list 21,22 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_DELETE"},
			}, []string{"AlarmService.SearchWhiteList", "AlarmService.DeleteWhiteList"}, policyDetectionWhitelistDeleteConfirmToken),
			policyDetectionWhitelistWriteManifestCommand("tanswer policy detection-whitelist-from-alarm", "chaitin-cli tanswer policy detection-whitelist-from-alarm", "从告警对象生成检测白名单。", []string{
				"已经确认一条威胁告警为误报，需要基于告警字段生成候选检测白名单。",
				"需要执行前查看告警对象、建议白名单匹配范围和风险提示。",
			}, policyDetectionWhitelistFromAlarmManifestFlags(), "policy_detection_whitelist_from_alarm_result", []ManifestExample{
				{Description: "预览从告警生成检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-from-alarm --id '<doc_id>' --valid-time 3600 --preview"},
				{Description: "确认从告警生成检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-from-alarm --id '<doc_id>' --remark 已确认误报 --valid-time 3600 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_FROM_ALARM"},
			}, []string{"AlarmService.GetAlarm", "AlarmService.CreateWhiteList"}, policyDetectionWhitelistFromAlarmConfirmToken),
			policyFileExportManifestCommand("tanswer policy detection-whitelist-export", "chaitin-cli tanswer policy detection-whitelist-export", "导出检测白名单文件。", "detection_whitelist_export_file", []ManifestExample{
				{Description: "导出全部检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-export --output ./detection-whitelist.xlsx"},
				{Description: "导出选中检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-export --id-list 21,22 --output ./selected-detection-whitelist.xlsx"},
			}, []string{"AlarmDownloadService.ExportWhiteList"}),
			policyFileImportManifestCommand("tanswer policy detection-whitelist-import", "chaitin-cli tanswer policy detection-whitelist-import", "导入检测白名单文件。", "policy_detection_whitelist_import_result", policyDetectionWhitelistImportConfirmToken, []ManifestExample{
				{Description: "预览导入检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-import --file ./detection-whitelist.xlsx --preview"},
				{Description: "确认导入检测白名单", Command: "chaitin-cli tanswer policy detection-whitelist-import --file ./detection-whitelist.xlsx --confirm CONFIRM_POLICY_DETECTION_WHITELIST_IMPORT"},
			}, []string{"AlarmUploadService.ImportWhiteList"}),
			{
				Name:        "tanswer policy custom-intelligence",
				FullCommand: "chaitin-cli tanswer policy custom-intelligence",
				Layer:       "semantic_shortcut",
				Summary:     "查询自定义情报，用于查看 IOC 检测配置。",
				UseWhen: []string{
					"需要查看当前自定义 IOC 情报列表、类型、状态或备注。",
					"需要按情报名称、IOC、IOC 类型或启用状态筛选自定义情报。",
				},
				DoNotUseWhen: []string{
					"需要新增、编辑、启停、删除、导入或导出自定义情报时，后续使用带确认保护的写入命令。",
					"需要查询外部情报授权、版本或节点状态时，不使用该命令。",
				},
				Flags:      policyCustomIntelligenceManifestFlags(),
				OutputType: "policy_custom_intelligence_list",
				OutputFields: []string{
					"total",
					"page",
					"page_size",
					"current_count",
					"has_more",
					"custom_intelligence",
				},
				RiskLevel:            "read",
				RequiresConfirmation: false,
				Examples: []ManifestExample{
					{Description: "查询自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence --page-size 10"},
					{Description: "按 IOC 查询启用的域名情报", Command: "chaitin-cli tanswer policy custom-intelligence --ioc evil.example --type domain --status enabled"},
				},
				Backend: map[string]interface{}{
					"rpc_methods": []string{"AlarmService.SearchAlarmCustomIntelligenceList"},
				},
			},
			policyCustomIntelligenceWriteManifestCommand("tanswer policy custom-intelligence-create", "chaitin-cli tanswer policy custom-intelligence-create", "新增自定义情报。", []string{
				"需要新增自定义 IOC 情报，用于后续检测命中。",
				"已经明确 IOC 类型、IOC 值、状态和备注。",
			}, []ManifestFlag{
				{Name: "--name", Type: "string", Required: true, Description: "custom intelligence name"},
				{Name: "--ioc", Type: "csv_string", Required: true, Description: "IOC values"},
				{Name: "--type", Type: "enum", Required: true, Enum: []string{"ip", "domain", "url", "md5", "sha1", "sha256", "1", "2", "3", "4", "5", "6"}, Description: "IOC type"},
				{Name: "--status", Type: "enum", Required: false, Default: "enabled", Enum: []string{"enabled", "disabled", "1", "0"}, Description: "custom intelligence status"},
				{Name: "--remarks", Type: "string", Required: false, Description: "remarks"},
				{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
				{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_POLICY_CUSTOM_INTELLIGENCE_CREATE to execute"},
			}, "policy_custom_intelligence_create_result", []ManifestExample{
				{Description: "预览新增自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence-create --name 恶意域名 --ioc evil.example --type domain --preview"},
				{Description: "确认新增自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence-create --name 恶意域名 --ioc evil.example --type domain --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_CREATE"},
			}, []string{"AlarmService.CreateAlarmCustomIntelligence"}, policyCustomIntelligenceCreateConfirmToken),
			policyCustomIntelligenceWriteManifestCommand("tanswer policy custom-intelligence-update", "chaitin-cli tanswer policy custom-intelligence-update", "编辑自定义情报。", []string{
				"需要更新单条自定义 IOC 情报配置。",
				"需要执行前查看当前情报与目标配置差异。",
			}, []ManifestFlag{
				{Name: "--id", Type: "integer", Required: true, Description: "custom intelligence id"},
				{Name: "--name", Type: "string", Required: true, Description: "custom intelligence name"},
				{Name: "--ioc", Type: "csv_string", Required: true, Description: "IOC values"},
				{Name: "--type", Type: "enum", Required: true, Enum: []string{"ip", "domain", "url", "md5", "sha1", "sha256", "1", "2", "3", "4", "5", "6"}, Description: "IOC type"},
				{Name: "--status", Type: "enum", Required: false, Default: "enabled", Enum: []string{"enabled", "disabled", "1", "0"}, Description: "custom intelligence status"},
				{Name: "--remarks", Type: "string", Required: false, Description: "remarks"},
				{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
				{Name: "--confirm", Type: "string", Required: false, Description: "must equal CONFIRM_POLICY_CUSTOM_INTELLIGENCE_UPDATE to execute"},
			}, "policy_custom_intelligence_update_result", []ManifestExample{
				{Description: "预览编辑自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence-update --id 12 --name 新情报 --ioc evil.example --type domain --preview"},
				{Description: "确认编辑自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence-update --id 12 --name 新情报 --ioc evil.example --type domain --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_UPDATE"},
			}, []string{"AlarmService.SearchAlarmCustomIntelligenceInfo", "AlarmService.UpdateAlarmCustomIntelligence"}, policyCustomIntelligenceUpdateConfirmToken),
			policyCustomIntelligenceWriteManifestCommand("tanswer policy custom-intelligence-enable", "chaitin-cli tanswer policy custom-intelligence-enable", "启用自定义情报。", []string{
				"需要启用一条或多条自定义 IOC 情报。",
			}, policyCustomIntelligenceStatusManifestFlags("CONFIRM_POLICY_CUSTOM_INTELLIGENCE_ENABLE"), "policy_custom_intelligence_enable_result", []ManifestExample{
				{Description: "确认启用自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence-enable --id-list 12,13 --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_ENABLE"},
			}, []string{"AlarmService.SearchAlarmCustomIntelligenceList", "AlarmService.UpdateAlarmCustomIntelligenceStatus"}, policyCustomIntelligenceEnableConfirmToken),
			policyCustomIntelligenceWriteManifestCommand("tanswer policy custom-intelligence-disable", "chaitin-cli tanswer policy custom-intelligence-disable", "禁用自定义情报。", []string{
				"需要禁用一条或多条自定义 IOC 情报。",
			}, policyCustomIntelligenceStatusManifestFlags("CONFIRM_POLICY_CUSTOM_INTELLIGENCE_DISABLE"), "policy_custom_intelligence_disable_result", []ManifestExample{
				{Description: "确认禁用自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence-disable --id-list 12,13 --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_DISABLE"},
			}, []string{"AlarmService.SearchAlarmCustomIntelligenceList", "AlarmService.UpdateAlarmCustomIntelligenceStatus"}, policyCustomIntelligenceDisableConfirmToken),
			policyCustomIntelligenceWriteManifestCommand("tanswer policy custom-intelligence-delete", "chaitin-cli tanswer policy custom-intelligence-delete", "删除自定义情报。", []string{
				"需要删除一条或多条自定义 IOC 情报。",
			}, policyCustomIntelligenceStatusManifestFlags("CONFIRM_POLICY_CUSTOM_INTELLIGENCE_DELETE"), "policy_custom_intelligence_delete_result", []ManifestExample{
				{Description: "确认删除自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence-delete --id-list 12,13 --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_DELETE"},
			}, []string{"AlarmService.SearchAlarmCustomIntelligenceList", "AlarmService.DeleteAlarmCustomIntelligence"}, policyCustomIntelligenceDeleteConfirmToken),
			policyFileExportManifestCommand("tanswer policy custom-intelligence-export", "chaitin-cli tanswer policy custom-intelligence-export", "导出自定义情报文件。", "custom_intelligence_export_file", []ManifestExample{
				{Description: "导出全部自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence-export --output ./custom-intelligence.xlsx"},
				{Description: "导出选中自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence-export --id-list 12,13 --output ./selected-custom-intelligence.xlsx"},
			}, []string{"AlarmDownloadService.ExportAlarmCustomIntelligence"}),
			policyFileImportManifestCommand("tanswer policy custom-intelligence-import", "chaitin-cli tanswer policy custom-intelligence-import", "导入自定义情报文件。", "policy_custom_intelligence_import_result", policyCustomIntelligenceImportConfirmToken, []ManifestExample{
				{Description: "预览导入自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence-import --file ./custom-intelligence.xlsx --preview"},
				{Description: "确认导入自定义情报", Command: "chaitin-cli tanswer policy custom-intelligence-import --file ./custom-intelligence.xlsx --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_IMPORT"},
			}, []string{"AlarmUploadService.ImportAlarmCustomIntelligence"}),
			responseManifestCommand("tanswer response block-policies", "chaitin-cli tanswer response block-policies", "查询旁路阻断策略。", []string{
				"需要查看当前旁路阻断策略、阻断对象、关联自动响应策略、失效时间或策略状态。",
				"需要按策略 ID、自动响应策略 ID、策略名称、阻断对象或状态筛选阻断策略。",
			}, []string{
				"需要新增、编辑、启停、删除或下发阻断策略时，后续使用带确认保护的写入命令。",
			}, responseBlockPolicyManifestFlags(), "response_block_policy_list", "block_policies", []ManifestExample{
				{Description: "查询旁路阻断策略", Command: "chaitin-cli tanswer response block-policies --page-size 10"},
				{Description: "按阻断对象查询策略", Command: "chaitin-cli tanswer response block-policies --object 198.51.100.10 --status enabled"},
			}, []string{"RulesService.SearchBlockRules"}),
			responseWriteManifestCommand("tanswer response block-policy-create", "chaitin-cli tanswer response block-policy-create", "新增旁路阻断策略。", []string{
				"需要对确认恶意对象新增旁路阻断策略。",
				"已经明确策略名称、阻断对象、对象类型、阻断时长或失效时间。",
			}, responseBlockPolicyWriteManifestFlags(false, responseBlockPolicyCreateConfirmToken), "response_block_policy_create_result", []ManifestExample{
				{Description: "预览新增旁路阻断策略", Command: "chaitin-cli tanswer response block-policy-create --name block-bad-ip --object 198.51.100.10 --preview"},
				{Description: "确认新增旁路阻断策略", Command: "chaitin-cli tanswer response block-policy-create --name block-bad-ip --object 198.51.100.10 --duration 3600 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_CREATE"},
			}, []string{"RulesService.CreateBlockRules"}, responseBlockPolicyCreateConfirmToken),
			responseWriteManifestCommand("tanswer response block-policy-update", "chaitin-cli tanswer response block-policy-update", "编辑旁路阻断策略。", []string{
				"需要更新单条旁路阻断策略。",
				"需要执行前查看当前策略与目标配置差异。",
			}, responseBlockPolicyWriteManifestFlags(true, responseBlockPolicyUpdateConfirmToken), "response_block_policy_update_result", []ManifestExample{
				{Description: "预览编辑旁路阻断策略", Command: "chaitin-cli tanswer response block-policy-update --id 7 --name new-block --object 198.51.100.10 --preview"},
				{Description: "确认编辑旁路阻断策略", Command: "chaitin-cli tanswer response block-policy-update --id 7 --name new-block --object 198.51.100.10 --expire 1784277612410 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_UPDATE"},
			}, []string{"RulesService.SearchBlockRules", "RulesService.UpdateBlockRules"}, responseBlockPolicyUpdateConfirmToken),
			responseWriteManifestCommand("tanswer response block-policy-enable", "chaitin-cli tanswer response block-policy-enable", "启用旁路阻断策略。", []string{
				"需要启用一条或多条旁路阻断策略。",
			}, responseBlockPolicyActionManifestFlags(responseBlockPolicyEnableConfirmToken), "response_block_policy_enable_result", []ManifestExample{
				{Description: "确认启用旁路阻断策略", Command: "chaitin-cli tanswer response block-policy-enable --id-list 7,8 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_ENABLE"},
			}, []string{"RulesService.SearchBlockRules", "RulesService.UpdateBlockRulesStatus"}, responseBlockPolicyEnableConfirmToken),
			responseWriteManifestCommand("tanswer response block-policy-disable", "chaitin-cli tanswer response block-policy-disable", "停用旁路阻断策略。", []string{
				"需要停用一条或多条旁路阻断策略。",
			}, responseBlockPolicyActionManifestFlags(responseBlockPolicyDisableConfirmToken), "response_block_policy_disable_result", []ManifestExample{
				{Description: "确认停用旁路阻断策略", Command: "chaitin-cli tanswer response block-policy-disable --id-list 7,8 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_DISABLE"},
			}, []string{"RulesService.SearchBlockRules", "RulesService.UpdateBlockRulesStatus"}, responseBlockPolicyDisableConfirmToken),
			responseWriteManifestCommand("tanswer response block-policy-delete", "chaitin-cli tanswer response block-policy-delete", "删除旁路阻断策略。", []string{
				"需要删除一条或多条旁路阻断策略。",
			}, responseBlockPolicyActionManifestFlags(responseBlockPolicyDeleteConfirmToken), "response_block_policy_delete_result", []ManifestExample{
				{Description: "确认删除旁路阻断策略", Command: "chaitin-cli tanswer response block-policy-delete --id-list 7,8 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_DELETE"},
			}, []string{"RulesService.SearchBlockRules", "RulesService.UpdateBlockRulesStatus"}, responseBlockPolicyDeleteConfirmToken),
			responseManifestCommand("tanswer response block-records", "chaitin-cli tanswer response block-records", "查询旁路阻断记录。", []string{
				"需要确认旁路阻断是否在产品记录中命中。",
				"需要按时间、源/目的 IP、源/目的端口、阻断策略或自动响应策略筛选阻断记录。",
			}, []string{
				"需要验证第三方设备或网络层真实阻断效果时，不使用该命令直接下结论。",
			}, responseBlockRecordManifestFlags(), "response_block_record_list", "block_records", []ManifestExample{
				{Description: "查询近 24 小时旁路阻断记录", Command: "chaitin-cli tanswer response block-records --time 24h --page-size 10"},
			}, []string{"RulesService.SearchTapBlockRecordList"}),
			responseManifestCommand("tanswer response whitelist", "chaitin-cli tanswer response whitelist", "查询响应白名单。", []string{
				"需要查看不会被响应处置影响的 IP 或 URL 对象、阻断方式、有效期或状态。",
				"需要按白名单对象、类型、策略状态、有效期或更新时间筛选响应白名单。",
			}, []string{
				"需要新增、编辑、启停或删除响应白名单时，后续使用带确认保护的写入命令。",
				"需要查询检测白名单时，使用 tanswer policy detection-whitelist。",
			}, responseWhitelistManifestFlags(), "response_whitelist_list", "response_whitelists", []ManifestExample{
				{Description: "查询响应白名单", Command: "chaitin-cli tanswer response whitelist --page-size 10"},
				{Description: "按对象查询响应白名单", Command: "chaitin-cli tanswer response whitelist --object 198.51.100.10 --type ip"},
			}, []string{"FirewallService.SearchWhiteList"}),
			responseWriteManifestCommand("tanswer response whitelist-create", "chaitin-cli tanswer response whitelist-create", "新增响应白名单。", []string{
				"需要将确认可信对象加入响应处置白名单。",
				"已经明确白名单对象、对象类型、有效期、阻断方式和 IP 方向。",
			}, responseWhitelistWriteManifestFlags(false, responseWhitelistCreateConfirmToken), "response_whitelist_create_result", []ManifestExample{
				{Description: "预览新增响应白名单", Command: "chaitin-cli tanswer response whitelist-create --type ip --object 198.51.100.10 --expire 1784277612410 --preview"},
				{Description: "确认新增响应白名单", Command: "chaitin-cli tanswer response whitelist-create --type ip --object 198.51.100.10 --expire 1784277612410 --confirm CONFIRM_RESPONSE_WHITELIST_CREATE"},
			}, []string{"FirewallService.CreateWhiteList"}, responseWhitelistCreateConfirmToken),
			responseWriteManifestCommand("tanswer response whitelist-update", "chaitin-cli tanswer response whitelist-update", "编辑响应白名单。", []string{
				"需要更新单条响应处置白名单。",
				"需要执行前查看当前白名单与目标配置差异。",
			}, responseWhitelistWriteManifestFlags(true, responseWhitelistUpdateConfirmToken), "response_whitelist_update_result", []ManifestExample{
				{Description: "预览编辑响应白名单", Command: "chaitin-cli tanswer response whitelist-update --id 3 --type ip --object 198.51.100.10 --expire 1784277612410 --preview"},
				{Description: "确认编辑响应白名单", Command: "chaitin-cli tanswer response whitelist-update --id 3 --type url --object http://example.com/a --expire 1784277612410 --confirm CONFIRM_RESPONSE_WHITELIST_UPDATE"},
			}, []string{"FirewallService.SearchWhiteList", "FirewallService.UpdateWhiteList"}, responseWhitelistUpdateConfirmToken),
			responseWriteManifestCommand("tanswer response whitelist-enable", "chaitin-cli tanswer response whitelist-enable", "启用响应白名单。", []string{
				"需要启用一条或多条响应处置白名单。",
			}, responseWhitelistActionManifestFlags(responseWhitelistEnableConfirmToken), "response_whitelist_enable_result", []ManifestExample{
				{Description: "确认启用响应白名单", Command: "chaitin-cli tanswer response whitelist-enable --id-list 3,4 --confirm CONFIRM_RESPONSE_WHITELIST_ENABLE"},
			}, []string{"FirewallService.SearchWhiteList", "FirewallService.UpdateWhiteListStatus"}, responseWhitelistEnableConfirmToken),
			responseWriteManifestCommand("tanswer response whitelist-disable", "chaitin-cli tanswer response whitelist-disable", "停用响应白名单。", []string{
				"需要停用一条或多条响应处置白名单。",
			}, responseWhitelistActionManifestFlags(responseWhitelistDisableConfirmToken), "response_whitelist_disable_result", []ManifestExample{
				{Description: "确认停用响应白名单", Command: "chaitin-cli tanswer response whitelist-disable --id-list 3,4 --confirm CONFIRM_RESPONSE_WHITELIST_DISABLE"},
			}, []string{"FirewallService.SearchWhiteList", "FirewallService.UpdateWhiteListStatus"}, responseWhitelistDisableConfirmToken),
			responseWriteManifestCommand("tanswer response whitelist-delete", "chaitin-cli tanswer response whitelist-delete", "删除响应白名单。", []string{
				"需要删除一条或多条响应处置白名单。",
			}, responseWhitelistActionManifestFlags(responseWhitelistDeleteConfirmToken), "response_whitelist_delete_result", []ManifestExample{
				{Description: "确认删除响应白名单", Command: "chaitin-cli tanswer response whitelist-delete --id-list 3,4 --confirm CONFIRM_RESPONSE_WHITELIST_DELETE"},
			}, []string{"FirewallService.SearchWhiteList", "FirewallService.DeleteWhiteList"}, responseWhitelistDeleteConfirmToken),
			responseWriteManifestCommand("tanswer response block-policy-from-alarm", "chaitin-cli tanswer response block-policy-from-alarm", "从告警生成旁路阻断策略。", []string{
				"需要基于已研判恶意告警生成候选旁路阻断策略。",
				"需要执行前查看告警上下文和候选阻断对象。",
			}, responseBlockPolicyFromAlarmManifestFlags(), "response_block_policy_from_alarm_result", []ManifestExample{
				{Description: "预览从告警生成旁路阻断策略", Command: "chaitin-cli tanswer response block-policy-from-alarm --id '<doc_id>' --target attacker --preview"},
				{Description: "确认从告警生成旁路阻断策略", Command: "chaitin-cli tanswer response block-policy-from-alarm --id '<doc_id>' --target flow --duration 3600 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_FROM_ALARM"},
			}, []string{"AlarmService.GetAlarm", "RulesService.CreateBlockRules"}, responseBlockPolicyFromAlarmConfirmToken),
			responseWriteManifestCommand("tanswer response whitelist-from-alarm", "chaitin-cli tanswer response whitelist-from-alarm", "从告警生成响应白名单。", []string{
				"需要基于已确认可信或误阻断风险的告警生成候选响应白名单。",
				"需要执行前查看告警上下文和候选白名单对象。",
			}, responseWhitelistFromAlarmManifestFlags(), "response_whitelist_from_alarm_result", []ManifestExample{
				{Description: "预览从告警生成响应白名单", Command: "chaitin-cli tanswer response whitelist-from-alarm --id '<doc_id>' --target victim --expire 1784277612410 --preview"},
				{Description: "确认从告警生成响应白名单", Command: "chaitin-cli tanswer response whitelist-from-alarm --id '<doc_id>' --target url --expire 1784277612410 --confirm CONFIRM_RESPONSE_WHITELIST_FROM_ALARM"},
			}, []string{"AlarmService.GetAlarm", "FirewallService.CreateWhiteList"}, responseWhitelistFromAlarmConfirmToken),
			responseManifestCommand("tanswer response devices", "chaitin-cli tanswer response devices", "查询联动设备配置。", []string{
				"需要查看第三方协同处置设备是否配置、设备状态、设备地址或更新时间。",
			}, []string{
				"需要测试、修改、删除联动设备配置时，不使用该只读命令。",
			}, responseDevicesManifestFlags(), "response_device_list", "devices", []ManifestExample{
				{Description: "查询联动设备", Command: "chaitin-cli tanswer response devices --page-size 10"},
			}, []string{"FirewallService.SearchFirewall"}),
			responseManifestCommand("tanswer response device-records", "chaitin-cli tanswer response device-records --device-id '<device_id>'", "查询联动设备处置记录。", []string{
				"需要查看某个联动设备的处置下发记录和状态。",
			}, []string{
				"没有 device-id 时，先使用 tanswer response devices 查询设备。",
				"需要验证第三方设备真实执行结果时，不使用该命令直接下结论。",
			}, []ManifestFlag{
				{Name: "--device-id", Type: "integer", Required: true, Description: "linkage device id"},
				{Name: "--page", Type: "integer", Required: false, Default: "1", Description: "page number, starts from 1"},
				{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "page size, 1-100"},
			}, "response_device_record_list", "device_records", []ManifestExample{
				{Description: "查询联动设备处置记录", Command: "chaitin-cli tanswer response device-records --device-id '<device_id>' --page-size 10"},
			}, []string{"FirewallService.SearchSendRecord"}),
			responseManifestCommand("tanswer response auto-policies", "chaitin-cli tanswer response auto-policies", "查询自动响应策略。", []string{
				"需要查看自动响应策略配置、启停状态、联动处置、处置方式或更新时间。",
			}, []string{
				"需要启停、编辑或删除自动响应策略时，不使用该只读命令。",
			}, responseAutoPolicyManifestFlags(), "response_auto_policy_list", "auto_policies", []ManifestExample{
				{Description: "查询自动响应策略", Command: "chaitin-cli tanswer response auto-policies --page-size 10"},
			}, []string{"FirewallService.SearchStrategy"}),
			responseManifestCommand("tanswer response auto-list", "chaitin-cli tanswer response auto-list", "查询自动响应处置名单。", []string{
				"需要查看自动响应策略生成的待处置或已处置对象。",
				"需要按 IP、状态、响应策略或时间范围筛选自动响应处置名单。",
			}, []string{
				"需要主动触发自动响应分析或处置时，不使用该只读命令。",
			}, responseAutoListManifestFlags(), "response_auto_list", "auto_list", []ManifestExample{
				{Description: "查询自动响应处置名单", Command: "chaitin-cli tanswer response auto-list --time 7d --page-size 10"},
			}, []string{"FirewallService.SearchBlackList"}),
		},
	}
	applySemanticProtectedWriteContract(&manifest)
	return manifest
}

var semanticProtectedWritePreviewFields = []string{
	"requires_confirmation",
	"confirmed",
	"operation_type",
	"risk_level",
	"target",
	"change_summary",
	"impact",
	"risk_warnings",
	"confirmation_token",
	"confirmation_note",
}

func applySemanticProtectedWriteContract(manifest *CommandManifest) {
	for index := range manifest.Commands {
		command := &manifest.Commands[index]
		if command.Layer != "semantic_shortcut" || !command.RequiresConfirmation {
			continue
		}
		if len(command.PreviewOutputFields) == 0 {
			command.PreviewOutputFields = append([]string(nil), semanticProtectedWritePreviewFields...)
		}
		if command.ConfirmationCondition == "" {
			command.ConfirmationCondition = "required after preview and explicit user confirmation"
		}
	}
}

func responseManifestCommand(name string, fullCommand string, summary string, useWhen []string, doNotUseWhen []string, flags []ManifestFlag, outputType string, dataKey string, examples []ManifestExample, rpcMethods []string) ManifestCommand {
	return ManifestCommand{
		Name:        name,
		FullCommand: fullCommand,
		Layer:       "semantic_shortcut",
		Summary:     summary,
		UseWhen:     useWhen,
		DoNotUseWhen: append(doNotUseWhen,
			"需要创建、启停、删除、下发、加白或阻断时，必须使用后续带预览和确认保护的写入命令。",
		),
		Flags:      flags,
		OutputType: outputType,
		OutputFields: []string{
			"total",
			"page",
			"page_size",
			"current_count",
			"has_more",
			dataKey,
		},
		RiskLevel:            "read",
		RequiresConfirmation: false,
		Examples:             examples,
		Backend: map[string]interface{}{
			"rpc_methods": rpcMethods,
		},
	}
}

func responseWriteManifestCommand(name string, fullCommand string, summary string, useWhen []string, flags []ManifestFlag, outputType string, examples []ManifestExample, rpcMethods []string, confirmToken string) ManifestCommand {
	return ManifestCommand{
		Name:        name,
		FullCommand: fullCommand,
		Layer:       "semantic_shortcut",
		Summary:     summary,
		UseWhen:     useWhen,
		DoNotUseWhen: []string{
			"只需要查询旁路阻断策略或阻断记录时，使用 response read-only 命令。",
			"没有明确处置意图或没有确认 token 时，不执行写操作。",
		},
		Flags:      flags,
		OutputType: outputType,
		OutputFields: []string{
			"preview.requires_confirmation",
			"preview.confirmation_token",
			"execution.confirmed",
			"execution.result",
			"execution.object",
			"execution.audit",
		},
		RiskLevel:            "write_high",
		RequiresConfirmation: true,
		Examples:             examples,
		Backend: map[string]interface{}{
			"rpc_methods":         rpcMethods,
			"confirmation_token":  confirmToken,
			"write_protection":    "preview_then_exact_confirm",
			"audit_actor_source":  "open_api_token",
			"requires_token_auth": true,
		},
	}
}

func policyFileExportManifestCommand(name string, fullCommand string, summary string, outputType string, examples []ManifestExample, rpcMethods []string) ManifestCommand {
	return ManifestCommand{
		Name:        name,
		FullCommand: fullCommand,
		Layer:       "semantic_shortcut",
		Summary:     summary,
		UseWhen: []string{
			"需要导出检测白名单或自定义情报文件用于备份或离线维护。",
			"需要按选中 ID 导出部分策略对象，或未指定 ID 时导出全部对象。",
		},
		DoNotUseWhen: []string{
			"需要新增、编辑、启停或删除策略对象时，使用对应受保护写命令。",
			"需要导入策略文件时，使用对应 import 命令。",
		},
		Flags: []ManifestFlag{
			{Name: "--id-list", Type: "csv_integer", Required: false, Description: "object IDs to export; empty exports all"},
			{Name: "--output", Type: "path", Required: false, Description: "output file path"},
		},
		OutputType: outputType,
		OutputFields: []string{
			"file_name",
			"file_path",
			"size_bytes",
			"status_code",
			"method",
			"download_query",
			"export_scope",
		},
		RiskLevel:            "read",
		RequiresConfirmation: false,
		Examples:             examples,
		Backend: map[string]interface{}{
			"rpc_methods":         rpcMethods,
			"requires_token_auth": true,
		},
	}
}

func policyFileImportManifestCommand(name string, fullCommand string, summary string, outputType string, confirmToken string, examples []ManifestExample, rpcMethods []string) ManifestCommand {
	return ManifestCommand{
		Name:        name,
		FullCommand: fullCommand,
		Layer:       "semantic_shortcut",
		Summary:     summary,
		UseWhen: []string{
			"需要通过产品导入模板批量维护检测白名单或自定义情报。",
			"已经准备好产品支持的导入文件，并需要执行前查看本地文件预览。",
		},
		DoNotUseWhen: []string{
			"只需要查询或导出策略对象时，不使用 import 命令。",
			"没有明确导入意图或没有确认 token 时，不执行上传。",
		},
		Flags: []ManifestFlag{
			{Name: "--file", Type: "path", Required: true, Description: "policy import template file path"},
			{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without uploading"},
			{Name: "--confirm", Type: "string", Required: false, Description: "must equal " + confirmToken + " to upload"},
		},
		OutputType: outputType,
		OutputFields: []string{
			"preview.requires_confirmation",
			"preview.confirmation_token",
			"execution.confirmed",
			"execution.result",
			"execution.object",
			"execution.audit",
			"execution.import_result",
		},
		RiskLevel:            "write_high",
		RequiresConfirmation: true,
		Examples:             examples,
		Backend: map[string]interface{}{
			"rpc_methods":         rpcMethods,
			"confirmation_token":  confirmToken,
			"write_protection":    "preview_then_exact_confirm",
			"audit_actor_source":  "open_api_token",
			"requires_token_auth": true,
		},
	}
}

func policyCustomIntelligenceWriteManifestCommand(name string, fullCommand string, summary string, useWhen []string, flags []ManifestFlag, outputType string, examples []ManifestExample, rpcMethods []string, confirmToken string) ManifestCommand {
	return ManifestCommand{
		Name:        name,
		FullCommand: fullCommand,
		Layer:       "semantic_shortcut",
		Summary:     summary,
		UseWhen:     useWhen,
		DoNotUseWhen: []string{
			"只需要查询自定义情报列表时，使用 tanswer policy custom-intelligence。",
			"需要导入或导出自定义情报文件时，使用后续文件型命令。",
		},
		Flags:      flags,
		OutputType: outputType,
		OutputFields: []string{
			"preview.requires_confirmation",
			"preview.confirmation_token",
			"execution.confirmed",
			"execution.result",
			"execution.object",
			"execution.audit",
		},
		RiskLevel:            "write_high",
		RequiresConfirmation: true,
		Examples:             examples,
		Backend: map[string]interface{}{
			"rpc_methods":         rpcMethods,
			"confirmation_token":  confirmToken,
			"write_protection":    "preview_then_exact_confirm",
			"audit_actor_source":  "open_api_token",
			"requires_token_auth": true,
		},
	}
}

func policyDetectionWhitelistWriteManifestCommand(name string, fullCommand string, summary string, useWhen []string, flags []ManifestFlag, outputType string, examples []ManifestExample, rpcMethods []string, confirmToken string) ManifestCommand {
	return ManifestCommand{
		Name:        name,
		FullCommand: fullCommand,
		Layer:       "semantic_shortcut",
		Summary:     summary,
		UseWhen:     useWhen,
		DoNotUseWhen: []string{
			"只需要查询检测白名单列表时，使用 tanswer policy detection-whitelist。",
			"需要响应白名单或旁路阻断策略时，使用 response 域命令。",
			"需要导入或导出检测白名单文件时，使用后续文件型命令。",
		},
		Flags: flags,
		InputConstraints: []ManifestInputConstraint{
			{
				Rule:        "at_least_one",
				Flags:       []string{"--expire", "--valid-time"},
				Description: "provide an expiry timestamp in milliseconds or a positive valid duration in seconds",
			},
		},
		OutputType: outputType,
		OutputFields: []string{
			"preview.requires_confirmation",
			"preview.confirmation_token",
			"execution.confirmed",
			"execution.result",
			"execution.object",
			"execution.audit",
		},
		RiskLevel:            "write_high",
		RequiresConfirmation: true,
		Examples:             examples,
		Backend: map[string]interface{}{
			"rpc_methods":         rpcMethods,
			"confirmation_token":  confirmToken,
			"write_protection":    "preview_then_exact_confirm",
			"audit_actor_source":  "open_api_token",
			"requires_token_auth": true,
		},
	}
}

func policyDetectionWhitelistWriteManifestFlags(includeID bool, confirmToken string) []ManifestFlag {
	flags := []ManifestFlag{}
	if includeID {
		flags = append(flags, ManifestFlag{Name: "--id", Type: "integer", Required: true, Description: "detection whitelist id"})
	}
	flags = append(flags,
		ManifestFlag{Name: "--name", Type: "string", Required: true, Description: "detection whitelist name"},
		ManifestFlag{Name: "--src-ip", Type: "string", Required: false, Description: "source IP"},
		ManifestFlag{Name: "--src-port", Type: "string", Required: false, Description: "source port"},
		ManifestFlag{Name: "--dest-ip", Type: "string", Required: false, Description: "destination IP"},
		ManifestFlag{Name: "--dest-port", Type: "string", Required: false, Description: "destination port"},
		ManifestFlag{Name: "--domain", Type: "string", Required: false, Description: "domain"},
		ManifestFlag{Name: "--url-path", Type: "string", Required: false, Description: "URL path"},
		ManifestFlag{Name: "--user-agent", Type: "string", Required: false, Description: "User-Agent"},
		ManifestFlag{Name: "--xff", Type: "string", Required: false, Description: "XFF"},
		ManifestFlag{Name: "--resp-code", Type: "string", Required: false, Description: "response status code"},
		ManifestFlag{Name: "--resp-body", Type: "string", Required: false, Description: "response body"},
		ManifestFlag{Name: "--threat", Type: "string", Required: false, Description: "threat type"},
		ManifestFlag{Name: "--rule-id", Type: "string", Required: false, Description: "detection rule id"},
		ManifestFlag{Name: "--storage", Type: "enum", Required: false, Default: "drop", Enum: []string{"drop", "ignore", "1", "2"}, Description: "handling mode"},
		ManifestFlag{Name: "--status", Type: "enum", Required: false, Default: "enabled", Enum: []string{"enabled", "disabled", "1", "0"}, Description: "status"},
		ManifestFlag{Name: "--mode", Type: "enum", Required: false, Default: "default", Enum: []string{"default", "advanced", "1", "2"}, Description: "match mode"},
		ManifestFlag{Name: "--expire", Type: "integer", Required: false, Description: "expiry timestamp in milliseconds; provide this or --valid-time; when set, must be greater than 0"},
		ManifestFlag{Name: "--valid-time", Type: "integer", Required: false, Description: "valid duration in seconds; provide this or --expire; must be greater than 0"},
		ManifestFlag{Name: "--ignore-history", Type: "boolean", Required: false, Description: "ignore matched historical alarms"},
		ManifestFlag{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
		ManifestFlag{Name: "--confirm", Type: "string", Required: false, Description: "must equal " + confirmToken + " to execute"},
	)
	return flags
}

func policyDetectionWhitelistFromAlarmManifestFlags() []ManifestFlag {
	flags := []ManifestFlag{
		{Name: "--id", Type: "string", Required: true, Description: "source alarm doc_id returned by alarm commands"},
	}
	flags = append(flags, policyDetectionWhitelistWriteManifestFlags(false, policyDetectionWhitelistFromAlarmConfirmToken)...)
	return flags
}

func policyDetectionWhitelistActionManifestFlags(confirmToken string) []ManifestFlag {
	return []ManifestFlag{
		{Name: "--id-list", Type: "csv_integer", Required: true, Description: "detection whitelist IDs"},
		{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
		{Name: "--confirm", Type: "string", Required: false, Description: "must equal " + confirmToken + " to execute"},
	}
}

func policyCustomIntelligenceStatusManifestFlags(confirmToken string) []ManifestFlag {
	return []ManifestFlag{
		{Name: "--id-list", Type: "csv_integer", Required: true, Description: "custom intelligence IDs"},
		{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
		{Name: "--confirm", Type: "string", Required: false, Description: "must equal " + confirmToken + " to execute"},
	}
}

func responseBlockPolicyManifestFlags() []ManifestFlag {
	return []ManifestFlag{
		{Name: "--page", Type: "integer", Required: false, Default: "1", Description: "page number, starts from 1"},
		{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "page size, 1-100"},
		{Name: "--id", Type: "integer", Required: false, Description: "block policy id"},
		{Name: "--strategy-id", Type: "integer", Required: false, Description: "auto response strategy id"},
		{Name: "--name", Type: "string", Required: false, Description: "block policy name filter"},
		{Name: "--object", Type: "string", Required: false, Description: "block object IP/CIDR/range filter"},
		{Name: "--status", Type: "enum", Required: false, Enum: []string{"enabled", "disabled", "1", "0"}, Description: "policy status filter"},
	}
}

func responseBlockPolicyWriteManifestFlags(includeID bool, confirmToken string) []ManifestFlag {
	flags := []ManifestFlag{}
	if includeID {
		flags = append(flags, ManifestFlag{Name: "--id", Type: "integer", Required: true, Description: "block policy id"})
	}
	flags = append(flags,
		ManifestFlag{Name: "--name", Type: "string", Required: true, Description: "block policy name"},
		ManifestFlag{Name: "--object", Type: "csv_string", Required: true, Description: "block objects"},
		ManifestFlag{Name: "--object-type", Type: "enum", Required: false, Default: "ip", Enum: []string{"ip", "ip-port", "tuple", "url", "1", "2", "3", "4"}, Description: "block object type"},
		ManifestFlag{Name: "--ip-type", Type: "enum", Required: false, Default: "both", Enum: []string{"both", "source", "dest", "0", "1", "2"}, Description: "IP direction for ip object type"},
		ManifestFlag{Name: "--status", Type: "enum", Required: false, Default: "enabled", Enum: []string{"enabled", "disabled", "1", "0"}, Description: "policy status"},
		ManifestFlag{Name: "--block-time-type", Type: "enum", Required: false, Default: "duration", Enum: []string{"duration", "expire", "1", "2"}, Description: "block time type"},
		ManifestFlag{Name: "--duration", Type: "integer", Required: false, Default: "3600", Description: "block duration in seconds"},
		ManifestFlag{Name: "--expire", Type: "integer", Required: false, Description: "expire timestamp in milliseconds"},
		ManifestFlag{Name: "--remark", Type: "string", Required: false, Description: "remark"},
		ManifestFlag{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
		ManifestFlag{Name: "--confirm", Type: "string", Required: false, Description: "must equal " + confirmToken + " to execute"},
	)
	return flags
}

func responseBlockPolicyActionManifestFlags(confirmToken string) []ManifestFlag {
	return []ManifestFlag{
		{Name: "--id-list", Type: "csv_integer", Required: true, Description: "block policy IDs"},
		{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
		{Name: "--confirm", Type: "string", Required: false, Description: "must equal " + confirmToken + " to execute"},
	}
}

func responseBlockRecordManifestFlags() []ManifestFlag {
	flags := responseTimePageManifestFlags()
	flags = append(flags,
		ManifestFlag{Name: "--src-ip", Type: "string", Required: false, Description: "source IP filter"},
		ManifestFlag{Name: "--src-port", Type: "integer", Required: false, Description: "source port filter"},
		ManifestFlag{Name: "--dest-ip", Type: "string", Required: false, Description: "destination IP filter"},
		ManifestFlag{Name: "--dest-port", Type: "integer", Required: false, Description: "destination port filter"},
		ManifestFlag{Name: "--policy-id", Type: "integer", Required: false, Description: "block policy id"},
		ManifestFlag{Name: "--policy-name", Type: "string", Required: false, Description: "block policy name"},
		ManifestFlag{Name: "--strategy-id", Type: "integer", Required: false, Description: "auto response strategy id"},
		ManifestFlag{Name: "--type", Type: "integer", Required: false, Description: "block record type"},
	)
	return flags
}

func responseWhitelistManifestFlags() []ManifestFlag {
	flags := responseTimePageManifestFlags()
	flags = append(flags,
		ManifestFlag{Name: "--type", Type: "enum", Required: false, Enum: []string{"ip", "url"}, Description: "whitelist object type"},
		ManifestFlag{Name: "--object", Type: "string", Required: false, Description: "whitelist object filter"},
		ManifestFlag{Name: "--status", Type: "enum", Required: false, Enum: []string{"enabled", "disabled", "1", "0"}, Description: "whitelist status filter"},
		ManifestFlag{Name: "--expire-after", Type: "integer", Required: false, Description: "expire timestamp lower bound in milliseconds"},
		ManifestFlag{Name: "--block-method", Type: "string", Required: false, Description: "block method filter"},
		ManifestFlag{Name: "--ip-type", Type: "string", Required: false, Description: "IP type filter"},
		ManifestFlag{Name: "--remark", Type: "string", Required: false, Description: "remark filter"},
	)
	return flags
}

func responseWhitelistWriteManifestFlags(includeID bool, confirmToken string) []ManifestFlag {
	flags := []ManifestFlag{}
	if includeID {
		flags = append(flags, ManifestFlag{Name: "--id", Type: "integer", Required: true, Description: "response whitelist id"})
	}
	flags = append(flags,
		ManifestFlag{Name: "--type", Type: "enum", Required: false, Default: "ip", Enum: []string{"ip", "url"}, Description: "whitelist object type"},
		ManifestFlag{Name: "--object", Type: "csv_string", Required: true, Description: "whitelist objects"},
		ManifestFlag{Name: "--status", Type: "enum", Required: false, Default: "enabled", Enum: []string{"enabled", "disabled", "2", "1"}, Description: "whitelist status"},
		ManifestFlag{Name: "--expire", Type: "integer", Required: true, Description: "expire timestamp in milliseconds"},
		ManifestFlag{Name: "--block-method", Type: "csv_enum", Required: false, Enum: []string{"Bypass", "Third_party"}, Description: "IP whitelist block methods"},
		ManifestFlag{Name: "--ip-type", Type: "enum", Required: false, Default: "both", Enum: []string{"both", "source", "dest", "SRC_OR_DST", "SRC_IP", "DST_IP"}, Description: "IP direction for ip whitelist"},
		ManifestFlag{Name: "--remark", Type: "string", Required: false, Description: "remark"},
		ManifestFlag{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
		ManifestFlag{Name: "--confirm", Type: "string", Required: false, Description: "must equal " + confirmToken + " to execute"},
	)
	return flags
}

func responseWhitelistActionManifestFlags(confirmToken string) []ManifestFlag {
	return []ManifestFlag{
		{Name: "--id-list", Type: "csv_integer", Required: true, Description: "response whitelist IDs"},
		{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
		{Name: "--confirm", Type: "string", Required: false, Description: "must equal " + confirmToken + " to execute"},
	}
}

func responseBlockPolicyFromAlarmManifestFlags() []ManifestFlag {
	return []ManifestFlag{
		{Name: "--id", Type: "string", Required: true, Description: "alarm doc_id from alarm list or alarm detail"},
		{Name: "--target", Type: "enum", Required: false, Default: "attacker", Enum: []string{"attacker", "victim", "flow"}, Description: "candidate block object source from alarm"},
		{Name: "--name", Type: "string", Required: false, Description: "block policy name; defaults from alarm name"},
		{Name: "--object", Type: "csv_string", Required: false, Description: "override block objects instead of deriving from alarm"},
		{Name: "--object-type", Type: "enum", Required: false, Default: "ip", Enum: []string{"ip", "ip-port", "tuple", "url", "1", "2", "3", "4"}, Description: "block object type"},
		{Name: "--ip-type", Type: "enum", Required: false, Default: "both", Enum: []string{"both", "source", "dest", "0", "1", "2"}, Description: "IP direction for ip object type"},
		{Name: "--status", Type: "enum", Required: false, Default: "enabled", Enum: []string{"enabled", "disabled", "1", "0"}, Description: "policy status"},
		{Name: "--block-time-type", Type: "enum", Required: false, Default: "duration", Enum: []string{"duration", "expire", "1", "2"}, Description: "block time type"},
		{Name: "--duration", Type: "integer", Required: false, Default: "3600", Description: "block duration in seconds"},
		{Name: "--expire", Type: "integer", Required: false, Description: "expire timestamp in milliseconds"},
		{Name: "--remark", Type: "string", Required: false, Description: "remark"},
		{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
		{Name: "--confirm", Type: "string", Required: false, Description: "must equal " + responseBlockPolicyFromAlarmConfirmToken + " to execute"},
	}
}

func responseWhitelistFromAlarmManifestFlags() []ManifestFlag {
	return []ManifestFlag{
		{Name: "--id", Type: "string", Required: true, Description: "alarm doc_id from alarm list or alarm detail"},
		{Name: "--target", Type: "enum", Required: false, Default: "victim", Enum: []string{"attacker", "victim", "url"}, Description: "candidate whitelist object source from alarm"},
		{Name: "--type", Type: "enum", Required: false, Default: "ip", Enum: []string{"ip", "url"}, Description: "whitelist object type"},
		{Name: "--object", Type: "csv_string", Required: false, Description: "override whitelist objects instead of deriving from alarm"},
		{Name: "--status", Type: "enum", Required: false, Default: "enabled", Enum: []string{"enabled", "disabled", "2", "1"}, Description: "whitelist status"},
		{Name: "--expire", Type: "integer", Required: true, Description: "expire timestamp in milliseconds"},
		{Name: "--block-method", Type: "csv_enum", Required: false, Enum: []string{"Bypass", "Third_party"}, Description: "IP whitelist block methods"},
		{Name: "--ip-type", Type: "enum", Required: false, Default: "both", Enum: []string{"both", "source", "dest", "SRC_OR_DST", "SRC_IP", "DST_IP"}, Description: "IP direction for ip whitelist"},
		{Name: "--remark", Type: "string", Required: false, Description: "remark"},
		{Name: "--preview", Type: "boolean", Required: false, Description: "return write preview without executing"},
		{Name: "--confirm", Type: "string", Required: false, Description: "must equal " + responseWhitelistFromAlarmConfirmToken + " to execute"},
	}
}

func responseDevicesManifestFlags() []ManifestFlag {
	return []ManifestFlag{
		{Name: "--page", Type: "integer", Required: false, Default: "1", Description: "page number, starts from 1"},
		{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "page size, 1-100"},
		{Name: "--id", Type: "csv_integer", Required: false, Description: "device id filter"},
		{Name: "--device-type", Type: "integer", Required: false, Description: "device type filter"},
		{Name: "--status", Type: "enum", Required: false, Enum: []string{"enabled", "disabled", "1", "0"}, Description: "device status filter"},
		{Name: "--remark", Type: "string", Required: false, Description: "remark filter"},
	}
}

func responseAutoPolicyManifestFlags() []ManifestFlag {
	flags := responseTimePageManifestFlags()
	flags = append(flags,
		ManifestFlag{Name: "--id", Type: "integer", Required: false, Description: "auto response policy id"},
		ManifestFlag{Name: "--name", Type: "string", Required: false, Description: "auto response policy name filter"},
		ManifestFlag{Name: "--device-id", Type: "csv_integer", Required: false, Description: "linkage device id filter"},
		ManifestFlag{Name: "--status", Type: "enum", Required: false, Enum: []string{"enabled", "disabled", "1", "0"}, Description: "policy status filter"},
		ManifestFlag{Name: "--punish-type", Type: "integer", Required: false, Description: "punish type filter"},
	)
	return flags
}

func responseAutoListManifestFlags() []ManifestFlag {
	flags := responseTimePageManifestFlags()
	flags = append(flags,
		ManifestFlag{Name: "--ip", Type: "string", Required: false, Description: "IP filter"},
		ManifestFlag{Name: "--status", Type: "enum", Required: false, Enum: []string{"enabled", "disabled", "1", "0"}, Description: "auto list status filter"},
		ManifestFlag{Name: "--strategy-id", Type: "integer", Required: false, Description: "auto response strategy id"},
		ManifestFlag{Name: "--block-time-type", Type: "integer", Required: false, Description: "block time type filter"},
	)
	return flags
}

func responseTimePageManifestFlags() []ManifestFlag {
	return []ManifestFlag{
		{Name: "--time", Type: "string", Required: false, Default: "today", Enum: []string{"today", "24h", "7d"}, Description: "preset time range"},
		{Name: "--start", Type: "datetime", Required: false, Description: "custom start time, format 2006-01-02 15:04:05"},
		{Name: "--end", Type: "datetime", Required: false, Description: "custom end time, format 2006-01-02 15:04:05"},
		{Name: "--page", Type: "integer", Required: false, Default: "1", Description: "page number, starts from 1"},
		{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "page size, 1-100"},
	}
}

func policyDetectionWhitelistManifestFlags() []ManifestFlag {
	return []ManifestFlag{
		{Name: "--page", Type: "integer", Required: false, Default: "1", Description: "page number, starts from 1"},
		{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "page size, 1-100"},
		{Name: "--name", Type: "string", Required: false, Description: "detection whitelist name filter"},
		{Name: "--src-ip", Type: "csv_string", Required: false, Description: "source IP filter"},
		{Name: "--src-port", Type: "csv_string", Required: false, Description: "source port filter"},
		{Name: "--dest-ip", Type: "csv_string", Required: false, Description: "destination IP filter"},
		{Name: "--dest-port", Type: "csv_string", Required: false, Description: "destination port filter"},
		{Name: "--domain", Type: "csv_string", Required: false, Description: "domain filter"},
		{Name: "--url-path", Type: "csv_string", Required: false, Description: "URL path filter"},
		{Name: "--user-agent", Type: "csv_string", Required: false, Description: "User-Agent filter"},
		{Name: "--xff", Type: "csv_string", Required: false, Description: "XFF filter"},
		{Name: "--resp-code", Type: "csv_string", Required: false, Description: "response status code filter"},
		{Name: "--resp-body", Type: "csv_string", Required: false, Description: "response body filter"},
		{Name: "--threat", Type: "string", Required: false, Description: "threat type filter"},
		{Name: "--rule-id", Type: "string", Required: false, Description: "detection rule id filter"},
		{Name: "--status", Type: "enum", Required: false, Enum: []string{"enabled", "disabled", "1", "0"}, Description: "rule status filter"},
	}
}

func policyCustomIntelligenceManifestFlags() []ManifestFlag {
	return []ManifestFlag{
		{Name: "--page", Type: "integer", Required: false, Default: "1", Description: "page number, starts from 1"},
		{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "page size, 1-100"},
		{Name: "--id", Type: "integer", Required: false, Description: "custom intelligence id filter"},
		{Name: "--name", Type: "string", Required: false, Description: "custom intelligence name filter"},
		{Name: "--ioc", Type: "csv_string", Required: false, Description: "IOC filter"},
		{Name: "--type", Type: "enum", Required: false, Enum: []string{"ip", "domain", "url", "md5", "sha1", "sha256", "1", "2", "3", "4", "5", "6"}, Description: "IOC type filter"},
		{Name: "--status", Type: "enum", Required: false, Enum: []string{"enabled", "disabled", "1", "0"}, Description: "custom intelligence status filter"},
		{Name: "--remarks", Type: "string", Required: false, Description: "remarks filter"},
	}
}

func metadataListManifestFlags(search bool) []ManifestFlag {
	flags := []ManifestFlag{
		{Name: "--time", Type: "string", Required: false, Default: "today", Enum: []string{"today", "24h", "7d"}, Description: "preset time range"},
		{Name: "--start", Type: "datetime", Required: false, Description: "custom start time, format 2006-01-02 15:04:05"},
		{Name: "--end", Type: "datetime", Required: false, Description: "custom end time, format 2006-01-02 15:04:05"},
		{Name: "--page", Type: "integer", Required: false, Default: "1", Description: "page number, starts from 1"},
		{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "page size, 1-100"},
		{Name: "--protocol", Type: "enum", Required: true, Enum: []string{"http", "http2", "dns", "tcp", "udp", "other"}, Description: "metadata protocol or event_type"},
		{Name: "--src-ip", Type: "csv_string", Required: false, Description: "source IP filter"},
		{Name: "--dest-ip", Type: "csv_string", Required: false, Description: "destination IP filter"},
		{Name: "--src-port", Type: "csv_integer", Required: false, Description: "source port filter"},
		{Name: "--dest-port", Type: "csv_integer", Required: false, Description: "destination port filter"},
		{Name: "--http-url", Type: "csv_string", Required: false, Description: "HTTP URL filter"},
		{Name: "--dns-rrname", Type: "csv_string", Required: false, Description: "DNS rrname filter"},
	}
	if search {
		flags = append(flags, ManifestFlag{Name: "--advanced-query", Type: "string", Required: true, Description: "advanced query expression"})
	} else {
		flags = append(flags, ManifestFlag{Name: "--advanced-query", Type: "string", Required: false, Description: "advanced query expression"})
	}
	return flags
}

func assetListManifestFlags() []ManifestFlag {
	return []ManifestFlag{
		{Name: "--page", Type: "integer", Required: false, Default: "1", Description: "page number, starts from 1"},
		{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "page size, 1-100"},
		{Name: "--id", Type: "integer", Required: false, Description: "asset id"},
		{Name: "--name", Type: "string", Required: false, Description: "asset name fuzzy filter"},
		{Name: "--ip", Type: "string", Required: false, Description: "asset IP, CIDR, or range fuzzy filter"},
		{Name: "--mac", Type: "string", Required: false, Description: "asset MAC fuzzy filter"},
		{Name: "--asset-type", Type: "string", Required: false, Description: "asset type fuzzy filter"},
		{Name: "--importance", Type: "enum", Required: false, Enum: []string{"important", "normal", "重点", "普通", "1", "2"}, Description: "asset importance"},
		{Name: "--tag-id", Type: "csv_integer", Required: false, Description: "asset tag id filter"},
		{Name: "--group-id", Type: "integer", Required: false, Description: "asset group id filter"},
	}
}

func fileAlarmOverviewManifestFlags() []ManifestFlag {
	return []ManifestFlag{
		{Name: "--time", Type: "string", Required: false, Default: "today", Enum: []string{"today", "24h", "7d"}, Description: "preset time range"},
		{Name: "--start", Type: "datetime", Required: false, Description: "custom start time, format 2006-01-02 15:04:05"},
		{Name: "--end", Type: "datetime", Required: false, Description: "custom end time, format 2006-01-02 15:04:05"},
		{Name: "--severity", Type: "csv_enum", Required: false, Enum: []string{"critical", "high", "medium", "low", "1", "2", "3", "4"}, Description: "severity filter"},
		{Name: "--file-type", Type: "csv_string", Required: false, Description: "file detection type filter, for example virus,backdoor,php_webshell"},
		{Name: "--keyword", Type: "csv_string", Required: false, Description: "keyword filter"},
	}
}

func fileAlarmListManifestFlags(tagDefault string) []ManifestFlag {
	flags := fileAlarmOverviewManifestFlags()
	flags = append(flags,
		ManifestFlag{Name: "--page", Type: "integer", Required: false, Default: "1", Description: "page number, starts from 1"},
		ManifestFlag{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "page size, 1-100"},
		ManifestFlag{Name: "--tag", Type: "csv_enum", Required: false, Default: tagDefault, Enum: []string{"elf", "webshell"}, Description: "file alarm tag filter"},
		ManifestFlag{Name: "--src-ip", Type: "csv_string", Required: false, Description: "source IP filter"},
		ManifestFlag{Name: "--dest-ip", Type: "csv_string", Required: false, Description: "destination IP filter"},
		ManifestFlag{Name: "--src-port", Type: "csv_integer", Required: false, Description: "source port filter"},
		ManifestFlag{Name: "--dest-port", Type: "csv_integer", Required: false, Description: "destination port filter"},
		ManifestFlag{Name: "--app-proto", Type: "csv_string", Required: false, Description: "application protocol filter"},
	)
	return flags
}

func alarmSubjectManifestFlags(subjectFlag string) []ManifestFlag {
	flags := []ManifestFlag{
		{Name: "--time", Type: "string", Required: false, Default: "today", Enum: []string{"today", "24h", "7d"}, Description: "preset time range"},
		{Name: "--start", Type: "datetime", Required: false, Description: "custom start time, format 2006-01-02 15:04:05"},
		{Name: "--end", Type: "datetime", Required: false, Description: "custom end time, format 2006-01-02 15:04:05"},
		{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "number of alarm summaries to return, 1-100"},
		{Name: "--severity", Type: "csv_enum", Required: false, Enum: []string{"critical", "high", "medium", "low", "1", "2", "3", "4"}, Description: "severity filter"},
		{Name: "--result", Type: "csv_enum", Required: false, Enum: []string{"success", "control", "failed", "unknown"}, Description: "attack result filter"},
		{Name: "--direction", Type: "csv_enum", Required: false, Enum: []string{"in", "lateral", "out", "other"}, Description: "traffic direction filter"},
	}
	switch subjectFlag {
	case "--attacker":
		flags = append(flags, ManifestFlag{Name: "--attacker", Type: "string", Required: true, Description: "attacker IP"})
	case "--victim":
		flags = append(flags, ManifestFlag{Name: "--victim", Type: "string", Required: true, Description: "victim IP or object"})
	}
	return flags
}

func alarmThreatManifestFlags() []ManifestFlag {
	flags := alarmSubjectManifestFlags("")
	flags = append(flags,
		ManifestFlag{Name: "--name", Type: "string", Required: false, Description: "threat name filter"},
		ManifestFlag{Name: "--tag", Type: "string", Required: false, Description: "threat type/tag filter"},
		ManifestFlag{Name: "--phase", Type: "csv_enum", Required: false, Enum: []string{"recon", "intrustion", "delivery", "control", "lateral", "goal", "other"}, Description: "attack phase filter"},
	)
	return flags
}

func alarmRankManifestFlags() []ManifestFlag {
	return []ManifestFlag{
		{Name: "--time", Type: "string", Required: false, Default: "today", Enum: []string{"today", "24h", "7d"}, Description: "preset time range"},
		{Name: "--start", Type: "datetime", Required: false, Description: "custom start time, format 2006-01-02 15:04:05"},
		{Name: "--end", Type: "datetime", Required: false, Description: "custom end time, format 2006-01-02 15:04:05"},
		{Name: "--top", Type: "integer", Required: false, Default: "10", Description: "ranking size, 1-100"},
		{Name: "--severity", Type: "csv_enum", Required: false, Enum: []string{"critical", "high", "medium", "low", "1", "2", "3", "4"}, Description: "severity filter"},
		{Name: "--result", Type: "csv_enum", Required: false, Enum: []string{"success", "control", "failed", "unknown"}, Description: "attack result filter"},
		{Name: "--phase", Type: "csv_enum", Required: false, Enum: []string{"recon", "intrustion", "delivery", "control", "lateral", "goal", "other"}, Description: "attack phase filter"},
	}
}

func alarmRelatedManifestFlags() []ManifestFlag {
	return []ManifestFlag{
		{Name: "--id", Type: "string", Required: true, Description: "source alarm doc_id returned by alarm list or high-priority"},
		{Name: "--window", Type: "duration", Required: false, Default: "30m", Description: "time window before and after the source alarm, for example 10m, 30m, 1h; max 24h"},
		{Name: "--relation", Type: "enum", Required: false, Default: "both", Enum: []string{"both", "attacker", "victim"}, Description: "related alarm scope"},
		{Name: "--limit", Type: "integer", Required: false, Default: "20", Description: "maximum related alarms to return after de-duplication, 1-100"},
	}
}

func alarmTimelineManifestFlags() []ManifestFlag {
	flags := []ManifestFlag{
		{Name: "--time", Type: "string", Required: false, Default: "today", Enum: []string{"today", "24h", "7d"}, Description: "preset time range"},
		{Name: "--start", Type: "datetime", Required: false, Description: "custom start time, format 2006-01-02 15:04:05"},
		{Name: "--end", Type: "datetime", Required: false, Description: "custom end time, format 2006-01-02 15:04:05"},
		{Name: "--interval", Type: "duration", Required: false, Description: "chart interval, for example 5m, 1h, 24h; empty means backend auto"},
	}
	for _, flag := range alarmListManifestFlags(false) {
		switch flag.Name {
		case "--time", "--start", "--end", "--page", "--page-size":
			continue
		default:
			flags = append(flags, flag)
		}
	}
	return flags
}

func alarmListManifestFlags(highPriority bool) []ManifestFlag {
	severityDefault := ""
	resultDefault := ""
	if highPriority {
		severityDefault = "critical,high"
		resultDefault = "success,control"
	}
	return []ManifestFlag{
		{Name: "--time", Type: "string", Required: false, Default: "today", Enum: []string{"today", "24h", "7d"}, Description: "preset time range"},
		{Name: "--start", Type: "datetime", Required: false, Description: "custom start time, format 2006-01-02 15:04:05"},
		{Name: "--end", Type: "datetime", Required: false, Description: "custom end time, format 2006-01-02 15:04:05"},
		{Name: "--page", Type: "integer", Required: false, Default: "1", Description: "page number, starts from 1"},
		{Name: "--page-size", Type: "integer", Required: false, Default: "10", Description: "page size, 1-100"},
		{Name: "--severity", Type: "csv_enum", Required: false, Default: severityDefault, Enum: []string{"critical", "high", "medium", "low", "1", "2", "3", "4"}, Description: "severity filter"},
		{Name: "--result", Type: "csv_enum", Required: false, Default: resultDefault, Enum: []string{"success", "control", "failed", "unknown"}, Description: "attack result filter"},
		{Name: "--phase", Type: "csv_enum", Required: false, Enum: []string{"recon", "intrustion", "delivery", "control", "lateral", "goal", "other"}, Description: "attack phase filter"},
		{Name: "--attacker", Type: "csv_string", Required: false, Description: "attacker IP filter"},
		{Name: "--victim", Type: "csv_string", Required: false, Description: "victim IP filter"},
		{Name: "--asset-ip", Type: "csv_string", Required: false, Description: "asset IP filter matching attacker or victim"},
		{Name: "--keyword", Type: "csv_string", Required: false, Description: "keyword filter"},
		{Name: "--name", Type: "csv_string", Required: false, Description: "threat name or attack summary filter"},
		{Name: "--tag", Type: "csv_string", Required: false, Description: "threat type/tag filter"},
		{Name: "--direction", Type: "csv_enum", Required: false, Enum: []string{"in", "lateral", "out", "other"}, Description: "traffic direction filter"},
		{Name: "--app-proto", Type: "csv_string", Required: false, Description: "application protocol filter"},
		{Name: "--http-url", Type: "csv_string", Required: false, Description: "HTTP URL filter"},
		{Name: "--host", Type: "csv_string", Required: false, Description: "Host/hostname filter"},
		{Name: "--xff", Type: "csv_string", Required: false, Description: "XFF filter"},
		{Name: "--src-port", Type: "csv_integer", Required: false, Description: "source port filter"},
		{Name: "--dest-port", Type: "csv_integer", Required: false, Description: "destination port filter"},
	}
}
