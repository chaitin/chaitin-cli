package cli

import (
	"encoding/json"
	"fmt"

	"github.com/chaitin/chaitin-cli/products/xray/client/plan"
	"github.com/chaitin/chaitin-cli/products/xray/models"
	"github.com/spf13/cobra"
)

func makeOperationPlanUpdateQuickCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "PostPlanUpdateQuick",
		Short: "快速修改扫描任务的排期",
		Long: `只修改已有任务的排期，不修改名称、目标、引擎、扫描配置或执行窗口，也不会立即触发扫描。

排期参数：
  CLOCKED     --exec-at 使用带时区 RFC3339，例如 2026-08-05T10:30:00+08:00
  DAY         --exec-at 使用 HH:MM 或 HH:MM:SS
  WEEK        额外要求 --weekday，1=周一 ... 7=周日
  MONTH       额外要求 --day-of-month，范围 1-31
  MONTH_WEEK  额外要求 --week-of-month 1-4 和 --weekday 1-7

省略 --plan-type 时沿用任务当前类型。--exec-at 表示任务触发时间，不是扫描结束时间。

示例：
  xray plan PostPlanUpdateQuick --id=123 --exec-at=11:00
  xray plan PostPlanUpdateQuick --id=123 --plan-type=DAY --exec-at=10:30
  xray plan PostPlanUpdateQuick --id=123 --plan-type=WEEK --weekday=5 --exec-at=10:30
  xray plan PostPlanUpdateQuick --id=123 --plan-type=MONTH --day-of-month=15 --exec-at=10:30
  xray plan PostPlanUpdateQuick --id=123 --plan-type=MONTH_WEEK --week-of-month=2 --weekday=5 --exec-at=10:30`,
		RunE: runOperationPlanUpdateQuick,
	}
	cmd.Flags().Int64("id", 0, "任务计划 ID (必填)")
	registerQuickScheduleFlags(cmd, "")
	return cmd, nil
}

func runOperationPlanUpdateQuick(cmd *cobra.Command, args []string) error {
	id, err := cmd.Flags().GetInt64("id")
	if err != nil {
		return err
	}
	if id <= 0 {
		return fmt.Errorf("--id is required and must be greater than zero")
	}
	input, err := readQuickScheduleInput(cmd)
	if err != nil {
		return err
	}
	if !input.changed() {
		return fmt.Errorf("at least one scheduling flag must be provided")
	}

	appCli, err := makeClient(cmd, args)
	if err != nil {
		return err
	}
	getResult, err := appCli.Plan.GetPlanID(plan.NewGetPlanIDParams().WithID(id), nil)
	if err != nil {
		return fmt.Errorf("failed to fetch plan %d: %w", id, err)
	}
	if getResult.Payload == nil || getResult.Payload.Data == nil {
		return fmt.Errorf("plan %d returned an empty response", id)
	}
	current := getResult.Payload.Data
	basicSetting, err := cloneObject(current.BasicSetting)
	if err != nil {
		return fmt.Errorf("plan %d has invalid basic_setting: %w", id, err)
	}
	existingPlanSetting, _ := basicSetting["planSetting"].(map[string]any)
	planSetting, err := buildPlanSetting(input, existingPlanSetting)
	if err != nil {
		return err
	}
	if current.Strategy != nil && current.Strategy.BasicSetting != nil {
		if err := validateSupportedPlanType(planSetting["planType"].(string), supportedPlanTypes(current.Strategy.BasicSetting.JSONSchema)); err != nil {
			return err
		}
	}
	basicSetting["planSetting"] = planSetting

	execRightNow := false
	params := plan.NewPostPlanUpdateParams()
	params.Body = &models.UpdatePlanBody{ID: &id, ExecRightNow: &execRightNow, BasicSetting: basicSetting}
	if dryRun {
		logDebugf("dry-run flag specified. Skip sending request.")
		if body, marshalErr := json.MarshalIndent(params.Body, "", "  "); marshalErr == nil {
			logDebugf("Request body: %s", body)
		}
		return nil
	}

	message, err := parseOperationPlanPostPlanUpdateResult(appCli.Plan.PostPlanUpdate(params, nil))
	if err != nil {
		return err
	}
	if !debug {
		fmt.Println(message)
	}
	return nil
}

func cloneObject(value any) (map[string]any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(encoded, &result); err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("expected JSON object")
	}
	return result, nil
}
