package tanswer

import "fmt"

type ProtectedWriteSpec struct {
	OperationType     string
	ConfirmationToken string
	Target            any
	ChangeSummary     []string
	Impact            []string
	RiskWarnings      []string
}

type ProtectedWritePreview struct {
	RequiresConfirmation bool     `json:"requires_confirmation"`
	Confirmed            bool     `json:"confirmed"`
	OperationType        string   `json:"operation_type"`
	Target               any      `json:"target"`
	ChangeSummary        []string `json:"change_summary"`
	Impact               []string `json:"impact"`
	RiskWarnings         []string `json:"risk_warnings,omitempty"`
	ConfirmationToken    string   `json:"confirmation_token"`
}

func BuildProtectedWritePreview(spec ProtectedWriteSpec, confirm string) ProtectedWritePreview {
	return ProtectedWritePreview{
		RequiresConfirmation: true,
		Confirmed:            confirm == spec.ConfirmationToken,
		OperationType:        spec.OperationType,
		Target:               spec.Target,
		ChangeSummary:        spec.ChangeSummary,
		Impact:               spec.Impact,
		RiskWarnings:         spec.RiskWarnings,
		ConfirmationToken:    spec.ConfirmationToken,
	}
}

func ValidateConfirmation(spec ProtectedWriteSpec, confirm string) error {
	if confirm != spec.ConfirmationToken {
		return fmt.Errorf("confirmation token mismatch: expected %s", spec.ConfirmationToken)
	}
	return nil
}
