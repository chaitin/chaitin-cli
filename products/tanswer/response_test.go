package tanswer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseBlockPoliciesHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "response", "block-policies", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查询旁路阻断策略", "阻断对象", "--page-size", "--status", "输出"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildResponseBlockPoliciesRequestMapsFilters(t *testing.T) {
	req := buildResponseBlockPoliciesRequest(responseBlockPoliciesOptions{
		page:       2,
		pageSize:   20,
		id:         "7",
		strategyID: "3",
		name:       "block bad ip",
		object:     "198.51.100.10",
		status:     "enabled",
	})

	if req["offset"] != int64(20) || req["count"] != int64(20) {
		t.Fatalf("pagination mismatch: %#v", req)
	}
	if got := req["id"].([]map[string]any)[0]["target"]; got != 7 {
		t.Fatalf("id target = %#v", got)
	}
	if got := req["strategy_id"].([]map[string]any)[0]["target"]; got != 3 {
		t.Fatalf("strategy_id target = %#v", got)
	}
	if got := req["name"].([]map[string]any)[0]["target"]; got != "block bad ip" {
		t.Fatalf("name target = %#v", got)
	}
	if req["ip_search"] != "198.51.100.10" {
		t.Fatalf("ip_search = %#v", req["ip_search"])
	}
	if got := req["status"].([]map[string]any)[0]["target"]; got != 1 {
		t.Fatalf("status target = %#v", got)
	}
}

func TestResponseBlockPoliciesCommandCallsSearchBlockRules(t *testing.T) {
	server := newResponseRPCServer(t, "RulesService.SearchBlockRules", `{"data":[{"id":7,"name":"block bad ip","ips":"198.51.100.10","strategy_id":3,"strategy_name":"auto","status":1,"expire":1784277612410,"created_at":1784277000000,"remark":"manual"}],"total":1}`, func(params map[string]any) {
		if params["count"] != float64(5) || params["offset"] != float64(0) {
			t.Fatalf("params = %#v", params)
		}
	})
	defer server.Close()

	env := runResponseCommand(t, server.URL, "response", "block-policies", "--page-size", "5")
	if env.Task != "查询旁路阻断策略" {
		t.Fatalf("Task = %q", env.Task)
	}
	requireResponseDataKey(t, env, "block_policies")
}

func TestResponseBlockRecordsCommandCallsTapBlockRecordList(t *testing.T) {
	server := newResponseRPCServer(t, "RulesService.SearchTapBlockRecordList", `{"records":[{"id":"r1","timestamp":1784277612410,"src_ip":"198.51.100.10","src_port":12345,"dest_ip":"192.0.2.10","dest_port":80,"policy_id":7,"policy_name":"block bad ip","type":1,"block_times":2}],"total":1}`, func(params map[string]any) {
		if params["count"] != float64(5) || params["offset"] != float64(0) || params["src_ip"] != "198.51.100.10" {
			t.Fatalf("params = %#v", params)
		}
	})
	defer server.Close()

	env := runResponseCommand(t, server.URL, "response", "block-records", "--time", "24h", "--page-size", "5", "--src-ip", "198.51.100.10")
	if env.Task != "查询旁路阻断记录" {
		t.Fatalf("Task = %q", env.Task)
	}
	requireResponseDataKey(t, env, "block_records")
}

func TestResponseWhitelistCommandCallsFirewallWhiteList(t *testing.T) {
	server := newResponseRPCServer(t, "FirewallService.SearchWhiteList", `{"data":[{"id":3,"type":"ip","values":["198.51.100.10"],"remark":"trusted","status":1,"expire":"2026-07-17T10:00:00Z","updated_at":"2026-07-17T09:00:00Z","block_method":["Bypass"],"ip_type":"src"}],"total":1}`, nil)
	defer server.Close()

	env := runResponseCommand(t, server.URL, "response", "whitelist", "--page-size", "5", "--object", "198.51.100.10")
	if env.Task != "查询响应白名单" {
		t.Fatalf("Task = %q", env.Task)
	}
	requireResponseDataKey(t, env, "response_whitelists")
}

func TestResponseWhitelistCreateHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "response", "whitelist-create", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"新增响应白名单", "--type", "--object", "--expire", "--block-method", "--ip-type", "--confirm", "CONFIRM_RESPONSE_WHITELIST_CREATE", "输出预览"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildResponseWhitelistWriteRequestMapsFields(t *testing.T) {
	req, err := buildResponseWhitelistWriteRequest(responseWhitelistWriteOptions{
		objectType:  "ip",
		objects:     "198.51.100.10,203.0.113.20",
		status:      "enabled",
		expire:      "1784277612410",
		blockMethod: "bypass,third",
		ipType:      "source",
		remark:      "trusted responder source",
	})
	if err != nil {
		t.Fatalf("buildResponseWhitelistWriteRequest returned error: %v", err)
	}
	if req["type"] != "ip" || req["status"] != 2 || req["expire"] != int64(1784277612410) || req["remark"] != "trusted responder source" || req["ip_type"] != "SRC_IP" {
		t.Fatalf("request = %#v", req)
	}
	values := req["values"].([]string)
	if len(values) != 2 || values[0] != "198.51.100.10" || values[1] != "203.0.113.20" {
		t.Fatalf("values = %#v", values)
	}
	methods := req["block_method"].([]string)
	if len(methods) != 2 || methods[0] != "Bypass" || methods[1] != "Third_party" {
		t.Fatalf("block_method = %#v", methods)
	}
}

func TestResponseWhitelistCreatePreviewDoesNotCallRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("preview must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "response", "whitelist-create", "--type", "ip", "--object", "198.51.100.10", "--expire", "1784277612410", "--preview"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend RPC was called during preview")
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "新增响应白名单预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmation_token"] != "CONFIRM_RESPONSE_WHITELIST_CREATE" || data["confirmed"] != false {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestResponseWhitelistCreateConfirmedCallsCreate(t *testing.T) {
	server := newResponseRPCSequenceServer(t, []responseExpectedRPC{{
		method: "FirewallService.CreateWhiteList",
		check: func(t *testing.T, params map[string]any) {
			if params["type"] != "ip" || params["status"] != float64(2) || params["expire"] != float64(1784277612410) || params["ip_type"] != "SRC_OR_DST" {
				t.Fatalf("create params = %#v", params)
			}
			values := params["values"].([]any)
			if len(values) != 1 || values[0] != "198.51.100.10" {
				t.Fatalf("values = %#v", values)
			}
			methods := params["block_method"].([]any)
			if len(methods) != 2 || methods[0] != "Bypass" || methods[1] != "Third_party" {
				t.Fatalf("block_method = %#v", methods)
			}
		},
		result: `{"id": 3}`,
	}})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "response", "whitelist-create", "--type", "ip", "--object", "198.51.100.10", "--expire", "1784277612410", "--confirm", "CONFIRM_RESPONSE_WHITELIST_CREATE"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 1 {
		t.Fatalf("request count = %d", server.requestCount)
	}
}

func TestResponseWhitelistUpdateConfirmedReadsBeforeAndCallsUpdate(t *testing.T) {
	server := newResponseRPCSequenceServer(t, []responseExpectedRPC{
		{method: "FirewallService.SearchWhiteList", result: `{"data":[{"id":3,"type":"ip","values":["198.51.100.10"],"remark":"old","status":2,"expire":"2026-07-17T10:00:00Z","block_method":["Bypass"],"ip_type":"SRC_OR_DST"}],"total":1}`},
		{
			method: "FirewallService.UpdateWhiteList",
			check: func(t *testing.T, params map[string]any) {
				if params["id"] != float64(3) || params["type"] != "url" || params["status"] != float64(1) || params["expire"] != float64(1784277612410) {
					t.Fatalf("update params = %#v", params)
				}
				values := params["values"].([]any)
				if len(values) != 1 || values[0] != "http://example.com/a" {
					t.Fatalf("values = %#v", values)
				}
				if _, ok := params["block_method"]; ok {
					t.Fatalf("url whitelist must not send block_method: %#v", params)
				}
			},
			result: `{}`,
		},
	})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "response", "whitelist-update", "--id", "3", "--type", "url", "--object", "http://example.com/a", "--status", "disabled", "--expire", "1784277612410", "--confirm", "CONFIRM_RESPONSE_WHITELIST_UPDATE"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 2 {
		t.Fatalf("request count = %d", server.requestCount)
	}
}

func TestResponseWhitelistStatusAndDeleteConfirmedCallRPC(t *testing.T) {
	for _, tc := range []struct {
		args       []string
		method     string
		wantStatus float64
		task       string
	}{
		{[]string{"response", "whitelist-enable", "--id-list", "3,4", "--confirm", "CONFIRM_RESPONSE_WHITELIST_ENABLE"}, "FirewallService.UpdateWhiteListStatus", 2, "启用响应白名单"},
		{[]string{"response", "whitelist-disable", "--id-list", "3,4", "--confirm", "CONFIRM_RESPONSE_WHITELIST_DISABLE"}, "FirewallService.UpdateWhiteListStatus", 1, "停用响应白名单"},
		{[]string{"response", "whitelist-delete", "--id-list", "3,4", "--confirm", "CONFIRM_RESPONSE_WHITELIST_DELETE"}, "FirewallService.DeleteWhiteList", 0, "删除响应白名单"},
	} {
		t.Run(tc.task, func(t *testing.T) {
			server := newResponseRPCSequenceServer(t, []responseExpectedRPC{
				{method: "FirewallService.SearchWhiteList", result: `{"data":[{"id":3,"type":"ip","values":["198.51.100.10"],"status":2},{"id":4,"type":"url","values":["http://example.com/a"],"status":1}],"total":2}`},
				{
					method: tc.method,
					check: func(t *testing.T, params map[string]any) {
						ids := params["ids"].([]any)
						if len(ids) != 2 || ids[0] != float64(3) || ids[1] != float64(4) {
							t.Fatalf("ids = %#v", ids)
						}
						if tc.wantStatus != 0 && params["status"] != tc.wantStatus {
							t.Fatalf("status = %#v", params["status"])
						}
					},
					result: `{}`,
				},
			})
			defer server.Close()

			var out bytes.Buffer
			cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
			cmd.SetArgs(append([]string{"tanswer", "--url", server.URL, "--api-key", "token-123"}, tc.args...))

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if server.requestCount != 2 {
				t.Fatalf("request count = %d", server.requestCount)
			}
		})
	}
}

func TestResponseBlockPolicyFromAlarmHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "response", "block-policy-from-alarm", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"从告警生成旁路阻断策略", "--id", "doc_id", "--target", "--duration", "--confirm", "CONFIRM_RESPONSE_BLOCK_POLICY_FROM_ALARM", "输出预览"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildResponseBlockPolicyFromAlarmRequestMapsAlarmFields(t *testing.T) {
	req, err := buildResponseBlockPolicyFromAlarmRequest(map[string]any{
		"doc_id":    "doc-1",
		"name":      "SQL注入",
		"src_ip":    "198.51.100.10",
		"src_port":  float64(12345),
		"dest_ip":   "192.0.2.10",
		"dest_port": float64(80),
	}, responseFromAlarmOptions{
		id:     "doc-1",
		target: "flow",
		block: responseBlockPolicyWriteOptions{
			duration: "7200",
			remark:   "confirmed malicious flow",
		},
	})
	if err != nil {
		t.Fatalf("buildResponseBlockPolicyFromAlarmRequest returned error: %v", err)
	}
	if req["block_type"] != uint(3) || req["block_time_value"] != int64(7200) || req["remark"] != "confirmed malicious flow" {
		t.Fatalf("request = %#v", req)
	}
	ips := req["ips"].([]string)
	if len(ips) != 1 || ips[0] != "198.51.100.10:12345-192.0.2.10:80" {
		t.Fatalf("ips = %#v", ips)
	}
	if !strings.Contains(req["name"].(string), "SQL注入") {
		t.Fatalf("name = %#v", req["name"])
	}
}

func TestResponseBlockPolicyFromAlarmPreviewReadsAlarmOnly(t *testing.T) {
	server := newResponseRPCSequenceServer(t, []responseExpectedRPC{{
		method: "AlarmService.GetAlarm",
		check: func(t *testing.T, params map[string]any) {
			if params["doc_id"] != "doc-1" {
				t.Fatalf("doc_id = %#v", params["doc_id"])
			}
		},
		result: `{"data":{"doc_id":"doc-1","name":"SQL注入","src_ip":"198.51.100.10","dest_ip":"192.0.2.10","dest_port":80}}`,
	}})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "response", "block-policy-from-alarm", "--id", "doc-1", "--preview"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 1 {
		t.Fatalf("request count = %d", server.requestCount)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "从告警生成旁路阻断策略预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmation_token"] != "CONFIRM_RESPONSE_BLOCK_POLICY_FROM_ALARM" || data["confirmed"] != false {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestResponseBlockPolicyFromAlarmConfirmedCreatesBlockPolicy(t *testing.T) {
	server := newResponseRPCSequenceServer(t, []responseExpectedRPC{
		{
			method: "AlarmService.GetAlarm",
			result: `{"data":{"doc_id":"doc-1","name":"SQL注入","src_ip":"198.51.100.10","dest_ip":"192.0.2.10","dest_port":80}}`,
		},
		{
			method: "RulesService.CreateBlockRules",
			check: func(t *testing.T, params map[string]any) {
				if params["block_type"] != float64(1) || params["block_time_value"] != float64(3600) {
					t.Fatalf("create params = %#v", params)
				}
				ips := params["ips"].([]any)
				if len(ips) != 1 || ips[0] != "198.51.100.10" {
					t.Fatalf("ips = %#v", ips)
				}
			},
			result: `{"id": 17}`,
		},
	})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "response", "block-policy-from-alarm", "--id", "doc-1", "--target", "attacker", "--confirm", "CONFIRM_RESPONSE_BLOCK_POLICY_FROM_ALARM"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 2 {
		t.Fatalf("request count = %d", server.requestCount)
	}
}

func TestBuildResponseWhitelistFromAlarmRequestMapsAlarmFields(t *testing.T) {
	req, err := buildResponseWhitelistFromAlarmRequest(map[string]any{
		"doc_id":  "doc-1",
		"name":    "SQL注入",
		"dest_ip": "192.0.2.10",
		"appbrief": map[string]any{
			"http": map[string]any{"url": "http://example.com/a"},
		},
	}, responseFromAlarmOptions{
		id:     "doc-1",
		target: "url",
		white: responseWhitelistWriteOptions{
			expire: "1784277612410",
			remark: "trusted url",
		},
	})
	if err != nil {
		t.Fatalf("buildResponseWhitelistFromAlarmRequest returned error: %v", err)
	}
	if req["type"] != "url" || req["status"] != 2 || req["expire"] != int64(1784277612410) || req["remark"] != "trusted url" {
		t.Fatalf("request = %#v", req)
	}
	values := req["values"].([]string)
	if len(values) != 1 || values[0] != "http://example.com/a" {
		t.Fatalf("values = %#v", values)
	}
}

func TestResponseWhitelistFromAlarmConfirmedCreatesWhitelist(t *testing.T) {
	server := newResponseRPCSequenceServer(t, []responseExpectedRPC{
		{
			method: "AlarmService.GetAlarm",
			result: `{"data":{"doc_id":"doc-1","name":"SQL注入","src_ip":"198.51.100.10","dest_ip":"192.0.2.10"}}`,
		},
		{
			method: "FirewallService.CreateWhiteList",
			check: func(t *testing.T, params map[string]any) {
				if params["type"] != "ip" || params["status"] != float64(2) || params["expire"] != float64(1784277612410) {
					t.Fatalf("create params = %#v", params)
				}
				values := params["values"].([]any)
				if len(values) != 1 || values[0] != "192.0.2.10" {
					t.Fatalf("values = %#v", values)
				}
			},
			result: `{"id": 18}`,
		},
	})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "response", "whitelist-from-alarm", "--id", "doc-1", "--target", "victim", "--expire", "1784277612410", "--confirm", "CONFIRM_RESPONSE_WHITELIST_FROM_ALARM"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 2 {
		t.Fatalf("request count = %d", server.requestCount)
	}
}

func TestResponseDevicesCommandCallsFirewallSearch(t *testing.T) {
	server := newResponseRPCServer(t, "FirewallService.SearchFirewall", `{"data":[{"id":2,"device_type":1,"remark":"fw","status":1,"addr":"192.0.2.11","updated_at":"2026-07-17T09:00:00Z"}],"total":1}`, nil)
	defer server.Close()

	env := runResponseCommand(t, server.URL, "response", "devices", "--page-size", "5")
	if env.Task != "查询联动设备配置" {
		t.Fatalf("Task = %q", env.Task)
	}
	requireResponseDataKey(t, env, "devices")
}

func TestResponseDeviceRecordsCommandCallsSendRecord(t *testing.T) {
	server := newResponseRPCServer(t, "FirewallService.SearchSendRecord", `{"data":[{"id":9,"ip":"198.51.100.10","last_try_time":"2026-07-17 09:00:00","last_result":1}],"total":1}`, func(params map[string]any) {
		if params["firewall_id"] != float64(2) {
			t.Fatalf("firewall_id = %#v", params["firewall_id"])
		}
	})
	defer server.Close()

	env := runResponseCommand(t, server.URL, "response", "device-records", "--device-id", "2", "--page-size", "5")
	if env.Task != "查询联动设备处置记录" {
		t.Fatalf("Task = %q", env.Task)
	}
	requireResponseDataKey(t, env, "device_records")
}

func TestResponseAutoPoliciesCommandCallsSearchStrategy(t *testing.T) {
	server := newResponseRPCServer(t, "FirewallService.SearchStrategy", `{"data":[{"id":5,"name":"auto block","firewall_id":[2],"status":1,"punish_type":1,"updated_at":"2026-07-17T09:00:00Z","block_type":["Bypass"],"remark":"auto"}],"total":1}`, nil)
	defer server.Close()

	env := runResponseCommand(t, server.URL, "response", "auto-policies", "--page-size", "5")
	if env.Task != "查询自动响应策略" {
		t.Fatalf("Task = %q", env.Task)
	}
	requireResponseDataKey(t, env, "auto_policies")
}

func TestResponseAutoListCommandCallsSearchBlackList(t *testing.T) {
	server := newResponseRPCServer(t, "FirewallService.SearchBlackList", `{"data":[{"id":8,"ip":"198.51.100.10","port":0,"status":1,"block_time_type":1,"block_time_value":3600,"strategy_id":5,"strategy_name":"auto block","update_time":"2026-07-17T09:00:00Z"}],"total":1}`, nil)
	defer server.Close()

	env := runResponseCommand(t, server.URL, "response", "auto-list", "--time", "7d", "--page-size", "5")
	if env.Task != "查询自动响应处置名单" {
		t.Fatalf("Task = %q", env.Task)
	}
	requireResponseDataKey(t, env, "auto_list")
}

func TestResponseBlockPolicyCreateHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "response", "block-policy-create", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"新增旁路阻断策略", "--name", "--object", "--object-type", "--duration", "--confirm", "CONFIRM_RESPONSE_BLOCK_POLICY_CREATE", "输出预览"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildResponseBlockPolicyWriteRequestMapsFields(t *testing.T) {
	req, err := buildResponseBlockPolicyWriteRequest(responseBlockPolicyWriteOptions{
		name:          "block bad ip",
		objects:       "198.51.100.10,203.0.113.20",
		objectType:    "ip",
		ipType:        "source",
		status:        "enabled",
		blockTimeType: "duration",
		duration:      "3600",
		remark:        "confirmed malicious",
	})
	if err != nil {
		t.Fatalf("buildResponseBlockPolicyWriteRequest returned error: %v", err)
	}
	if req["name"] != "block bad ip" || req["block_type"] != uint(1) || req["status"] != uint(1) || req["block_time_type"] != int32(1) || req["block_time_value"] != int64(3600) {
		t.Fatalf("request = %#v", req)
	}
	ips := req["ips"].([]string)
	if len(ips) != 2 || ips[0] != "198.51.100.10" || ips[1] != "203.0.113.20" {
		t.Fatalf("ips = %#v", ips)
	}
	ipType := req["ip_type"].(*uint)
	if *ipType != uint(1) {
		t.Fatalf("ip_type = %#v", *ipType)
	}
}

func TestResponseBlockPolicyCreatePreviewDoesNotCallRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("preview must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "response", "block-policy-create", "--name", "block bad ip", "--object", "198.51.100.10", "--preview"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend RPC was called during preview")
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "新增旁路阻断策略预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmation_token"] != "CONFIRM_RESPONSE_BLOCK_POLICY_CREATE" || data["confirmed"] != false {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestResponseBlockPolicyCreateConfirmedCallsCreate(t *testing.T) {
	server := newResponseRPCSequenceServer(t, []responseExpectedRPC{{
		method: "RulesService.CreateBlockRules",
		check: func(t *testing.T, params map[string]any) {
			if params["name"] != "block bad ip" || params["block_type"] != float64(1) || params["block_time_type"] != float64(1) || params["block_time_value"] != float64(3600) {
				t.Fatalf("create params = %#v", params)
			}
			ips := params["ips"].([]any)
			if len(ips) != 1 || ips[0] != "198.51.100.10" {
				t.Fatalf("ips = %#v", ips)
			}
		},
		result: `{"id": 7}`,
	}})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "response", "block-policy-create", "--name", "block bad ip", "--object", "198.51.100.10", "--duration", "3600", "--confirm", "CONFIRM_RESPONSE_BLOCK_POLICY_CREATE"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 1 {
		t.Fatalf("request count = %d", server.requestCount)
	}
}

func TestResponseBlockPolicyUpdateConfirmedReadsBeforeAndCallsUpdate(t *testing.T) {
	server := newResponseRPCSequenceServer(t, []responseExpectedRPC{
		{
			method: "RulesService.SearchBlockRules",
			check: func(t *testing.T, params map[string]any) {
				idFilters := params["id"].([]any)
				if len(idFilters) != 1 {
					t.Fatalf("id filters = %#v", idFilters)
				}
				first := idFilters[0].(map[string]any)
				if first["oper"] != "=" || first["target"] != float64(7) {
					t.Fatalf("id filter = %#v", first)
				}
			},
			result: `{"data":[{"id":7,"name":"old","ips":"198.51.100.10","status":1,"block_type":1,"expire":1784277612410}],"total":1}`,
		},
		{
			method: "RulesService.UpdateBlockRules",
			check: func(t *testing.T, params map[string]any) {
				if params["id"] != float64(7) || params["name"] != "new block" || params["expire"] != float64(1784277612410) {
					t.Fatalf("update params = %#v", params)
				}
			},
			result: `{"id": 7}`,
		},
	})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "response", "block-policy-update", "--id", "7", "--name", "new block", "--object", "198.51.100.10", "--expire", "1784277612410", "--confirm", "CONFIRM_RESPONSE_BLOCK_POLICY_UPDATE"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 2 {
		t.Fatalf("request count = %d", server.requestCount)
	}
}

func TestResponseBlockPolicyStatusAndDeleteConfirmedCallRPC(t *testing.T) {
	for _, tc := range []struct {
		args   []string
		action string
		task   string
	}{
		{[]string{"response", "block-policy-enable", "--id-list", "7,8", "--confirm", "CONFIRM_RESPONSE_BLOCK_POLICY_ENABLE"}, "show", "启用旁路阻断策略"},
		{[]string{"response", "block-policy-disable", "--id-list", "7,8", "--confirm", "CONFIRM_RESPONSE_BLOCK_POLICY_DISABLE"}, "hide", "停用旁路阻断策略"},
		{[]string{"response", "block-policy-delete", "--id-list", "7,8", "--confirm", "CONFIRM_RESPONSE_BLOCK_POLICY_DELETE"}, "delete", "删除旁路阻断策略"},
	} {
		t.Run(tc.task, func(t *testing.T) {
			server := newResponseRPCSequenceServer(t, []responseExpectedRPC{
				{
					method: "RulesService.SearchBlockRules",
					check: func(t *testing.T, params map[string]any) {
						idFilters := params["id"].([]any)
						if len(idFilters) != 2 {
							t.Fatalf("id filters = %#v", idFilters)
						}
						first := idFilters[0].(map[string]any)
						second := idFilters[1].(map[string]any)
						if first["oper"] != "=" || first["target"] != float64(7) || second["oper"] != "=" || second["target"] != float64(8) {
							t.Fatalf("id filters = %#v", idFilters)
						}
					},
					result: `{"data":[{"id":7,"name":"A","ips":"198.51.100.10","status":1},{"id":8,"name":"B","ips":"203.0.113.20","status":1}],"total":2}`,
				},
				{
					method: "RulesService.UpdateBlockRulesStatus",
					check: func(t *testing.T, params map[string]any) {
						ids := params["ids"].([]any)
						if len(ids) != 2 || ids[0] != float64(7) || ids[1] != float64(8) {
							t.Fatalf("ids = %#v", ids)
						}
						if params["action"] != tc.action {
							t.Fatalf("action = %#v", params["action"])
						}
					},
					result: `{"ids":[7,8]}`,
				},
			})
			defer server.Close()

			var out bytes.Buffer
			cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
			cmd.SetArgs(append([]string{"tanswer", "--url", server.URL, "--api-key", "token-123"}, tc.args...))

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if server.requestCount != 2 {
				t.Fatalf("request count = %d", server.requestCount)
			}
		})
	}
}

func newResponseRPCServer(t *testing.T, wantMethod string, result string, check func(map[string]any)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != wantMethod {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if check != nil {
			check(params)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(result),
		})
	}))
}

type responseExpectedRPC struct {
	method string
	check  func(t *testing.T, params map[string]any)
	result string
}

type responseRPCSequenceServer struct {
	*httptest.Server
	requestCount int
}

func newResponseRPCSequenceServer(t *testing.T, expected []responseExpectedRPC) *responseRPCSequenceServer {
	t.Helper()
	state := &responseRPCSequenceServer{}
	state.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if state.requestCount >= len(expected) {
			t.Fatalf("unexpected extra RPC request")
		}
		want := expected[state.requestCount]
		state.requestCount++
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != want.method {
			t.Fatalf("method = %q, want %q", req.Method, want.method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if want.check != nil {
			want.check(t, params)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(want.result),
		})
	}))
	return state
}

func runResponseCommand(t *testing.T, addr string, args ...string) SuccessEnvelope {
	t.Helper()
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs(append([]string{"tanswer", "--url", addr, "--api-key", "token-123"}, args...))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	return env
}

func requireResponseDataKey(t *testing.T, env SuccessEnvelope, key string) {
	t.Helper()
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if _, ok := data[key]; !ok {
		t.Fatalf("missing data key %q in %#v", key, data)
	}
}
