package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaitin/chaitin-cli/products/xray/client"
	"github.com/chaitin/chaitin-cli/products/xray/client/plan"
	"github.com/chaitin/chaitin-cli/products/xray/client/template"
	"github.com/chaitin/chaitin-cli/products/xray/models"

	"github.com/spf13/cobra"
)

const defaultBuiltinTemplateName = "基础服务漏洞扫描"

// makeOperationPlanCreateQuickCmd returns a command to handle quick plan creation
func makeOperationPlanCreateQuickCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "PostPlanCreateQuick",
		Short: "快速创建立即或定时扫描任务",
		Long: fmt.Sprintf(`快速创建立即、单次定时或周期扫描任务。

排期参数：
  NOW         不需要排期参数，创建后立即执行
  CLOCKED     --exec-at 使用带时区 RFC3339，例如 2026-08-05T10:30:00+08:00
  DAY         --exec-at 使用 HH:MM 或 HH:MM:SS
  WEEK        额外要求 --weekday，1=周一 ... 7=周日
  MONTH       额外要求 --day-of-month，范围 1-31
  MONTH_WEEK  额外要求 --week-of-month 1-4 和 --weekday 1-7

--exec-at 表示任务触发时间，不是扫描结束时间。

示例：
  xray plan PostPlanCreateQuick --targets=example.com --engines=engine1 --project-id=1
  xray plan PostPlanCreateQuick --targets=example.com --engines=engine1 --project-id=1 --plan-type=CLOCKED --exec-at=2026-08-05T10:30:00+08:00
  xray plan PostPlanCreateQuick --targets=example.com --engines=engine1 --project-id=1 --plan-type=DAY --exec-at=10:30
  xray plan PostPlanCreateQuick --targets=example.com --engines=engine1 --project-id=1 --plan-type=WEEK --weekday=5 --exec-at=10:30
  xray plan PostPlanCreateQuick --targets=example.com --engines=engine1 --project-id=1 --plan-type=MONTH --day-of-month=15 --exec-at=10:30
  xray plan PostPlanCreateQuick --targets=example.com --engines=engine1 --project-id=1 --plan-type=MONTH_WEEK --week-of-month=2 --weekday=5 --exec-at=10:30

默认使用"基础服务漏洞扫描"(BUILTIN)模板。`),
		RunE: runOperationPlanQuick,
	}

	cmd.Flags().String("targets", "", "目标地址，逗号分隔 (必填)")
	cmd.Flags().String("engines", "", "引擎 ID 列表，逗号分隔 (必填)")
	cmd.Flags().String("name", "quick-scan", "任务名称")
	cmd.Flags().Int64("project-id", 0, "工作区 ID (必填)")
	cmd.Flags().String("template-name", "", fmt.Sprintf("模板名称 (默认: %s)", defaultBuiltinTemplateName))
	registerQuickScheduleFlags(cmd, planTypeNow)

	return cmd, nil
}

func runOperationPlanQuick(cmd *cobra.Command, args []string) error {
	appCli, err := makeClient(cmd, args)
	if err != nil {
		return err
	}

	targetsStr, _ := cmd.Flags().GetString("targets")
	enginesStr, _ := cmd.Flags().GetString("engines")
	name, _ := cmd.Flags().GetString("name")
	projectID, _ := cmd.Flags().GetInt64("project-id")
	templateNameFlag, _ := cmd.Flags().GetString("template-name")
	scheduleInput, err := readQuickScheduleInput(cmd)
	if err != nil {
		return err
	}
	planSetting, err := buildPlanSetting(scheduleInput, nil)
	if err != nil {
		return err
	}

	targets := parseCommaSeparated(targetsStr)
	engines := parseCommaSeparated(enginesStr)

	if len(targets) == 0 {
		return fmt.Errorf("targets is required")
	}
	if len(engines) == 0 {
		return fmt.Errorf("engines is required")
	}
	if projectID == 0 {
		return fmt.Errorf("project-id is required")
	}

	// Find template and get task_setting
	templateName := templateNameFlag
	if templateName == "" {
		templateName = defaultBuiltinTemplateName
	}

	templateID, taskSetting, allowedPlanTypes, err := findBuiltinTemplateWithTaskSetting(appCli, templateName)
	if err != nil {
		return err
	}
	if templateID == 0 {
		return fmt.Errorf("未找到模板: %s", templateName)
	}
	if err := validateSupportedPlanType(scheduleInput.planType, allowedPlanTypes); err != nil {
		return err
	}
	logDebugf("Found template ID %d for '%s'", templateID, templateName)

	// Build basic_setting
	basicSetting := map[string]interface{}{
		"remark": "",
		"taskTarget": map[string]interface{}{
			"targetType": "MANUAL",
			"target":     targets,
		},
		"globalWhiteList": []interface{}{},
		"templateSync":    false,
		"executionSetting": map[string]interface{}{
			"enabled":    false,
			"rule":       "ALLOW",
			"timeRanges": []interface{}{map[string]interface{}{}},
			"timeType":   "DAY",
		},
		"planSetting":  planSetting,
		"engineChoice": engines,
		"taskName":     name,
	}

	active := true
	execRightNow := scheduleInput.planType == planTypeNow

	body := &models.CreatePlanBody{
		Active:               &active,
		BasicSetting:         basicSetting,
		DisabledWhitelistIds: []int64{}, // empty array instead of nil to satisfy API validation
		ExecRightNow:         execRightNow,
		ProjectID:            projectID,
		TaskSetting:          taskSetting,
		TaskTemplateID:       &templateID,
	}

	params := plan.NewPostPlanCreateParams()
	params.Body = body

	if dryRun {
		logDebugf("dry-run flag specified. Skip sending request.")
		debugBytes, _ := json.MarshalIndent(body, "", "  ")
		logDebugf("Request body: %v", string(debugBytes))
		return nil
	}

	msgStr, err := parseOperationPlanPostPlanCreateResult(appCli.Plan.PostPlanCreate(params, nil))
	if err != nil {
		return err
	}

	if !debug {
		fmt.Println(msgStr)
	}

	return nil
}

// findBuiltinTemplateWithTaskSetting finds a matching built-in template and
// returns the values needed to build and validate a plan request.
func findBuiltinTemplateWithTaskSetting(appCli *client.OPENAPI, name string) (int64, interface{}, []string, error) {
	params := template.NewGetTemplateSummaryParams()
	params.Limit = 100
	params.Offset = 0

	result, err := appCli.Template.GetTemplateSummary(params, nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to fetch templates: %w", err)
	}

	if result.Payload == nil || result.Payload.Data == nil || result.Payload.Data.Content == nil {
		return 0, nil, nil, fmt.Errorf("empty template response")
	}

	for _, t := range result.Payload.Data.Content {
		if t.Name != nil && t.TemplateType != nil {
			if *t.TemplateType == "BUILTIN" && strings.Contains(*t.Name, name) {
				if t.ID != nil {
					// Fetch full template to get task_setting
					taskSetting, allowedPlanTypes, err := getTemplateSettings(appCli, *t.ID)
					if err != nil {
						return 0, nil, nil, err
					}
					return *t.ID, taskSetting, allowedPlanTypes, nil
				}
			}
		}
	}

	return 0, nil, nil, nil
}

// getTemplateSettings fetches task settings and supported plan types for a template.
func getTemplateSettings(appCli *client.OPENAPI, templateID int64) (interface{}, []string, error) {
	params := template.NewGetTemplateIDParams()
	params.ID = templateID

	result, err := appCli.Template.GetTemplateID(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch template %d: %w", templateID, err)
	}

	if result.Payload == nil || result.Payload.Data == nil {
		return nil, nil, fmt.Errorf("template %d not found", templateID)
	}

	var allowedPlanTypes []string
	if strategy := result.Payload.Data.Strategy; strategy != nil && strategy.BasicSetting != nil {
		allowedPlanTypes = supportedPlanTypes(strategy.BasicSetting.JSONSchema)
	}
	taskSetting := result.Payload.Data.TaskSetting
	logDebugf("Got task_setting from template API (type: %T)", taskSetting)
	return taskSetting, allowedPlanTypes, nil
}

func parseCommaSeparated(value string) []string {
	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}
