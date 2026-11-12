package tanswer

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestBuildWritePreviewRequiresExplicitConfirmation(t *testing.T) {
	preview := BuildWritePreview(WriteOperationSpec{
		Task:          "新增资产",
		Command:       "chaitin-cli tanswer asset create",
		OperationType: "asset_create",
		RiskLevel:     "write_high",
		Target:        map[string]any{"ip": "192.0.2.10"},
		ChangeSummary: map[string]any{"name": "core-db"},
		Impact:        map[string]any{"asset_count": 1},
		RiskWarnings:  []string{"将新增资产配置。"},
		ConfirmToken:  "CONFIRM_ASSET_CREATE",
	})

	if preview["requires_confirmation"] != true {
		t.Fatalf("requires_confirmation = %#v", preview["requires_confirmation"])
	}
	if preview["confirmation_token"] != "CONFIRM_ASSET_CREATE" {
		t.Fatalf("confirmation_token = %#v", preview["confirmation_token"])
	}
	if preview["confirmed"] != false {
		t.Fatalf("confirmed = %#v", preview["confirmed"])
	}
	warnings := preview["risk_warnings"].([]string)
	if len(warnings) != 1 || warnings[0] != "将新增资产配置。" {
		t.Fatalf("risk_warnings = %#v", warnings)
	}
}

func TestValidateWriteConfirmationRejectsSilentExecution(t *testing.T) {
	err := ValidateWriteConfirmation("", "CONFIRM_ASSET_DELETE")
	if err == nil {
		t.Fatal("expected confirmation error")
	}
	if !strings.Contains(err.Error(), "CONFIRM_ASSET_DELETE") {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateWriteConfirmationAcceptsExactTokenOnly(t *testing.T) {
	if err := ValidateWriteConfirmation("CONFIRM_ASSET_DELETE", "CONFIRM_ASSET_DELETE"); err != nil {
		t.Fatalf("ValidateWriteConfirmation returned error: %v", err)
	}
	if err := ValidateWriteConfirmation("confirm_asset_delete", "CONFIRM_ASSET_DELETE"); err == nil {
		t.Fatal("expected case-sensitive confirmation error")
	}
}

func TestBuildWriteExecutionResultIncludesAuditFields(t *testing.T) {
	now := time.Date(2026, 7, 17, 18, 30, 0, 0, time.UTC)
	result := BuildWriteExecutionResult(WriteExecutionSpec{
		OperationType: "asset_create",
		Object:        map[string]any{"id": 7, "ip": "192.0.2.10"},
		Action:        "create",
		Environment:   "https://tanswer.test",
		Actor:         "open_api_token",
		BeforeAfter:   map[string]any{"before": nil, "after": map[string]any{"ip": "192.0.2.10"}},
		Result:        "success",
		ExecutedAt:    now,
	})

	raw, err := RenderJSON(result)
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode result: %v\n%s", err, string(raw))
	}
	if got["confirmed"] != true || got["operation_type"] != "asset_create" || got["result"] != "success" {
		t.Fatalf("result = %#v", got)
	}
	audit := got["audit"].(map[string]any)
	if audit["actor"] != "open_api_token" || audit["environment"] != "https://tanswer.test" {
		t.Fatalf("audit = %#v", audit)
	}
	if audit["executed_at"] != "2026-07-17T18:30:00Z" {
		t.Fatalf("executed_at = %#v", audit["executed_at"])
	}
}
