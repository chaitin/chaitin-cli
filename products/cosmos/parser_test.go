package cosmos

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestBuildRequestBodyNormalizesTimeQueryShorthand(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := []flagDef{{
		name:      "created_at",
		typeName:  "string",
		queryType: timeQueryType,
	}}
	values := make(map[string]interface{})
	bindFlag(cmd, flags[0], values)

	if err := cmd.Flags().Set("created_at", "0-1781501714000"); err != nil {
		t.Fatal(err)
	}

	params := buildRequestBody(cmd, values, flags)
	createdAt, ok := params["created_at"].([]map[string]string)
	if !ok {
		t.Fatalf("created_at type = %T, want []map[string]string", params["created_at"])
	}
	if len(createdAt) != 1 {
		t.Fatalf("created_at len = %d, want 1", len(createdAt))
	}
	if createdAt[0]["oper"] != "eq_in" {
		t.Fatalf("created_at oper = %q, want eq_in", createdAt[0]["oper"])
	}
	if createdAt[0]["target"] != "0-1781501714000" {
		t.Fatalf("created_at target = %q, want 0-1781501714000", createdAt[0]["target"])
	}
}

func TestBuildRequestBodyPreservesTimeQueryJSON(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := []flagDef{{
		name:      "created_at",
		typeName:  "string",
		queryType: timeQueryType,
	}}
	values := make(map[string]interface{})
	bindFlag(cmd, flags[0], values)

	if err := cmd.Flags().Set("created_at", `[{"oper":"eq_in","target":"0-1781501714000"}]`); err != nil {
		t.Fatal(err)
	}

	params := buildRequestBody(cmd, values, flags)
	createdAt, ok := params["created_at"].([]interface{})
	if !ok {
		t.Fatalf("created_at type = %T, want []interface{}", params["created_at"])
	}
	if len(createdAt) != 1 {
		t.Fatalf("created_at len = %d, want 1", len(createdAt))
	}
	item, ok := createdAt[0].(map[string]interface{})
	if !ok {
		t.Fatalf("created_at[0] type = %T, want map[string]interface{}", createdAt[0])
	}
	if item["oper"] != "eq_in" || item["target"] != "0-1781501714000" {
		t.Fatalf("created_at[0] = %#v", item)
	}
}

func TestBuildRequestBodyParsesJSONQueryType(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := []flagDef{
		{
			name:      "configs",
			typeName:  "string",
			queryType: "JSON",
		},
		{
			name:      "params",
			typeName:  "string",
			queryType: "JSON",
		},
	}
	values := make(map[string]interface{})
	for _, f := range flags {
		bindFlag(cmd, f, values)
	}

	if err := cmd.Flags().Set("configs", `[{"ips":["1.2.3.4"],"dev_ids":["dev-1"]}]`); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("params", `{"title":"告警","count":1}`); err != nil {
		t.Fatal(err)
	}

	request := buildRequestBody(cmd, values, flags)
	configs, ok := request["configs"].([]interface{})
	if !ok {
		t.Fatalf("configs type = %T, want []interface{}", request["configs"])
	}
	if len(configs) != 1 {
		t.Fatalf("configs len = %d, want 1", len(configs))
	}
	params, ok := request["params"].(map[string]interface{})
	if !ok {
		t.Fatalf("params type = %T, want map[string]interface{}", request["params"])
	}
	if params["title"] != "告警" || params["count"] != float64(1) {
		t.Fatalf("params = %#v", params)
	}
}

func TestBuildRequestBodyParsesQueryTypeJSONForArrayNumberAndBoolFields(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := []flagDef{
		{
			name:      "attack_info",
			typeName:  "string",
			isArray:   true,
			queryType: "StringInArrayQuery",
		},
		{
			name:      "organization_id",
			typeName:  "integer",
			queryType: "UintQuery",
		},
		{
			name:      "focused",
			typeName:  "bool",
			queryType: "BoolQuery",
		},
	}
	values := make(map[string]interface{})
	for _, f := range flags {
		bindFlag(cmd, f, values)
	}

	if err := cmd.Flags().Set("attack_info", `[{"oper":"array_in","target":"T1190,T1059"}]`); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("organization_id", `[{"oper":"=","target":1}]`); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("focused", `[{"oper":"=","target":true}]`); err != nil {
		t.Fatal(err)
	}

	params := buildRequestBody(cmd, values, flags)
	attackInfo := firstQueryItem(t, params, "attack_info")
	if attackInfo["oper"] != "array_in" || attackInfo["target"] != "T1190,T1059" {
		t.Fatalf("attack_info query = %#v", attackInfo)
	}
	organizationID := firstQueryItem(t, params, "organization_id")
	if organizationID["oper"] != "=" || organizationID["target"] != float64(1) {
		t.Fatalf("organization_id query = %#v", organizationID)
	}
	focused := firstQueryItem(t, params, "focused")
	if focused["oper"] != "=" || focused["target"] != true {
		t.Fatalf("focused query = %#v", focused)
	}
}

func TestRegisterOperationsMergesExistingServiceCommand(t *testing.T) {
	parent := &cobra.Command{Use: "cosmos"}

	registerOperations(parent, []parsedOp{{
		serviceName: "AlarmService",
		methodName:  "GetAlarmInfo",
		serviceCmd:  "alarm",
		methodCmd:   "get-alarm-info",
		comment:     "获取告警详情",
	}})
	registerOperations(parent, []parsedOp{{
		serviceName: "AlarmService",
		methodName:  "TriggerPlaybookGeneral",
		serviceCmd:  "alarm",
		methodCmd:   "trigger-playbook-general",
		comment:     "触发 SOAR 剧本（通用）",
	}})

	alarmCommands := 0
	var alarmCmd *cobra.Command
	for _, cmd := range parent.Commands() {
		if commandUseName(cmd) == "alarm" {
			alarmCommands++
			alarmCmd = cmd
		}
	}
	if alarmCommands != 1 {
		t.Fatalf("alarm command count = %d, want 1", alarmCommands)
	}
	if _, _, err := alarmCmd.Find([]string{"get-alarm-info"}); err != nil {
		t.Fatalf("missing get-alarm-info: %v", err)
	}
	if _, _, err := alarmCmd.Find([]string{"trigger-playbook-general"}); err != nil {
		t.Fatalf("missing trigger-playbook-general: %v", err)
	}
}

func TestValidateAlarmListRequiresCreatedOrUpdatedAt(t *testing.T) {
	err := validateRequestParams(alarmListRPCMethod, map[string]interface{}{
		"count":  1,
		"offset": 0,
	})
	if err == nil {
		t.Fatal("expected missing time filter error")
	}
	if !strings.Contains(err.Error(), "requires --created_at or --updated_at") {
		t.Fatalf("unexpected error: %v", err)
	}

	validCases := []map[string]interface{}{
		{"created_at": []map[string]string{{"oper": "eq_in", "target": "0-1781501714000"}}},
		{"updated_at": []map[string]string{{"oper": "eq_in", "target": "0-1781501714000"}}},
		{"workflow_id": 1},
		{"exclude_time_filter": true},
	}
	for _, params := range validCases {
		if err := validateRequestParams(alarmListRPCMethod, params); err != nil {
			t.Fatalf("validateRequestParams(%#v) error = %v", params, err)
		}
	}
}

func TestBuildRequestBodyKeepsExplicitEmptyStringForWriteFields(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := []flagDef{{
		name:     "business_ext",
		typeName: "string",
	}}
	values := make(map[string]interface{})
	bindFlag(cmd, flags[0], values)

	if err := cmd.Flags().Set("business_ext", ""); err != nil {
		t.Fatal(err)
	}

	params := buildRequestBody(cmd, values, flags)
	value, ok := params["business_ext"].(string)
	if !ok {
		t.Fatalf("business_ext type = %T, want string", params["business_ext"])
	}
	if value != "" {
		t.Fatalf("business_ext = %q, want empty string", value)
	}
}

func TestBuildRequestBodyConvertsExplicitEmptyStringSliceToEmptySlice(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := []flagDef{{
		name:     "attack_info",
		typeName: "string",
		isArray:  true,
	}}
	values := make(map[string]interface{})
	bindFlag(cmd, flags[0], values)

	if err := cmd.Flags().Set("attack_info", ""); err != nil {
		t.Fatal(err)
	}

	params := buildRequestBody(cmd, values, flags)
	value, ok := params["attack_info"].([]string)
	if !ok {
		t.Fatalf("attack_info type = %T, want []string", params["attack_info"])
	}
	if len(value) != 0 {
		t.Fatalf("attack_info = %#v, want empty slice", value)
	}
}

func TestBuildRequestBodyNestsAIFields(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := []flagDef{
		{name: "ai_fields.alarm_id", typeName: "string", required: true},
		{name: "ai_fields.analysis_report", typeName: "string"},
		{name: "ai_fields.confidence_score", typeName: "integer"},
		{name: "ai_fields.ai_analysis", typeName: "integer"},
		{name: "ai_fields.disposal_report", typeName: "string"},
		{name: "ai_fields.ai_disposal", typeName: "integer"},
	}
	values := make(map[string]interface{})
	for _, f := range flags {
		bindFlag(cmd, f, values)
	}

	for name, value := range map[string]string{
		"ai_fields.alarm_id":         "alarm-1",
		"ai_fields.analysis_report":  "analysis",
		"ai_fields.confidence_score": "85",
		"ai_fields.ai_analysis":      "1",
		"ai_fields.disposal_report":  "disposal",
		"ai_fields.ai_disposal":      "1",
	} {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}

	params := buildRequestBody(cmd, values, flags)
	aiFields, ok := params["ai_fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("ai_fields type = %T, want map[string]interface{}", params["ai_fields"])
	}
	for name, want := range map[string]interface{}{
		"alarm_id":         "alarm-1",
		"analysis_report":  "analysis",
		"confidence_score": 85,
		"ai_analysis":      1,
		"disposal_report":  "disposal",
		"ai_disposal":      1,
	} {
		if got := aiFields[name]; got != want {
			t.Errorf("ai_fields.%s = %#v, want %#v", name, got, want)
		}
	}
}

func firstQueryItem(t *testing.T, params map[string]interface{}, key string) map[string]interface{} {
	t.Helper()

	items, ok := params[key].([]interface{})
	if !ok {
		t.Fatalf("%s type = %T, want []interface{}", key, params[key])
	}
	if len(items) != 1 {
		t.Fatalf("%s len = %d, want 1", key, len(items))
	}
	item, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("%s[0] type = %T, want map[string]interface{}", key, items[0])
	}
	return item
}
