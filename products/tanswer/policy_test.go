package tanswer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPolicyHelpListsCustomIntelligenceWriteCommands(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "policy", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"预览和精确确认", "custom-intelligence-create", "custom-intelligence-update", "custom-intelligence-delete"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestPolicyDetectionWhitelistHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "policy", "detection-whitelist", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查询检测白名单", "误报抑制", "--page-size", "--src-ip", "--status", "输出"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildPolicyDetectionWhitelistRequestMapsFilters(t *testing.T) {
	req := buildPolicyDetectionWhitelistRequest(policyDetectionWhitelistOptions{
		page:     2,
		pageSize: 20,
		name:     "误报",
		srcIP:    "198.51.100.10",
		destIP:   "192.0.2.10",
		status:   "enabled",
		ruleID:   "sid-1",
		threat:   "SQL注入",
	})

	if req["offset"] != int64(20) || req["count"] != int64(20) {
		t.Fatalf("pagination mismatch: %#v", req)
	}
	if got := req["name"].([]map[string]any)[0]["target"]; got != "误报" {
		t.Fatalf("name target = %#v", got)
	}
	if got := req["src_ip"].([]map[string]any)[0]["target"]; got != "198.51.100.10" {
		t.Fatalf("src_ip target = %#v", got)
	}
	if got := req["status"].([]map[string]any)[0]["target"]; got != 1 {
		t.Fatalf("status target = %#v", got)
	}
	if req["sid"] != "sid-1" || req["type"] != "SQL注入" {
		t.Fatalf("sid/type mismatch: %#v", req)
	}
}

func TestPolicyDetectionWhitelistCommandCallsSearchWhiteList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AlarmService.SearchWhiteList" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["count"] != float64(5) || params["offset"] != float64(0) {
			t.Fatalf("pagination params = %#v", params)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"data":[{"id":7,"name":"误报白名单","src_ip":"198.51.100.10","dest_ip":"192.0.2.10","domain":"example.com","url_path":"/login","user_agent":"curl","xff":"203.0.113.20","resp_status_code":"200","resp_body":"ok","type":"SQL注入","sid":"1001","updated_at":"2026-07-17T10:00:00Z","expire":1784277612410,"status":1,"remark":"case closed"}],
				"total":1
			}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "policy", "detection-whitelist", "--page-size", "5", "--status", "enabled"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查询检测白名单" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	items := data["detection_whitelists"].([]any)
	item := items[0].(map[string]any)
	if item["id"] != float64(7) || item["name"] != "误报白名单" || item["domain"] != "example.com" {
		t.Fatalf("summary mismatch: %#v", item)
	}
}

func TestPolicyDetectionWhitelistCreateHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "policy", "detection-whitelist-create", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"新增检测白名单", "--name", "--src-ip", "--dest-ip", "--storage", "--confirm", "CONFIRM_POLICY_DETECTION_WHITELIST_CREATE", "输出预览"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildPolicyDetectionWhitelistWriteRequestMapsFields(t *testing.T) {
	req, err := buildPolicyDetectionWhitelistWriteRequest(policyDetectionWhitelistWriteOptions{
		name:        "登录误报",
		srcIP:       "198.51.100.10",
		srcPort:     "443",
		destIP:      "192.0.2.10",
		destPort:    "8443",
		domain:      "example.com",
		urlPath:     "login",
		userAgent:   "curl",
		xff:         "203.0.113.20",
		respCode:    "40X",
		respBody:    "ok",
		threat:      "SQL注入",
		ruleID:      "1001",
		status:      "enabled",
		storage:     "ignore",
		defaultMode: "default",
		expire:      "1784277612000",
		validTime:   "3600",
		ignore:      true,
		remark:      "case closed",
	})
	if err != nil {
		t.Fatalf("buildPolicyDetectionWhitelistWriteRequest returned error: %v", err)
	}
	if req["name"] != "登录误报" || req["src_ip"] != "198.51.100.10" || req["dest_ip"] != "192.0.2.10" {
		t.Fatalf("request fields = %#v", req)
	}
	if req["url_path"] != "/login" || req["resp_status_code"] != "40x" {
		t.Fatalf("normalized fields = %#v", req)
	}
	if req["status"] != uint(1) || req["storage"] != uint(2) || req["default_mode"] != uint(1) || req["expire"] != int64(1784277612000) || req["valid_time"] != int64(3600) || req["ignore"] != true {
		t.Fatalf("numeric fields = %#v", req)
	}
}

func TestBuildPolicyDetectionWhitelistWriteRequestRejectsNonPositiveValidTime(t *testing.T) {
	for _, validTime := range []string{"0", "-1"} {
		t.Run(validTime, func(t *testing.T) {
			_, err := buildPolicyDetectionWhitelistWriteRequest(policyDetectionWhitelistWriteOptions{
				name:      "登录误报",
				srcIP:     "198.51.100.10",
				validTime: validTime,
			})
			if err == nil || !strings.Contains(err.Error(), "valid-time") {
				t.Fatalf("valid-time %q error = %v", validTime, err)
			}
		})
	}
}

func TestBuildPolicyDetectionWhitelistWriteRequestRequiresExpiry(t *testing.T) {
	_, err := buildPolicyDetectionWhitelistWriteRequest(policyDetectionWhitelistWriteOptions{
		name:  "登录误报",
		srcIP: "198.51.100.10",
	})
	if err == nil || !strings.Contains(err.Error(), "--expire or --valid-time") {
		t.Fatalf("missing expiry error = %v", err)
	}
}

func TestBuildPolicyDetectionWhitelistWriteRequestAcceptsExpireWithoutValidTime(t *testing.T) {
	req, err := buildPolicyDetectionWhitelistWriteRequest(policyDetectionWhitelistWriteOptions{
		name:   "登录误报",
		srcIP:  "198.51.100.10",
		expire: "1784277612000",
	})
	if err != nil {
		t.Fatalf("buildPolicyDetectionWhitelistWriteRequest returned error: %v", err)
	}
	if req["expire"] != int64(1784277612000) || req["valid_time"] != int64(0) {
		t.Fatalf("expiry fields = %#v", req)
	}
}

func TestPolicyDetectionWhitelistCreatePreviewDoesNotCallRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("preview must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "detection-whitelist-create",
		"--name", "登录误报",
		"--src-ip", "198.51.100.10",
		"--valid-time", "3600",
		"--preview",
	})

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
	if env.Task != "新增检测白名单预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["requires_confirmation"] != true || data["confirmed"] != false || data["confirmation_token"] != "CONFIRM_POLICY_DETECTION_WHITELIST_CREATE" {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestPolicyDetectionWhitelistCreateRequiresExactConfirmBeforeRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("invalid confirmation must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "detection-whitelist-create",
		"--name", "登录误报",
		"--src-ip", "198.51.100.10",
		"--valid-time", "3600",
		"--confirm", "confirm_policy_detection_whitelist_create",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend RPC was called with invalid confirmation")
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Error.Code != "POLICY_DETECTION_WHITELIST_CREATE_CONFIRMATION_REQUIRED" {
		t.Fatalf("error = %#v", env.Error)
	}
}

func TestPolicyDetectionWhitelistCreateConfirmedCallsCreate(t *testing.T) {
	server := newPolicyRPCSequenceServer(t, []policyExpectedRPC{
		{
			method: "AlarmService.CreateWhiteList",
			check: func(t *testing.T, params map[string]any) {
				if params["name"] != "登录误报" || params["src_ip"] != "198.51.100.10" || params["default_mode"] != float64(1) {
					t.Fatalf("params = %#v", params)
				}
			},
			result: `{"id": 21}`,
		},
	})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "detection-whitelist-create",
		"--name", "登录误报",
		"--src-ip", "198.51.100.10",
		"--valid-time", "3600",
		"--confirm", "CONFIRM_POLICY_DETECTION_WHITELIST_CREATE",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "新增检测白名单" {
		t.Fatalf("Task = %q", env.Task)
	}
}

func TestPolicyDetectionWhitelistUpdateConfirmedReadsBeforeAndCallsUpdate(t *testing.T) {
	server := newPolicyRPCSequenceServer(t, []policyExpectedRPC{
		{
			method: "AlarmService.SearchWhiteList",
			check: func(t *testing.T, params map[string]any) {
				if params["count"] != float64(1) || params["offset"] != float64(0) {
					t.Fatalf("search params = %#v", params)
				}
			},
			result: `{"data":[{"id":21,"name":"旧白名单","src_ip":"198.51.100.10","status":1,"storage":1,"default_mode":1}],"total":1}`,
		},
		{
			method: "AlarmService.UpdateWhiteList",
			check: func(t *testing.T, params map[string]any) {
				if params["id"] != float64(21) || params["name"] != "新白名单" || params["src_ip"] != "198.51.100.11" {
					t.Fatalf("update params = %#v", params)
				}
			},
			result: `{"id":21}`,
		},
	})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "detection-whitelist-update",
		"--id", "21",
		"--name", "新白名单",
		"--src-ip", "198.51.100.11",
		"--valid-time", "3600",
		"--confirm", "CONFIRM_POLICY_DETECTION_WHITELIST_UPDATE",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 2 {
		t.Fatalf("request count = %d", server.requestCount)
	}
}

func TestPolicyDetectionWhitelistStatusAndDeleteConfirmedCallRPC(t *testing.T) {
	for _, tc := range []struct {
		args   []string
		action any
		task   string
	}{
		{[]string{"policy", "detection-whitelist-enable", "--id-list", "21,22", "--confirm", "CONFIRM_POLICY_DETECTION_WHITELIST_ENABLE"}, "show", "启用检测白名单"},
		{[]string{"policy", "detection-whitelist-disable", "--id-list", "21,22", "--confirm", "CONFIRM_POLICY_DETECTION_WHITELIST_DISABLE"}, "hide", "禁用检测白名单"},
		{[]string{"policy", "detection-whitelist-delete", "--id-list", "21,22", "--confirm", "CONFIRM_POLICY_DETECTION_WHITELIST_DELETE"}, "delete", "删除检测白名单"},
	} {
		t.Run(tc.task, func(t *testing.T) {
			server := newPolicyRPCSequenceServer(t, []policyExpectedRPC{
				{
					method: "AlarmService.SearchWhiteList",
					result: `{"data":[{"id":21,"name":"白名单A","src_ip":"198.51.100.10","status":1},{"id":22,"name":"白名单B","src_ip":"198.51.100.11","status":1}],"total":2}`,
				},
				{
					method: "AlarmService.DeleteWhiteList",
					check: func(t *testing.T, params map[string]any) {
						ids := params["ids"].([]any)
						if len(ids) != 2 || ids[0] != float64(21) || ids[1] != float64(22) {
							t.Fatalf("ids = %#v", ids)
						}
						if params["action"] != tc.action {
							t.Fatalf("action = %#v", params["action"])
						}
					},
					result: `{"ids":[21,22]}`,
				},
			})
			defer server.Close()

			var out bytes.Buffer
			cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
			allArgs := append([]string{"tanswer", "--url", server.URL, "--api-key", "token-123"}, tc.args...)
			cmd.SetArgs(allArgs)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if server.requestCount != 2 {
				t.Fatalf("request count = %d", server.requestCount)
			}
		})
	}
}

func TestPolicyDetectionWhitelistFromAlarmHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "policy", "detection-whitelist-from-alarm", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"从告警对象生成检测白名单", "--id", "doc_id", "--confirm", "CONFIRM_POLICY_DETECTION_WHITELIST_FROM_ALARM", "输出预览"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildPolicyDetectionWhitelistFromAlarmRequestMapsAlarmFields(t *testing.T) {
	req, err := buildPolicyDetectionWhitelistFromAlarmRequest(map[string]any{
		"doc_id":    "doc-1",
		"name":      "SQL注入",
		"src_ip":    "198.51.100.10",
		"src_port":  float64(12345),
		"dest_ip":   "192.0.2.10",
		"dest_port": float64(80),
		"sid":       "1001",
		"tag":       "SQL注入",
		"xff":       "203.0.113.20",
		"appbrief": map[string]any{
			"http": map[string]any{
				"hostname":   "example.com",
				"url":        "/login",
				"user_agent": "curl",
			},
		},
	}, policyDetectionWhitelistFromAlarmOptions{
		id: "doc-1",
		write: policyDetectionWhitelistWriteOptions{
			remark:    "confirmed false positive",
			validTime: "3600",
		},
	})
	if err != nil {
		t.Fatalf("buildPolicyDetectionWhitelistFromAlarmRequest returned error: %v", err)
	}
	for key, want := range map[string]any{
		"src_ip":     "198.51.100.10",
		"src_port":   "12345",
		"dest_ip":    "192.0.2.10",
		"dest_port":  "80",
		"domain":     "example.com",
		"url_path":   "/login",
		"user_agent": "curl",
		"xff":        "203.0.113.20",
		"sid":        "1001",
		"type":       "SQL注入",
		"remark":     "confirmed false positive",
	} {
		if req[key] != want {
			t.Fatalf("%s = %#v, want %#v; req=%#v", key, req[key], want, req)
		}
	}
	if !strings.Contains(fmt.Sprint(req["name"]), "SQL注入") {
		t.Fatalf("name = %#v", req["name"])
	}
}

func TestPolicyDetectionWhitelistFromAlarmPreviewReadsAlarmOnly(t *testing.T) {
	server := newPolicyRPCSequenceServer(t, []policyExpectedRPC{
		{
			method: "AlarmService.GetAlarm",
			check: func(t *testing.T, params map[string]any) {
				if params["doc_id"] != "doc-1" {
					t.Fatalf("doc_id = %#v", params["doc_id"])
				}
			},
			result: `{"data":{"doc_id":"doc-1","name":"SQL注入","src_ip":"198.51.100.10","dest_ip":"192.0.2.10","dest_port":80,"sid":"1001","tag":"SQL注入"}}`,
		},
	})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "detection-whitelist-from-alarm",
		"--id", "doc-1",
		"--valid-time", "3600",
		"--preview",
	})

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
	if env.Task != "从告警对象生成检测白名单预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmation_token"] != "CONFIRM_POLICY_DETECTION_WHITELIST_FROM_ALARM" || data["confirmed"] != false {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestPolicyDetectionWhitelistFromAlarmConfirmedCreatesWhitelist(t *testing.T) {
	server := newPolicyRPCSequenceServer(t, []policyExpectedRPC{
		{
			method: "AlarmService.GetAlarm",
			result: `{"data":{"doc_id":"doc-1","name":"SQL注入","src_ip":"198.51.100.10","dest_ip":"192.0.2.10","dest_port":80,"sid":"1001","tag":"SQL注入"}}`,
		},
		{
			method: "AlarmService.CreateWhiteList",
			check: func(t *testing.T, params map[string]any) {
				if params["src_ip"] != "198.51.100.10" || params["dest_ip"] != "192.0.2.10" || params["sid"] != "1001" || params["type"] != "SQL注入" {
					t.Fatalf("create params = %#v", params)
				}
			},
			result: `{"id": 31}`,
		},
	})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "detection-whitelist-from-alarm",
		"--id", "doc-1",
		"--valid-time", "3600",
		"--confirm", "CONFIRM_POLICY_DETECTION_WHITELIST_FROM_ALARM",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 2 {
		t.Fatalf("request count = %d", server.requestCount)
	}
}

func TestPolicyDetectionWhitelistExportDownloadsFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "detection-whitelist.xlsx")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertPolicyDownloadRequest(t, r, "AlarmDownloadService.ExportWhiteList", []float64{21, 22})
		w.Header().Set("Content-Disposition", `attachment; filename="detection-whitelist.xlsx";`)
		_, _ = io.WriteString(w, "export-bytes")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "detection-whitelist-export",
		"--id-list", "21,22",
		"--output", outputPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	rawFile, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(rawFile) != "export-bytes" {
		t.Fatalf("file content = %q", string(rawFile))
	}
}

func TestPolicyCustomIntelligenceImportConfirmedUploadsFileWithImportPattern(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertPolicyUploadRequest(t, r, "AlarmUploadService.ImportAlarmCustomIntelligence", "ioc.xlsx", `{"pattern":"import"}`)
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      "upload",
			Result:  json.RawMessage(`{"record_count":{"record_total":2,"success_total":1,"fail_total":1,"cover_total":0,"failure_reason":[{"row":2,"reason":"invalid ioc"}]}}`),
		})
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "ioc.xlsx")
	if err := os.WriteFile(path, []byte("xlsx-bytes"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "custom-intelligence-import",
		"--file", path,
		"--confirm", "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_IMPORT",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	result := data["import_result"].(map[string]any)
	if result["record_total"] != float64(2) || result["fail_total"] != float64(1) {
		t.Fatalf("import_result = %#v", result)
	}
}

func TestPolicyCustomIntelligenceCommandCallsSearchList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AlarmService.SearchAlarmCustomIntelligenceList" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["count"] != float64(10) || params["offset"] != float64(0) {
			t.Fatalf("pagination params = %#v", params)
		}
		if got := params["ioc"].([]any)[0].(map[string]any)["target"]; got != "evil.example" {
			t.Fatalf("ioc target = %#v", got)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":[{"id":3,"name":"恶意域名","ioc":["evil.example"],"type":2,"updated_at":"1784277612","status":1,"remarks":"manual"}],"total":1}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "policy", "custom-intelligence", "--ioc", "evil.example", "--status", "enabled"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查询自定义情报" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	items := data["custom_intelligence"].([]any)
	item := items[0].(map[string]any)
	if item["id"] != float64(3) || item["name"] != "恶意域名" {
		t.Fatalf("summary mismatch: %#v", item)
	}
}

func TestPolicyCustomIntelligenceCreateHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "policy", "custom-intelligence-create", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"新增自定义情报", "--name", "--ioc", "--type", "--confirm", "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_CREATE", "输出预览"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildPolicyCustomIntelligenceWriteRequestMapsFields(t *testing.T) {
	req, err := buildPolicyCustomIntelligenceWriteRequest(policyCustomIntelligenceWriteOptions{
		name:    "恶意域名",
		ioc:     "evil.example, c2.example",
		iocType: "domain",
		status:  "enabled",
		remarks: "manual",
	})
	if err != nil {
		t.Fatalf("buildPolicyCustomIntelligenceWriteRequest returned error: %v", err)
	}
	if req["name"] != "恶意域名" || req["type"] != 2 || req["status"] != uint(1) || req["remarks"] != "manual" {
		t.Fatalf("request fields = %#v", req)
	}
	iocs := req["ioc"].([]string)
	if len(iocs) != 2 || iocs[0] != "evil.example" || iocs[1] != "c2.example" {
		t.Fatalf("ioc = %#v", iocs)
	}
}

func TestPolicyCustomIntelligenceCreatePreviewDoesNotCallRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("preview must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "custom-intelligence-create",
		"--name", "恶意域名",
		"--ioc", "evil.example",
		"--type", "domain",
		"--preview",
	})

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
	if env.Task != "新增自定义情报预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["requires_confirmation"] != true || data["confirmed"] != false || data["confirmation_token"] != "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_CREATE" {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestPolicyCustomIntelligenceCreateRequiresExactConfirmBeforeRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("invalid confirmation must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "custom-intelligence-create",
		"--name", "恶意域名",
		"--ioc", "evil.example",
		"--type", "domain",
		"--confirm", "confirm_policy_custom_intelligence_create",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend RPC was called with invalid confirmation")
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Error.Code != "POLICY_CUSTOM_INTELLIGENCE_CREATE_CONFIRMATION_REQUIRED" {
		t.Fatalf("error = %#v", env.Error)
	}
}

func TestPolicyCustomIntelligenceCreateConfirmedCallsCreate(t *testing.T) {
	server := newPolicyRPCSequenceServer(t, []policyExpectedRPC{
		{
			method: "AlarmService.CreateAlarmCustomIntelligence",
			check: func(t *testing.T, params map[string]any) {
				if params["name"] != "恶意域名" || params["type"] != float64(2) || params["status"] != float64(1) {
					t.Fatalf("params = %#v", params)
				}
				iocs := params["ioc"].([]any)
				if len(iocs) != 1 || iocs[0] != "evil.example" {
					t.Fatalf("ioc = %#v", iocs)
				}
			},
			result: `{"id": 12}`,
		},
	})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "custom-intelligence-create",
		"--name", "恶意域名",
		"--ioc", "evil.example",
		"--type", "domain",
		"--confirm", "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_CREATE",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "新增自定义情报" {
		t.Fatalf("Task = %q", env.Task)
	}
}

func TestPolicyCustomIntelligenceUpdateConfirmedReadsBeforeAndCallsUpdate(t *testing.T) {
	server := newPolicyRPCSequenceServer(t, []policyExpectedRPC{
		{
			method: "AlarmService.SearchAlarmCustomIntelligenceInfo",
			check: func(t *testing.T, params map[string]any) {
				if params["id"] != float64(12) {
					t.Fatalf("info params = %#v", params)
				}
			},
			result: `{"data":{"id":12,"name":"旧情报","ioc":["old.example"],"type":2,"status":1,"remarks":"old"}}`,
		},
		{
			method: "AlarmService.UpdateAlarmCustomIntelligence",
			check: func(t *testing.T, params map[string]any) {
				if params["id"] != float64(12) || params["name"] != "新情报" || params["type"] != float64(2) {
					t.Fatalf("update params = %#v", params)
				}
			},
			result: `{}`,
		},
	})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"policy", "custom-intelligence-update",
		"--id", "12",
		"--name", "新情报",
		"--ioc", "evil.example",
		"--type", "domain",
		"--confirm", "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_UPDATE",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 2 {
		t.Fatalf("request count = %d", server.requestCount)
	}
}

func TestPolicyCustomIntelligenceStatusAndDeleteConfirmedCallRPC(t *testing.T) {
	for _, tc := range []struct {
		args   []string
		method string
		status any
		task   string
	}{
		{[]string{"policy", "custom-intelligence-enable", "--id-list", "12,13", "--confirm", "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_ENABLE"}, "AlarmService.UpdateAlarmCustomIntelligenceStatus", float64(1), "启用自定义情报"},
		{[]string{"policy", "custom-intelligence-disable", "--id-list", "12,13", "--confirm", "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_DISABLE"}, "AlarmService.UpdateAlarmCustomIntelligenceStatus", float64(0), "禁用自定义情报"},
		{[]string{"policy", "custom-intelligence-delete", "--id-list", "12,13", "--confirm", "CONFIRM_POLICY_CUSTOM_INTELLIGENCE_DELETE"}, "AlarmService.DeleteAlarmCustomIntelligence", nil, "删除自定义情报"},
	} {
		t.Run(tc.task, func(t *testing.T) {
			server := newPolicyRPCSequenceServer(t, []policyExpectedRPC{
				{
					method: "AlarmService.SearchAlarmCustomIntelligenceList",
					check: func(t *testing.T, params map[string]any) {
						idFilters := params["id"].([]any)
						if len(idFilters) != 2 {
							t.Fatalf("id filters = %#v", idFilters)
						}
						first := idFilters[0].(map[string]any)
						second := idFilters[1].(map[string]any)
						if first["oper"] != "=" || first["target"] != float64(12) || second["oper"] != "=" || second["target"] != float64(13) {
							t.Fatalf("id filters = %#v", idFilters)
						}
					},
					result: `{"data":[{"id":12,"name":"情报A","ioc":["a.example"],"type":2,"status":1},{"id":13,"name":"情报B","ioc":["b.example"],"type":2,"status":1}],"total":2}`,
				},
				{
					method: tc.method,
					check: func(t *testing.T, params map[string]any) {
						ids := params["ids"].([]any)
						if len(ids) != 2 || ids[0] != float64(12) || ids[1] != float64(13) {
							t.Fatalf("ids = %#v", ids)
						}
						if tc.status != nil && params["status"] != tc.status {
							t.Fatalf("status = %#v", params["status"])
						}
					},
					result: `{}`,
				},
			})
			defer server.Close()

			var out bytes.Buffer
			cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
			allArgs := append([]string{"tanswer", "--url", server.URL, "--api-key", "token-123"}, tc.args...)
			cmd.SetArgs(allArgs)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if server.requestCount != 2 {
				t.Fatalf("request count = %d", server.requestCount)
			}
		})
	}
}

type policyExpectedRPC struct {
	method string
	check  func(t *testing.T, params map[string]any)
	result string
}

type policyRPCSequenceServer struct {
	*httptest.Server
	requestCount int
}

func newPolicyRPCSequenceServer(t *testing.T, expected []policyExpectedRPC) *policyRPCSequenceServer {
	t.Helper()
	state := &policyRPCSequenceServer{}
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

func assertPolicyDownloadRequest(t *testing.T, r *http.Request, method string, wantIDs []float64) {
	t.Helper()
	if r.URL.Path != "/api/download" {
		t.Fatalf("path = %q", r.URL.Path)
	}
	if r.URL.Query().Get("id") != method {
		t.Fatalf("id query = %q", r.URL.Query().Get("id"))
	}
	if r.Header.Get("Api-Token") != "token-123" {
		t.Fatalf("Api-Token header = %q", r.Header.Get("Api-Token"))
	}
	raw, err := base64.StdEncoding.DecodeString(r.URL.Query().Get("query"))
	if err != nil {
		t.Fatalf("decode query: %v", err)
	}
	var query map[string]any
	if err := json.Unmarshal(raw, &query); err != nil {
		t.Fatalf("decode query json: %v", err)
	}
	ids := query["ids"].([]any)
	if len(ids) != len(wantIDs) {
		t.Fatalf("ids = %#v", ids)
	}
	for i, want := range wantIDs {
		if ids[i] != want {
			t.Fatalf("ids[%d] = %#v, want %#v", i, ids[i], want)
		}
	}
}

func assertPolicyUploadRequest(t *testing.T, r *http.Request, method string, fileName string, param string) {
	t.Helper()
	if r.URL.Path != "/api/upload" {
		t.Fatalf("path = %q", r.URL.Path)
	}
	if r.URL.Query().Get("id") != method {
		t.Fatalf("id query = %q", r.URL.Query().Get("id"))
	}
	if r.Header.Get("Api-Token") != "token-123" {
		t.Fatalf("Api-Token header = %q", r.Header.Get("Api-Token"))
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	if got := r.MultipartForm.Value["param"]; len(got) != 1 || got[0] != param {
		t.Fatalf("param field = %#v", got)
	}
	files := r.MultipartForm.File["file"]
	if len(files) != 1 || files[0].Filename != fileName {
		t.Fatalf("file headers = %#v", files)
	}
}
