package tanswer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAlarmRelatedHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "alarm", "related", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查看相关告警", "--id", "--window", "--relation", "30m"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildAlarmRelatedListRequestUsesWindowAndRelation(t *testing.T) {
	original := map[string]any{
		"doc_id":    "doc-1",
		"timestamp": float64(1000000),
		"attacker":  "198.51.100.10",
		"victim":    "203.0.113.20",
	}
	req, err := buildAlarmRelatedListRequest(original, "attacker", 10*60*1000, 20)
	if err != nil {
		t.Fatalf("buildAlarmRelatedListRequest returned error: %v", err)
	}
	if req["time_range_start"] != int64(400000) || req["time_range_end"] != int64(1600000) {
		t.Fatalf("time range mismatch: %#v", req)
	}
	if req["count"] != int64(20) {
		t.Fatalf("count = %#v", req["count"])
	}
	attacker := req["attacker"].([]map[string]any)
	if attacker[0]["target"] != "198.51.100.10" {
		t.Fatalf("attacker filter = %#v", attacker)
	}
}

func TestAlarmRelatedCommandFetchesDetailAndRelatedAlarms(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		switch len(methods) {
		case 1:
			if req.Method != "AlarmService.GetAlarm" {
				t.Fatalf("first method = %q", req.Method)
			}
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":{"doc_id":"doc-1","timestamp":1000000,"attacker":"198.51.100.10","victim":"203.0.113.20","name":"SQL注入","severity":2}}`),
			})
		case 2:
			if req.Method != "AlarmService.SearchAlarmList" {
				t.Fatalf("second method = %q", req.Method)
			}
			raw, _ := json.Marshal(req.Params)
			var params map[string]any
			if err := json.Unmarshal(raw, &params); err != nil {
				t.Fatalf("decode params: %v", err)
			}
			if params["time_range_start"] != float64(-800000) || params["time_range_end"] != float64(2800000) {
				t.Fatalf("window params = %#v", params)
			}
			if _, ok := params["attacker"]; !ok {
				t.Fatalf("missing attacker filter: %#v", params)
			}
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"data":[
						{"doc_id":"doc-1","timestamp":1000000,"attacker":"198.51.100.10","victim":"203.0.113.20","name":"original"},
						{"doc_id":"doc-2","timestamp":1100000,"attacker":"198.51.100.10","victim":"203.0.113.21","name":"same attacker"}
					],
					"total": 2
				}`),
			})
		case 3:
			if req.Method != "AlarmService.SearchAlarmList" {
				t.Fatalf("third method = %q", req.Method)
			}
			raw, _ := json.Marshal(req.Params)
			var params map[string]any
			if err := json.Unmarshal(raw, &params); err != nil {
				t.Fatalf("decode params: %v", err)
			}
			if _, ok := params["victim"]; !ok {
				t.Fatalf("missing victim filter: %#v", params)
			}
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"data":[
						{"doc_id":"doc-2","timestamp":1100000,"attacker":"198.51.100.10","victim":"203.0.113.21","name":"duplicate"},
						{"doc_id":"doc-3","timestamp":1200000,"attacker":"203.0.113.30","victim":"203.0.113.20","name":"same victim"}
					],
					"total": 2
				}`),
			})
		default:
			t.Fatalf("unexpected extra method: %s", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"alarm", "related",
		"--id", "doc-1",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查看相关告警" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["related_total"] != float64(2) || data["current_count"] != float64(2) {
		t.Fatalf("related counts = %#v", data)
	}
	alarms := data["alarms"].([]any)
	first := alarms[0].(map[string]any)
	if first["doc_id"] == "doc-1" {
		t.Fatalf("original alarm should be excluded: %#v", alarms)
	}
	if first["matched_by"] != "same_attacker,same_victim" {
		t.Fatalf("matched_by = %#v", first["matched_by"])
	}
}
