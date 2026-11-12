package tanswer

import (
	"fmt"
	"strings"
	"time"
)

type WriteOperationSpec struct {
	Task          string
	Command       string
	OperationType string
	RiskLevel     string
	Target        any
	ChangeSummary any
	Impact        any
	RiskWarnings  []string
	ConfirmToken  string
}

type WriteExecutionSpec struct {
	OperationType string
	Object        any
	Action        string
	Environment   string
	Actor         string
	BeforeAfter   any
	Result        string
	FailureReason string
	ExecutedAt    time.Time
}

func BuildWritePreview(spec WriteOperationSpec) map[string]any {
	return map[string]any{
		"operation_type":        spec.OperationType,
		"risk_level":            spec.RiskLevel,
		"target":                spec.Target,
		"change_summary":        spec.ChangeSummary,
		"impact":                spec.Impact,
		"risk_warnings":         spec.RiskWarnings,
		"requires_confirmation": true,
		"confirmed":             false,
		"confirmation_token":    spec.ConfirmToken,
		"confirmation_note":     fmt.Sprintf("Re-run with --confirm %s to execute after reviewing this preview.", spec.ConfirmToken),
	}
}

func ValidateWriteConfirmation(got string, expected string) error {
	got = strings.TrimSpace(got)
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return fmt.Errorf("missing expected confirmation token")
	}
	if got != expected {
		return fmt.Errorf("write operation requires explicit confirmation token %q", expected)
	}
	return nil
}

func BuildWriteExecutionResult(spec WriteExecutionSpec) map[string]any {
	executedAt := spec.ExecutedAt
	if executedAt.IsZero() {
		executedAt = time.Now().UTC()
	}
	return map[string]any{
		"operation_type": spec.OperationType,
		"confirmed":      true,
		"result":         spec.Result,
		"object":         spec.Object,
		"audit": map[string]any{
			"actor":          spec.Actor,
			"executed_at":    executedAt.UTC().Format(time.RFC3339),
			"environment":    spec.Environment,
			"object":         spec.Object,
			"action":         spec.Action,
			"before_after":   spec.BeforeAfter,
			"result":         spec.Result,
			"failure_reason": spec.FailureReason,
		},
	}
}
