package tanswer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAlarmByAttackerHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "alarm", "by-attacker", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查询指定攻击源相关告警", "--attacker", "--time", "受害对象 Top"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAlarmByThreatHelpExplainsAtLeastOneSelector(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "alarm", "by-threat", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(out.String(), "至少提供 --name、--tag 或 --phase 中的一项") {
		t.Fatalf("help does not explain selector requirement:\n%s", out.String())
	}
}

func TestBuildAlarmSubjectRequestMapsAttacker(t *testing.T) {
	rng := TimeRange{Start: 1000, End: 2000}
	req := buildAlarmSubjectListRequest(rng, alarmSubjectOptions{
		time:     "today",
		pageSize: 20,
		subject:  "198.51.100.10",
	}, alarmSubjectAttacker)

	if req["time_range_start"] != int64(1000) || req["time_range_end"] != int64(2000) {
		t.Fatalf("time range mismatch: %#v", req)
	}
	if req["count"] != int64(20) {
		t.Fatalf("count = %#v", req["count"])
	}
	got := req["attacker"].([]map[string]any)
	if got[0]["target"] != "198.51.100.10" {
		t.Fatalf("attacker filter = %#v", got)
	}
}

func TestAlarmByAttackerCommandCallsSearchAlarmList(t *testing.T) {
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
		if params["count"] != float64(10) {
			t.Fatalf("count = %#v", params["count"])
		}
		attacker := params["attacker"].([]any)[0].(map[string]any)
		if attacker["target"] != "198.51.100.10" {
			t.Fatalf("attacker filter = %#v", params["attacker"])
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"data":[
					{"doc_id":"a1","name":"SQL注入","severity":2,"attacker":"198.51.100.10","victim":"203.0.113.20","result":"success","phase":"intrustion","tag":"代码执行"},
					{"doc_id":"a2","name":"弱口令","severity":3,"attacker":"198.51.100.10","victim":"203.0.113.21","result":"control","phase":"lateral","tag":"弱口令"}
				],
				"total": 2,
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
		"alarm", "by-attacker",
		"--attacker", "198.51.100.10",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查询指定攻击源相关告警" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["related_total"] != float64(2) || data["success_control_count"] != float64(2) {
		t.Fatalf("summary mismatch: %#v", data)
	}
	if len(data["victim_top"].([]any)) != 2 {
		t.Fatalf("victim_top = %#v", data["victim_top"])
	}
}

func TestAlarmRankCommandCallsSearchAlarmAggTop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AlarmService.SearchAlarmAggTop" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["agg"] != "attacker" || params["top"] != float64(5) {
			t.Fatalf("rank params = %#v", params)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":[{"key":"198.51.100.10","doc_count":7}]}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"alarm", "attacker-rank",
		"--top", "5",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查看攻击源排行" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["rank_type"] != "attacker" {
		t.Fatalf("rank_type = %#v", data["rank_type"])
	}
}
