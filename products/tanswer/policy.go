package tanswer

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const policyCustomIntelligenceCreateConfirmToken = "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_CREATE"
const policyCustomIntelligenceUpdateConfirmToken = "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_UPDATE"
const policyCustomIntelligenceEnableConfirmToken = "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_ENABLE"
const policyCustomIntelligenceDisableConfirmToken = "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_DISABLE"
const policyCustomIntelligenceDeleteConfirmToken = "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_DELETE"
const policyDetectionWhitelistCreateConfirmToken = "CONFIRM_POLICY_DETECTION_WHITELIST_CREATE"
const policyDetectionWhitelistUpdateConfirmToken = "CONFIRM_POLICY_DETECTION_WHITELIST_UPDATE"
const policyDetectionWhitelistEnableConfirmToken = "CONFIRM_POLICY_DETECTION_WHITELIST_ENABLE"
const policyDetectionWhitelistDisableConfirmToken = "CONFIRM_POLICY_DETECTION_WHITELIST_DISABLE"
const policyDetectionWhitelistDeleteConfirmToken = "CONFIRM_POLICY_DETECTION_WHITELIST_DELETE"
const policyDetectionWhitelistFromAlarmConfirmToken = "CONFIRM_POLICY_DETECTION_WHITELIST_FROM_ALARM"
const policyDetectionWhitelistImportConfirmToken = "CONFIRM_POLICY_DETECTION_WHITELIST_IMPORT"
const policyCustomIntelligenceImportConfirmToken = "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_IMPORT"

const policyDetectionWhitelistExportMethod = "AlarmDownloadService.ExportWhiteList"
const policyDetectionWhitelistImportMethod = "AlarmUploadService.ImportWhiteList"
const policyCustomIntelligenceExportMethod = "AlarmDownloadService.ExportAlarmCustomIntelligence"
const policyCustomIntelligenceImportMethod = "AlarmUploadService.ImportAlarmCustomIntelligence"

type policyDetectionWhitelistOptions struct {
	page      int
	pageSize  int
	name      string
	srcIP     string
	srcPort   string
	destIP    string
	destPort  string
	domain    string
	urlPath   string
	userAgent string
	xff       string
	respCode  string
	respBody  string
	threat    string
	ruleID    string
	status    string
}

type policyCustomIntelligenceOptions struct {
	page     int
	pageSize int
	id       string
	name     string
	ioc      string
	iocType  string
	status   string
	remarks  string
}

type policyCustomIntelligenceWriteOptions struct {
	id      string
	idList  string
	name    string
	ioc     string
	iocType string
	status  string
	remarks string
	preview bool
	confirm string
}

type policyDetectionWhitelistWriteOptions struct {
	id           string
	idList       string
	name         string
	srcIP        string
	srcPort      string
	destIP       string
	destPort     string
	domain       string
	urlPath      string
	userAgent    string
	xff          string
	respCode     string
	respBody     string
	threat       string
	ruleID       string
	status       string
	storage      string
	defaultMode  string
	expire       string
	validTime    string
	ignore       bool
	remark       string
	srcAdvanced  string
	destAdvanced string
	sidAdvanced  string
	typeAdvanced string
	preview      bool
	confirm      string
}

type policyDetectionWhitelistFromAlarmOptions struct {
	id    string
	write policyDetectionWhitelistWriteOptions
}

type policyFileOptions struct {
	idList  string
	output  string
	file    string
	preview bool
	confirm string
}

type policyListRPCResult struct {
	Data  []map[string]any `json:"data"`
	Total int64            `json:"total"`
}

func newPolicyCommand(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "安全策略调优语义命令",
		Long:  "安全策略调优语义命令。用于查询检测白名单和自定义情报，并通过预览和精确确认保护维护检测白名单和自定义情报；导入和导出后续按确认保护单独实现。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newPolicyDetectionWhitelistCommand(opts))
	cmd.AddCommand(newPolicyDetectionWhitelistCreateCommand(opts))
	cmd.AddCommand(newPolicyDetectionWhitelistUpdateCommand(opts))
	cmd.AddCommand(newPolicyDetectionWhitelistEnableCommand(opts))
	cmd.AddCommand(newPolicyDetectionWhitelistDisableCommand(opts))
	cmd.AddCommand(newPolicyDetectionWhitelistDeleteCommand(opts))
	cmd.AddCommand(newPolicyDetectionWhitelistFromAlarmCommand(opts))
	cmd.AddCommand(newPolicyDetectionWhitelistExportCommand(opts))
	cmd.AddCommand(newPolicyDetectionWhitelistImportCommand(opts))
	cmd.AddCommand(newPolicyCustomIntelligenceCommand(opts))
	cmd.AddCommand(newPolicyCustomIntelligenceCreateCommand(opts))
	cmd.AddCommand(newPolicyCustomIntelligenceUpdateCommand(opts))
	cmd.AddCommand(newPolicyCustomIntelligenceEnableCommand(opts))
	cmd.AddCommand(newPolicyCustomIntelligenceDisableCommand(opts))
	cmd.AddCommand(newPolicyCustomIntelligenceDeleteCommand(opts))
	cmd.AddCommand(newPolicyCustomIntelligenceExportCommand(opts))
	cmd.AddCommand(newPolicyCustomIntelligenceImportCommand(opts))
	return cmd
}

func newPolicyDetectionWhitelistCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyDetectionWhitelistOptions
	cmd := &cobra.Command{
		Use:   "detection-whitelist",
		Short: "查询检测白名单",
		Long:  "查询检测白名单，用于查看当前误报抑制策略。该命令只读取已有规则，返回列表展示字段和分页信息；不新增、编辑、启停或删除白名单。\n\n输出：实际筛选条件、total、page、page_size、current_count、has_more、detection_whitelists。",
		Example: "  chaitin-cli tanswer policy detection-whitelist --page-size 10\n" +
			"  chaitin-cli tanswer policy detection-whitelist --src-ip 198.51.100.10 --status enabled",
		RunE: func(cmd *cobra.Command, args []string) error {
			if policyOpts.page < 1 {
				return writePolicyError(cmd, "查询检测白名单", "chaitin-cli tanswer policy detection-whitelist", "INVALID_PAGE", "page must be greater than or equal to 1", false)
			}
			if policyOpts.pageSize < 1 || policyOpts.pageSize > 100 {
				return writePolicyError(cmd, "查询检测白名单", "chaitin-cli tanswer policy detection-whitelist", "INVALID_PAGE_SIZE", "page-size must be between 1 and 100", false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			req := buildPolicyDetectionWhitelistRequest(policyOpts)
			client := NewClient(cfg)
			var result policyListRPCResult
			if err := client.CallRPC(cmd.Context(), "AlarmService.SearchWhiteList", req, &result); err != nil {
				return writePolicyError(cmd, "查询检测白名单", "chaitin-cli tanswer policy detection-whitelist", "POLICY_DETECTION_WHITELIST_FAILED", err.Error(), true)
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查询检测白名单",
				Command: "chaitin-cli tanswer policy detection-whitelist",
				Query: map[string]any{
					"filters":   policyDetectionWhitelistFilters(policyOpts),
					"page":      policyOpts.page,
					"page_size": policyOpts.pageSize,
				},
				Data: map[string]any{
					"total":                result.Total,
					"page":                 policyOpts.page,
					"page_size":            policyOpts.pageSize,
					"current_count":        len(result.Data),
					"has_more":             int64(policyOpts.page*policyOpts.pageSize) < result.Total,
					"detection_whitelists": summarizePolicyDetectionWhitelists(result.Data),
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	addPolicyDetectionWhitelistFlags(cmd, &policyOpts)
	return cmd
}

func newPolicyDetectionWhitelistCreateCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyDetectionWhitelistWriteOptions
	cmd := &cobra.Command{
		Use:   "detection-whitelist-create",
		Short: "新增检测白名单",
		Long: "新增检测白名单，用于维护误报抑制规则。该命令是高影响写操作：默认只返回写入预览，必须使用 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_CREATE 才会调用后端创建接口。\n\n" +
			"有效期必须提供 --expire（毫秒时间戳）或 --valid-time（正整数秒）中的至少一项。\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer policy detection-whitelist-create --name 登录误报 --src-ip 198.51.100.10 --valid-time 3600 --preview\n" +
			"  chaitin-cli tanswer policy detection-whitelist-create --name 登录误报 --src-ip 198.51.100.10 --valid-time 3600 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_CREATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildPolicyDetectionWhitelistWriteRequest(policyOpts)
			if err != nil {
				return writePolicyError(cmd, "新增检测白名单", "chaitin-cli tanswer policy detection-whitelist-create", "INVALID_POLICY_DETECTION_WHITELIST_CREATE_REQUEST", err.Error(), false)
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "新增检测白名单预览",
				Command:       "chaitin-cli tanswer policy detection-whitelist-create",
				OperationType: "policy_detection_whitelist_create",
				RiskLevel:     "write_high",
				Target:        map[string]any{"name": req["name"], "match": policyDetectionWhitelistMatchSummary(req)},
				ChangeSummary: map[string]any{"before": nil, "after": req},
				Impact:        map[string]any{"detection_whitelist_count": 1},
				RiskWarnings: []string{
					"将新增检测白名单，匹配范围内的告警可能被抑制或标记忽略。",
					"如果匹配条件过宽，可能影响后续告警可见性。",
				},
				ConfirmToken: policyDetectionWhitelistCreateConfirmToken,
			})
			if policyOpts.preview || strings.TrimSpace(policyOpts.confirm) == "" {
				return writePolicySuccess(cmd, "新增检测白名单预览", "chaitin-cli tanswer policy detection-whitelist-create", req, preview)
			}
			if err := ValidateWriteConfirmation(policyOpts.confirm, policyDetectionWhitelistCreateConfirmToken); err != nil {
				return writePolicyError(cmd, "新增检测白名单", "chaitin-cli tanswer policy detection-whitelist-create", "POLICY_DETECTION_WHITELIST_CREATE_CONFIRMATION_REQUIRED", err.Error(), false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var result struct {
				ID uint `json:"id"`
			}
			if err := client.CallRPC(cmd.Context(), "AlarmService.CreateWhiteList", req, &result); err != nil {
				return writePolicyError(cmd, "新增检测白名单", "chaitin-cli tanswer policy detection-whitelist-create", "POLICY_DETECTION_WHITELIST_CREATE_FAILED", err.Error(), true)
			}
			object := map[string]any{"id": result.ID, "name": req["name"], "status": req["status"], "storage": req["storage"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "policy_detection_whitelist_create",
				Object:        object,
				Action:        "create",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   map[string]any{"before": nil, "after": req},
				Result:        "success",
			})
			return writePolicySuccess(cmd, "新增检测白名单", "chaitin-cli tanswer policy detection-whitelist-create", req, data)
		},
	}
	addPolicyDetectionWhitelistWriteFlags(cmd, &policyOpts, false)
	return cmd
}

func newPolicyDetectionWhitelistUpdateCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyDetectionWhitelistWriteOptions
	cmd := &cobra.Command{
		Use:   "detection-whitelist-update",
		Short: "编辑检测白名单",
		Long: "编辑检测白名单，用于更新单条误报抑制规则。该命令是高影响写操作：预览阶段会读取当前白名单并返回 before/after；必须使用 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_UPDATE 才会调用后端更新接口。\n\n" +
			"有效期必须提供 --expire（毫秒时间戳）或 --valid-time（正整数秒）中的至少一项。\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer policy detection-whitelist-update --id 21 --name 新白名单 --src-ip 198.51.100.11 --valid-time 3600 --preview\n" +
			"  chaitin-cli tanswer policy detection-whitelist-update --id 21 --name 新白名单 --src-ip 198.51.100.11 --valid-time 3600 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_UPDATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildPolicyDetectionWhitelistUpdateRequest(policyOpts)
			if err != nil {
				return writePolicyError(cmd, "编辑检测白名单", "chaitin-cli tanswer policy detection-whitelist-update", "INVALID_POLICY_DETECTION_WHITELIST_UPDATE_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(policyOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(policyOpts.confirm, policyDetectionWhitelistUpdateConfirmToken); err != nil {
					return writePolicyError(cmd, "编辑检测白名单", "chaitin-cli tanswer policy detection-whitelist-update", "POLICY_DETECTION_WHITELIST_UPDATE_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			before, err := fetchPolicyDetectionWhitelistsByIDs(cmd, client, []uint{req["id"].(uint)})
			if err != nil {
				return writePolicyError(cmd, "编辑检测白名单", "chaitin-cli tanswer policy detection-whitelist-update", "POLICY_DETECTION_WHITELIST_UPDATE_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{"before": before, "after": req}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "编辑检测白名单预览",
				Command:       "chaitin-cli tanswer policy detection-whitelist-update",
				OperationType: "policy_detection_whitelist_update",
				RiskLevel:     "write_high",
				Target:        map[string]any{"id": req["id"]},
				ChangeSummary: changeSummary,
				Impact:        map[string]any{"detection_whitelist_count": 1},
				RiskWarnings:  []string{"将更新检测白名单，匹配范围变化可能影响告警抑制结果。"},
				ConfirmToken:  policyDetectionWhitelistUpdateConfirmToken,
			})
			if policyOpts.preview || strings.TrimSpace(policyOpts.confirm) == "" {
				return writePolicySuccess(cmd, "编辑检测白名单预览", "chaitin-cli tanswer policy detection-whitelist-update", req, preview)
			}
			var result struct {
				ID uint `json:"id"`
			}
			if err := client.CallRPC(cmd.Context(), "AlarmService.UpdateWhiteList", req, &result); err != nil {
				return writePolicyError(cmd, "编辑检测白名单", "chaitin-cli tanswer policy detection-whitelist-update", "POLICY_DETECTION_WHITELIST_UPDATE_FAILED", err.Error(), true)
			}
			object := map[string]any{"id": req["id"], "name": req["name"], "status": req["status"], "storage": req["storage"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "policy_detection_whitelist_update",
				Object:        object,
				Action:        "update",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   changeSummary,
				Result:        "success",
			})
			return writePolicySuccess(cmd, "编辑检测白名单", "chaitin-cli tanswer policy detection-whitelist-update", req, data)
		},
	}
	addPolicyDetectionWhitelistWriteFlags(cmd, &policyOpts, true)
	return cmd
}

func newPolicyDetectionWhitelistEnableCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyDetectionWhitelistWriteOptions
	return newPolicyDetectionWhitelistActionCommand(opts, &policyOpts, "detection-whitelist-enable", "启用检测白名单", "enable", "show", policyDetectionWhitelistEnableConfirmToken)
}

func newPolicyDetectionWhitelistDisableCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyDetectionWhitelistWriteOptions
	return newPolicyDetectionWhitelistActionCommand(opts, &policyOpts, "detection-whitelist-disable", "禁用检测白名单", "disable", "hide", policyDetectionWhitelistDisableConfirmToken)
}

func newPolicyDetectionWhitelistDeleteCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyDetectionWhitelistWriteOptions
	return newPolicyDetectionWhitelistActionCommand(opts, &policyOpts, "detection-whitelist-delete", "删除检测白名单", "delete", "delete", policyDetectionWhitelistDeleteConfirmToken)
}

func newPolicyDetectionWhitelistFromAlarmCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyDetectionWhitelistFromAlarmOptions
	commandText := "chaitin-cli tanswer policy detection-whitelist-from-alarm"
	cmd := &cobra.Command{
		Use:   "detection-whitelist-from-alarm",
		Short: "从告警对象生成检测白名单",
		Long: "从告警对象生成检测白名单，用于把已确认误报的威胁告警转换为候选误报抑制规则。该命令是高影响写操作：预览阶段会读取告警详情并生成候选白名单；必须使用 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_FROM_ALARM 才会调用后端创建接口。\n\n" +
			"有效期必须提供 --expire（毫秒时间戳）或 --valid-time（正整数秒）中的至少一项。\n" +
			"输出预览：source_alarm、suggested_whitelist、requires_confirmation、confirmation_token、risk_warnings。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer policy detection-whitelist-from-alarm --id '<doc_id>' --valid-time 3600 --preview\n" +
			"  chaitin-cli tanswer policy detection-whitelist-from-alarm --id '<doc_id>' --remark 已确认误报 --valid-time 3600 --confirm CONFIRM_POLICY_DETECTION_WHITELIST_FROM_ALARM",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(policyOpts.id) == "" {
				return writePolicyError(cmd, "从告警对象生成检测白名单", commandText, "MISSING_ALARM_ID", "missing alarm doc_id: set --id", false)
			}
			if _, _, err := policyDetectionWhitelistExpiry(policyOpts.write); err != nil {
				return writePolicyError(cmd, "从告警对象生成检测白名单", commandText, "INVALID_POLICY_DETECTION_WHITELIST_FROM_ALARM_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(policyOpts.write.confirm) != "" {
				if err := ValidateWriteConfirmation(policyOpts.write.confirm, policyDetectionWhitelistFromAlarmConfirmToken); err != nil {
					return writePolicyError(cmd, "从告警对象生成检测白名单", commandText, "POLICY_DETECTION_WHITELIST_FROM_ALARM_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var alarm alarmDetailRPCResult
			if err := client.CallRPC(cmd.Context(), "AlarmService.GetAlarm", map[string]any{"doc_id": strings.TrimSpace(policyOpts.id)}, &alarm); err != nil {
				return writePolicyError(cmd, "从告警对象生成检测白名单", commandText, "POLICY_DETECTION_WHITELIST_FROM_ALARM_READ_FAILED", err.Error(), true)
			}
			if len(alarm.Data) == 0 {
				return writePolicyError(cmd, "从告警对象生成检测白名单", commandText, "ALARM_NOT_FOUND", "alarm detail is empty", false)
			}
			req, err := buildPolicyDetectionWhitelistFromAlarmRequest(alarm.Data, policyOpts)
			if err != nil {
				return writePolicyError(cmd, "从告警对象生成检测白名单", commandText, "POLICY_DETECTION_WHITELIST_FROM_ALARM_FIELDS_INSUFFICIENT", err.Error(), false)
			}
			sourceAlarm := summarizeSourceAlarm(alarm.Data)
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "从告警对象生成检测白名单预览",
				Command:       commandText,
				OperationType: "policy_detection_whitelist_from_alarm",
				RiskLevel:     "write_high",
				Target: map[string]any{
					"source_alarm":        sourceAlarm,
					"suggested_whitelist": policyDetectionWhitelistMatchSummary(req),
					"name":                req["name"],
				},
				ChangeSummary: map[string]any{"before": sourceAlarm, "after": req},
				Impact:        map[string]any{"detection_whitelist_count": 1, "source_alarm_doc_id": strings.TrimSpace(policyOpts.id)},
				RiskWarnings: []string{
					"将基于单条告警创建检测白名单，匹配范围内的后续告警可能被抑制或标记忽略。",
					"候选范围来自告警字段，执行前应确认源、目的、URL、威胁类型和检测规则没有过宽。",
				},
				ConfirmToken: policyDetectionWhitelistFromAlarmConfirmToken,
			})
			if policyOpts.write.preview || strings.TrimSpace(policyOpts.write.confirm) == "" {
				return writePolicySuccess(cmd, "从告警对象生成检测白名单预览", commandText, map[string]any{"doc_id": strings.TrimSpace(policyOpts.id)}, preview)
			}
			var result struct {
				ID uint `json:"id"`
			}
			if err := client.CallRPC(cmd.Context(), "AlarmService.CreateWhiteList", req, &result); err != nil {
				return writePolicyError(cmd, "从告警对象生成检测白名单", commandText, "POLICY_DETECTION_WHITELIST_FROM_ALARM_CREATE_FAILED", err.Error(), true)
			}
			object := map[string]any{"id": result.ID, "source_alarm_doc_id": strings.TrimSpace(policyOpts.id), "name": req["name"], "status": req["status"], "storage": req["storage"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "policy_detection_whitelist_from_alarm",
				Object:        object,
				Action:        "create_from_alarm",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   map[string]any{"before": sourceAlarm, "after": req},
				Result:        "success",
			})
			return writePolicySuccess(cmd, "从告警对象生成检测白名单", commandText, map[string]any{"doc_id": strings.TrimSpace(policyOpts.id)}, data)
		},
	}
	cmd.Flags().StringVar(&policyOpts.id, "id", "", "alarm doc_id from alarm list or alarm detail, required")
	addPolicyDetectionWhitelistWriteFlags(cmd, &policyOpts.write, false)
	return cmd
}

func newPolicyDetectionWhitelistExportCommand(opts *RootOptions) *cobra.Command {
	return newPolicyFileExportCommand(opts, "detection-whitelist-export", "导出检测白名单", policyDetectionWhitelistExportMethod, "detection-whitelist-export.xlsx", "detection_whitelist")
}

func newPolicyCustomIntelligenceExportCommand(opts *RootOptions) *cobra.Command {
	return newPolicyFileExportCommand(opts, "custom-intelligence-export", "导出自定义情报", policyCustomIntelligenceExportMethod, "custom-intelligence-export.xlsx", "custom_intelligence")
}

func newPolicyFileExportCommand(opts *RootOptions, use string, task string, method string, defaultName string, scopeName string) *cobra.Command {
	var policyOpts policyFileOptions
	commandText := "chaitin-cli tanswer policy " + use
	cmd := &cobra.Command{
		Use:     use,
		Short:   task,
		Long:    fmt.Sprintf("%s，用于批量备份或离线审计安全策略文件。未指定 --id-list 时导出全部；指定 --id-list 时只导出选中对象。该命令只下载文件，不修改策略。\n\n输出：file_name、file_path、size_bytes、status_code、method、download_query、export_scope。", task),
		Example: fmt.Sprintf("  %s --output ./%s\n  %s --id-list 21,22 --output ./selected-%s", commandText, defaultName, commandText, defaultName),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			query, scope, err := buildPolicyExportRequest(policyOpts)
			if err != nil {
				return writePolicyError(cmd, task, commandText, "INVALID_POLICY_EXPORT_REQUEST", err.Error(), false)
			}
			return downloadPolicyFile(cmd, cfg, task, commandText, method, query, policyOpts.output, defaultName, scope, scopeName)
		},
	}
	cmd.Flags().StringVar(&policyOpts.output, "output", "", "output file path; defaults to downloaded filename")
	cmd.Flags().StringVar(&policyOpts.idList, "id-list", "", "comma-separated IDs to export; empty exports all")
	return cmd
}

func newPolicyDetectionWhitelistImportCommand(opts *RootOptions) *cobra.Command {
	return newPolicyFileImportCommand(opts, "detection-whitelist-import", "导入检测白名单", policyDetectionWhitelistImportMethod, policyDetectionWhitelistImportConfirmToken, "policy_detection_whitelist_import", "detection_whitelist")
}

func newPolicyCustomIntelligenceImportCommand(opts *RootOptions) *cobra.Command {
	return newPolicyFileImportCommand(opts, "custom-intelligence-import", "导入自定义情报", policyCustomIntelligenceImportMethod, policyCustomIntelligenceImportConfirmToken, "policy_custom_intelligence_import", "custom_intelligence")
}

func newPolicyFileImportCommand(opts *RootOptions, use string, task string, method string, confirmToken string, operationType string, scopeName string) *cobra.Command {
	var policyOpts policyFileOptions
	commandText := "chaitin-cli tanswer policy " + use
	cmd := &cobra.Command{
		Use:     use,
		Short:   task,
		Long:    fmt.Sprintf("%s，用于上传产品导入模板文件批量维护安全策略。该命令是高影响写操作：预览阶段只读取本地文件元信息，不上传文件；必须使用 --confirm %s 才会调用后端导入接口。\n\n输出预览：requires_confirmation、confirmation_token、file_name、file_path、size_bytes。\n执行输出：confirmed、result、object、audit、import_result。", task, confirmToken),
		Example: fmt.Sprintf("  %s --file ./policy.xlsx --preview\n  %s --file ./policy.xlsx --confirm %s", commandText, commandText, confirmToken),
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildPolicyImportRequest(policyOpts, method)
			if err != nil {
				return writePolicyError(cmd, task, commandText, "INVALID_POLICY_IMPORT_REQUEST", err.Error(), false)
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          task + "预览",
				Command:       commandText,
				OperationType: operationType,
				RiskLevel:     "write_high",
				Target:        map[string]any{"file_name": req["file_name"], "file_path": req["file_path"], "scope": scopeName},
				ChangeSummary: map[string]any{"before": nil, "after": req},
				Impact:        map[string]any{"file_size_bytes": req["size_bytes"], "scope": scopeName},
				RiskWarnings: []string{
					"将上传策略导入文件并批量创建或更新安全策略配置。",
					"导入结果由后端按模板内容返回，可能包含行级失败原因。",
				},
				ConfirmToken: confirmToken,
			})
			if policyOpts.preview || strings.TrimSpace(policyOpts.confirm) == "" {
				return writePolicySuccess(cmd, task+"预览", commandText, req, preview)
			}
			if err := ValidateWriteConfirmation(policyOpts.confirm, confirmToken); err != nil {
				return writePolicyError(cmd, task, commandText, "POLICY_IMPORT_CONFIRMATION_REQUIRED", err.Error(), false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			file, err := os.Open(req["file_path"].(string))
			if err != nil {
				return writePolicyError(cmd, task, commandText, "POLICY_IMPORT_FILE_OPEN_FAILED", err.Error(), false)
			}
			defer file.Close()
			client := NewClient(cfg)
			resp, err := client.UploadFileWithFields(cmd.Context(), method, req["file_name"].(string), file, map[string]string{"param": `{"pattern":"import"}`})
			if err != nil {
				return writePolicyError(cmd, task, commandText, "POLICY_IMPORT_FAILED", err.Error(), true)
			}
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return writePolicyError(cmd, task, commandText, "POLICY_IMPORT_FAILED", fmt.Sprintf("upload returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(resp.Body))), true)
			}
			importResult, err := parseAssetImportResult(resp.Body)
			if err != nil {
				return writePolicyError(cmd, task, commandText, "POLICY_IMPORT_RESPONSE_INVALID", err.Error(), true)
			}
			object := map[string]any{"file_name": req["file_name"], "file_path": req["file_path"], "scope": scopeName}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: operationType,
				Object:        object,
				Action:        "import",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   map[string]any{"before": nil, "after": req},
				Result:        "success",
			})
			data["import_result"] = importResult
			return writePolicySuccess(cmd, task, commandText, req, data)
		},
	}
	cmd.Flags().StringVar(&policyOpts.file, "file", "", "policy import template file path, required")
	cmd.Flags().BoolVar(&policyOpts.preview, "preview", false, "return write preview without uploading")
	cmd.Flags().StringVar(&policyOpts.confirm, "confirm", "", "exact confirmation token required to upload")
	return cmd
}

func newPolicyDetectionWhitelistActionCommand(opts *RootOptions, policyOpts *policyDetectionWhitelistWriteOptions, use string, task string, action string, backendAction string, confirmToken string) *cobra.Command {
	commandText := "chaitin-cli tanswer policy " + use
	cmd := &cobra.Command{
		Use:     use,
		Short:   task,
		Long:    fmt.Sprintf("%s，用于控制检测白名单状态或删除检测白名单。该命令是高影响写操作：预览阶段会读取目标白名单摘要；必须使用 --confirm %s 才会调用后端批量操作接口。\n\n输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n执行输出：confirmed、result、object、audit。", task, confirmToken),
		Example: fmt.Sprintf("  %s --id-list 21,22 --preview\n  %s --id-list 21,22 --confirm %s", commandText, commandText, confirmToken),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseUintCSV(policyOpts.idList)
			if err != nil || len(ids) == 0 {
				return writePolicyError(cmd, task, commandText, "INVALID_POLICY_DETECTION_WHITELIST_ACTION_REQUEST", "missing or invalid id-list: set --id-list", false)
			}
			if strings.TrimSpace(policyOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(policyOpts.confirm, confirmToken); err != nil {
					return writePolicyError(cmd, task, commandText, "POLICY_DETECTION_WHITELIST_"+strings.ToUpper(action)+"_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			before, err := fetchPolicyDetectionWhitelistsByIDs(cmd, client, ids)
			if err != nil {
				return writePolicyError(cmd, task, commandText, "POLICY_DETECTION_WHITELIST_ACTION_PREVIEW_FAILED", err.Error(), true)
			}
			req := map[string]any{"ids": ids, "action": backendAction}
			changeSummary := map[string]any{"before": before, "after": req}
			if backendAction == "delete" {
				changeSummary["after"] = nil
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          task + "预览",
				Command:       commandText,
				OperationType: "policy_detection_whitelist_" + action,
				RiskLevel:     "write_high",
				Target:        map[string]any{"ids": ids},
				ChangeSummary: changeSummary,
				Impact:        map[string]any{"detection_whitelist_count": len(ids), "backend_action": backendAction},
				RiskWarnings:  []string{"将改变检测白名单生效状态或删除规则，可能影响告警抑制结果。"},
				ConfirmToken:  confirmToken,
			})
			if policyOpts.preview || strings.TrimSpace(policyOpts.confirm) == "" {
				return writePolicySuccess(cmd, task+"预览", commandText, req, preview)
			}
			var result struct {
				IDs []uint `json:"ids"`
			}
			if err := client.CallRPC(cmd.Context(), "AlarmService.DeleteWhiteList", req, &result); err != nil {
				return writePolicyError(cmd, task, commandText, "POLICY_DETECTION_WHITELIST_"+strings.ToUpper(action)+"_FAILED", err.Error(), true)
			}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "policy_detection_whitelist_" + action,
				Object:        map[string]any{"ids": ids, "action": backendAction},
				Action:        action,
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   changeSummary,
				Result:        "success",
			})
			return writePolicySuccess(cmd, task, commandText, req, data)
		},
	}
	cmd.Flags().StringVar(&policyOpts.idList, "id-list", "", "detection whitelist IDs, comma separated, required")
	cmd.Flags().BoolVar(&policyOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&policyOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newPolicyCustomIntelligenceCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyCustomIntelligenceOptions
	cmd := &cobra.Command{
		Use:   "custom-intelligence",
		Short: "查询自定义情报",
		Long:  "查询自定义情报，用于查看当前 IOC 检测配置。该命令只读取已有自定义情报，返回列表展示字段和分页信息；不新增、编辑、启停、删除、导入或导出情报。\n\n输出：实际筛选条件、total、page、page_size、current_count、has_more、custom_intelligence。",
		Example: "  chaitin-cli tanswer policy custom-intelligence --page-size 10\n" +
			"  chaitin-cli tanswer policy custom-intelligence --ioc evil.example --type domain --status enabled",
		RunE: func(cmd *cobra.Command, args []string) error {
			if policyOpts.page < 1 {
				return writePolicyError(cmd, "查询自定义情报", "chaitin-cli tanswer policy custom-intelligence", "INVALID_PAGE", "page must be greater than or equal to 1", false)
			}
			if policyOpts.pageSize < 1 || policyOpts.pageSize > 100 {
				return writePolicyError(cmd, "查询自定义情报", "chaitin-cli tanswer policy custom-intelligence", "INVALID_PAGE_SIZE", "page-size must be between 1 and 100", false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			req, err := buildPolicyCustomIntelligenceRequest(policyOpts)
			if err != nil {
				return writePolicyError(cmd, "查询自定义情报", "chaitin-cli tanswer policy custom-intelligence", "INVALID_POLICY_FILTER", err.Error(), false)
			}
			client := NewClient(cfg)
			var result policyListRPCResult
			if err := client.CallRPC(cmd.Context(), "AlarmService.SearchAlarmCustomIntelligenceList", req, &result); err != nil {
				return writePolicyError(cmd, "查询自定义情报", "chaitin-cli tanswer policy custom-intelligence", "POLICY_CUSTOM_INTELLIGENCE_FAILED", err.Error(), true)
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查询自定义情报",
				Command: "chaitin-cli tanswer policy custom-intelligence",
				Query: map[string]any{
					"filters":   policyCustomIntelligenceFilters(policyOpts),
					"page":      policyOpts.page,
					"page_size": policyOpts.pageSize,
				},
				Data: map[string]any{
					"total":               result.Total,
					"page":                policyOpts.page,
					"page_size":           policyOpts.pageSize,
					"current_count":       len(result.Data),
					"has_more":            int64(policyOpts.page*policyOpts.pageSize) < result.Total,
					"custom_intelligence": summarizePolicyCustomIntelligence(result.Data),
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	addPolicyCustomIntelligenceFlags(cmd, &policyOpts)
	return cmd
}

func newPolicyCustomIntelligenceCreateCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyCustomIntelligenceWriteOptions
	cmd := &cobra.Command{
		Use:   "custom-intelligence-create",
		Short: "新增自定义情报",
		Long: "新增自定义情报，用于维护自定义 IOC 检测配置。该命令是高影响写操作：默认只返回写入预览，必须使用 --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_CREATE 才会调用后端创建接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer policy custom-intelligence-create --name 恶意域名 --ioc evil.example --type domain --preview\n" +
			"  chaitin-cli tanswer policy custom-intelligence-create --name 恶意域名 --ioc evil.example --type domain --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_CREATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildPolicyCustomIntelligenceWriteRequest(policyOpts)
			if err != nil {
				return writePolicyError(cmd, "新增自定义情报", "chaitin-cli tanswer policy custom-intelligence-create", "INVALID_POLICY_CUSTOM_INTELLIGENCE_CREATE_REQUEST", err.Error(), false)
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "新增自定义情报预览",
				Command:       "chaitin-cli tanswer policy custom-intelligence-create",
				OperationType: "policy_custom_intelligence_create",
				RiskLevel:     "write_high",
				Target: map[string]any{
					"name": req["name"],
					"ioc":  req["ioc"],
					"type": req["type"],
				},
				ChangeSummary: map[string]any{"before": nil, "after": req},
				Impact:        map[string]any{"custom_intelligence_count": 1},
				RiskWarnings: []string{
					"将新增自定义 IOC 情报，可能影响后续告警命中。",
					"如果 IOC 与已有情报重复或字段不合法，后端会拒绝写入。",
				},
				ConfirmToken: policyCustomIntelligenceCreateConfirmToken,
			})
			if policyOpts.preview || strings.TrimSpace(policyOpts.confirm) == "" {
				return writePolicySuccess(cmd, "新增自定义情报预览", "chaitin-cli tanswer policy custom-intelligence-create", req, preview)
			}
			if err := ValidateWriteConfirmation(policyOpts.confirm, policyCustomIntelligenceCreateConfirmToken); err != nil {
				return writePolicyError(cmd, "新增自定义情报", "chaitin-cli tanswer policy custom-intelligence-create", "POLICY_CUSTOM_INTELLIGENCE_CREATE_CONFIRMATION_REQUIRED", err.Error(), false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var result struct {
				ID uint `json:"id"`
			}
			if err := client.CallRPC(cmd.Context(), "AlarmService.CreateAlarmCustomIntelligence", req, &result); err != nil {
				return writePolicyError(cmd, "新增自定义情报", "chaitin-cli tanswer policy custom-intelligence-create", "POLICY_CUSTOM_INTELLIGENCE_CREATE_FAILED", err.Error(), true)
			}
			object := map[string]any{"id": result.ID, "name": req["name"], "ioc": req["ioc"], "type": req["type"], "status": req["status"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "policy_custom_intelligence_create",
				Object:        object,
				Action:        "create",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   map[string]any{"before": nil, "after": req},
				Result:        "success",
			})
			return writePolicySuccess(cmd, "新增自定义情报", "chaitin-cli tanswer policy custom-intelligence-create", req, data)
		},
	}
	addPolicyCustomIntelligenceWriteFlags(cmd, &policyOpts, false, false)
	return cmd
}

func newPolicyCustomIntelligenceUpdateCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyCustomIntelligenceWriteOptions
	cmd := &cobra.Command{
		Use:   "custom-intelligence-update",
		Short: "编辑自定义情报",
		Long: "编辑自定义情报，用于更新单条自定义 IOC 检测配置。该命令是高影响写操作：预览阶段会读取当前情报并返回 before/after；必须使用 --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_UPDATE 才会调用后端更新接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer policy custom-intelligence-update --id 12 --name 新情报 --ioc evil.example --type domain --preview\n" +
			"  chaitin-cli tanswer policy custom-intelligence-update --id 12 --name 新情报 --ioc evil.example --type domain --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_UPDATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildPolicyCustomIntelligenceUpdateRequest(policyOpts)
			if err != nil {
				return writePolicyError(cmd, "编辑自定义情报", "chaitin-cli tanswer policy custom-intelligence-update", "INVALID_POLICY_CUSTOM_INTELLIGENCE_UPDATE_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(policyOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(policyOpts.confirm, policyCustomIntelligenceUpdateConfirmToken); err != nil {
					return writePolicyError(cmd, "编辑自定义情报", "chaitin-cli tanswer policy custom-intelligence-update", "POLICY_CUSTOM_INTELLIGENCE_UPDATE_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			before, err := fetchPolicyCustomIntelligenceInfo(cmd, client, req["id"].(uint))
			if err != nil {
				return writePolicyError(cmd, "编辑自定义情报", "chaitin-cli tanswer policy custom-intelligence-update", "POLICY_CUSTOM_INTELLIGENCE_UPDATE_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{"before": before, "after": req}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "编辑自定义情报预览",
				Command:       "chaitin-cli tanswer policy custom-intelligence-update",
				OperationType: "policy_custom_intelligence_update",
				RiskLevel:     "write_high",
				Target:        map[string]any{"id": req["id"]},
				ChangeSummary: changeSummary,
				Impact:        map[string]any{"custom_intelligence_count": 1},
				RiskWarnings: []string{
					"将更新自定义 IOC 情报，可能影响后续告警命中。",
					"如果 IOC 与其他情报冲突或字段不合法，后端会拒绝写入。",
				},
				ConfirmToken: policyCustomIntelligenceUpdateConfirmToken,
			})
			if policyOpts.preview || strings.TrimSpace(policyOpts.confirm) == "" {
				return writePolicySuccess(cmd, "编辑自定义情报预览", "chaitin-cli tanswer policy custom-intelligence-update", req, preview)
			}
			if err := client.CallRPC(cmd.Context(), "AlarmService.UpdateAlarmCustomIntelligence", req, &struct{}{}); err != nil {
				return writePolicyError(cmd, "编辑自定义情报", "chaitin-cli tanswer policy custom-intelligence-update", "POLICY_CUSTOM_INTELLIGENCE_UPDATE_FAILED", err.Error(), true)
			}
			object := map[string]any{"id": req["id"], "name": req["name"], "ioc": req["ioc"], "type": req["type"], "status": req["status"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "policy_custom_intelligence_update",
				Object:        object,
				Action:        "update",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   changeSummary,
				Result:        "success",
			})
			return writePolicySuccess(cmd, "编辑自定义情报", "chaitin-cli tanswer policy custom-intelligence-update", req, data)
		},
	}
	addPolicyCustomIntelligenceWriteFlags(cmd, &policyOpts, true, false)
	return cmd
}

func newPolicyCustomIntelligenceEnableCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyCustomIntelligenceWriteOptions
	return newPolicyCustomIntelligenceStatusCommand(opts, &policyOpts, "custom-intelligence-enable", "启用自定义情报", "enable", uint(1), policyCustomIntelligenceEnableConfirmToken)
}

func newPolicyCustomIntelligenceDisableCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyCustomIntelligenceWriteOptions
	return newPolicyCustomIntelligenceStatusCommand(opts, &policyOpts, "custom-intelligence-disable", "禁用自定义情报", "disable", uint(0), policyCustomIntelligenceDisableConfirmToken)
}

func newPolicyCustomIntelligenceStatusCommand(opts *RootOptions, policyOpts *policyCustomIntelligenceWriteOptions, use string, task string, action string, status uint, confirmToken string) *cobra.Command {
	commandText := "chaitin-cli tanswer policy " + use
	cmd := &cobra.Command{
		Use:     use,
		Short:   task,
		Long:    fmt.Sprintf("%s，用于控制自定义 IOC 情报启停状态。该命令是高影响写操作：预览阶段会读取目标情报摘要；必须使用 --confirm %s 才会调用后端状态更新接口。\n\n输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n执行输出：confirmed、result、object、audit。", task, confirmToken),
		Example: fmt.Sprintf("  %s --id-list 12,13 --preview\n  %s --id-list 12,13 --confirm %s", commandText, commandText, confirmToken),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseUintCSV(policyOpts.idList)
			if err != nil || len(ids) == 0 {
				return writePolicyError(cmd, task, commandText, "INVALID_POLICY_CUSTOM_INTELLIGENCE_STATUS_REQUEST", "missing or invalid id-list: set --id-list", false)
			}
			if strings.TrimSpace(policyOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(policyOpts.confirm, confirmToken); err != nil {
					return writePolicyError(cmd, task, commandText, "POLICY_CUSTOM_INTELLIGENCE_"+strings.ToUpper(action)+"_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			before, err := fetchPolicyCustomIntelligenceListByIDs(cmd, client, ids)
			if err != nil {
				return writePolicyError(cmd, task, commandText, "POLICY_CUSTOM_INTELLIGENCE_STATUS_PREVIEW_FAILED", err.Error(), true)
			}
			req := map[string]any{"ids": ids, "status": status}
			changeSummary := map[string]any{"before": before, "after": req}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          task + "预览",
				Command:       commandText,
				OperationType: "policy_custom_intelligence_" + action,
				RiskLevel:     "write_high",
				Target:        map[string]any{"ids": ids},
				ChangeSummary: changeSummary,
				Impact:        map[string]any{"custom_intelligence_count": len(ids), "target_status": status},
				RiskWarnings:  []string{"将改变自定义 IOC 情报状态，可能影响后续告警命中。"},
				ConfirmToken:  confirmToken,
			})
			if policyOpts.preview || strings.TrimSpace(policyOpts.confirm) == "" {
				return writePolicySuccess(cmd, task+"预览", commandText, req, preview)
			}
			if err := client.CallRPC(cmd.Context(), "AlarmService.UpdateAlarmCustomIntelligenceStatus", req, &struct{}{}); err != nil {
				return writePolicyError(cmd, task, commandText, "POLICY_CUSTOM_INTELLIGENCE_"+strings.ToUpper(action)+"_FAILED", err.Error(), true)
			}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "policy_custom_intelligence_" + action,
				Object:        map[string]any{"ids": ids, "status": status},
				Action:        action,
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   changeSummary,
				Result:        "success",
			})
			return writePolicySuccess(cmd, task, commandText, req, data)
		},
	}
	cmd.Flags().StringVar(&policyOpts.idList, "id-list", "", "custom intelligence IDs, comma separated, required")
	cmd.Flags().BoolVar(&policyOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&policyOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newPolicyCustomIntelligenceDeleteCommand(opts *RootOptions) *cobra.Command {
	var policyOpts policyCustomIntelligenceWriteOptions
	commandText := "chaitin-cli tanswer policy custom-intelligence-delete"
	cmd := &cobra.Command{
		Use:   "custom-intelligence-delete",
		Short: "删除自定义情报",
		Long: "删除自定义情报，用于移除不再需要的自定义 IOC 检测配置。该命令是高影响写操作：预览阶段会读取目标情报摘要；必须使用 --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_DELETE 才会调用后端删除接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer policy custom-intelligence-delete --id-list 12,13 --preview\n" +
			"  chaitin-cli tanswer policy custom-intelligence-delete --id-list 12,13 --confirm CONFIRM_POLICY_CUSTOM_INTELLIGENCE_DELETE",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseUintCSV(policyOpts.idList)
			if err != nil || len(ids) == 0 {
				return writePolicyError(cmd, "删除自定义情报", commandText, "INVALID_POLICY_CUSTOM_INTELLIGENCE_DELETE_REQUEST", "missing or invalid id-list: set --id-list", false)
			}
			if strings.TrimSpace(policyOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(policyOpts.confirm, policyCustomIntelligenceDeleteConfirmToken); err != nil {
					return writePolicyError(cmd, "删除自定义情报", commandText, "POLICY_CUSTOM_INTELLIGENCE_DELETE_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			before, err := fetchPolicyCustomIntelligenceListByIDs(cmd, client, ids)
			if err != nil {
				return writePolicyError(cmd, "删除自定义情报", commandText, "POLICY_CUSTOM_INTELLIGENCE_DELETE_PREVIEW_FAILED", err.Error(), true)
			}
			req := map[string]any{"ids": ids}
			changeSummary := map[string]any{"before": before, "after": nil}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "删除自定义情报预览",
				Command:       commandText,
				OperationType: "policy_custom_intelligence_delete",
				RiskLevel:     "write_high",
				Target:        map[string]any{"ids": ids},
				ChangeSummary: changeSummary,
				Impact:        map[string]any{"custom_intelligence_count": len(ids)},
				RiskWarnings:  []string{"将删除自定义 IOC 情报，可能影响后续告警命中。"},
				ConfirmToken:  policyCustomIntelligenceDeleteConfirmToken,
			})
			if policyOpts.preview || strings.TrimSpace(policyOpts.confirm) == "" {
				return writePolicySuccess(cmd, "删除自定义情报预览", commandText, req, preview)
			}
			if err := client.CallRPC(cmd.Context(), "AlarmService.DeleteAlarmCustomIntelligence", req, &struct{}{}); err != nil {
				return writePolicyError(cmd, "删除自定义情报", commandText, "POLICY_CUSTOM_INTELLIGENCE_DELETE_FAILED", err.Error(), true)
			}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "policy_custom_intelligence_delete",
				Object:        map[string]any{"ids": ids},
				Action:        "delete",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   changeSummary,
				Result:        "success",
			})
			return writePolicySuccess(cmd, "删除自定义情报", commandText, req, data)
		},
	}
	cmd.Flags().StringVar(&policyOpts.idList, "id-list", "", "custom intelligence IDs, comma separated, required")
	cmd.Flags().BoolVar(&policyOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&policyOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func addPolicyDetectionWhitelistFlags(cmd *cobra.Command, opts *policyDetectionWhitelistOptions) {
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().StringVar(&opts.name, "name", "", "detection whitelist name filter")
	cmd.Flags().StringVar(&opts.srcIP, "src-ip", "", "source IP filter, comma separated")
	cmd.Flags().StringVar(&opts.srcPort, "src-port", "", "source port filter, comma separated")
	cmd.Flags().StringVar(&opts.destIP, "dest-ip", "", "destination IP filter, comma separated")
	cmd.Flags().StringVar(&opts.destPort, "dest-port", "", "destination port filter, comma separated")
	cmd.Flags().StringVar(&opts.domain, "domain", "", "domain filter, comma separated")
	cmd.Flags().StringVar(&opts.urlPath, "url-path", "", "URL path filter, comma separated")
	cmd.Flags().StringVar(&opts.userAgent, "user-agent", "", "User-Agent filter, comma separated")
	cmd.Flags().StringVar(&opts.xff, "xff", "", "XFF filter, comma separated")
	cmd.Flags().StringVar(&opts.respCode, "resp-code", "", "response status code filter, comma separated")
	cmd.Flags().StringVar(&opts.respBody, "resp-body", "", "response body filter, comma separated")
	cmd.Flags().StringVar(&opts.threat, "threat", "", "threat type filter")
	cmd.Flags().StringVar(&opts.ruleID, "rule-id", "", "detection rule id filter")
	cmd.Flags().StringVar(&opts.status, "status", "", "status filter: enabled, disabled, 1, 0")
}

func addPolicyCustomIntelligenceFlags(cmd *cobra.Command, opts *policyCustomIntelligenceOptions) {
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().StringVar(&opts.id, "id", "", "custom intelligence id filter")
	cmd.Flags().StringVar(&opts.name, "name", "", "custom intelligence name filter")
	cmd.Flags().StringVar(&opts.ioc, "ioc", "", "IOC filter, comma separated")
	cmd.Flags().StringVar(&opts.iocType, "type", "", "IOC type: ip, domain, url, md5, sha1, sha256, or 1-6")
	cmd.Flags().StringVar(&opts.status, "status", "", "status filter: enabled, disabled, 1, 0")
	cmd.Flags().StringVar(&opts.remarks, "remarks", "", "remarks filter")
}

func addPolicyCustomIntelligenceWriteFlags(cmd *cobra.Command, opts *policyCustomIntelligenceWriteOptions, includeID bool, includeIDList bool) {
	if includeID {
		cmd.Flags().StringVar(&opts.id, "id", "", "custom intelligence ID, required")
	}
	if includeIDList {
		cmd.Flags().StringVar(&opts.idList, "id-list", "", "custom intelligence IDs, comma separated")
	}
	cmd.Flags().StringVar(&opts.name, "name", "", "custom intelligence name, required")
	cmd.Flags().StringVar(&opts.ioc, "ioc", "", "IOC values, comma separated, required")
	cmd.Flags().StringVar(&opts.iocType, "type", "", "IOC type: ip, domain, url, md5, sha1, sha256, or 1-6, required")
	cmd.Flags().StringVar(&opts.status, "status", "enabled", "custom intelligence status: enabled, disabled, 1, or 0")
	cmd.Flags().StringVar(&opts.remarks, "remarks", "", "custom intelligence remarks")
	cmd.Flags().BoolVar(&opts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&opts.confirm, "confirm", "", "exact confirmation token required to execute")
}

func addPolicyDetectionWhitelistWriteFlags(cmd *cobra.Command, opts *policyDetectionWhitelistWriteOptions, includeID bool) {
	if includeID {
		cmd.Flags().StringVar(&opts.id, "id", "", "detection whitelist ID, required")
	}
	cmd.Flags().StringVar(&opts.name, "name", "", "detection whitelist name, required")
	cmd.Flags().StringVar(&opts.srcIP, "src-ip", "", "source IP")
	cmd.Flags().StringVar(&opts.srcPort, "src-port", "", "source port")
	cmd.Flags().StringVar(&opts.destIP, "dest-ip", "", "destination IP")
	cmd.Flags().StringVar(&opts.destPort, "dest-port", "", "destination port")
	cmd.Flags().StringVar(&opts.domain, "domain", "", "domain")
	cmd.Flags().StringVar(&opts.urlPath, "url-path", "", "URL path")
	cmd.Flags().StringVar(&opts.userAgent, "user-agent", "", "User-Agent")
	cmd.Flags().StringVar(&opts.xff, "xff", "", "XFF")
	cmd.Flags().StringVar(&opts.respCode, "resp-code", "", "response status code")
	cmd.Flags().StringVar(&opts.respBody, "resp-body", "", "response body")
	cmd.Flags().StringVar(&opts.threat, "threat", "", "threat type")
	cmd.Flags().StringVar(&opts.ruleID, "rule-id", "", "detection rule id")
	cmd.Flags().StringVar(&opts.status, "status", "enabled", "status: enabled, disabled, 1, or 0")
	cmd.Flags().StringVar(&opts.storage, "storage", "drop", "handling mode: drop, ignore, 1, or 2")
	cmd.Flags().StringVar(&opts.defaultMode, "mode", "default", "match mode: default, advanced, 1, or 2")
	cmd.Flags().StringVar(&opts.expire, "expire", "", "expire timestamp in milliseconds; provide this or --valid-time")
	cmd.Flags().StringVar(&opts.validTime, "valid-time", "", "valid duration in seconds; must be greater than 0; provide this or --expire")
	cmd.Flags().BoolVar(&opts.ignore, "ignore-history", false, "ignore matched historical alarms")
	cmd.Flags().StringVar(&opts.remark, "remark", "", "remark")
	cmd.Flags().StringVar(&opts.srcAdvanced, "src-advanced", "", "advanced source IP:port rules, comma separated")
	cmd.Flags().StringVar(&opts.destAdvanced, "dest-advanced", "", "advanced destination IP:port rules, comma separated")
	cmd.Flags().StringVar(&opts.sidAdvanced, "sid-advanced", "", "advanced detection rule IDs, comma separated")
	cmd.Flags().StringVar(&opts.typeAdvanced, "type-advanced", "", "advanced threat types, comma separated")
	cmd.Flags().BoolVar(&opts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&opts.confirm, "confirm", "", "exact confirmation token required to execute")
}

func buildPolicyDetectionWhitelistRequest(opts policyDetectionWhitelistOptions) map[string]any {
	if opts.page < 1 {
		opts.page = 1
	}
	if opts.pageSize < 1 {
		opts.pageSize = 10
	}
	req := map[string]any{
		"offset": int64((opts.page - 1) * opts.pageSize),
		"count":  int64(opts.pageSize),
	}
	for key, value := range policyDetectionWhitelistFilters(opts) {
		req[key] = value
	}
	return req
}

func policyDetectionWhitelistFilters(opts policyDetectionWhitelistOptions) map[string]any {
	filters := map[string]any{}
	addStringFilter(filters, "name", opts.name, strings.TrimSpace)
	addStringFilter(filters, "src_ip", opts.srcIP, strings.TrimSpace)
	addStringFilter(filters, "src_port", opts.srcPort, strings.TrimSpace)
	addStringFilter(filters, "dest_ip", opts.destIP, strings.TrimSpace)
	addStringFilter(filters, "dest_port", opts.destPort, strings.TrimSpace)
	addStringFilter(filters, "domain", opts.domain, strings.TrimSpace)
	addStringFilter(filters, "url_path", opts.urlPath, strings.TrimSpace)
	addStringFilter(filters, "user_agent", opts.userAgent, strings.TrimSpace)
	addStringFilter(filters, "xff", opts.xff, strings.TrimSpace)
	addStringFilter(filters, "resp_status_code", opts.respCode, strings.TrimSpace)
	addStringFilter(filters, "resp_body", opts.respBody, strings.TrimSpace)
	addIntFilter(filters, "status", opts.status, statusValue)
	if value := strings.TrimSpace(opts.ruleID); value != "" {
		filters["sid"] = value
	}
	if value := strings.TrimSpace(opts.threat); value != "" {
		filters["type"] = value
	}
	return filters
}

func buildPolicyCustomIntelligenceRequest(opts policyCustomIntelligenceOptions) (map[string]any, error) {
	if opts.page < 1 {
		opts.page = 1
	}
	if opts.pageSize < 1 {
		opts.pageSize = 10
	}
	req := map[string]any{
		"offset": int64((opts.page - 1) * opts.pageSize),
		"count":  int64(opts.pageSize),
	}
	filters, err := policyCustomIntelligenceRequestFilters(opts)
	if err != nil {
		return nil, err
	}
	for key, value := range filters {
		req[key] = value
	}
	return req, nil
}

func policyCustomIntelligenceFilters(opts policyCustomIntelligenceOptions) map[string]any {
	filters := map[string]any{}
	if opts.id != "" {
		filters["id"] = opts.id
	}
	if opts.name != "" {
		filters["name"] = opts.name
	}
	if opts.ioc != "" {
		filters["ioc"] = opts.ioc
	}
	if opts.iocType != "" {
		filters["type"] = opts.iocType
	}
	if opts.status != "" {
		filters["status"] = opts.status
	}
	if opts.remarks != "" {
		filters["remarks"] = opts.remarks
	}
	return filters
}

func policyCustomIntelligenceRequestFilters(opts policyCustomIntelligenceOptions) (map[string]any, error) {
	filters := map[string]any{}
	addIntFilter(filters, "id", opts.id, portValue)
	addStringFilter(filters, "name", opts.name, strings.TrimSpace)
	addStringFilter(filters, "ioc", opts.ioc, strings.TrimSpace)
	addIntFilter(filters, "status", opts.status, statusValue)
	addStringFilter(filters, "remarks", opts.remarks, strings.TrimSpace)
	if strings.TrimSpace(opts.iocType) != "" {
		iocType, err := intelligenceTypeValue(opts.iocType)
		if err != nil {
			return nil, err
		}
		filters["type"] = []map[string]any{{"oper": "=", "target": iocType}}
	}
	return filters, nil
}

func buildPolicyDetectionWhitelistWriteRequest(opts policyDetectionWhitelistWriteOptions) (map[string]any, error) {
	name := strings.TrimSpace(opts.name)
	if name == "" {
		return nil, fmt.Errorf("missing detection whitelist name: set --name")
	}
	status := uint(statusValue(opts.status))
	if status != 0 && status != 1 {
		return nil, fmt.Errorf("unsupported status %q, expected enabled, disabled, 1, or 0", opts.status)
	}
	storage, err := policyDetectionWhitelistStorageValue(opts.storage)
	if err != nil {
		return nil, err
	}
	defaultMode, err := policyDetectionWhitelistDefaultModeValue(opts.defaultMode)
	if err != nil {
		return nil, err
	}
	expire, validTime, err := policyDetectionWhitelistExpiry(opts)
	if err != nil {
		return nil, err
	}
	req := map[string]any{
		"name":             name,
		"src_ip":           strings.TrimSpace(opts.srcIP),
		"src_port":         strings.TrimSpace(opts.srcPort),
		"dest_ip":          strings.TrimSpace(opts.destIP),
		"dest_port":        strings.TrimSpace(opts.destPort),
		"domain":           strings.TrimSpace(opts.domain),
		"url_path":         normalizeURLPath(opts.urlPath),
		"user_agent":       strings.TrimSpace(opts.userAgent),
		"xff":              strings.TrimSpace(opts.xff),
		"resp_status_code": strings.ToLower(strings.TrimSpace(opts.respCode)),
		"resp_body":        strings.TrimSpace(opts.respBody),
		"type":             strings.TrimSpace(opts.threat),
		"sid":              strings.TrimSpace(opts.ruleID),
		"remark":           strings.TrimSpace(opts.remark),
		"expire":           expire,
		"ignore":           opts.ignore,
		"storage":          storage,
		"status":           status,
		"valid_time":       validTime,
		"default_mode":     defaultMode,
		"src_advanced":     parseCSV(opts.srcAdvanced),
		"dest_advanced":    parseCSV(opts.destAdvanced),
		"sid_advanced":     parseCSV(opts.sidAdvanced),
		"type_advanced":    parseCSV(opts.typeAdvanced),
	}
	if !policyDetectionWhitelistHasMatch(req) {
		return nil, fmt.Errorf("missing detection whitelist match condition: set at least one simple or advanced match field")
	}
	return req, nil
}

func policyDetectionWhitelistExpiry(opts policyDetectionWhitelistWriteOptions) (int64, int64, error) {
	expire, err := optionalInt64(opts.expire, "expire")
	if err != nil {
		return 0, 0, err
	}
	if strings.TrimSpace(opts.expire) != "" && expire <= 0 {
		return 0, 0, fmt.Errorf("expire must be greater than 0")
	}
	validTime, err := optionalInt64(opts.validTime, "valid-time")
	if err != nil {
		return 0, 0, err
	}
	if strings.TrimSpace(opts.validTime) != "" && validTime <= 0 {
		return 0, 0, fmt.Errorf("valid-time must be greater than 0")
	}
	if expire <= 0 && validTime <= 0 {
		return 0, 0, fmt.Errorf("missing detection whitelist expiry: set --expire or --valid-time")
	}
	return expire, validTime, nil
}

func buildPolicyDetectionWhitelistUpdateRequest(opts policyDetectionWhitelistWriteOptions) (map[string]any, error) {
	id, err := parseRequiredPolicyObjectID(opts.id, "detection whitelist")
	if err != nil {
		return nil, err
	}
	req, err := buildPolicyDetectionWhitelistWriteRequest(opts)
	if err != nil {
		return nil, err
	}
	req["id"] = id
	return req, nil
}

func buildPolicyDetectionWhitelistFromAlarmRequest(alarm map[string]any, opts policyDetectionWhitelistFromAlarmOptions) (map[string]any, error) {
	writeOpts := opts.write
	alarmName := firstPolicyString(alarm, "name", "msg", "tag")
	if strings.TrimSpace(writeOpts.name) == "" {
		nameBase := alarmName
		if nameBase == "" {
			nameBase = strings.TrimSpace(opts.id)
		}
		if nameBase == "" {
			nameBase = "告警"
		}
		writeOpts.name = "误报白名单-" + nameBase
	}
	if strings.TrimSpace(writeOpts.srcIP) == "" {
		writeOpts.srcIP = firstPolicyString(alarm, "src_ip", "attacker")
	}
	if strings.TrimSpace(writeOpts.srcPort) == "" {
		writeOpts.srcPort = firstPolicyScalarString(alarm, "src_port", "attacker_port")
	}
	if strings.TrimSpace(writeOpts.destIP) == "" {
		writeOpts.destIP = firstPolicyString(alarm, "dest_ip", "victim")
	}
	if strings.TrimSpace(writeOpts.destPort) == "" {
		writeOpts.destPort = firstPolicyScalarString(alarm, "dest_port", "victim_port")
	}
	if strings.TrimSpace(writeOpts.domain) == "" {
		writeOpts.domain = firstPolicyString(alarm, "domain", "host", "hostname")
		if writeOpts.domain == "" {
			writeOpts.domain = nestedPolicyString(alarm, "appbrief", "http", "hostname")
		}
	}
	if strings.TrimSpace(writeOpts.urlPath) == "" {
		writeOpts.urlPath = firstPolicyString(alarm, "url", "url_path")
		if writeOpts.urlPath == "" {
			writeOpts.urlPath = nestedPolicyString(alarm, "appbrief", "http", "url")
		}
	}
	if strings.TrimSpace(writeOpts.userAgent) == "" {
		writeOpts.userAgent = firstPolicyString(alarm, "user_agent")
		if writeOpts.userAgent == "" {
			writeOpts.userAgent = nestedPolicyString(alarm, "appbrief", "http", "user_agent")
		}
	}
	if strings.TrimSpace(writeOpts.xff) == "" {
		writeOpts.xff = firstPolicyString(alarm, "xff")
	}
	if strings.TrimSpace(writeOpts.respCode) == "" {
		writeOpts.respCode = firstPolicyScalarString(alarm, "resp_status_code", "status_code")
		if writeOpts.respCode == "" {
			writeOpts.respCode = nestedPolicyScalarString(alarm, "appbrief", "http", "status_code")
		}
	}
	if strings.TrimSpace(writeOpts.respBody) == "" {
		writeOpts.respBody = firstPolicyString(alarm, "resp_body")
	}
	if strings.TrimSpace(writeOpts.threat) == "" {
		writeOpts.threat = firstPolicyString(alarm, "tag", "type", "name")
	}
	if strings.TrimSpace(writeOpts.ruleID) == "" {
		writeOpts.ruleID = firstPolicyString(alarm, "sid")
	}
	return buildPolicyDetectionWhitelistWriteRequest(writeOpts)
}

func buildPolicyExportRequest(opts policyFileOptions) (map[string]any, string, error) {
	ids, err := parseUintCSV(opts.idList)
	if err != nil {
		return nil, "", fmt.Errorf("invalid id-list: %w", err)
	}
	scope := "all"
	if len(ids) > 0 {
		scope = "selected"
	}
	return map[string]any{"ids": ids}, scope, nil
}

func buildPolicyImportRequest(opts policyFileOptions, method string) (map[string]any, error) {
	path := strings.TrimSpace(opts.file)
	if path == "" {
		return nil, fmt.Errorf("missing import file: set --file")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("import file must not be a directory")
	}
	return map[string]any{
		"file_path":  path,
		"file_name":  filepath.Base(path),
		"size_bytes": info.Size(),
		"method":     method,
	}, nil
}

func downloadPolicyFile(cmd *cobra.Command, cfg Config, task string, command string, method string, query map[string]any, outputPath string, defaultName string, exportScope string, scopeName string) error {
	client := NewClient(cfg)
	file, err := client.Download(cmd.Context(), method, query)
	if err != nil {
		return writePolicyError(cmd, task, command, "POLICY_FILE_DOWNLOAD_FAILED", err.Error(), true)
	}
	if file.StatusCode < 200 || file.StatusCode >= 300 {
		return writePolicyError(cmd, task, command, "POLICY_FILE_DOWNLOAD_FAILED", fmt.Sprintf("download returned HTTP %d: %s", file.StatusCode, strings.TrimSpace(string(file.Body))), true)
	}
	target, fileName := resolveDownloadOutputPath(outputPath, file.FileName, defaultName)
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return writePolicyError(cmd, task, command, "POLICY_FILE_WRITE_FAILED", err.Error(), false)
	}
	if err := os.WriteFile(target, file.Body, 0644); err != nil {
		return writePolicyError(cmd, task, command, "POLICY_FILE_WRITE_FAILED", err.Error(), false)
	}
	data := map[string]any{
		"file_name":      fileName,
		"file_path":      target,
		"size_bytes":     len(file.Body),
		"status_code":    file.StatusCode,
		"method":         method,
		"download_query": query,
		"export_scope":   exportScope,
		"scope":          scopeName,
	}
	return writePolicySuccess(cmd, task, command, query, data)
}

func normalizeURLPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "/") {
		return value
	}
	return "/" + value
}

func optionalInt64(value string, name string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("invalid %s: %q", name, value)
	}
	return parsed, nil
}

func policyDetectionWhitelistStorageValue(value string) (uint, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "drop", "no-store", "not-store", "1":
		return 1, nil
	case "ignore", "store-ignore", "store", "2":
		return 2, nil
	default:
		return 0, fmt.Errorf("unsupported storage %q, expected drop, ignore, 1, or 2", value)
	}
}

func policyDetectionWhitelistDefaultModeValue(value string) (uint, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "default", "normal", "simple", "1":
		return 1, nil
	case "advanced", "2":
		return 2, nil
	default:
		return 0, fmt.Errorf("unsupported mode %q, expected default, advanced, 1, or 2", value)
	}
}

func policyDetectionWhitelistHasMatch(req map[string]any) bool {
	for _, key := range []string{"src_ip", "src_port", "dest_ip", "dest_port", "domain", "url_path", "user_agent", "xff", "resp_status_code", "resp_body", "type", "sid"} {
		if strings.TrimSpace(fmt.Sprint(req[key])) != "" {
			return true
		}
	}
	for _, key := range []string{"src_advanced", "dest_advanced", "sid_advanced", "type_advanced"} {
		if values, ok := req[key].([]string); ok && len(values) > 0 {
			return true
		}
	}
	return false
}

func policyDetectionWhitelistMatchSummary(req map[string]any) map[string]any {
	out := map[string]any{}
	for _, key := range []string{"src_ip", "src_port", "dest_ip", "dest_port", "domain", "url_path", "user_agent", "xff", "resp_status_code", "resp_body", "type", "sid", "src_advanced", "dest_advanced", "sid_advanced", "type_advanced"} {
		if value, ok := req[key]; ok {
			out[key] = value
		}
	}
	return out
}

func firstPolicyString(item map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(fmt.Sprint(item[key])); value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func firstPolicyScalarString(item map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := policyScalarString(item[key]); value != "" {
			return value
		}
	}
	return ""
}

func nestedPolicyString(item map[string]any, keys ...string) string {
	value := nestedPolicyValue(item, keys...)
	if text := strings.TrimSpace(fmt.Sprint(value)); text != "" && text != "<nil>" {
		return text
	}
	return ""
}

func nestedPolicyScalarString(item map[string]any, keys ...string) string {
	return policyScalarString(nestedPolicyValue(item, keys...))
}

func nestedPolicyValue(item map[string]any, keys ...string) any {
	var current any = item
	for _, key := range keys {
		asMap, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = asMap[key]
	}
	return current
}

func policyScalarString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case int:
		if typed == 0 {
			return ""
		}
		return strconv.Itoa(typed)
	case int64:
		if typed == 0 {
			return ""
		}
		return strconv.FormatInt(typed, 10)
	case uint:
		if typed == 0 {
			return ""
		}
		return strconv.FormatUint(uint64(typed), 10)
	case float64:
		if typed == 0 {
			return ""
		}
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(typed, 'f', 6, 64), "0"), ".")
	default:
		text := strings.TrimSpace(fmt.Sprint(value))
		if text == "<nil>" || text == "0" {
			return ""
		}
		return text
	}
}

func buildPolicyCustomIntelligenceWriteRequest(opts policyCustomIntelligenceWriteOptions) (map[string]any, error) {
	name := strings.TrimSpace(opts.name)
	if name == "" {
		return nil, fmt.Errorf("missing custom intelligence name: set --name")
	}
	iocs := parseCSV(opts.ioc)
	if len(iocs) == 0 {
		return nil, fmt.Errorf("missing IOC: set --ioc")
	}
	iocType, err := intelligenceTypeValue(opts.iocType)
	if err != nil {
		return nil, err
	}
	status := uint(statusValue(opts.status))
	if status != 0 && status != 1 {
		return nil, fmt.Errorf("unsupported status %q, expected enabled, disabled, 1, or 0", opts.status)
	}
	return map[string]any{
		"name":    name,
		"ioc":     iocs,
		"type":    iocType,
		"status":  status,
		"remarks": strings.TrimSpace(opts.remarks),
	}, nil
}

func buildPolicyCustomIntelligenceUpdateRequest(opts policyCustomIntelligenceWriteOptions) (map[string]any, error) {
	id, err := parseRequiredPolicyID(opts.id)
	if err != nil {
		return nil, err
	}
	req, err := buildPolicyCustomIntelligenceWriteRequest(opts)
	if err != nil {
		return nil, err
	}
	req["id"] = id
	return req, nil
}

func parseRequiredPolicyID(value string) (uint, error) {
	return parseRequiredPolicyObjectID(value, "custom intelligence")
}

func parseRequiredPolicyObjectID(value string, objectName string) (uint, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("missing %s id: set --id", objectName)
	}
	parsed, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("invalid %s id: %q", objectName, value)
	}
	return uint(parsed), nil
}

func fetchPolicyCustomIntelligenceInfo(cmd *cobra.Command, client *Client, id uint) (map[string]any, error) {
	var result struct {
		Data map[string]any `json:"data"`
	}
	if err := client.CallRPC(cmd.Context(), "AlarmService.SearchAlarmCustomIntelligenceInfo", map[string]any{"id": id}, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func fetchPolicyCustomIntelligenceListByIDs(cmd *cobra.Command, client *Client, ids []uint) ([]map[string]any, error) {
	req := map[string]any{
		"offset": int64(0),
		"count":  int64(len(ids)),
		"id":     uintEqualityQuery(ids),
	}
	var result policyListRPCResult
	if err := client.CallRPC(cmd.Context(), "AlarmService.SearchAlarmCustomIntelligenceList", req, &result); err != nil {
		return nil, err
	}
	return summarizePolicyCustomIntelligence(result.Data), nil
}

func fetchPolicyDetectionWhitelistsByIDs(cmd *cobra.Command, client *Client, ids []uint) ([]map[string]any, error) {
	req := map[string]any{
		"offset": int64(0),
		"count":  int64(len(ids)),
		"id":     uintEqualityQuery(ids),
	}
	var result policyListRPCResult
	if err := client.CallRPC(cmd.Context(), "AlarmService.SearchWhiteList", req, &result); err != nil {
		return nil, err
	}
	return summarizePolicyDetectionWhitelists(result.Data), nil
}

func summarizePolicyDetectionWhitelists(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		summary := map[string]any{}
		for _, key := range []string{
			"id",
			"name",
			"src_ip",
			"src_port",
			"dest_ip",
			"dest_port",
			"domain",
			"url_path",
			"user_agent",
			"xff",
			"resp_status_code",
			"resp_body",
			"type",
			"sid",
			"updated_at",
			"expire",
			"expired",
			"status",
			"remark",
			"storage",
			"default_mode",
			"src_advanced",
			"dest_advanced",
			"sid_advanced",
			"type_advanced",
		} {
			copyIfPresent(summary, item, key)
		}
		out = append(out, summary)
	}
	return out
}

func summarizePolicyCustomIntelligence(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		summary := map[string]any{}
		for _, key := range []string{
			"id",
			"name",
			"ioc",
			"type",
			"remarks",
			"updated_at",
			"created_at",
			"status",
		} {
			copyIfPresent(summary, item, key)
		}
		out = append(out, summary)
	}
	return out
}

func statusValue(value string) int {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "enabled", "enable", "启用", "1":
		return 1
	case "disabled", "disable", "禁用", "0":
		return 0
	default:
		n, _ := strconv.Atoi(strings.TrimSpace(value))
		return n
	}
}

func intelligenceTypeValue(value string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "ip", "ip地址", "1":
		return 1, nil
	case "domain", "domain_name", "域名", "2":
		return 2, nil
	case "url", "3":
		return 3, nil
	case "md5", "4":
		return 4, nil
	case "sha1", "5":
		return 5, nil
	case "sha256", "6":
		return 6, nil
	default:
		return 0, fmt.Errorf("unsupported IOC type %q, expected ip, domain, url, md5, sha1, sha256, or 1-6", value)
	}
}

func writePolicySuccess(cmd *cobra.Command, task string, command string, query any, data any) error {
	raw, err := RenderJSON(SuccessEnvelope{
		Success: true,
		Task:    task,
		Command: command,
		Query:   query,
		Data:    data,
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func writePolicyError(cmd *cobra.Command, task string, command string, code string, message string, retryable bool) error {
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
