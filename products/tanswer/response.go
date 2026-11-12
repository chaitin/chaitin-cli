package tanswer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const responseBlockPolicyCreateConfirmToken = "CONFIRM_RESPONSE_BLOCK_POLICY_CREATE"
const responseBlockPolicyUpdateConfirmToken = "CONFIRM_RESPONSE_BLOCK_POLICY_UPDATE"
const responseBlockPolicyEnableConfirmToken = "CONFIRM_RESPONSE_BLOCK_POLICY_ENABLE"
const responseBlockPolicyDisableConfirmToken = "CONFIRM_RESPONSE_BLOCK_POLICY_DISABLE"
const responseBlockPolicyDeleteConfirmToken = "CONFIRM_RESPONSE_BLOCK_POLICY_DELETE"
const responseWhitelistCreateConfirmToken = "CONFIRM_RESPONSE_WHITELIST_CREATE"
const responseWhitelistUpdateConfirmToken = "CONFIRM_RESPONSE_WHITELIST_UPDATE"
const responseWhitelistEnableConfirmToken = "CONFIRM_RESPONSE_WHITELIST_ENABLE"
const responseWhitelistDisableConfirmToken = "CONFIRM_RESPONSE_WHITELIST_DISABLE"
const responseWhitelistDeleteConfirmToken = "CONFIRM_RESPONSE_WHITELIST_DELETE"
const responseBlockPolicyFromAlarmConfirmToken = "CONFIRM_RESPONSE_BLOCK_POLICY_FROM_ALARM"
const responseWhitelistFromAlarmConfirmToken = "CONFIRM_RESPONSE_WHITELIST_FROM_ALARM"

type responseBlockPoliciesOptions struct {
	page       int
	pageSize   int
	id         string
	strategyID string
	name       string
	object     string
	status     string
}

type responseBlockRecordsOptions struct {
	time       string
	start      string
	end        string
	page       int
	pageSize   int
	srcIP      string
	srcPort    string
	destIP     string
	destPort   string
	policyID   string
	policyName string
	strategyID string
	blockType  string
}

type responseWhitelistOptions struct {
	time        string
	start       string
	end         string
	page        int
	pageSize    int
	objectType  string
	object      string
	status      string
	expireAfter string
	blockMethod string
	ipType      string
	remark      string
}

type responseDevicesOptions struct {
	page       int
	pageSize   int
	id         string
	deviceType string
	status     string
	remark     string
}

type responseDeviceRecordsOptions struct {
	page     int
	pageSize int
	deviceID string
}

type responseAutoPoliciesOptions struct {
	time       string
	start      string
	end        string
	page       int
	pageSize   int
	id         string
	name       string
	deviceID   string
	status     string
	punishType string
}

type responseAutoListOptions struct {
	time          string
	start         string
	end           string
	page          int
	pageSize      int
	ip            string
	status        string
	strategyID    string
	blockTimeType string
}

type responseBlockPolicyWriteOptions struct {
	id            string
	idList        string
	name          string
	objects       string
	objectType    string
	ipType        string
	status        string
	blockTimeType string
	duration      string
	expire        string
	remark        string
	preview       bool
	confirm       string
}

type responseWhitelistWriteOptions struct {
	id          string
	idList      string
	objectType  string
	objects     string
	status      string
	expire      string
	blockMethod string
	ipType      string
	remark      string
	preview     bool
	confirm     string
}

type responseFromAlarmOptions struct {
	id     string
	target string
	block  responseBlockPolicyWriteOptions
	white  responseWhitelistWriteOptions
}

type responseListRPCResult struct {
	Data  []map[string]any `json:"data"`
	Total int64            `json:"total"`
}

type responseBlockRecordsRPCResult struct {
	Records []map[string]any `json:"records"`
	Total   int64            `json:"total"`
}

func newResponseCommand(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "response",
		Short: "响应处置语义命令",
		Long:  "响应处置语义命令。用于只读查询旁路阻断策略、阻断记录、响应白名单、联动设备、设备处置记录、自动响应策略和自动响应处置名单；不创建、启停、删除或下发处置。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newResponseBlockPoliciesCommand(opts))
	cmd.AddCommand(newResponseBlockPolicyCreateCommand(opts))
	cmd.AddCommand(newResponseBlockPolicyUpdateCommand(opts))
	cmd.AddCommand(newResponseBlockPolicyEnableCommand(opts))
	cmd.AddCommand(newResponseBlockPolicyDisableCommand(opts))
	cmd.AddCommand(newResponseBlockPolicyDeleteCommand(opts))
	cmd.AddCommand(newResponseBlockRecordsCommand(opts))
	cmd.AddCommand(newResponseWhitelistCommand(opts))
	cmd.AddCommand(newResponseWhitelistCreateCommand(opts))
	cmd.AddCommand(newResponseWhitelistUpdateCommand(opts))
	cmd.AddCommand(newResponseWhitelistEnableCommand(opts))
	cmd.AddCommand(newResponseWhitelistDisableCommand(opts))
	cmd.AddCommand(newResponseWhitelistDeleteCommand(opts))
	cmd.AddCommand(newResponseBlockPolicyFromAlarmCommand(opts))
	cmd.AddCommand(newResponseWhitelistFromAlarmCommand(opts))
	cmd.AddCommand(newResponseDevicesCommand(opts))
	cmd.AddCommand(newResponseDeviceRecordsCommand(opts))
	cmd.AddCommand(newResponseAutoPoliciesCommand(opts))
	cmd.AddCommand(newResponseAutoListCommand(opts))
	return cmd
}

func newResponseBlockPoliciesCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseBlockPoliciesOptions
	cmd := &cobra.Command{
		Use:   "block-policies",
		Short: "查询旁路阻断策略",
		Long:  "查询旁路阻断策略，用于查看当前阻断策略配置、阻断对象、关联自动响应策略、失效时间和状态。该命令只读取策略列表，不启停、删除或下发阻断。\n\n输出：实际筛选条件、total、page、page_size、current_count、has_more、block_policies。",
		Example: "  chaitin-cli tanswer response block-policies --page-size 10\n" +
			"  chaitin-cli tanswer response block-policies --object 198.51.100.10 --status enabled",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResponseDataListCommand(cmd, opts, "查询旁路阻断策略", "chaitin-cli tanswer response block-policies", "RulesService.SearchBlockRules", buildResponseBlockPoliciesRequest(responseOpts), responseBlockPoliciesFilters(responseOpts), responseOpts.page, responseOpts.pageSize, "block_policies", summarizeResponseBlockPolicies)
		},
	}
	addResponseBlockPoliciesFlags(cmd, &responseOpts)
	return cmd
}

func newResponseBlockPolicyCreateCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseBlockPolicyWriteOptions
	commandText := "chaitin-cli tanswer response block-policy-create"
	cmd := &cobra.Command{
		Use:   "block-policy-create",
		Short: "新增旁路阻断策略",
		Long: "新增旁路阻断策略，用于对确认恶意对象建立旁路阻断策略。该命令是高影响写操作：默认只返回写入预览，必须使用 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_CREATE 才会调用后端创建接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer response block-policy-create --name block-bad-ip --object 198.51.100.10 --preview\n" +
			"  chaitin-cli tanswer response block-policy-create --name block-bad-ip --object 198.51.100.10 --duration 3600 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_CREATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildResponseBlockPolicyWriteRequest(responseOpts)
			if err != nil {
				return writeResponseError(cmd, "新增旁路阻断策略", commandText, "INVALID_RESPONSE_BLOCK_POLICY_CREATE_REQUEST", err.Error(), false)
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "新增旁路阻断策略预览",
				Command:       commandText,
				OperationType: "response_block_policy_create",
				RiskLevel:     "write_high",
				Target:        map[string]any{"name": req["name"], "ips": req["ips"], "block_type": req["block_type"]},
				ChangeSummary: map[string]any{"before": nil, "after": req},
				Impact:        map[string]any{"block_policy_count": 1},
				RiskWarnings:  []string{"将新增旁路阻断策略，匹配对象的后续流量可能被阻断。", "执行前应确认阻断对象、时长和方向没有过宽。"},
				ConfirmToken:  responseBlockPolicyCreateConfirmToken,
			})
			if responseOpts.preview || strings.TrimSpace(responseOpts.confirm) == "" {
				return writeResponseSuccess(cmd, "新增旁路阻断策略预览", commandText, req, preview)
			}
			if err := ValidateWriteConfirmation(responseOpts.confirm, responseBlockPolicyCreateConfirmToken); err != nil {
				return writeResponseError(cmd, "新增旁路阻断策略", commandText, "RESPONSE_BLOCK_POLICY_CREATE_CONFIRMATION_REQUIRED", err.Error(), false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var result struct {
				ID uint `json:"id"`
				Id uint `json:"Id"`
			}
			if err := client.CallRPC(cmd.Context(), "RulesService.CreateBlockRules", req, &result); err != nil {
				return writeResponseError(cmd, "新增旁路阻断策略", commandText, "RESPONSE_BLOCK_POLICY_CREATE_FAILED", err.Error(), true)
			}
			id := result.ID
			if id == 0 {
				id = result.Id
			}
			object := map[string]any{"id": id, "name": req["name"], "ips": req["ips"], "status": req["status"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "response_block_policy_create", Object: object, Action: "create", Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: map[string]any{"before": nil, "after": req}, Result: "success"})
			return writeResponseSuccess(cmd, "新增旁路阻断策略", commandText, req, data)
		},
	}
	addResponseBlockPolicyWriteFlags(cmd, &responseOpts, false)
	return cmd
}

func newResponseBlockPolicyUpdateCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseBlockPolicyWriteOptions
	commandText := "chaitin-cli tanswer response block-policy-update"
	cmd := &cobra.Command{
		Use:   "block-policy-update",
		Short: "编辑旁路阻断策略",
		Long:  "编辑旁路阻断策略，用于更新单条旁路阻断策略。该命令是高影响写操作：预览阶段会读取当前策略并返回 before/after；必须使用 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_UPDATE 才会调用后端更新接口。\n\n输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer response block-policy-update --id 7 --name new-block --object 198.51.100.10 --preview\n" +
			"  chaitin-cli tanswer response block-policy-update --id 7 --name new-block --object 198.51.100.10 --expire 1784277612410 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_UPDATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildResponseBlockPolicyUpdateRequest(responseOpts)
			if err != nil {
				return writeResponseError(cmd, "编辑旁路阻断策略", commandText, "INVALID_RESPONSE_BLOCK_POLICY_UPDATE_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(responseOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(responseOpts.confirm, responseBlockPolicyUpdateConfirmToken); err != nil {
					return writeResponseError(cmd, "编辑旁路阻断策略", commandText, "RESPONSE_BLOCK_POLICY_UPDATE_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			before, err := fetchResponseBlockPoliciesByIDs(cmd, client, []uint{req["id"].(uint)})
			if err != nil {
				return writeResponseError(cmd, "编辑旁路阻断策略", commandText, "RESPONSE_BLOCK_POLICY_UPDATE_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{"before": before, "after": req}
			preview := BuildWritePreview(WriteOperationSpec{Task: "编辑旁路阻断策略预览", Command: commandText, OperationType: "response_block_policy_update", RiskLevel: "write_high", Target: map[string]any{"id": req["id"]}, ChangeSummary: changeSummary, Impact: map[string]any{"block_policy_count": 1}, RiskWarnings: []string{"将更新旁路阻断策略，阻断对象或有效期变化可能影响处置结果。"}, ConfirmToken: responseBlockPolicyUpdateConfirmToken})
			if responseOpts.preview || strings.TrimSpace(responseOpts.confirm) == "" {
				return writeResponseSuccess(cmd, "编辑旁路阻断策略预览", commandText, req, preview)
			}
			var result struct {
				ID uint `json:"id"`
			}
			if err := client.CallRPC(cmd.Context(), "RulesService.UpdateBlockRules", req, &result); err != nil {
				return writeResponseError(cmd, "编辑旁路阻断策略", commandText, "RESPONSE_BLOCK_POLICY_UPDATE_FAILED", err.Error(), true)
			}
			object := map[string]any{"id": req["id"], "name": req["name"], "ips": req["ips"], "status": req["status"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "response_block_policy_update", Object: object, Action: "update", Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: changeSummary, Result: "success"})
			return writeResponseSuccess(cmd, "编辑旁路阻断策略", commandText, req, data)
		},
	}
	addResponseBlockPolicyWriteFlags(cmd, &responseOpts, true)
	return cmd
}

func newResponseBlockPolicyEnableCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseBlockPolicyWriteOptions
	return newResponseBlockPolicyActionCommand(opts, &responseOpts, "block-policy-enable", "启用旁路阻断策略", "enable", "show", responseBlockPolicyEnableConfirmToken)
}

func newResponseBlockPolicyDisableCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseBlockPolicyWriteOptions
	return newResponseBlockPolicyActionCommand(opts, &responseOpts, "block-policy-disable", "停用旁路阻断策略", "disable", "hide", responseBlockPolicyDisableConfirmToken)
}

func newResponseBlockPolicyDeleteCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseBlockPolicyWriteOptions
	return newResponseBlockPolicyActionCommand(opts, &responseOpts, "block-policy-delete", "删除旁路阻断策略", "delete", "delete", responseBlockPolicyDeleteConfirmToken)
}

func newResponseBlockPolicyActionCommand(opts *RootOptions, responseOpts *responseBlockPolicyWriteOptions, use string, task string, action string, backendAction string, confirmToken string) *cobra.Command {
	commandText := "chaitin-cli tanswer response " + use
	cmd := &cobra.Command{
		Use:     use,
		Short:   task,
		Long:    fmt.Sprintf("%s，用于控制旁路阻断策略状态或删除策略。该命令是高影响写操作：预览阶段会读取目标策略摘要；必须使用 --confirm %s 才会调用后端批量操作接口。\n\n输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n执行输出：confirmed、result、object、audit。", task, confirmToken),
		Example: fmt.Sprintf("  %s --id-list 7,8 --preview\n  %s --id-list 7,8 --confirm %s", commandText, commandText, confirmToken),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseUintCSV(responseOpts.idList)
			if err != nil || len(ids) == 0 {
				return writeResponseError(cmd, task, commandText, "INVALID_RESPONSE_BLOCK_POLICY_ACTION_REQUEST", "missing or invalid id-list: set --id-list", false)
			}
			if strings.TrimSpace(responseOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(responseOpts.confirm, confirmToken); err != nil {
					return writeResponseError(cmd, task, commandText, "RESPONSE_BLOCK_POLICY_"+strings.ToUpper(action)+"_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			before, err := fetchResponseBlockPoliciesByIDs(cmd, client, ids)
			if err != nil {
				return writeResponseError(cmd, task, commandText, "RESPONSE_BLOCK_POLICY_ACTION_PREVIEW_FAILED", err.Error(), true)
			}
			req := map[string]any{"ids": ids, "action": backendAction}
			changeSummary := map[string]any{"before": before, "after": req}
			if backendAction == "delete" {
				changeSummary["after"] = nil
			}
			preview := BuildWritePreview(WriteOperationSpec{Task: task + "预览", Command: commandText, OperationType: "response_block_policy_" + action, RiskLevel: "write_high", Target: map[string]any{"ids": ids}, ChangeSummary: changeSummary, Impact: map[string]any{"block_policy_count": len(ids), "backend_action": backendAction}, RiskWarnings: []string{"将改变旁路阻断策略状态或删除策略，可能影响当前处置结果。"}, ConfirmToken: confirmToken})
			if responseOpts.preview || strings.TrimSpace(responseOpts.confirm) == "" {
				return writeResponseSuccess(cmd, task+"预览", commandText, req, preview)
			}
			var result struct {
				IDs []uint `json:"ids"`
			}
			if err := client.CallRPC(cmd.Context(), "RulesService.UpdateBlockRulesStatus", req, &result); err != nil {
				return writeResponseError(cmd, task, commandText, "RESPONSE_BLOCK_POLICY_"+strings.ToUpper(action)+"_FAILED", err.Error(), true)
			}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "response_block_policy_" + action, Object: map[string]any{"ids": ids, "action": backendAction}, Action: action, Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: changeSummary, Result: "success"})
			return writeResponseSuccess(cmd, task, commandText, req, data)
		},
	}
	cmd.Flags().StringVar(&responseOpts.idList, "id-list", "", "block policy IDs, comma separated, required")
	cmd.Flags().BoolVar(&responseOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&responseOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newResponseBlockRecordsCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseBlockRecordsOptions
	cmd := &cobra.Command{
		Use:   "block-records",
		Short: "查询旁路阻断记录",
		Long:  "查询旁路阻断记录，用于确认产品记录中的阻断命中情况。该命令只读取已有记录，不验证网络层真实阻断效果。\n\n输出：查询时间范围、实际筛选条件、total、page、page_size、current_count、has_more、block_records。",
		Example: "  chaitin-cli tanswer response block-records --time 24h --page-size 10\n" +
			"  chaitin-cli tanswer response block-records --src-ip 198.51.100.10 --dest-ip 192.0.2.10",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateResponsePage(responseOpts.page, responseOpts.pageSize); err != nil {
				return writeResponseError(cmd, "查询旁路阻断记录", "chaitin-cli tanswer response block-records", "INVALID_PAGE", err.Error(), false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			rng, err := ParseTimeRange(TimeRangeOptions{Time: responseOpts.time, Start: responseOpts.start, End: responseOpts.end})
			if err != nil {
				return writeResponseError(cmd, "查询旁路阻断记录", "chaitin-cli tanswer response block-records", "INVALID_TIME_RANGE", err.Error(), false)
			}
			req := buildResponseBlockRecordsRequest(rng, responseOpts)
			client := NewClient(cfg)
			var result responseBlockRecordsRPCResult
			if err := client.CallRPC(cmd.Context(), "RulesService.SearchTapBlockRecordList", req, &result); err != nil {
				return writeResponseError(cmd, "查询旁路阻断记录", "chaitin-cli tanswer response block-records", "RESPONSE_BLOCK_RECORDS_FAILED", err.Error(), true)
			}
			return writeResponseSuccess(cmd, "查询旁路阻断记录", "chaitin-cli tanswer response block-records", map[string]any{"time_range": rng, "filters": responseBlockRecordsFilters(responseOpts), "page": responseOpts.page, "page_size": responseOpts.pageSize}, responseListData(result.Total, responseOpts.page, responseOpts.pageSize, len(result.Records), "block_records", summarizeResponseBlockRecords(result.Records)))
		},
	}
	addResponseBlockRecordsFlags(cmd, &responseOpts)
	return cmd
}

func newResponseWhitelistCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseWhitelistOptions
	cmd := &cobra.Command{
		Use:   "whitelist",
		Short: "查询响应白名单",
		Long:  "查询响应白名单，用于查看不会被响应处置影响的 IP 或 URL 对象、阻断方式、有效期和状态。该命令只读取白名单，不新增、编辑、启停或删除。\n\n输出：查询时间范围、实际筛选条件、total、page、page_size、current_count、has_more、response_whitelists。",
		Example: "  chaitin-cli tanswer response whitelist --page-size 10\n" +
			"  chaitin-cli tanswer response whitelist --object 198.51.100.10 --type ip --status enabled",
		RunE: func(cmd *cobra.Command, args []string) error {
			rng, req, err := buildResponseWhitelistTimedRequest(responseOpts)
			if err != nil {
				return writeResponseError(cmd, "查询响应白名单", "chaitin-cli tanswer response whitelist", "INVALID_TIME_RANGE", err.Error(), false)
			}
			query := map[string]any{"time_range": rng, "filters": responseWhitelistFilters(responseOpts), "page": responseOpts.page, "page_size": responseOpts.pageSize}
			return runResponseDataListCommand(cmd, opts, "查询响应白名单", "chaitin-cli tanswer response whitelist", "FirewallService.SearchWhiteList", req, query, responseOpts.page, responseOpts.pageSize, "response_whitelists", summarizeResponseWhitelists)
		},
	}
	addResponseWhitelistFlags(cmd, &responseOpts)
	return cmd
}

func newResponseWhitelistCreateCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseWhitelistWriteOptions
	commandText := "chaitin-cli tanswer response whitelist-create"
	cmd := &cobra.Command{
		Use:   "whitelist-create",
		Short: "新增响应白名单",
		Long: "新增响应白名单，用于将确认可信对象加入响应处置白名单，避免被旁路阻断或第三方协同阻断影响。该命令是高影响写操作：默认只返回写入预览，必须使用 --confirm CONFIRM_RESPONSE_WHITELIST_CREATE 才会调用后端创建接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer response whitelist-create --type ip --object 198.51.100.10 --expire 1784277612410 --preview\n" +
			"  chaitin-cli tanswer response whitelist-create --type ip --object 198.51.100.10 --expire 1784277612410 --confirm CONFIRM_RESPONSE_WHITELIST_CREATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildResponseWhitelistWriteRequest(responseOpts)
			if err != nil {
				return writeResponseError(cmd, "新增响应白名单", commandText, "INVALID_RESPONSE_WHITELIST_CREATE_REQUEST", err.Error(), false)
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "新增响应白名单预览",
				Command:       commandText,
				OperationType: "response_whitelist_create",
				RiskLevel:     "write_high",
				Target:        map[string]any{"type": req["type"], "values": req["values"]},
				ChangeSummary: map[string]any{"before": nil, "after": req},
				Impact:        map[string]any{"whitelist_count": len(req["values"].([]string))},
				RiskWarnings:  []string{"将新增响应白名单，匹配对象后续可能不会被响应处置阻断。", "执行前应确认白名单对象、方向、阻断方式和失效时间没有过宽。"},
				ConfirmToken:  responseWhitelistCreateConfirmToken,
			})
			if responseOpts.preview || strings.TrimSpace(responseOpts.confirm) == "" {
				return writeResponseSuccess(cmd, "新增响应白名单预览", commandText, req, preview)
			}
			if err := ValidateWriteConfirmation(responseOpts.confirm, responseWhitelistCreateConfirmToken); err != nil {
				return writeResponseError(cmd, "新增响应白名单", commandText, "RESPONSE_WHITELIST_CREATE_CONFIRMATION_REQUIRED", err.Error(), false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var result struct {
				ID uint `json:"id"`
				Id uint `json:"Id"`
			}
			if err := client.CallRPC(cmd.Context(), "FirewallService.CreateWhiteList", req, &result); err != nil {
				return writeResponseError(cmd, "新增响应白名单", commandText, "RESPONSE_WHITELIST_CREATE_FAILED", err.Error(), true)
			}
			id := result.ID
			if id == 0 {
				id = result.Id
			}
			object := map[string]any{"id": id, "type": req["type"], "values": req["values"], "status": req["status"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "response_whitelist_create", Object: object, Action: "create", Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: map[string]any{"before": nil, "after": req}, Result: "success"})
			return writeResponseSuccess(cmd, "新增响应白名单", commandText, req, data)
		},
	}
	addResponseWhitelistWriteFlags(cmd, &responseOpts, false)
	return cmd
}

func newResponseWhitelistUpdateCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseWhitelistWriteOptions
	commandText := "chaitin-cli tanswer response whitelist-update"
	cmd := &cobra.Command{
		Use:   "whitelist-update",
		Short: "编辑响应白名单",
		Long:  "编辑响应白名单，用于更新单条响应处置白名单。该命令是高影响写操作：预览阶段会读取当前白名单并返回 before/after；必须使用 --confirm CONFIRM_RESPONSE_WHITELIST_UPDATE 才会调用后端更新接口。\n\n输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer response whitelist-update --id 3 --type ip --object 198.51.100.10 --expire 1784277612410 --preview\n" +
			"  chaitin-cli tanswer response whitelist-update --id 3 --type url --object http://example.com/a --expire 1784277612410 --confirm CONFIRM_RESPONSE_WHITELIST_UPDATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildResponseWhitelistUpdateRequest(responseOpts)
			if err != nil {
				return writeResponseError(cmd, "编辑响应白名单", commandText, "INVALID_RESPONSE_WHITELIST_UPDATE_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(responseOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(responseOpts.confirm, responseWhitelistUpdateConfirmToken); err != nil {
					return writeResponseError(cmd, "编辑响应白名单", commandText, "RESPONSE_WHITELIST_UPDATE_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			before, err := fetchResponseWhitelistsByIDs(cmd, client, []uint{req["id"].(uint)})
			if err != nil {
				return writeResponseError(cmd, "编辑响应白名单", commandText, "RESPONSE_WHITELIST_UPDATE_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{"before": before, "after": req}
			preview := BuildWritePreview(WriteOperationSpec{Task: "编辑响应白名单预览", Command: commandText, OperationType: "response_whitelist_update", RiskLevel: "write_high", Target: map[string]any{"id": req["id"]}, ChangeSummary: changeSummary, Impact: map[string]any{"whitelist_count": 1}, RiskWarnings: []string{"将更新响应白名单，匹配对象、方向或有效期变化可能影响处置结果。"}, ConfirmToken: responseWhitelistUpdateConfirmToken})
			if responseOpts.preview || strings.TrimSpace(responseOpts.confirm) == "" {
				return writeResponseSuccess(cmd, "编辑响应白名单预览", commandText, req, preview)
			}
			var result struct{}
			if err := client.CallRPC(cmd.Context(), "FirewallService.UpdateWhiteList", req, &result); err != nil {
				return writeResponseError(cmd, "编辑响应白名单", commandText, "RESPONSE_WHITELIST_UPDATE_FAILED", err.Error(), true)
			}
			object := map[string]any{"id": req["id"], "type": req["type"], "values": req["values"], "status": req["status"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "response_whitelist_update", Object: object, Action: "update", Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: changeSummary, Result: "success"})
			return writeResponseSuccess(cmd, "编辑响应白名单", commandText, req, data)
		},
	}
	addResponseWhitelistWriteFlags(cmd, &responseOpts, true)
	return cmd
}

func newResponseWhitelistEnableCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseWhitelistWriteOptions
	return newResponseWhitelistActionCommand(opts, &responseOpts, "whitelist-enable", "启用响应白名单", "enable", 2, "FirewallService.UpdateWhiteListStatus", responseWhitelistEnableConfirmToken)
}

func newResponseWhitelistDisableCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseWhitelistWriteOptions
	return newResponseWhitelistActionCommand(opts, &responseOpts, "whitelist-disable", "停用响应白名单", "disable", 1, "FirewallService.UpdateWhiteListStatus", responseWhitelistDisableConfirmToken)
}

func newResponseWhitelistDeleteCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseWhitelistWriteOptions
	return newResponseWhitelistActionCommand(opts, &responseOpts, "whitelist-delete", "删除响应白名单", "delete", 0, "FirewallService.DeleteWhiteList", responseWhitelistDeleteConfirmToken)
}

func newResponseWhitelistActionCommand(opts *RootOptions, responseOpts *responseWhitelistWriteOptions, use string, task string, action string, status int, method string, confirmToken string) *cobra.Command {
	commandText := "chaitin-cli tanswer response " + use
	cmd := &cobra.Command{
		Use:     use,
		Short:   task,
		Long:    fmt.Sprintf("%s，用于控制响应白名单状态或删除白名单。该命令是高影响写操作：预览阶段会读取目标白名单摘要；必须使用 --confirm %s 才会调用后端批量操作接口。\n\n输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n执行输出：confirmed、result、object、audit。", task, confirmToken),
		Example: fmt.Sprintf("  %s --id-list 3,4 --preview\n  %s --id-list 3,4 --confirm %s", commandText, commandText, confirmToken),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseUintCSV(responseOpts.idList)
			if err != nil || len(ids) == 0 {
				return writeResponseError(cmd, task, commandText, "INVALID_RESPONSE_WHITELIST_ACTION_REQUEST", "missing or invalid id-list: set --id-list", false)
			}
			if strings.TrimSpace(responseOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(responseOpts.confirm, confirmToken); err != nil {
					return writeResponseError(cmd, task, commandText, "RESPONSE_WHITELIST_"+strings.ToUpper(action)+"_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			before, err := fetchResponseWhitelistsByIDs(cmd, client, ids)
			if err != nil {
				return writeResponseError(cmd, task, commandText, "RESPONSE_WHITELIST_ACTION_PREVIEW_FAILED", err.Error(), true)
			}
			req := map[string]any{"ids": ids}
			changeSummary := map[string]any{"before": before, "after": req}
			if status != 0 {
				req["status"] = status
			} else {
				changeSummary["after"] = nil
			}
			preview := BuildWritePreview(WriteOperationSpec{Task: task + "预览", Command: commandText, OperationType: "response_whitelist_" + action, RiskLevel: "write_high", Target: map[string]any{"ids": ids}, ChangeSummary: changeSummary, Impact: map[string]any{"whitelist_count": len(ids)}, RiskWarnings: []string{"将改变响应白名单状态或删除白名单，可能影响后续响应处置结果。"}, ConfirmToken: confirmToken})
			if responseOpts.preview || strings.TrimSpace(responseOpts.confirm) == "" {
				return writeResponseSuccess(cmd, task+"预览", commandText, req, preview)
			}
			var result struct{}
			if err := client.CallRPC(cmd.Context(), method, req, &result); err != nil {
				return writeResponseError(cmd, task, commandText, "RESPONSE_WHITELIST_"+strings.ToUpper(action)+"_FAILED", err.Error(), true)
			}
			object := map[string]any{"ids": ids, "action": action}
			if status != 0 {
				object["status"] = status
			}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "response_whitelist_" + action, Object: object, Action: action, Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: changeSummary, Result: "success"})
			return writeResponseSuccess(cmd, task, commandText, req, data)
		},
	}
	cmd.Flags().StringVar(&responseOpts.idList, "id-list", "", "response whitelist IDs, comma separated, required")
	cmd.Flags().BoolVar(&responseOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&responseOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newResponseBlockPolicyFromAlarmCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseFromAlarmOptions
	commandText := "chaitin-cli tanswer response block-policy-from-alarm"
	cmd := &cobra.Command{
		Use:   "block-policy-from-alarm",
		Short: "从告警生成旁路阻断策略",
		Long: "从告警生成旁路阻断策略，用于基于已研判恶意告警生成候选阻断策略。该命令是高影响写操作：预览阶段会读取告警详情并生成候选策略；必须使用 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_FROM_ALARM 才会调用后端创建接口。\n\n" +
			"输出预览：source_alarm、suggested_block_policy、requires_confirmation、confirmation_token、risk_warnings。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer response block-policy-from-alarm --id '<doc_id>' --target attacker --preview\n" +
			"  chaitin-cli tanswer response block-policy-from-alarm --id '<doc_id>' --target flow --duration 3600 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_FROM_ALARM",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(responseOpts.id) == "" {
				return writeResponseError(cmd, "从告警生成旁路阻断策略", commandText, "MISSING_ALARM_ID", "missing alarm doc_id: set --id", false)
			}
			if strings.TrimSpace(responseOpts.block.confirm) != "" {
				if err := ValidateWriteConfirmation(responseOpts.block.confirm, responseBlockPolicyFromAlarmConfirmToken); err != nil {
					return writeResponseError(cmd, "从告警生成旁路阻断策略", commandText, "RESPONSE_BLOCK_POLICY_FROM_ALARM_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var alarm alarmDetailRPCResult
			if err := client.CallRPC(cmd.Context(), "AlarmService.GetAlarm", map[string]any{"doc_id": strings.TrimSpace(responseOpts.id)}, &alarm); err != nil {
				return writeResponseError(cmd, "从告警生成旁路阻断策略", commandText, "RESPONSE_BLOCK_POLICY_FROM_ALARM_READ_FAILED", err.Error(), true)
			}
			if len(alarm.Data) == 0 {
				return writeResponseError(cmd, "从告警生成旁路阻断策略", commandText, "ALARM_NOT_FOUND", "alarm detail is empty", false)
			}
			req, err := buildResponseBlockPolicyFromAlarmRequest(alarm.Data, responseOpts)
			if err != nil {
				return writeResponseError(cmd, "从告警生成旁路阻断策略", commandText, "RESPONSE_BLOCK_POLICY_FROM_ALARM_FIELDS_INSUFFICIENT", err.Error(), false)
			}
			sourceAlarm := summarizeSourceAlarm(alarm.Data)
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "从告警生成旁路阻断策略预览",
				Command:       commandText,
				OperationType: "response_block_policy_from_alarm",
				RiskLevel:     "write_high",
				Target:        map[string]any{"source_alarm": sourceAlarm, "suggested_block_policy": map[string]any{"name": req["name"], "ips": req["ips"], "block_type": req["block_type"]}},
				ChangeSummary: map[string]any{"before": sourceAlarm, "after": req},
				Impact:        map[string]any{"block_policy_count": 1, "source_alarm_doc_id": strings.TrimSpace(responseOpts.id)},
				RiskWarnings: []string{
					"将基于单条告警创建旁路阻断策略，匹配对象后续流量可能被阻断。",
					"候选对象来自告警字段，执行前应确认攻击源、受害对象、端口和阻断时长没有过宽。",
				},
				ConfirmToken: responseBlockPolicyFromAlarmConfirmToken,
			})
			if responseOpts.block.preview || strings.TrimSpace(responseOpts.block.confirm) == "" {
				return writeResponseSuccess(cmd, "从告警生成旁路阻断策略预览", commandText, map[string]any{"doc_id": strings.TrimSpace(responseOpts.id), "target": responseFromAlarmTarget(responseOpts.target, "attacker")}, preview)
			}
			var result struct {
				ID uint `json:"id"`
				Id uint `json:"Id"`
			}
			if err := client.CallRPC(cmd.Context(), "RulesService.CreateBlockRules", req, &result); err != nil {
				return writeResponseError(cmd, "从告警生成旁路阻断策略", commandText, "RESPONSE_BLOCK_POLICY_FROM_ALARM_CREATE_FAILED", err.Error(), true)
			}
			id := result.ID
			if id == 0 {
				id = result.Id
			}
			object := map[string]any{"id": id, "source_alarm_doc_id": strings.TrimSpace(responseOpts.id), "name": req["name"], "ips": req["ips"], "status": req["status"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "response_block_policy_from_alarm", Object: object, Action: "create_from_alarm", Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: map[string]any{"before": sourceAlarm, "after": req}, Result: "success"})
			return writeResponseSuccess(cmd, "从告警生成旁路阻断策略", commandText, map[string]any{"doc_id": strings.TrimSpace(responseOpts.id), "target": responseFromAlarmTarget(responseOpts.target, "attacker")}, data)
		},
	}
	addResponseBlockPolicyFromAlarmFlags(cmd, &responseOpts)
	return cmd
}

func newResponseWhitelistFromAlarmCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseFromAlarmOptions
	commandText := "chaitin-cli tanswer response whitelist-from-alarm"
	cmd := &cobra.Command{
		Use:   "whitelist-from-alarm",
		Short: "从告警生成响应白名单",
		Long: "从告警生成响应白名单，用于基于已确认可信或误阻断风险的告警生成候选响应白名单。该命令是高影响写操作：预览阶段会读取告警详情并生成候选白名单；必须使用 --confirm CONFIRM_RESPONSE_WHITELIST_FROM_ALARM 才会调用后端创建接口。\n\n" +
			"输出预览：source_alarm、suggested_whitelist、requires_confirmation、confirmation_token、risk_warnings。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer response whitelist-from-alarm --id '<doc_id>' --target victim --expire 1784277612410 --preview\n" +
			"  chaitin-cli tanswer response whitelist-from-alarm --id '<doc_id>' --target url --expire 1784277612410 --confirm CONFIRM_RESPONSE_WHITELIST_FROM_ALARM",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(responseOpts.id) == "" {
				return writeResponseError(cmd, "从告警生成响应白名单", commandText, "MISSING_ALARM_ID", "missing alarm doc_id: set --id", false)
			}
			if strings.TrimSpace(responseOpts.white.confirm) != "" {
				if err := ValidateWriteConfirmation(responseOpts.white.confirm, responseWhitelistFromAlarmConfirmToken); err != nil {
					return writeResponseError(cmd, "从告警生成响应白名单", commandText, "RESPONSE_WHITELIST_FROM_ALARM_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var alarm alarmDetailRPCResult
			if err := client.CallRPC(cmd.Context(), "AlarmService.GetAlarm", map[string]any{"doc_id": strings.TrimSpace(responseOpts.id)}, &alarm); err != nil {
				return writeResponseError(cmd, "从告警生成响应白名单", commandText, "RESPONSE_WHITELIST_FROM_ALARM_READ_FAILED", err.Error(), true)
			}
			if len(alarm.Data) == 0 {
				return writeResponseError(cmd, "从告警生成响应白名单", commandText, "ALARM_NOT_FOUND", "alarm detail is empty", false)
			}
			req, err := buildResponseWhitelistFromAlarmRequest(alarm.Data, responseOpts)
			if err != nil {
				return writeResponseError(cmd, "从告警生成响应白名单", commandText, "RESPONSE_WHITELIST_FROM_ALARM_FIELDS_INSUFFICIENT", err.Error(), false)
			}
			sourceAlarm := summarizeSourceAlarm(alarm.Data)
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "从告警生成响应白名单预览",
				Command:       commandText,
				OperationType: "response_whitelist_from_alarm",
				RiskLevel:     "write_high",
				Target:        map[string]any{"source_alarm": sourceAlarm, "suggested_whitelist": map[string]any{"type": req["type"], "values": req["values"]}},
				ChangeSummary: map[string]any{"before": sourceAlarm, "after": req},
				Impact:        map[string]any{"response_whitelist_count": 1, "source_alarm_doc_id": strings.TrimSpace(responseOpts.id)},
				RiskWarnings: []string{
					"将基于单条告警创建响应白名单，匹配对象后续可能不会被响应处置阻断。",
					"候选对象来自告警字段，执行前应确认白名单对象、方向、阻断方式和失效时间没有过宽。",
				},
				ConfirmToken: responseWhitelistFromAlarmConfirmToken,
			})
			if responseOpts.white.preview || strings.TrimSpace(responseOpts.white.confirm) == "" {
				return writeResponseSuccess(cmd, "从告警生成响应白名单预览", commandText, map[string]any{"doc_id": strings.TrimSpace(responseOpts.id), "target": responseFromAlarmTarget(responseOpts.target, "victim")}, preview)
			}
			var result struct {
				ID uint `json:"id"`
				Id uint `json:"Id"`
			}
			if err := client.CallRPC(cmd.Context(), "FirewallService.CreateWhiteList", req, &result); err != nil {
				return writeResponseError(cmd, "从告警生成响应白名单", commandText, "RESPONSE_WHITELIST_FROM_ALARM_CREATE_FAILED", err.Error(), true)
			}
			id := result.ID
			if id == 0 {
				id = result.Id
			}
			object := map[string]any{"id": id, "source_alarm_doc_id": strings.TrimSpace(responseOpts.id), "type": req["type"], "values": req["values"], "status": req["status"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "response_whitelist_from_alarm", Object: object, Action: "create_from_alarm", Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: map[string]any{"before": sourceAlarm, "after": req}, Result: "success"})
			return writeResponseSuccess(cmd, "从告警生成响应白名单", commandText, map[string]any{"doc_id": strings.TrimSpace(responseOpts.id), "target": responseFromAlarmTarget(responseOpts.target, "victim")}, data)
		},
	}
	addResponseWhitelistFromAlarmFlags(cmd, &responseOpts)
	return cmd
}

func newResponseDevicesCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseDevicesOptions
	cmd := &cobra.Command{
		Use:   "devices",
		Short: "查询联动设备配置",
		Long:  "查询联动设备配置，用于查看第三方协同处置设备是否已配置、状态、地址和更新时间。该命令只读取设备配置，不测试或修改设备连接。\n\n输出：实际筛选条件、total、page、page_size、current_count、has_more、devices。",
		Example: "  chaitin-cli tanswer response devices --page-size 10\n" +
			"  chaitin-cli tanswer response devices --status enabled",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := buildResponseDevicesRequest(responseOpts)
			return runResponseDataListCommand(cmd, opts, "查询联动设备配置", "chaitin-cli tanswer response devices", "FirewallService.SearchFirewall", req, responseDevicesFilters(responseOpts), responseOpts.page, responseOpts.pageSize, "devices", summarizeResponseDevices)
		},
	}
	addResponseDevicesFlags(cmd, &responseOpts)
	return cmd
}

func newResponseDeviceRecordsCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseDeviceRecordsOptions
	cmd := &cobra.Command{
		Use:     "device-records",
		Short:   "查询联动设备处置记录",
		Long:    "查询联动设备处置记录，用于查看某个联动设备的处置下发结果。该命令只读取产品记录，不验证第三方设备真实网络状态。\n\n输出：device_id、total、page、page_size、current_count、has_more、device_records。",
		Example: "  chaitin-cli tanswer response device-records --device-id 2 --page-size 10",
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID, err := strconv.Atoi(strings.TrimSpace(responseOpts.deviceID))
			if err != nil || deviceID <= 0 {
				return writeResponseError(cmd, "查询联动设备处置记录", "chaitin-cli tanswer response device-records", "MISSING_DEVICE_ID", "missing device id: set --device-id", false)
			}
			req := buildResponseDeviceRecordsRequest(responseOpts, deviceID)
			query := map[string]any{"device_id": deviceID, "page": responseOpts.page, "page_size": responseOpts.pageSize}
			return runResponseDataListCommand(cmd, opts, "查询联动设备处置记录", "chaitin-cli tanswer response device-records", "FirewallService.SearchSendRecord", req, query, responseOpts.page, responseOpts.pageSize, "device_records", summarizeResponseDeviceRecords)
		},
	}
	cmd.Flags().StringVar(&responseOpts.deviceID, "device-id", "", "linkage device id")
	cmd.Flags().IntVar(&responseOpts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&responseOpts.pageSize, "page-size", 10, "page size, 1-100")
	return cmd
}

func newResponseAutoPoliciesCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseAutoPoliciesOptions
	cmd := &cobra.Command{
		Use:   "auto-policies",
		Short: "查询自动响应策略",
		Long:  "查询自动响应策略，用于查看自动响应策略配置、启停状态、处置方式和联动处置设备。该命令只读取策略，不启停或修改策略。\n\n输出：查询时间范围、实际筛选条件、total、page、page_size、current_count、has_more、auto_policies。",
		Example: "  chaitin-cli tanswer response auto-policies --page-size 10\n" +
			"  chaitin-cli tanswer response auto-policies --status enabled",
		RunE: func(cmd *cobra.Command, args []string) error {
			rng, req, err := buildResponseAutoPoliciesTimedRequest(responseOpts)
			if err != nil {
				return writeResponseError(cmd, "查询自动响应策略", "chaitin-cli tanswer response auto-policies", "INVALID_TIME_RANGE", err.Error(), false)
			}
			query := map[string]any{"time_range": rng, "filters": responseAutoPoliciesFilters(responseOpts), "page": responseOpts.page, "page_size": responseOpts.pageSize}
			return runResponseDataListCommand(cmd, opts, "查询自动响应策略", "chaitin-cli tanswer response auto-policies", "FirewallService.SearchStrategy", req, query, responseOpts.page, responseOpts.pageSize, "auto_policies", summarizeResponseAutoPolicies)
		},
	}
	addResponseAutoPoliciesFlags(cmd, &responseOpts)
	return cmd
}

func newResponseAutoListCommand(opts *RootOptions) *cobra.Command {
	var responseOpts responseAutoListOptions
	cmd := &cobra.Command{
		Use:   "auto-list",
		Short: "查询自动响应处置名单",
		Long:  "查询自动响应处置名单，用于查看自动响应策略生成的待处置或已处置对象。该命令只读取名单，不触发新的自动响应分析或处置。\n\n输出：查询时间范围、实际筛选条件、total、page、page_size、current_count、has_more、auto_list。",
		Example: "  chaitin-cli tanswer response auto-list --time 7d --page-size 10\n" +
			"  chaitin-cli tanswer response auto-list --ip 198.51.100.10 --status enabled",
		RunE: func(cmd *cobra.Command, args []string) error {
			rng, req, err := buildResponseAutoListTimedRequest(responseOpts)
			if err != nil {
				return writeResponseError(cmd, "查询自动响应处置名单", "chaitin-cli tanswer response auto-list", "INVALID_TIME_RANGE", err.Error(), false)
			}
			query := map[string]any{"time_range": rng, "filters": responseAutoListFilters(responseOpts), "page": responseOpts.page, "page_size": responseOpts.pageSize}
			return runResponseDataListCommand(cmd, opts, "查询自动响应处置名单", "chaitin-cli tanswer response auto-list", "FirewallService.SearchBlackList", req, query, responseOpts.page, responseOpts.pageSize, "auto_list", summarizeResponseAutoList)
		},
	}
	addResponseAutoListFlags(cmd, &responseOpts)
	return cmd
}

func runResponseDataListCommand(cmd *cobra.Command, opts *RootOptions, task string, command string, method string, req map[string]any, query any, page int, pageSize int, dataKey string, summarize func([]map[string]any) []map[string]any) error {
	if err := validateResponsePage(page, pageSize); err != nil {
		return writeResponseError(cmd, task, command, "INVALID_PAGE", err.Error(), false)
	}
	cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
	if err != nil {
		return err
	}
	client := NewClient(cfg)
	var result responseListRPCResult
	if err := client.CallRPC(cmd.Context(), method, req, &result); err != nil {
		return writeResponseError(cmd, task, command, "RESPONSE_LIST_FAILED", err.Error(), true)
	}
	return writeResponseSuccess(cmd, task, command, query, responseListData(result.Total, page, pageSize, len(result.Data), dataKey, summarize(result.Data)))
}

func responseListData(total int64, page int, pageSize int, currentCount int, dataKey string, items []map[string]any) map[string]any {
	return map[string]any{
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
		"current_count": currentCount,
		"has_more":      int64(page*pageSize) < total,
		dataKey:         items,
	}
}

func buildResponseBlockPoliciesRequest(opts responseBlockPoliciesOptions) map[string]any {
	req := responsePageRequest(opts.page, opts.pageSize)
	for key, value := range responseBlockPoliciesFilters(opts) {
		req[key] = value
	}
	return req
}

func responseBlockPoliciesFilters(opts responseBlockPoliciesOptions) map[string]any {
	filters := map[string]any{}
	addIntFilter(filters, "id", opts.id, portValue)
	addIntFilter(filters, "strategy_id", opts.strategyID, portValue)
	addStringFilter(filters, "name", opts.name, strings.TrimSpace)
	addIntFilter(filters, "status", opts.status, statusValue)
	if value := strings.TrimSpace(opts.object); value != "" {
		filters["ip_search"] = value
	}
	return filters
}

func buildResponseBlockPolicyWriteRequest(opts responseBlockPolicyWriteOptions) (map[string]any, error) {
	name := strings.TrimSpace(opts.name)
	if name == "" {
		return nil, fmt.Errorf("missing block policy name: set --name")
	}
	objects := parseCSV(opts.objects)
	if len(objects) == 0 {
		return nil, fmt.Errorf("missing block object: set --object")
	}
	blockType, err := responseBlockObjectTypeValue(opts.objectType)
	if err != nil {
		return nil, err
	}
	status := uint(statusValue(opts.status))
	if status != 0 && status != 1 {
		return nil, fmt.Errorf("unsupported status %q, expected enabled, disabled, 1, or 0", opts.status)
	}
	blockTimeType, blockTimeValue, err := responseBlockTimeValue(opts)
	if err != nil {
		return nil, err
	}
	req := map[string]any{
		"name":             name,
		"ips":              objects,
		"status":           status,
		"type":             uint(0),
		"block_type":       blockType,
		"block_time_type":  blockTimeType,
		"block_time_value": blockTimeValue,
		"block_target":     []string{"answer"},
		"remark":           strings.TrimSpace(opts.remark),
	}
	if blockType == 1 {
		ipType, err := responseBlockIPTypeValue(opts.ipType)
		if err != nil {
			return nil, err
		}
		req["ip_type"] = &ipType
	}
	return req, nil
}

func buildResponseBlockPolicyUpdateRequest(opts responseBlockPolicyWriteOptions) (map[string]any, error) {
	id, err := parseRequiredPolicyObjectID(opts.id, "block policy")
	if err != nil {
		return nil, err
	}
	req, err := buildResponseBlockPolicyWriteRequest(opts)
	if err != nil {
		return nil, err
	}
	expire, err := optionalInt64(opts.expire, "expire")
	if err != nil {
		return nil, err
	}
	if expire == 0 {
		_, value, err := responseBlockTimeValue(opts)
		if err != nil {
			return nil, err
		}
		expire = value
	}
	req["id"] = id
	req["expire"] = expire
	req["block_time_type"] = int32(2)
	delete(req, "block_time_value")
	return req, nil
}

func responseBlockObjectTypeValue(value string) (uint, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "ip", "cidr", "1":
		return 1, nil
	case "ip-port", "ip_port", "port", "2":
		return 2, nil
	case "tuple", "quadruple", "four-tuple", "4tuple", "3":
		return 3, nil
	case "url", "4":
		return 4, nil
	default:
		return 0, fmt.Errorf("unsupported object-type %q, expected ip, ip-port, tuple, url, or 1-4", value)
	}
}

func responseBlockIPTypeValue(value string) (uint, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "both", "src_or_dst", "source-or-destination", "0":
		return 0, nil
	case "source", "src", "1":
		return 1, nil
	case "dest", "destination", "dst", "2":
		return 2, nil
	default:
		return 0, fmt.Errorf("unsupported ip-type %q, expected both, source, dest, 0, 1, or 2", value)
	}
}

func responseBlockTimeValue(opts responseBlockPolicyWriteOptions) (int32, int64, error) {
	switch strings.ToLower(strings.TrimSpace(opts.blockTimeType)) {
	case "", "duration", "1":
		duration := strings.TrimSpace(opts.duration)
		if duration == "" {
			duration = "3600"
		}
		value, err := optionalInt64(duration, "duration")
		if err != nil {
			return 0, 0, err
		}
		if value <= 0 {
			return 0, 0, fmt.Errorf("duration must be greater than 0")
		}
		return 1, value, nil
	case "expire", "time", "2":
		value, err := optionalInt64(opts.expire, "expire")
		if err != nil {
			return 0, 0, err
		}
		if value <= 0 {
			return 0, 0, fmt.Errorf("expire must be greater than 0 when block-time-type is expire")
		}
		return 2, value, nil
	default:
		return 0, 0, fmt.Errorf("unsupported block-time-type %q, expected duration, expire, 1, or 2", opts.blockTimeType)
	}
}

func fetchResponseBlockPoliciesByIDs(cmd *cobra.Command, client *Client, ids []uint) ([]map[string]any, error) {
	req := map[string]any{
		"offset": int64(0),
		"count":  int64(len(ids)),
		"id":     uintEqualityQuery(ids),
	}
	var result responseListRPCResult
	if err := client.CallRPC(cmd.Context(), "RulesService.SearchBlockRules", req, &result); err != nil {
		return nil, err
	}
	return summarizeResponseBlockPolicies(result.Data), nil
}

func buildResponseBlockRecordsRequest(rng TimeRange, opts responseBlockRecordsOptions) map[string]any {
	req := responsePageRequest(opts.page, opts.pageSize)
	req["start_time"] = rng.Start
	req["end_time"] = rng.End
	copyPlainString(req, "src_ip", opts.srcIP)
	copyPlainString(req, "dest_ip", opts.destIP)
	copyPlainString(req, "policy_name", opts.policyName)
	copyPlainInt(req, "src_port", opts.srcPort)
	copyPlainInt(req, "dest_port", opts.destPort)
	copyPlainInt(req, "policy_id", opts.policyID)
	copyPlainInt(req, "strategy_id", opts.strategyID)
	copyPlainInt(req, "type", opts.blockType)
	return req
}

func responseBlockRecordsFilters(opts responseBlockRecordsOptions) map[string]any {
	return nonEmptyFilterMap(map[string]string{"src_ip": opts.srcIP, "src_port": opts.srcPort, "dest_ip": opts.destIP, "dest_port": opts.destPort, "policy_id": opts.policyID, "policy_name": opts.policyName, "strategy_id": opts.strategyID, "type": opts.blockType})
}

func buildResponseWhitelistTimedRequest(opts responseWhitelistOptions) (TimeRange, map[string]any, error) {
	rng, err := ParseTimeRange(TimeRangeOptions{Time: opts.time, Start: opts.start, End: opts.end})
	if err != nil {
		return TimeRange{}, nil, err
	}
	req := responsePageRequest(opts.page, opts.pageSize)
	req["start_time"] = rng.Start
	req["end_time"] = rng.End
	for key, value := range responseWhitelistFilters(opts) {
		req[key] = value
	}
	copyPlainInt(req, "expire_after", opts.expireAfter)
	return rng, req, nil
}

func responseWhitelistFilters(opts responseWhitelistOptions) map[string]any {
	filters := map[string]any{}
	copyPlainString(filters, "type", opts.objectType)
	copyPlainString(filters, "value", opts.object)
	copyPlainString(filters, "ip_search", opts.object)
	copyPlainString(filters, "remark", opts.remark)
	copyPlainString(filters, "block_method", opts.blockMethod)
	copyPlainString(filters, "ip_type", opts.ipType)
	copyPlainInt(filters, "status", statusString(opts.status))
	return filters
}

func buildResponseWhitelistWriteRequest(opts responseWhitelistWriteOptions) (map[string]any, error) {
	objectType, err := responseWhitelistObjectTypeValue(opts.objectType)
	if err != nil {
		return nil, err
	}
	values := parseCSV(opts.objects)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing whitelist object: set --object")
	}
	status, err := responseWhitelistStatusValue(opts.status)
	if err != nil {
		return nil, err
	}
	expire, err := optionalInt64(opts.expire, "expire")
	if err != nil {
		return nil, err
	}
	if expire <= 0 {
		return nil, fmt.Errorf("missing whitelist expire timestamp: set --expire")
	}
	req := map[string]any{
		"type":   objectType,
		"values": values,
		"status": status,
		"remark": strings.TrimSpace(opts.remark),
		"expire": expire,
	}
	if objectType == "ip" {
		methods, err := responseWhitelistBlockMethods(opts.blockMethod)
		if err != nil {
			return nil, err
		}
		ipType, err := responseWhitelistIPTypeValue(opts.ipType)
		if err != nil {
			return nil, err
		}
		req["block_method"] = methods
		req["ip_type"] = ipType
	}
	return req, nil
}

func buildResponseWhitelistUpdateRequest(opts responseWhitelistWriteOptions) (map[string]any, error) {
	id, err := parseRequiredPolicyObjectID(opts.id, "response whitelist")
	if err != nil {
		return nil, err
	}
	req, err := buildResponseWhitelistWriteRequest(opts)
	if err != nil {
		return nil, err
	}
	req["id"] = id
	return req, nil
}

func responseWhitelistObjectTypeValue(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "ip":
		return "ip", nil
	case "url":
		return "url", nil
	default:
		return "", fmt.Errorf("unsupported whitelist type %q, expected ip or url", value)
	}
}

func responseWhitelistStatusValue(value string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "enabled", "enable", "active", "启用", "2":
		return 2, nil
	case "disabled", "disable", "inactive", "停用", "1":
		return 1, nil
	default:
		return 0, fmt.Errorf("unsupported whitelist status %q, expected enabled, disabled, 2, or 1", value)
	}
}

func responseWhitelistBlockMethods(value string) ([]string, error) {
	values := parseCSV(value)
	if len(values) == 0 {
		return []string{"Bypass", "Third_party"}, nil
	}
	out := make([]string, 0, len(values))
	for _, item := range values {
		switch strings.ToLower(strings.TrimSpace(item)) {
		case "bypass", "answer", "旁路阻断":
			out = append(out, "Bypass")
		case "third", "third_party", "third-party", "thirdparty", "第三方协同阻断":
			out = append(out, "Third_party")
		default:
			return nil, fmt.Errorf("unsupported block-method %q, expected Bypass or Third_party", item)
		}
	}
	return out, nil
}

func responseWhitelistIPTypeValue(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "both", "src_or_dst", "src-or-dst", "source-or-destination", "0":
		return "SRC_OR_DST", nil
	case "source", "src", "src_ip", "src-ip", "1":
		return "SRC_IP", nil
	case "dest", "destination", "dst", "dst_ip", "dst-ip", "2":
		return "DST_IP", nil
	default:
		return "", fmt.Errorf("unsupported ip-type %q, expected both, source, dest, SRC_OR_DST, SRC_IP, or DST_IP", value)
	}
}

func fetchResponseWhitelistsByIDs(cmd *cobra.Command, client *Client, ids []uint) ([]map[string]any, error) {
	req := map[string]any{
		"offset": int64(0),
		"count":  int64(len(ids)),
		"ids":    ids,
	}
	var result responseListRPCResult
	if err := client.CallRPC(cmd.Context(), "FirewallService.SearchWhiteList", req, &result); err != nil {
		return nil, err
	}
	return summarizeResponseWhitelists(result.Data), nil
}

func buildResponseBlockPolicyFromAlarmRequest(alarm map[string]any, opts responseFromAlarmOptions) (map[string]any, error) {
	writeOpts := opts.block
	alarmName := firstPolicyString(alarm, "name", "msg", "tag")
	if strings.TrimSpace(writeOpts.name) == "" {
		nameBase := alarmName
		if nameBase == "" {
			nameBase = strings.TrimSpace(opts.id)
		}
		if nameBase == "" {
			nameBase = "告警"
		}
		writeOpts.name = "告警阻断-" + nameBase
	}
	if strings.TrimSpace(writeOpts.objects) == "" {
		target := responseFromAlarmTarget(opts.target, "attacker")
		object, objectType, err := responseBlockObjectFromAlarm(alarm, target)
		if err != nil {
			return nil, err
		}
		writeOpts.objects = object
		if strings.TrimSpace(writeOpts.objectType) == "" || strings.TrimSpace(writeOpts.objectType) == "ip" {
			writeOpts.objectType = objectType
		}
	}
	return buildResponseBlockPolicyWriteRequest(writeOpts)
}

func buildResponseWhitelistFromAlarmRequest(alarm map[string]any, opts responseFromAlarmOptions) (map[string]any, error) {
	writeOpts := opts.white
	if strings.TrimSpace(writeOpts.objects) == "" {
		target := responseFromAlarmTarget(opts.target, "victim")
		object, objectType, err := responseWhitelistObjectFromAlarm(alarm, target)
		if err != nil {
			return nil, err
		}
		writeOpts.objects = object
		if strings.TrimSpace(writeOpts.objectType) == "" || strings.TrimSpace(writeOpts.objectType) == "ip" {
			writeOpts.objectType = objectType
		}
	}
	return buildResponseWhitelistWriteRequest(writeOpts)
}

func responseFromAlarmTarget(value string, defaultValue string) string {
	target := strings.ToLower(strings.TrimSpace(value))
	if target == "" {
		return defaultValue
	}
	return target
}

func responseBlockObjectFromAlarm(alarm map[string]any, target string) (string, string, error) {
	switch target {
	case "attacker", "src", "source":
		value := firstPolicyString(alarm, "src_ip", "attacker")
		if value == "" {
			return "", "", fmt.Errorf("alarm missing attacker/source IP")
		}
		return value, "ip", nil
	case "victim", "dest", "destination":
		value := firstPolicyString(alarm, "dest_ip", "victim")
		if value == "" {
			return "", "", fmt.Errorf("alarm missing victim/destination IP")
		}
		return value, "ip", nil
	case "flow", "tuple", "quadruple":
		srcIP := firstPolicyString(alarm, "src_ip", "attacker")
		srcPort := firstPolicyScalarString(alarm, "src_port", "attacker_port")
		destIP := firstPolicyString(alarm, "dest_ip", "victim")
		destPort := firstPolicyScalarString(alarm, "dest_port", "victim_port")
		if srcIP == "" || srcPort == "" || destIP == "" || destPort == "" {
			return "", "", fmt.Errorf("alarm missing source/destination IP or port for flow target")
		}
		return fmt.Sprintf("%s:%s-%s:%s", srcIP, srcPort, destIP, destPort), "tuple", nil
	default:
		return "", "", fmt.Errorf("unsupported target %q, expected attacker, victim, or flow", target)
	}
}

func responseWhitelistObjectFromAlarm(alarm map[string]any, target string) (string, string, error) {
	switch target {
	case "attacker", "src", "source":
		value := firstPolicyString(alarm, "src_ip", "attacker")
		if value == "" {
			return "", "", fmt.Errorf("alarm missing attacker/source IP")
		}
		return value, "ip", nil
	case "victim", "dest", "destination":
		value := firstPolicyString(alarm, "dest_ip", "victim")
		if value == "" {
			return "", "", fmt.Errorf("alarm missing victim/destination IP")
		}
		return value, "ip", nil
	case "url":
		value := firstPolicyString(alarm, "url", "url_path")
		if value == "" {
			value = nestedPolicyString(alarm, "appbrief", "http", "url")
		}
		if strings.HasPrefix(value, "/") {
			host := firstPolicyString(alarm, "domain", "host", "hostname")
			if host == "" {
				host = nestedPolicyString(alarm, "appbrief", "http", "hostname")
			}
			if host != "" {
				value = "http://" + strings.TrimRight(host, "/") + value
			}
		}
		if value == "" {
			return "", "", fmt.Errorf("alarm missing URL")
		}
		return value, "url", nil
	default:
		return "", "", fmt.Errorf("unsupported target %q, expected attacker, victim, or url", target)
	}
}

func buildResponseDevicesRequest(opts responseDevicesOptions) map[string]any {
	req := responsePageRequest(opts.page, opts.pageSize)
	copyPlainInt(req, "device_type", opts.deviceType)
	copyPlainInt(req, "status", statusString(opts.status))
	copyPlainString(req, "remark", opts.remark)
	if ids, err := parseIntCSV(opts.id); err == nil && len(ids) > 0 {
		req["ids"] = ids
	}
	return req
}

func responseDevicesFilters(opts responseDevicesOptions) map[string]any {
	return nonEmptyFilterMap(map[string]string{"id": opts.id, "device_type": opts.deviceType, "status": opts.status, "remark": opts.remark})
}

func buildResponseDeviceRecordsRequest(opts responseDeviceRecordsOptions, deviceID int) map[string]any {
	req := responsePageRequest(opts.page, opts.pageSize)
	req["firewall_id"] = deviceID
	req["order"] = "last_result"
	req["order_desc"] = "desc"
	return req
}

func buildResponseAutoPoliciesTimedRequest(opts responseAutoPoliciesOptions) (TimeRange, map[string]any, error) {
	rng, err := ParseTimeRange(TimeRangeOptions{Time: opts.time, Start: opts.start, End: opts.end})
	if err != nil {
		return TimeRange{}, nil, err
	}
	req := responsePageRequest(opts.page, opts.pageSize)
	req["start_time"] = rng.Start
	req["end_time"] = rng.End
	copyPlainInt(req, "id", opts.id)
	copyPlainString(req, "name", opts.name)
	copyPlainInt(req, "status", statusString(opts.status))
	copyPlainInt(req, "punish_type", opts.punishType)
	if ids, err := parseUintCSV(opts.deviceID); err == nil && len(ids) > 0 {
		req["firewall_id"] = ids
	}
	return rng, req, nil
}

func responseAutoPoliciesFilters(opts responseAutoPoliciesOptions) map[string]any {
	return nonEmptyFilterMap(map[string]string{"id": opts.id, "name": opts.name, "device_id": opts.deviceID, "status": opts.status, "punish_type": opts.punishType})
}

func buildResponseAutoListTimedRequest(opts responseAutoListOptions) (TimeRange, map[string]any, error) {
	rng, err := ParseTimeRange(TimeRangeOptions{Time: opts.time, Start: opts.start, End: opts.end})
	if err != nil {
		return TimeRange{}, nil, err
	}
	req := responsePageRequest(opts.page, opts.pageSize)
	req["start_time"] = rng.Start
	req["end_time"] = rng.End
	copyPlainString(req, "ip", opts.ip)
	copyPlainInt(req, "strategy_id", opts.strategyID)
	copyPlainInt(req, "block_time_type", opts.blockTimeType)
	if strings.TrimSpace(opts.status) != "" {
		req["status"] = []uint{uint(statusValue(opts.status))}
	}
	return rng, req, nil
}

func responseAutoListFilters(opts responseAutoListOptions) map[string]any {
	return nonEmptyFilterMap(map[string]string{"ip": opts.ip, "status": opts.status, "strategy_id": opts.strategyID, "block_time_type": opts.blockTimeType})
}

func responsePageRequest(page int, pageSize int) map[string]any {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return map[string]any{"offset": int64((page - 1) * pageSize), "count": int64(pageSize)}
}

func validateResponsePage(page int, pageSize int) error {
	if page < 1 {
		return fmt.Errorf("page must be greater than or equal to 1")
	}
	if pageSize < 1 || pageSize > 100 {
		return fmt.Errorf("page-size must be between 1 and 100")
	}
	return nil
}

func addResponseBlockPoliciesFlags(cmd *cobra.Command, opts *responseBlockPoliciesOptions) {
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().StringVar(&opts.id, "id", "", "block policy id")
	cmd.Flags().StringVar(&opts.strategyID, "strategy-id", "", "auto response strategy id")
	cmd.Flags().StringVar(&opts.name, "name", "", "block policy name filter")
	cmd.Flags().StringVar(&opts.object, "object", "", "block object IP/CIDR/range filter")
	cmd.Flags().StringVar(&opts.status, "status", "", "status filter: enabled, disabled, 1, 0")
}

func addResponseBlockPolicyWriteFlags(cmd *cobra.Command, opts *responseBlockPolicyWriteOptions, includeID bool) {
	if includeID {
		cmd.Flags().StringVar(&opts.id, "id", "", "block policy ID, required")
	}
	cmd.Flags().StringVar(&opts.name, "name", "", "block policy name, required")
	cmd.Flags().StringVar(&opts.objects, "object", "", "block objects, comma separated, required")
	cmd.Flags().StringVar(&opts.objectType, "object-type", "ip", "block object type: ip, ip-port, tuple, url, 1, 2, 3, or 4")
	cmd.Flags().StringVar(&opts.ipType, "ip-type", "both", "IP direction for ip object type: both, source, dest, 0, 1, or 2")
	cmd.Flags().StringVar(&opts.status, "status", "enabled", "policy status: enabled, disabled, 1, or 0")
	cmd.Flags().StringVar(&opts.blockTimeType, "block-time-type", "duration", "block time type: duration, expire, 1, or 2")
	cmd.Flags().StringVar(&opts.duration, "duration", "3600", "block duration in seconds when block-time-type is duration")
	cmd.Flags().StringVar(&opts.expire, "expire", "", "expire timestamp in milliseconds when block-time-type is expire; required for update")
	cmd.Flags().StringVar(&opts.remark, "remark", "", "remark")
	cmd.Flags().BoolVar(&opts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&opts.confirm, "confirm", "", "exact confirmation token required to execute")
}

func addResponseBlockRecordsFlags(cmd *cobra.Command, opts *responseBlockRecordsOptions) {
	cmd.Flags().StringVar(&opts.time, "time", "today", "time range: today, 24h, 7d")
	cmd.Flags().StringVar(&opts.start, "start", "", "custom start time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&opts.end, "end", "", "custom end time, format: 2006-01-02 15:04:05")
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().StringVar(&opts.srcIP, "src-ip", "", "source IP filter")
	cmd.Flags().StringVar(&opts.srcPort, "src-port", "", "source port filter")
	cmd.Flags().StringVar(&opts.destIP, "dest-ip", "", "destination IP filter")
	cmd.Flags().StringVar(&opts.destPort, "dest-port", "", "destination port filter")
	cmd.Flags().StringVar(&opts.policyID, "policy-id", "", "block policy id")
	cmd.Flags().StringVar(&opts.policyName, "policy-name", "", "block policy name")
	cmd.Flags().StringVar(&opts.strategyID, "strategy-id", "", "auto response strategy id")
	cmd.Flags().StringVar(&opts.blockType, "type", "", "block record type")
}

func addResponseWhitelistFlags(cmd *cobra.Command, opts *responseWhitelistOptions) {
	cmd.Flags().StringVar(&opts.time, "time", "today", "updated time range: today, 24h, 7d")
	cmd.Flags().StringVar(&opts.start, "start", "", "custom start time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&opts.end, "end", "", "custom end time, format: 2006-01-02 15:04:05")
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().StringVar(&opts.objectType, "type", "", "whitelist type: ip or url")
	cmd.Flags().StringVar(&opts.object, "object", "", "whitelist object filter")
	cmd.Flags().StringVar(&opts.status, "status", "", "status filter: enabled, disabled, 1, 0")
	cmd.Flags().StringVar(&opts.expireAfter, "expire-after", "", "expire timestamp lower bound in milliseconds")
	cmd.Flags().StringVar(&opts.blockMethod, "block-method", "", "block method filter, for example Bypass or Third_party")
	cmd.Flags().StringVar(&opts.ipType, "ip-type", "", "IP type filter")
	cmd.Flags().StringVar(&opts.remark, "remark", "", "remark filter")
}

func addResponseWhitelistWriteFlags(cmd *cobra.Command, opts *responseWhitelistWriteOptions, includeID bool) {
	if includeID {
		cmd.Flags().StringVar(&opts.id, "id", "", "response whitelist ID, required")
	}
	cmd.Flags().StringVar(&opts.objectType, "type", "ip", "whitelist type: ip or url")
	cmd.Flags().StringVar(&opts.objects, "object", "", "whitelist objects, comma separated, required")
	cmd.Flags().StringVar(&opts.status, "status", "enabled", "whitelist status: enabled, disabled, 2, or 1")
	cmd.Flags().StringVar(&opts.expire, "expire", "", "expire timestamp in milliseconds, required")
	cmd.Flags().StringVar(&opts.blockMethod, "block-method", "", "IP whitelist block methods, comma separated: Bypass, Third_party")
	cmd.Flags().StringVar(&opts.ipType, "ip-type", "both", "IP direction for ip whitelist: both, source, dest, SRC_OR_DST, SRC_IP, or DST_IP")
	cmd.Flags().StringVar(&opts.remark, "remark", "", "remark")
	cmd.Flags().BoolVar(&opts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&opts.confirm, "confirm", "", "exact confirmation token required to execute")
}

func addResponseBlockPolicyFromAlarmFlags(cmd *cobra.Command, opts *responseFromAlarmOptions) {
	cmd.Flags().StringVar(&opts.id, "id", "", "alarm doc_id from alarm list or alarm detail, required")
	cmd.Flags().StringVar(&opts.target, "target", "attacker", "candidate target from alarm: attacker, victim, or flow")
	cmd.Flags().StringVar(&opts.block.name, "name", "", "block policy name; defaults from alarm name")
	cmd.Flags().StringVar(&opts.block.objects, "object", "", "override block object instead of deriving from alarm")
	cmd.Flags().StringVar(&opts.block.objectType, "object-type", "ip", "block object type: ip, ip-port, tuple, url, 1, 2, 3, or 4")
	cmd.Flags().StringVar(&opts.block.ipType, "ip-type", "both", "IP direction for ip object type: both, source, dest, 0, 1, or 2")
	cmd.Flags().StringVar(&opts.block.status, "status", "enabled", "policy status: enabled, disabled, 1, or 0")
	cmd.Flags().StringVar(&opts.block.blockTimeType, "block-time-type", "duration", "block time type: duration, expire, 1, or 2")
	cmd.Flags().StringVar(&opts.block.duration, "duration", "3600", "block duration in seconds when block-time-type is duration")
	cmd.Flags().StringVar(&opts.block.expire, "expire", "", "expire timestamp in milliseconds when block-time-type is expire")
	cmd.Flags().StringVar(&opts.block.remark, "remark", "", "remark")
	cmd.Flags().BoolVar(&opts.block.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&opts.block.confirm, "confirm", "", "exact confirmation token required to execute")
}

func addResponseWhitelistFromAlarmFlags(cmd *cobra.Command, opts *responseFromAlarmOptions) {
	cmd.Flags().StringVar(&opts.id, "id", "", "alarm doc_id from alarm list or alarm detail, required")
	cmd.Flags().StringVar(&opts.target, "target", "victim", "candidate target from alarm: attacker, victim, or url")
	cmd.Flags().StringVar(&opts.white.objectType, "type", "ip", "whitelist type: ip or url")
	cmd.Flags().StringVar(&opts.white.objects, "object", "", "override whitelist object instead of deriving from alarm")
	cmd.Flags().StringVar(&opts.white.status, "status", "enabled", "whitelist status: enabled, disabled, 2, or 1")
	cmd.Flags().StringVar(&opts.white.expire, "expire", "", "expire timestamp in milliseconds, required")
	cmd.Flags().StringVar(&opts.white.blockMethod, "block-method", "", "IP whitelist block methods, comma separated: Bypass, Third_party")
	cmd.Flags().StringVar(&opts.white.ipType, "ip-type", "both", "IP direction for ip whitelist: both, source, dest, SRC_OR_DST, SRC_IP, or DST_IP")
	cmd.Flags().StringVar(&opts.white.remark, "remark", "", "remark")
	cmd.Flags().BoolVar(&opts.white.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&opts.white.confirm, "confirm", "", "exact confirmation token required to execute")
}

func addResponseDevicesFlags(cmd *cobra.Command, opts *responseDevicesOptions) {
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().StringVar(&opts.id, "id", "", "device id filter, comma separated")
	cmd.Flags().StringVar(&opts.deviceType, "device-type", "", "device type integer filter")
	cmd.Flags().StringVar(&opts.status, "status", "", "status filter: enabled, disabled, 1, 0")
	cmd.Flags().StringVar(&opts.remark, "remark", "", "remark filter")
}

func addResponseAutoPoliciesFlags(cmd *cobra.Command, opts *responseAutoPoliciesOptions) {
	cmd.Flags().StringVar(&opts.time, "time", "today", "updated time range: today, 24h, 7d")
	cmd.Flags().StringVar(&opts.start, "start", "", "custom start time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&opts.end, "end", "", "custom end time, format: 2006-01-02 15:04:05")
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().StringVar(&opts.id, "id", "", "auto response policy id")
	cmd.Flags().StringVar(&opts.name, "name", "", "auto response policy name filter")
	cmd.Flags().StringVar(&opts.deviceID, "device-id", "", "linkage device id filter, comma separated")
	cmd.Flags().StringVar(&opts.status, "status", "", "status filter: enabled, disabled, 1, 0")
	cmd.Flags().StringVar(&opts.punishType, "punish-type", "", "punish type integer filter")
}

func addResponseAutoListFlags(cmd *cobra.Command, opts *responseAutoListOptions) {
	cmd.Flags().StringVar(&opts.time, "time", "today", "updated time range: today, 24h, 7d")
	cmd.Flags().StringVar(&opts.start, "start", "", "custom start time, format: 2006-01-02 15:04:05")
	cmd.Flags().StringVar(&opts.end, "end", "", "custom end time, format: 2006-01-02 15:04:05")
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().StringVar(&opts.ip, "ip", "", "IP filter")
	cmd.Flags().StringVar(&opts.status, "status", "", "status filter: enabled, disabled, 1, 0")
	cmd.Flags().StringVar(&opts.strategyID, "strategy-id", "", "auto response strategy id")
	cmd.Flags().StringVar(&opts.blockTimeType, "block-time-type", "", "block time type")
}

func summarizeResponseBlockPolicies(items []map[string]any) []map[string]any {
	return summarizeResponseItems(items, []string{"id", "name", "ips", "status", "block_type", "ip_type", "expire", "created_at", "type", "strategy_id", "strategy_name", "firewalls", "remark"})
}

func summarizeResponseBlockRecords(items []map[string]any) []map[string]any {
	return summarizeResponseItems(items, []string{"id", "timestamp", "event_type", "src_ip", "src_port", "dest_ip", "dest_port", "policy_id", "policy_name", "url", "strategy_id", "type", "block_times"})
}

func summarizeResponseWhitelists(items []map[string]any) []map[string]any {
	return summarizeResponseItems(items, []string{"id", "type", "values", "remark", "status", "expire", "updated_at", "block_method", "ip_type"})
}

func summarizeResponseDevices(items []map[string]any) []map[string]any {
	return summarizeResponseItems(items, []string{"id", "device_type", "remark", "status", "addr", "updated_at", "fw_config"})
}

func summarizeResponseDeviceRecords(items []map[string]any) []map[string]any {
	return summarizeResponseItems(items, []string{"id", "ip", "last_try_time", "last_result"})
}

func summarizeResponseAutoPolicies(items []map[string]any) []map[string]any {
	return summarizeResponseItems(items, []string{"id", "name", "firewall_id", "status", "punish_type", "updated_at", "logic", "option", "block_time_type", "block_time_value", "block_type", "rules", "firewalls", "remark"})
}

func summarizeResponseAutoList(items []map[string]any) []map[string]any {
	return summarizeResponseItems(items, []string{"id", "ip", "port", "status", "block_time_type", "block_time_value", "strategy_id", "strategy_name", "fire_wall_ids", "update_time"})
}

func summarizeResponseItems(items []map[string]any, keys []string) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		summary := map[string]any{}
		for _, key := range keys {
			copyIfPresent(summary, item, key)
		}
		out = append(out, summary)
	}
	return out
}

func copyPlainString(dst map[string]any, key string, value string) {
	if strings.TrimSpace(value) != "" {
		dst[key] = strings.TrimSpace(value)
	}
}

func copyPlainInt(dst map[string]any, key string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	n, _ := strconv.Atoi(strings.TrimSpace(value))
	dst[key] = n
}

func statusString(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return strconv.Itoa(statusValue(value))
}

func parseIntCSV(value string) ([]int, error) {
	values := parseCSV(value)
	out := make([]int, 0, len(values))
	for _, item := range values {
		n, err := strconv.Atoi(item)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func nonEmptyFilterMap(values map[string]string) map[string]any {
	out := map[string]any{}
	for key, value := range values {
		if strings.TrimSpace(value) != "" {
			out[key] = strings.TrimSpace(value)
		}
	}
	return out
}

func writeResponseSuccess(cmd *cobra.Command, task string, command string, query any, data map[string]any) error {
	raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: task, Command: command, Query: query, Data: data})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func writeResponseError(cmd *cobra.Command, task string, command string, code string, message string, retryable bool) error {
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
