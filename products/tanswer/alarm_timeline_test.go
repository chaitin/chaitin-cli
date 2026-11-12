package tanswer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAlarmTimelineHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "alarm", "timeline", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查看威胁告警趋势", "曲线", "--time", "--interval", "--severity", "输出"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAlarmTimelineCommandCallsSearchAlarmListChart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AlarmService.SearchAlarmListChart" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["interval"] != "1h" || params["range_mode"] != float64(0) {
			t.Fatalf("timeline params = %#v", params)
		}
		urlFilter := params["url"].([]any)
		firstURLFilter := urlFilter[0].(map[string]any)
		if firstURLFilter["target"] != "/login" {
			t.Fatalf("url filter = %#v", urlFilter)
		}
		severity := params["severity"].([]any)
		firstSeverity := severity[0].(map[string]any)
		if firstSeverity["target"] != float64(1) {
			t.Fatalf("severity filter = %#v", severity)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"data": [
					{"key": 1719392400000, "doc_count": 2, "severity": {"buckets": [{"key": 1, "doc_count": 1}, {"key": 2, "doc_count": 1}]}},
					{"key": 1719396000000, "doc_count": 0, "severity": {"buckets": []}}
				]
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
		"alarm", "timeline",
		"--time", "24h",
		"--interval", "1h",
		"--severity", "critical,high",
		"--http-url", "/login",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查看威胁告警趋势" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["interval"] != "1h" {
		t.Fatalf("interval = %#v", data["interval"])
	}
	points := data["points"].([]any)
	firstPoint := points[0].(map[string]any)
	if firstPoint["doc_count"] != float64(2) {
		t.Fatalf("first point = %#v", firstPoint)
	}
}
