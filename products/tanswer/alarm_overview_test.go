package tanswer

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAlarmOverviewHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "alarm", "overview", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查看威胁告警概览", "不返回原始告警列表", "--time", "--severity", "--result", "输出"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestFetchAlarmOverviewFallsBackToListWhenAggregationsNeedLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		switch req.Method {
		case "AlarmService.SearchAlarmCount":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &rpcError{Code: 1, Message: "需要登录"},
			})
		case "AlarmService.SearchAlarmList":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"total": 3,
					"page_total": 1,
					"data": [
						{"severity": 1, "result": "success", "phase": "lateral", "tag": "WebShell", "attacker": "198.51.100.10", "victim": "192.0.2.10"},
						{"severity": 2, "result": "success", "phase": "control", "tag": "SQL注入", "attacker": "198.51.100.10", "victim": "192.0.2.11"},
						{"severity": 2, "result": "control", "phase": "control", "tag": "SQL注入", "attacker": "203.0.113.20", "victim": "192.0.2.11"}
					]
				}`),
			})
		default:
			t.Fatalf("unexpected method %q", req.Method)
		}
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, APIToken: "token-123", Timeout: time.Second})
	data, err := fetchAlarmOverview(context.Background(), client, TimeRange{Start: 1, End: 2}, nil)
	if err != nil {
		t.Fatalf("fetchAlarmOverview returned error: %v", err)
	}

	summary, ok := data["summary"].(map[string]any)
	if !ok {
		t.Fatalf("summary = %#v", data["summary"])
	}
	if summary["alarm_total"] != int64(3) || summary["source"] != "list_fallback" {
		t.Fatalf("summary = %#v", summary)
	}
	if got := data["attacker_top"].([]map[string]any)[0]; got["key"] != "198.51.100.10" || got["count"] != 2 {
		t.Fatalf("attacker_top[0] = %#v", got)
	}
	if got := data["threat_type_top"].([]map[string]any)[0]; got["key"] != "SQL注入" || got["count"] != 2 {
		t.Fatalf("threat_type_top[0] = %#v", got)
	}
}
