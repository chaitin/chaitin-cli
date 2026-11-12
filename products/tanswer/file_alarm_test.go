package tanswer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFileAlarmHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "file-alarm", "malicious", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查询恶意文件告警", "文件名", "MD5", "SHA256", "--page-size", "--severity"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildFileAlarmListRequestMapsSemanticFilters(t *testing.T) {
	rng := TimeRange{Start: 1000, End: 2000}
	req := buildFileAlarmListRequest(rng, fileAlarmListOptions{
		page:     2,
		pageSize: 20,
		severity: "critical,high",
		tag:      "elf",
		fileType: "virus,backdoor",
		srcIP:    "198.51.100.10",
		destIP:   "192.0.2.10",
		appProto: "http",
		srcPort:  "12345",
		destPort: "443",
		keyword:  "shell.php",
	})

	if req["time_range_start"] != int64(1000) || req["time_range_end"] != int64(2000) {
		t.Fatalf("time range mismatch: %#v", req)
	}
	if req["offset"] != int64(20) || req["count"] != int64(20) {
		t.Fatalf("pagination mismatch: %#v", req)
	}
	if got := req["severity"].([]map[string]any)[0]["target"]; got != 1 {
		t.Fatalf("severity target = %#v", got)
	}
	if got := req["tag"].([]map[string]any)[0]["target"]; got != "elf" {
		t.Fatalf("tag target = %#v", got)
	}
	if got := req["type"].([]map[string]any)[1]["target"]; got != "backdoor" {
		t.Fatalf("type target = %#v", got)
	}
	if got := req["keyword"].([]string); len(got) != 1 || got[0] != "shell.php" {
		t.Fatalf("keyword = %#v", got)
	}
}

func TestFileAlarmMaliciousCommandCallsSearchAlarmFdetectList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AlarmService.SearchAlarmFdetectList" {
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
		tag := params["tag"].([]any)[0].(map[string]any)
		if tag["target"] != "elf" {
			t.Fatalf("tag filter = %#v", params["tag"])
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"data":[{"id":"fd-1","timestamp":1784282400,"severity":2,"tag":"elf","type":"virus","filename":"evil.exe","src_ip":"198.51.100.10","dest_ip":"192.0.2.10","app_proto":"http","md5":"m1","sha256":"s1","sandbox":{"score":9.5}}],
				"total": 1,
				"page_total": 1
			}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"file-alarm", "malicious",
		"--time", "24h",
		"--page-size", "5",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查询恶意文件告警" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	items := data["file_alarms"].([]any)
	item := items[0].(map[string]any)
	if item["filename"] != "evil.exe" || item["md5"] != "m1" || item["sha256"] != "s1" {
		t.Fatalf("summary mismatch: %#v", item)
	}
}

func TestFileAlarmWebshellCommandUsesWebshellFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AlarmService.SearchAlarmFdetectList" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		tag := params["tag"].([]any)[0].(map[string]any)
		if tag["target"] != "webshell" {
			t.Fatalf("tag filter = %#v", params["tag"])
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":[{"id":"ws-1","tag":"webshell","type":"php_webshell","filename":"cmd.php","md5":"m2","sha256":"s2"}],"total":1,"page_total":1}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "file-alarm", "webshell"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查询 Webshell 告警" {
		t.Fatalf("Task = %q", env.Task)
	}
}

func TestFileAlarmSandboxCommandCallsSearchSandboxAlarmFdetectList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AlarmService.SearchSandboxAlarmFdetectList" {
			t.Fatalf("method = %q", req.Method)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":[{"id":"sb-1","filename":"sample.doc","operating_env":"Windows 7","score":8.8,"src_ip":"198.51.100.10","dest_ip":"192.0.2.10"}],"total":1,"page_total":1}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "file-alarm", "sandbox"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查询沙箱检测告警" {
		t.Fatalf("Task = %q", env.Task)
	}
}

func TestFileAlarmDetailCommandCallsGetAlarmFdetect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AlarmService.GetAlarmFdetect" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["doc_id"] != "fd-1" {
			t.Fatalf("doc_id = %#v", params["doc_id"])
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":{"id":"fd-1","tag":"elf","filename":"evil.exe","md5":"m1","sha256":"s1","sandbox":{"info":{"score":9.5}}}}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "file-alarm", "detail", "--id", "fd-1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查看文件告警详情" {
		t.Fatalf("Task = %q", env.Task)
	}
}

func TestFileAlarmOverviewCommandCallsCountAndAggTop(t *testing.T) {
	seen := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		seen[req.Method] = true
		switch req.Method {
		case "AlarmService.SearchAlarmFdetectCount":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"total":3}`),
			})
		case "AlarmService.SearchAlarmFdetectAggTop":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":[{"key":"elf","doc_count":2},{"key":"webshell","doc_count":1}],"total":2,"page_total":2}`),
			})
		default:
			t.Fatalf("unexpected method %q", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "file-alarm", "overview"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !seen["AlarmService.SearchAlarmFdetectCount"] || !seen["AlarmService.SearchAlarmFdetectAggTop"] {
		t.Fatalf("missing expected RPC calls: %#v", seen)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查看文件告警概览" {
		t.Fatalf("Task = %q", env.Task)
	}
}
