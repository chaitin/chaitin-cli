package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaitin/workspace-cli/products/xray/client"
	"github.com/chaitin/workspace-cli/products/xray/client/plan"
	"github.com/chaitin/workspace-cli/products/xray/client/template"
	"github.com/chaitin/workspace-cli/products/xray/models"

	"github.com/spf13/cobra"
)

const defaultBuiltinTemplateName = "基础服务漏洞扫描"

// makeOperationPlanCreateQuickCmd returns a command to handle quick plan creation
func makeOperationPlanCreateQuickCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "PostPlanCreateQuick",
		Short: "Quick create a plan with targets and engines",
		Long: fmt.Sprintf(`Quick create a plan that runs immediately with specified targets and engines.

Example:
  xray plan quick --targets=example.com --engines=00000000000000000000000000000001
  xray plan quick --targets=example.com,example2.com --engines=engine1,engine2 --name=my-scan

The template-id defaults to the "%s" (BUILTIN) template.`, defaultBuiltinTemplateName),
		RunE: runOperationPlanQuick,
	}

	cmd.Flags().String("targets", "", "Comma-separated list of targets (required)")
	cmd.Flags().String("engines", "", "Comma-separated list of engine IDs (required)")
	cmd.Flags().String("name", "quick-scan", "Task name")
	cmd.Flags().Int64("project-id", 0, "Project ID")
	cmd.Flags().Int64("template-id", 0, fmt.Sprintf("Task template ID (defaults to %s)", defaultBuiltinTemplateName))
	cmd.Flags().String("template-name", "", fmt.Sprintf("Override template name search (default: %s)", defaultBuiltinTemplateName))

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
	templateID, _ := cmd.Flags().GetInt64("template-id")
	templateNameFlag, _ := cmd.Flags().GetString("template-name")

	targets := strings.Split(targetsStr, ",")
	engines := strings.Split(enginesStr, ",")

	if targetsStr == "" {
		return fmt.Errorf("targets is required")
	}
	if enginesStr == "" {
		return fmt.Errorf("engines is required")
	}
	if projectID == 0 {
		return fmt.Errorf("project-id is required")
	}

	// Find template and get task_setting
	var taskSetting interface{}
	if templateID == 0 {
		templateName := templateNameFlag
		if templateName == "" {
			templateName = defaultBuiltinTemplateName
		}

		templateID, taskSetting, err = findBuiltinTemplateWithTaskSetting(appCli, templateName)
		if err != nil {
			return err
		}
		if templateID == 0 {
			return fmt.Errorf("template '%s' not found. Use --template-id to specify explicitly", templateName)
		}
		logDebugf("Found template ID %d for '%s'", templateID, templateName)
	} else {
		// When template-id is explicitly specified, fetch its task_setting
		taskSetting, err = getTemplateTaskSetting(appCli, templateID)
		if err != nil {
			return err
		}
	}

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
		"planSetting": map[string]interface{}{
			"enabled":  true,
			"planType": "NOW",
		},
		"engineChoice": engines,
		"taskName":     name,
	}

	active := true
	execRightNow := true

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

// findBuiltinTemplateWithTaskSetting fetches templates and finds the one matching the given name with BUILTIN type, returns id and task_setting
func findBuiltinTemplateWithTaskSetting(appCli *client.OPENAPI, name string) (int64, interface{}, error) {
	params := template.NewGetTemplateSummaryParams()
	params.Limit = 100
	params.Offset = 0

	result, err := appCli.Template.GetTemplateSummary(params, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch templates: %w", err)
	}

	if result.Payload == nil || result.Payload.Data == nil || result.Payload.Data.Content == nil {
		return 0, nil, fmt.Errorf("empty template response")
	}

	for _, t := range result.Payload.Data.Content {
		if t.Name != nil && t.TemplateType != nil {
			if *t.TemplateType == "BUILTIN" && strings.Contains(*t.Name, name) {
				if t.ID != nil {
					// Fetch full template to get task_setting
					taskSetting, err := getTemplateTaskSetting(appCli, *t.ID)
					if err != nil {
						return 0, nil, err
					}
					return *t.ID, taskSetting, nil
				}
			}
		}
	}

	return 0, nil, nil
}

// getTemplateTaskSetting fetches the task_setting for a given template ID
func getTemplateTaskSetting(appCli *client.OPENAPI, templateID int64) (interface{}, error) {
	params := template.NewGetTemplateIDParams()
	params.ID = templateID

	result, err := appCli.Template.GetTemplateID(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template %d: %w", templateID, err)
	}

	if result.Payload == nil || result.Payload.Data == nil {
		return nil, fmt.Errorf("template %d not found", templateID)
	}

	taskSetting := result.Payload.Data.TaskSetting
	if taskSetting == nil {
		return nil, nil
	}

	// TaskSetting is already interface{} (map), go-swagger handles JSON correctly
	logDebugf("Got task_setting from template API (type: %T)", taskSetting)
	return taskSetting, nil
}
