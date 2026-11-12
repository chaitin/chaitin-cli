package tanswer

import "testing"

func TestProtectedWriteRequiresExactConfirmation(t *testing.T) {
	spec := ProtectedWriteSpec{
		OperationType:     "asset_create",
		ConfirmationToken: "CONFIRM_ASSET_CREATE",
		Target:            map[string]any{"name": "cli-test"},
		ChangeSummary:     []string{"create asset"},
		Impact:            []string{"asset inventory changes"},
	}
	preview := BuildProtectedWritePreview(spec, "")
	if preview.RequiresConfirmation != true || preview.Confirmed != false {
		t.Fatalf("preview confirmation flags are wrong: %#v", preview)
	}
	if err := ValidateConfirmation(spec, "WRONG_CONFIRM"); err == nil {
		t.Fatalf("wrong confirmation should fail")
	}
	if err := ValidateConfirmation(spec, "CONFIRM_ASSET_CREATE"); err != nil {
		t.Fatalf("exact confirmation should pass: %v", err)
	}
}
