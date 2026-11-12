package tanswer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAlarmListHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "alarm", "list", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查询威胁告警列表", "原始告警列表", "--page-size", "--severity", "--attacker", "--asset-ip"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildAlarmListRequestMapsSemanticFilters(t *testing.T) {
	rng := TimeRange{Start: 1000, End: 2000}
	req := buildAlarmListRequest(rng, alarmListOptions{
		page:      2,
		pageSize:  20,
		severity:  "critical,high",
		result:    "success,control",
		phase:     "recon",
		attacker:  "198.51.100.10",
		victim:    "203.0.113.20",
		assetIP:   "192.0.2.10,192.0.2.11",
		keyword:   "webshell,sql",
		name:      "SQL注入",
		tag:       "代码执行",
		direction: "in",
		appProto:  "http",
		url:       "/login",
		host:      "example.com",
		xff:       "203.0.113.30",
		srcPort:   "12345",
		destPort:  "443",
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
	if got := req["result"].([]map[string]any)[1]["target"]; got != "control" {
		t.Fatalf("result target = %#v", got)
	}
	if got := req["asset_ip"].([]string); len(got) != 2 || got[1] != "192.0.2.11" {
		t.Fatalf("asset_ip = %#v", got)
	}
	if got := req["keyword"].([]string); len(got) != 2 || got[0] != "webshell" {
		t.Fatalf("keyword = %#v", got)
	}
	if got := req["dest_port"].([]map[string]any)[0]["target"]; got != 443 {
		t.Fatalf("dest_port target = %#v", got)
	}
}

func TestAlarmListCommandCallsSearchAlarmList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AlarmService.SearchAlarmList" {
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
		urlFilter := params["url"].([]any)
		firstURLFilter := urlFilter[0].(map[string]any)
		if firstURLFilter["target"] != "/login" {
			t.Fatalf("url filter = %#v", urlFilter)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"data":[{"doc_id":"doc-1","name":"SQL注入","severity":2,"attacker":"198.51.100.10","victim":"203.0.113.20","result":"success","payload":"large-payload","appbrief":{"http":{"url":"/login","hostname":"example.com"}}}],
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
		"alarm", "list",
		"--time", "24h",
		"--page-size", "5",
		"--http-url", "/login",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查询威胁告警列表" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["total"] != float64(1) {
		t.Fatalf("total = %#v", data["total"])
	}
	alarms := data["alarms"].([]any)
	alarm := alarms[0].(map[string]any)
	if _, ok := alarm["payload"]; ok {
		t.Fatalf("list summary should not include payload: %#v", alarm)
	}
	if _, ok := alarm["appbrief"]; ok {
		t.Fatalf("list summary should not include raw appbrief: %#v", alarm)
	}
	if alarm["url"] != "/login" || alarm["host"] != "example.com" {
		t.Fatalf("summary url/host mismatch: %#v", alarm)
	}
}
