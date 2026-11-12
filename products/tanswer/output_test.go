package tanswer

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSuccessEnvelopeJSON(t *testing.T) {
	env := SuccessEnvelope{
		Success: true,
		Task:    "查看威胁告警概览",
		Command: "chaitin-cli tanswer alarm overview",
		Query:   map[string]any{"time": "today"},
		Data:    map[string]any{"alarm_total": 1},
	}

	raw, err := RenderJSON(env)
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got["success"] != true {
		t.Fatalf("success = %#v", got["success"])
	}
	if got["task"] != "查看威胁告警概览" {
		t.Fatalf("task = %#v", got["task"])
	}
}

func TestErrorEnvelopeJSON(t *testing.T) {
	env := ErrorEnvelope{
		Success: false,
		Task:    "查看威胁告警概览",
		Command: "chaitin-cli tanswer alarm overview",
		Error: CLIError{
			Code:      "INVALID_TIME_RANGE",
			Message:   "开始时间不能晚于结束时间。",
			Retryable: false,
		},
	}

	raw, err := RenderJSON(env)
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}

	var got struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got.Success {
		t.Fatal("success should be false")
	}
	if got.Error.Code != "INVALID_TIME_RANGE" {
		t.Fatalf("error.code = %q", got.Error.Code)
	}
}

func TestRenderJSONDoesNotHTMLEscapeCommandExamples(t *testing.T) {
	raw, err := RenderJSON(map[string]any{
		"command": "chaitin-cli tanswer api <METHOD> <PATH>",
	})
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}
	if strings.Contains(string(raw), "\\u003c") || strings.Contains(string(raw), "\\u003e") {
		t.Fatalf("RenderJSON HTML-escaped command example:\n%s", string(raw))
	}
	if !strings.Contains(string(raw), "<METHOD>") {
		t.Fatalf("RenderJSON missing readable command example:\n%s", string(raw))
	}
}
