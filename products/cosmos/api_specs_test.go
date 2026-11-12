package cosmos

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestAlarmJSON_Structure 验证 alarm.json 的结构完整性：
// - 7 个 API 全部存在
// - UpdateAlarmInfo 关键字段为必填
// - UpdateAlarmMoreInfo 包含新增的 5 个嵌套对象
// - UpdateAiFieldsContent 使用 ai_fields 包装专用 AI 字段
// - MultiJudgeAlert.remark 为必填
func TestAlarmJSON_Structure(t *testing.T) {
	ops := loadAPIOperations(t, "alarm.json", 7)
	byName := mapByName(ops)

	// 1) GetAlarmList 存在且 queryType 注解正确
	t.Run("GetAlarmList", func(t *testing.T) {
		op, ok := byName["AlarmService.GetAlarmList"]
		if !ok {
			t.Fatal("missing GetAlarmList")
		}
		// 关键时间字段必须有 queryType
		for _, f := range []string{"created_at", "log_start_at", "log_end_at"} {
			field, ok := op.Args[f]
			if !ok {
				t.Errorf("missing field %q", f)
			} else if field.QueryType != "TimeQuery" {
				t.Errorf("%q: QueryType=%q, want TimeQuery", f, field.QueryType)
			}
		}
		// count/offset 不应有 queryType
		for _, f := range []string{"count", "offset", "scenarios", "workflow_id"} {
			if field, ok := op.Args[f]; ok && field.QueryType != "" {
				t.Errorf("%q should not have queryType but got %q", f, field.QueryType)
			}
		}
		t.Logf("%d total args", len(op.Args))
	})

	// 2) UpdateAlarmInfo 必填字段
	t.Run("UpdateAlarmInfo", func(t *testing.T) {
		op, ok := byName["AlarmService.UpdateAlarmInfo"]
		if !ok {
			t.Fatal("missing UpdateAlarmInfo")
		}
		required := []string{"id", "alarm_name", "alarm_level"}
		for _, f := range required {
			field, ok := op.Args[f]
			if !ok {
				t.Errorf("missing field %q", f)
			} else if field.Optional {
				t.Errorf("%q should be required", f)
			}
		}
		// alarm_area_id 应可选
		if field, ok := op.Args["alarm_area_id"]; ok && !field.Optional {
			t.Error("alarm_area_id should be optional")
		}
		// attack_direction 后端有 omitempty，应可选
		if field, ok := op.Args["attack_direction"]; ok && !field.Optional {
			t.Error("attack_direction should be optional")
		}
	})

	// 3) MultiJudgeAlert - remark 必填
	t.Run("MultiJudgeAlert", func(t *testing.T) {
		op, ok := byName["JudgeService.MultiJudgeAlert"]
		if !ok {
			t.Fatal("missing MultiJudgeAlert")
		}
		if field, ok := op.Args["remark"]; !ok {
			t.Error("missing remark")
		} else if field.Optional {
			t.Error("remark should be required")
		}
	})

	// 4) UpdateAlarmMoreInfo - 新增 5 个嵌套对象
	t.Run("UpdateAlarmMoreInfo", func(t *testing.T) {
		op, ok := byName["AlarmService.UpdateAlarmMoreInfo"]
		if !ok {
			t.Fatal("missing UpdateAlarmMoreInfo")
		}
		want := []string{
			"host_security.endpoint_info",
			"malicious_file.file_info",
			"net_attack.dns_info",
			"user_abnormal_behavior.account_info",
			"vulnerability.compliance_baseline",
		}
		for _, f := range want {
			if _, ok := op.Args[f]; !ok {
				t.Errorf("missing new nested field %q", f)
			}
		}
	})

	// 5) GetAlarmList 核心字段改为可选（后端零值即"不过滤"）
	t.Run("GetAlarmList_optionality", func(t *testing.T) {
		op := byName["AlarmService.GetAlarmList"]
		for _, f := range []string{"scenarios", "workflow_id", "alarm_level_statistics", "attack_chain_statistics"} {
			field, ok := op.Args[f]
			if !ok {
				t.Errorf("missing field %q", f)
			} else if !field.Optional {
				t.Errorf("%q should be optional (zero value = no filter)", f)
			}
		}
	})

	// 6) 各 API 均补充了 created_at 分区键字段（可选，用于加速查询）
	t.Run("created_at_partition_key", func(t *testing.T) {
		for _, apiName := range []string{
			"AlarmService.GetAlarmInfo",
			"AlarmService.UpdateTag",
			"AlarmService.UpdateAlarmMoreInfo",
			"AlarmService.UpdateAlarmInfo",
			"JudgeService.MultiJudgeAlert",
		} {
			op := byName[apiName]
			field, ok := op.Args["created_at"]
			if !ok {
				t.Errorf("%s: missing created_at partition key", apiName)
			} else if !field.Optional {
				t.Errorf("%s: created_at should be optional", apiName)
			}
		}
	})

	// 7) GetAlarmInfo 新增 include_event_history
	t.Run("GetAlarmInfo_include_event_history", func(t *testing.T) {
		op := byName["AlarmService.GetAlarmInfo"]
		field, ok := op.Args["include_event_history"]
		if !ok {
			t.Error("missing include_event_history")
		} else if !field.Optional {
			t.Error("include_event_history should be optional")
		}
	})

	// 8) AI 字段只能通过专用接口写回，且必须包装在 ai_fields 中。
	t.Run("UpdateAiFieldsContent", func(t *testing.T) {
		op, ok := byName["AlarmService.UpdateAiFieldsContent"]
		if !ok {
			t.Fatal("missing UpdateAiFieldsContent")
		}

		if field, ok := op.Args["ai_fields.alarm_id"]; !ok {
			t.Fatal("missing ai_fields.alarm_id")
		} else if field.Optional {
			t.Error("ai_fields.alarm_id should be required")
		}

		for _, name := range []string{
			"ai_fields.analysis_report",
			"ai_fields.confidence_score",
			"ai_fields.ai_analysis",
			"ai_fields.ai_judgment_result",
			"ai_fields.ai_judgment_level",
			"ai_fields.ai_event_nature",
			"ai_fields.disposal_report",
			"ai_fields.ai_disposal",
			"ai_fields.created_at",
		} {
			field, ok := op.Args[name]
			if !ok {
				t.Errorf("missing %s", name)
				continue
			}
			if !field.Optional {
				t.Errorf("%s should be optional", name)
			}
		}
		assertOperationRegistered(t, op)
	})

	t.Log("All 7 APIs validated successfully")
}

func TestUpdateAiFieldsContentDryRun(t *testing.T) {
	oldServerURL := serverURL
	oldAPIToken := apiToken
	oldDryRun := dryRun
	t.Cleanup(func() {
		serverURL = oldServerURL
		apiToken = oldAPIToken
		dryRun = oldDryRun
	})

	serverURL = "https://cosmos.example.com"
	apiToken = "test-token"
	dryRun = true

	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{
		"alarm", "update-ai-fields-content",
		"--ai_fields.alarm_id", "alarm-1",
		"--ai_fields.analysis_report", "analysis",
		"--ai_fields.confidence_score", "85",
		"--ai_fields.ai_analysis", "1",
		"--ai_fields.ai_judgment_result", "误报",
		"--ai_fields.ai_judgment_level", "1",
		"--ai_fields.ai_event_nature", "业务测试",
		"--ai_fields.disposal_report", "disposal",
		"--ai_fields.ai_disposal", "1",
		"--raw",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute update-ai-fields-content: %v", err)
	}

	var summary dryRunRequestSummary
	if err := json.Unmarshal(out.Bytes(), &summary); err != nil {
		t.Fatalf("unmarshal dry-run output: %v\n%s", err, out.String())
	}
	if summary.Body.Method != "AlarmService.UpdateAiFieldsContent" {
		t.Fatalf("method = %q, want AlarmService.UpdateAiFieldsContent", summary.Body.Method)
	}

	params, ok := summary.Body.Params.(map[string]interface{})
	if !ok {
		t.Fatalf("params type = %T, want map[string]interface{}", summary.Body.Params)
	}
	aiFields, ok := params["ai_fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("ai_fields type = %T, want map[string]interface{}", params["ai_fields"])
	}
	for name, want := range map[string]interface{}{
		"alarm_id":           "alarm-1",
		"analysis_report":    "analysis",
		"confidence_score":   float64(85),
		"ai_analysis":        float64(1),
		"ai_judgment_result": "误报",
		"ai_judgment_level":  "1",
		"ai_event_nature":    "业务测试",
		"disposal_report":    "disposal",
		"ai_disposal":        float64(1),
	} {
		if got := aiFields[name]; got != want {
			t.Errorf("ai_fields.%s = %#v, want %#v", name, got, want)
		}
	}
}

func TestAssetJSONStructure(t *testing.T) {
	ops := loadAPIOperations(t, "asset.json", 8)
	byName := mapByName(ops)

	for _, name := range []string{
		"AssetService.SearchHostAsset",
		"AssetService.GetHostAsset",
		"AssetService.SaveHostAsset",
		"AssetService.SearchAssetScopeByIP",
		"AssetService.SearchOrganization",
		"AssetService.SearchWebAsset",
		"AssetService.GetWebAsset",
		"AssetService.FindRelationWebAsset",
	} {
		if _, ok := byName[name]; !ok {
			t.Fatalf("missing %s", name)
		}
	}

	t.Run("SearchHostAsset", func(t *testing.T) {
		op := byName["AssetService.SearchHostAsset"]
		for field, queryType := range map[string]string{
			"id":                          "UintQuery",
			"draft":                       "BoolQuery",
			"name":                        "StringQuery",
			"state":                       "StringQuery",
			"category_id":                 "UintQuery",
			"position_id":                 "UintQuery",
			"ip":                          "InetQuery",
			"asset_ip_type":               "UintQuery",
			"organization_id":             "UintQuery",
			"security_scope_id":           "UintQuery",
			"business_scope_id":           "UintQuery",
			"deploy_location_id":          "UintQuery",
			"security_deploy_location_id": "UintQuery",
			"business_deploy_location_id": "UintQuery",
			"tag_id":                      "UintQuery",
			"tags":                        "StringQuery",
			"asset_owner":                 "UintQuery",
			"asset_owner_name":            "StringQuery",
			"business_owner":              "UintQuery",
			"business_owner_name":         "StringQuery",
			"importance":                  "UintQuery",
			"value_lv":                    "StringQuery",
			"is_alarmed":                  "BoolQuery",
			"is_vulned":                   "BoolQuery",
			"external_accessible":         "BoolQuery",
			"port":                        "UintQuery",
			"app_name":                    "StringQuery",
			"subnet":                      "StringQuery",
			"live_port":                   "UintQuery",
			"source":                      "StringQuery",
			"find_by":                     "StringQuery",
			"find_by_source":              "StringQuery",
			"factor":                      "StringQuery",
			"country":                     "StringQuery",
			"province":                    "StringQuery",
			"city":                        "StringQuery",
			"ip_change_state":             "IntQuery",
			"port_change_state":           "IntQuery",
			"latest_alarmed_at":           "TimeQuery",
			"latest_vulned_at":            "TimeQuery",
			"last_modify_at":              "TimeQuery",
			"created_at":                  "TimeQuery",
			"find_at":                     "TimeQuery",
			"scan_vuln_level":             "IntQuery",
			"scan_vuln_time":              "TimeQuery",
			"alarm_fall":                  "BoolQuery",
			"alarm_level":                 "StringQuery",
			"related_web_asset":           "StringQuery",
			"condition_query":             "JSON",
			"group_id":                    "UintQuery",
		} {
			assertQueryType(t, op, field, queryType)
		}
		assertOptional(t, op, "count")
		assertOptional(t, op, "offset")
		assertOptional(t, op, "workflow_id")
		assertOptional(t, op, "is_zombie_network")
	})

	t.Run("SaveHostAsset", func(t *testing.T) {
		op := byName["AssetService.SaveHostAsset"]
		for _, field := range []string{"ip", "name", "organization_id"} {
			assertRequired(t, op, field)
		}
		for _, field := range []string{"id", "security_scope_id", "business_scope_id", "asset_ip_type", "importance"} {
			assertOptional(t, op, field)
		}
		assertArray(t, op, "tags")
		assertArray(t, op, "related_ip")
		assertArray(t, op, "related_web_asset_id")
		assertQueryType(t, op, "category_ids", "JSON")
		assertQueryType(t, op, "custom_common_attribute", "JSON")
		assertQueryType(t, op, "custom_category_attribute", "JSON")
	})

	t.Run("SearchAssetScopeByIP", func(t *testing.T) {
		op := byName["AssetService.SearchAssetScopeByIP"]
		assertRequired(t, op, "ip")
		assertArray(t, op, "ip")
	})

	t.Run("SearchOrganization", func(t *testing.T) {
		op := byName["AssetService.SearchOrganization"]
		for field, queryType := range map[string]string{
			"id":                    "UintQuery",
			"name":                  "StringQuery",
			"tag":                   "StringQuery",
			"category":              "UintQuery",
			"industry":              "UintQuery",
			"location_id":           "UintQuery",
			"location_name":         "StringQuery",
			"superior_organization": "UintQuery",
			"is_enabled":            "BoolQuery",
			"is_system":             "BoolQuery",
		} {
			assertQueryType(t, op, field, queryType)
		}
		assertOptional(t, op, "count")
		assertOptional(t, op, "offset")
	})

	t.Run("SearchWebAsset", func(t *testing.T) {
		op := byName["AssetService.SearchWebAsset"]
		for field, queryType := range map[string]string{
			"id":                          "UintQuery",
			"draft":                       "BoolQuery",
			"name":                        "StringQuery",
			"state":                       "StringQuery",
			"site_url":                    "StringQuery",
			"domain":                      "StringQuery",
			"netloc":                      "StringQuery",
			"site_md5":                    "StringQuery",
			"scheme":                      "StringQuery",
			"port":                        "UintQuery",
			"organization_id":             "UintQuery",
			"security_scope_id":           "UintQuery",
			"business_scope_id":           "UintQuery",
			"deploy_location_id":          "UintQuery",
			"security_deploy_location_id": "UintQuery",
			"business_deploy_location_id": "UintQuery",
			"category_id":                 "UintQuery",
			"group_id":                    "UintQuery",
			"tag_id":                      "UintQuery",
			"tags":                        "StringQuery",
			"position_id":                 "UintQuery",
			"asset_owner":                 "UintQuery",
			"asset_owner_name":            "StringQuery",
			"business_owner":              "UintQuery",
			"business_owner_name":         "StringQuery",
			"importance":                  "UintQuery",
			"value_lv":                    "StringQuery",
			"is_alarmed":                  "BoolQuery",
			"is_vulned":                   "BoolQuery",
			"external_accessible":         "BoolQuery",
			"find_by":                     "StringQuery",
			"find_by_source":              "StringQuery",
			"latest_alarmed_at":           "TimeQuery",
			"latest_vulned_at":            "TimeQuery",
			"last_modify_at":              "TimeQuery",
			"created_at":                  "TimeQuery",
			"find_at":                     "TimeQuery",
			"scan_vuln_level":             "IntQuery",
			"scan_vuln_time":              "TimeQuery",
			"alarm_fall":                  "BoolQuery",
			"alarm_level":                 "StringQuery",
			"condition_query":             "JSON",
		} {
			assertQueryType(t, op, field, queryType)
		}
		assertOptional(t, op, "workflow_id")
	})

	t.Run("FindRelationWebAsset", func(t *testing.T) {
		op := byName["AssetService.FindRelationWebAsset"]
		for field, queryType := range map[string]string{
			"id":                "UintQuery",
			"draft":             "BoolQuery",
			"name":              "StringQuery",
			"site_url":          "StringQuery",
			"domain":            "StringQuery",
			"organization_id":   "UintQuery",
			"security_scope_id": "UintQuery",
		} {
			assertQueryType(t, op, field, queryType)
		}
		assertOptional(t, op, "workflow_id")
	})

	assertOperationsRegistered(t, ops)
}

func TestDisposalJSONStructure(t *testing.T) {
	t.Run("IpBlock", func(t *testing.T) {
		ops := loadAPIOperations(t, "ipblock.json", 7)
		byName := mapByName(ops)

		for _, name := range []string{
			"IpBlockService.SearchBlockHistory",
			"IpBlockService.SearchBlockedList",
			"IpBlockService.CreateBlackIp",
			"IpBlockService.DeleteBlackIp",
			"IpBlockService.SearchDeviceList",
			"IpBlockService.SearchWhiteList",
			"IpBlockService.CreateWhiteIp",
		} {
			assertOperationRegistered(t, byName[name])
		}

		history := byName["IpBlockService.SearchBlockHistory"]
		for field, queryType := range map[string]string{
			"filter.ip":           "StringQuery",
			"filter.dev_id":       "StringQuery",
			"filter.block_mode":   "StringQuery",
			"filter.block_time":   "TimeQuery",
			"filter.unblock_mode": "StringQuery",
			"filter.unblock_time": "TimeQuery",
		} {
			assertQueryType(t, history, field, queryType)
		}
		assertArray(t, history, "filter.organization_id")

		blocked := byName["IpBlockService.SearchBlockedList"]
		for field, queryType := range map[string]string{
			"filter.ip":         "StringQuery",
			"filter.dev_id":     "StringQuery",
			"filter.block_mode": "StringQuery",
			"filter.block_time": "TimeQuery",
		} {
			assertQueryType(t, blocked, field, queryType)
		}
		assertArray(t, blocked, "filter.organization_id")

		device := byName["IpBlockService.SearchDeviceList"]
		assertQueryType(t, device, "filter.dev_name", "StringQuery")
		assertQueryType(t, device, "filter.dev_vendor", "StringQuery")
		assertArray(t, device, "filter.organization_id")

		white := byName["IpBlockService.SearchWhiteList"]
		for field, queryType := range map[string]string{
			"filter.ip":            "StringQuery",
			"filter.dev_id":        "StringQuery",
			"filter.remark":        "StringQuery",
			"filter.create_time":   "TimeQuery",
			"filter.update_time":   "TimeQuery",
			"filter.created_by_id": "UintQuery",
			"filter.type":          "UintQuery",
		} {
			assertQueryType(t, white, field, queryType)
		}
		assertOptional(t, white, "filter.expire_state")
		assertArray(t, white, "filter.organization_id")

		createBlack := byName["IpBlockService.CreateBlackIp"]
		assertQueryType(t, createBlack, "configs", "JSON")
		assertOptional(t, createBlack, "alarm_id")
		assertOptional(t, createBlack, "created_at")

		deleteBlack := byName["IpBlockService.DeleteBlackIp"]
		assertQueryType(t, deleteBlack, "black_list", "JSON")

		createWhite := byName["IpBlockService.CreateWhiteIp"]
		assertQueryType(t, createWhite, "configs", "JSON")
	})

	t.Run("NoticeStrategy", func(t *testing.T) {
		ops := loadAPIOperations(t, "notice.json", -1)
		createStrategy := mapByName(ops)["NoticeService.CreateStrategy"]
		assertOperationRegistered(t, createStrategy)
		assertQueryType(t, createStrategy, "trigger_soar_params", "JSON")
		assertArray(t, createStrategy, "email")
		assertArray(t, createStrategy, "sms")
		assertArray(t, createStrategy, "webhook")
	})

	t.Run("SOAR", func(t *testing.T) {
		ops := loadAPIOperations(t, "soar.json", 3)
		byName := mapByName(ops)

		playbookList := byName["AnalysisService.GetSOARPlaybookList"]
		assertOperationRegistered(t, playbookList)
		assertOptional(t, playbookList, "display_name_en")

		jobList := byName["AnalysisService.GetSoarJobExecInfoList"]
		assertOperationRegistered(t, jobList)
		for _, field := range []string{
			"user_id",
			"job_id",
			"model_id",
			"offset",
			"count",
			"trigger_way",
			"organize_id",
			"deploy_name_cn",
			"task_status",
			"start_time",
			"end_time",
		} {
			assertOptional(t, jobList, field)
		}

		trigger := byName["AlarmService.TriggerPlaybookGeneral"]
		assertOperationRegistered(t, trigger)
		assertRequired(t, trigger, "soar_id")
		assertQueryType(t, trigger, "params", "JSON")
	})
}

func TestIntelligenceJSONStructure(t *testing.T) {
	ops := loadAPIOperations(t, "intelligence.json", 5)
	byName := mapByName(ops)

	for _, name := range []string{
		"IntelligenceService.GetIPIocList",
		"IntelligenceService.GetIPIocByIp",
		"IntelligenceService.GetIPIocInfo",
		"IntelligenceService.GetIPIocTag",
		"IntelligenceService.CreateIpIoc",
	} {
		if _, ok := byName[name]; !ok {
			t.Fatalf("missing %s", name)
		}
	}

	t.Run("GetIPIocList", func(t *testing.T) {
		op := byName["IntelligenceService.GetIPIocList"]
		for field, queryType := range map[string]string{
			"ip":               "StringQuery",
			"ioc_src":          "StringQuery",
			"severity":         "IntQuery",
			"confidence_level": "IntQuery",
			"updated_at":       "TimeQuery",
			"invalid_date":     "TimeQuery",
			"is_hit":           "IntQuery",
			"is_block":         "IntQuery",
			"judgments":        "StringQuery",
			"tags":             "StringQuery",
			"state":            "IntQuery",
		} {
			assertQueryType(t, op, field, queryType)
		}
		assertOptional(t, op, "count")
		assertOptional(t, op, "offset")
	})

	t.Run("GetByIP", func(t *testing.T) {
		for _, name := range []string{
			"IntelligenceService.GetIPIocByIp",
			"IntelligenceService.GetIPIocInfo",
			"IntelligenceService.GetIPIocTag",
		} {
			assertRequired(t, byName[name], "ip")
		}
	})

	t.Run("CreateIpIoc", func(t *testing.T) {
		op := byName["IntelligenceService.CreateIpIoc"]
		for _, field := range []string{
			"ip_ioc.ip",
			"ip_ioc.threat_type",
			"ip_ioc.severity",
			"ip_ioc.confidence",
			"ip_ioc.confidence_level",
		} {
			assertRequired(t, op, field)
		}
		for _, field := range []string{
			"ip_ioc.todo_propose",
			"ip_ioc.intel_type",
			"ip_ioc.industry",
			"ip_ioc.ip_port",
			"ip_ioc.domains",
			"ip_ioc.whois",
			"ip_ioc.url",
			"ip_ioc.gangs",
			"ip_ioc.more_info",
			"ip_ioc.ip_country",
			"ip_ioc.ip_country_code",
			"ip_ioc.ip_region",
			"ip_ioc.ip_city",
			"ip_ioc.ip_lng",
			"ip_ioc.ip_lat",
			"ip_ioc.operator",
			"ip_ioc.asn_info",
			"ip_ioc.idc_name",
			"ip_ioc.asn_label",
			"ip_ioc.invalid_date",
			"ip_ioc.pattern",
		} {
			assertOptional(t, op, field)
		}
		assertQueryType(t, op, "threat_tags", "JSON")
		assertQueryType(t, op, "info_tags", "JSON")
		assertArray(t, op, "invalid_tags")
	})

	assertOperationsRegistered(t, ops)
}

func TestLogJSONStructure(t *testing.T) {
	ops := loadAPIOperations(t, "log.json", 5)
	byName := mapByName(ops)

	for _, name := range []string{
		"LogService.SearchLogList",
		"LogService.SearchLogInfo",
		"LogService.SearchAggregationStatistics",
		"LogService.SearchLogListTotal",
		"NonstandardLogApiService.SearchNonstandardLog",
	} {
		if _, ok := byName[name]; !ok {
			t.Fatalf("missing %s", name)
		}
	}

	t.Run("SearchLogList", func(t *testing.T) {
		op := byName["LogService.SearchLogList"]
		assertRequired(t, op, "time_range_start")
		assertRequired(t, op, "time_range_end")
		assertArray(t, op, "keyword")
		assertArray(t, op, "attack_chain_phase")
		assertQueryType(t, op, "filter", "JSON")
		assertQueryType(t, op, "advanced_query", "JSON")
		assertQueryType(t, op, "condition_query", "JSON")
		assertQueryType(t, op, "organization", "UintQuery")
	})

	t.Run("SearchAggregationStatistics", func(t *testing.T) {
		op := byName["LogService.SearchAggregationStatistics"]
		assertRequired(t, op, "time_range_start")
		assertRequired(t, op, "time_range_end")
		assertRequired(t, op, "key")
		assertArray(t, op, "key")
		assertQueryType(t, op, "filter", "JSON")
		assertQueryType(t, op, "advanced_query", "JSON")
		assertQueryType(t, op, "condition_query", "JSON")
		assertQueryType(t, op, "organization", "UintQuery")
	})

	t.Run("SearchLogListTotal", func(t *testing.T) {
		op := byName["LogService.SearchLogListTotal"]
		assertRequired(t, op, "time_range_start")
		assertRequired(t, op, "time_range_end")
		assertQueryType(t, op, "filter", "JSON")
		assertQueryType(t, op, "advanced_query", "JSON")
		assertQueryType(t, op, "condition_query", "JSON")
		assertQueryType(t, op, "organization", "UintQuery")
	})

	t.Run("SearchNonstandardLog", func(t *testing.T) {
		op := byName["NonstandardLogApiService.SearchNonstandardLog"]
		assertRequired(t, op, "start_time")
		assertRequired(t, op, "end_time")
		assertArray(t, op, "key_words")
		assertArray(t, op, "dev_type")
		assertQueryType(t, op, "log_id", "StringQuery")
		assertQueryType(t, op, "non_log_id", "StringQuery")
		assertQueryType(t, op, "device_name", "StringQuery")
		assertQueryType(t, op, "agent_name", "StringQuery")
		assertQueryType(t, op, "condition_query", "JSON")
	})

	assertOperationsRegistered(t, ops)
}

func TestAgentJSONSearchDeviceListStructure(t *testing.T) {
	ops := loadAPIOperations(t, "agent.json", 1)
	op := ops[0]
	if op.Name != "AgentService.SearchDeviceList" {
		t.Fatalf("unexpected API %q", op.Name)
	}

	for field, queryType := range map[string]string{
		"id":                "UintQuery",
		"ip":                "InetQuery",
		"name":              "StringQuery",
		"device_type":       "UintQuery",
		"product_name":      "StringQuery",
		"security_scope_id": "UintQuery",
		"last_receive_time": "TimeQuery",
		"created_time":      "TimeQuery",
		"is_monitoring":     "BoolQuery",
		"status":            "UintQuery",
		"owner_id":          "UintQuery",
		"order_by":          "JSON",
		"receive_check":     "BoolQuery",
	} {
		assertQueryType(t, op, field, queryType)
	}
	assertArray(t, op, "organization_ids")

	parsed := parseOneOperation(op)
	if parsed.skipped {
		t.Fatal("AgentService.SearchDeviceList should be registered, got skipped")
	}
}

func TestOpsJSONSystemMonitoringStructure(t *testing.T) {
	ops := loadAPIOperations(t, "ops.json", -1)
	byName := mapByName(ops)

	for _, name := range []string{
		"OpsService.BigdataNodeList",
		"OpsService.BigdataCpuAvg",
		"OpsService.BigdataMemAvg",
		"OpsService.BigdataDiskIO",
		"OpsService.BigdataNetwork",
		"OpsService.SearchSystemAlarm",
		"OpsService.SearchSystemAlarmModule",
		"DataMgrService.DataMonitor",
		"DataMgrService.GetDataMonitorStatus",
		"ExtraSettingService.GetNodeInfo",
	} {
		assertOperationRegistered(t, byName[name])
	}

	for _, name := range []string{
		"OpsService.BigdataCpuAvg",
		"OpsService.BigdataMemAvg",
		"OpsService.BigdataDiskIO",
		"OpsService.BigdataNetwork",
	} {
		op := byName[name]
		assertRequired(t, op, "start")
		assertRequired(t, op, "end")
		assertRequired(t, op, "ip")
		if op.ParamStyle != paramStyleArray {
			t.Fatalf("%s paramStyle = %q, want %q", op.Name, op.ParamStyle, paramStyleArray)
		}
	}

	systemAlarm := byName["OpsService.SearchSystemAlarm"]
	assertQueryType(t, systemAlarm, "alarmed_at", "TimeQuery")
	assertQueryType(t, systemAlarm, "level", "StringQuery")
	assertQueryType(t, systemAlarm, "module", "StringQuery")
	assertQueryType(t, systemAlarm, "platform", "StringQuery")

	systemAlarmModule := byName["OpsService.SearchSystemAlarmModule"]
	assertQueryType(t, systemAlarmModule, "platform", "StringQuery")

	dataMonitorStatus := byName["DataMgrService.GetDataMonitorStatus"]
	assertOptional(t, dataMonitorStatus, "time_unit")

	getNodeInfo := byName["ExtraSettingService.GetNodeInfo"]
	if len(getNodeInfo.Args) != 0 {
		t.Fatalf("ExtraSettingService.GetNodeInfo args len = %d, want 0", len(getNodeInfo.Args))
	}
}

func TestBigdataMetricDryRunUsesArrayParams(t *testing.T) {
	oldServerURL := serverURL
	oldAPIToken := apiToken
	oldDryRun := dryRun
	t.Cleanup(func() {
		serverURL = oldServerURL
		apiToken = oldAPIToken
		dryRun = oldDryRun
	})

	serverURL = "https://cosmos.example.com"
	apiToken = "test-token"
	dryRun = true

	cmd := &cobra.Command{Use: "test"}
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := doRequestWithParamStyle(cmd, "OpsService.BigdataCpuAvg", map[string]interface{}{
		"start": float64(1781515764),
		"end":   float64(1781519364),
		"ip":    "10.2.37.87",
	}, true, paramStyleArray)
	if err != nil {
		t.Fatalf("doRequestWithParamStyle() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		`"method":"OpsService.BigdataCpuAvg"`,
		`"params":[{"end":1781519364,"ip":"10.2.37.87","start":1781515764}]`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, got)
		}
	}
}

func TestSystemMonitoringCommandsRegistered(t *testing.T) {
	cmd := NewCommand()

	for _, path := range [][]string{
		{"ops", "bigdata-node-list"},
		{"ops", "bigdata-cpu-avg"},
		{"ops", "bigdata-mem-avg"},
		{"ops", "bigdata-disk-io"},
		{"ops", "bigdata-network"},
		{"ops", "search-system-alarm"},
		{"ops", "search-system-alarm-module"},
		{"notice", "get-strategy-list"},
		{"notice", "get-strategy"},
		{"data-mgr", "data-monitor"},
		{"data-mgr", "get-data-monitor-status"},
		{"extra-setting", "get-node-info"},
	} {
		if _, _, err := cmd.Find(path); err != nil {
			t.Fatalf("missing command %v: %v", path, err)
		}
	}
}

func TestVulnJSONStructure(t *testing.T) {
	ops := loadAPIOperations(t, "vuln.json", 6)
	byName := mapByName(ops)

	for _, name := range []string{
		"ScanVulnIpService.SearchScanVulnIpList",
		"ScanVulnIpService.SearchScanVulnIpDetail",
		"ScanVulnIpService.UpsertScanVulnIp",
		"ScanVulnIpService.SearchScanVulnWebList",
		"ScanVulnIpService.GetWebVulnDetail",
		"ScanVulnIpService.UpsertScanVulnWeb",
	} {
		if _, ok := byName[name]; !ok {
			t.Fatalf("missing %s", name)
		}
	}

	t.Run("SearchScanVulnIpList", func(t *testing.T) {
		op := byName["ScanVulnIpService.SearchScanVulnIpList"]
		assertRequired(t, op, "count")
		assertOptional(t, op, "offset")
		assertVulnListQueryTypes(t, op, "vuln_ip")
		assertQueryType(t, op, "port", "UintQuery")
		assertQueryType(t, op, "service", "StringQuery")
		assertOptional(t, op, "sort_field")
		assertOptional(t, op, "sort_order")
	})

	t.Run("SearchScanVulnWebList", func(t *testing.T) {
		op := byName["ScanVulnIpService.SearchScanVulnWebList"]
		assertOptional(t, op, "count")
		assertOptional(t, op, "offset")
		assertVulnListQueryTypes(t, op, "vuln_url")
		assertOptional(t, op, "sort_field")
		assertOptional(t, op, "sort_order")
	})

	t.Run("Details", func(t *testing.T) {
		for _, name := range []string{
			"ScanVulnIpService.SearchScanVulnIpDetail",
			"ScanVulnIpService.GetWebVulnDetail",
		} {
			op := byName[name]
			assertRequired(t, op, "vuln_id")
		}
	})

	t.Run("UpsertScanVulnIp", func(t *testing.T) {
		op := byName["ScanVulnIpService.UpsertScanVulnIp"]
		for _, field := range []string{
			"name",
			"organization_id",
			"vuln_major_category",
			"vuln_category",
			"vuln_ip",
			"description",
		} {
			assertRequired(t, op, field)
		}
		for _, field := range []string{"id", "find_by", "cvss", "vuln_data_type"} {
			assertOptional(t, op, field)
		}
		assertArray(t, op, "file_id")
		assertArray(t, op, "vuln_tag")
		assertQueryType(t, op, "extra_field", "JSON")
	})

	t.Run("UpsertScanVulnWeb", func(t *testing.T) {
		op := byName["ScanVulnIpService.UpsertScanVulnWeb"]
		for _, field := range []string{
			"name",
			"organization_id",
			"vuln_major_category",
			"vuln_category",
			"vuln_url",
			"description",
		} {
			assertRequired(t, op, field)
		}
		for _, field := range []string{"id", "find_by", "cvss_score", "vuln_location", "vuln_data_type"} {
			assertOptional(t, op, field)
		}
		assertArray(t, op, "file_id")
		assertArray(t, op, "vuln_tag")
		assertQueryType(t, op, "extra_field", "JSON")
	})

	assertOperationsRegistered(t, ops)
}

func loadAPIOperations(t *testing.T, filename string, wantLen int) []APIOperation {
	t.Helper()
	data, err := apiSpecs.ReadFile("apis/" + filename)
	if err != nil {
		t.Fatalf("failed to read %s: %v", filename, err)
	}

	var ops []APIOperation
	if err := json.Unmarshal(data, &ops); err != nil {
		t.Fatalf("failed to parse %s: %v", filename, err)
	}
	if wantLen >= 0 && len(ops) != wantLen {
		t.Fatalf("expected %d %s APIs, got %d", wantLen, strings.TrimSuffix(filename, ".json"), len(ops))
	}
	return ops
}

func mapByName(ops []APIOperation) map[string]APIOperation {
	byName := make(map[string]APIOperation, len(ops))
	for _, op := range ops {
		byName[op.Name] = op
	}
	return byName
}

func assertOperationsRegistered(t *testing.T, ops []APIOperation) {
	t.Helper()
	for _, parsed := range parseOperations(ops) {
		if parsed.skipped {
			t.Fatalf("%s.%s should be registered, got skipped", parsed.serviceName, parsed.methodName)
		}
	}
}

func assertOperationRegistered(t *testing.T, op APIOperation) {
	t.Helper()
	if op.Name == "" {
		t.Fatal("missing operation")
	}
	parsed := parseOneOperation(op)
	if parsed.skipped {
		t.Fatalf("%s should be registered, got skipped", op.Name)
	}
}

func assertRequired(t *testing.T, op APIOperation, field string) {
	t.Helper()
	arg, ok := op.Args[field]
	if !ok {
		t.Fatalf("%s missing field %q", op.Name, field)
	}
	if arg.Optional {
		t.Fatalf("%s field %q should be required", op.Name, field)
	}
}

func assertOptional(t *testing.T, op APIOperation, field string) {
	t.Helper()
	arg, ok := op.Args[field]
	if !ok {
		t.Fatalf("%s missing field %q", op.Name, field)
	}
	if !arg.Optional {
		t.Fatalf("%s field %q should be optional", op.Name, field)
	}
}

func assertArray(t *testing.T, op APIOperation, field string) {
	t.Helper()
	arg, ok := op.Args[field]
	if !ok {
		t.Fatalf("%s missing field %q", op.Name, field)
	}
	if arg.TypeClass != "array" {
		t.Fatalf("%s field %q typeclass = %q, want array", op.Name, field, arg.TypeClass)
	}
}

func assertQueryType(t *testing.T, op APIOperation, field string, want string) {
	t.Helper()
	arg, ok := op.Args[field]
	if !ok {
		t.Fatalf("%s missing field %q", op.Name, field)
	}
	if arg.QueryType != want {
		t.Fatalf("%s field %q queryType = %q, want %q", op.Name, field, arg.QueryType, want)
	}
	if resolveTypeName(arg.Type) != "string" {
		t.Fatalf("%s field %q type = %q, want string", op.Name, field, resolveTypeName(arg.Type))
	}
}

func assertVulnListQueryTypes(t *testing.T, op APIOperation, assetField string) {
	t.Helper()
	for field, queryType := range map[string]string{
		"id":                "UintQuery",
		"organization_id":   "UintQuery",
		"name":              "StringQuery",
		"vuln_category":     "StringQuery",
		"vuln_level":        "IntQuery",
		assetField:          "StringQuery",
		"merge_num":         "UintQuery",
		"updated_at":        "TimeQuery",
		"created_at":        "TimeQuery",
		"vuln_update_time":  "TimeQuery",
		"vuln_dispose_type": "UintQuery",
		"vuln_status":       "UintQuery",
		"vuln_tag":          "StringQuery",
		"task_name":         "StringQuery",
		"defect_founder":    "StringQuery",
		"defect_verifier":   "StringQuery",
		"defect_fixer":      "StringQuery",
		"defect_rechecker":  "StringQuery",
		"defect_closer":     "StringQuery",
		"found_time":        "TimeQuery",
		"verify_time":       "TimeQuery",
		"fix_time":          "TimeQuery",
		"recheck_time":      "TimeQuery",
		"close_time":        "TimeQuery",
		"is_match_vuln":     "BoolQuery",
		"is_match_asset":    "BoolQuery",
		"find_by":           "StringQuery",
		"user":              "StringQuery",
		"password":          "StringQuery",
		"vuln_data_type":    "UintQuery",
		"asset_owner_id":    "UintQuery",
		"condition_query":   "JSON",
	} {
		assertQueryType(t, op, field, queryType)
	}
}
