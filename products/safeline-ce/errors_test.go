package safelinece

import (
	"errors"
	"testing"
)

func TestCLIError(t *testing.T) {
	t.Run("error message without underlying error", func(t *testing.T) {
		err := &CLIError{
			Code:    ExitConfigError,
			Message: "config file not found",
		}
		if err.Error() != "config file not found" {
			t.Errorf("Error() = %q, want %q", err.Error(), "config file not found")
		}
	})

	t.Run("error message with underlying error", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := &CLIError{
			Code:    ExitNetworkError,
			Message: "connection failed",
			Err:     underlying,
		}
		want := "connection failed: underlying error"
		if err.Error() != want {
			t.Errorf("Error() = %q, want %q", err.Error(), want)
		}
	})
}

func TestNewConfigError(t *testing.T) {
	underlying := errors.New("file not found")
	err := NewConfigError("failed to load config", underlying)

	if err.Code != ExitConfigError {
		t.Errorf("Code = %d, want %d", err.Code, ExitConfigError)
	}
	if err.Message != "failed to load config" {
		t.Errorf("Message = %q, want %q", err.Message, "failed to load config")
	}
	if err.Err != underlying {
		t.Error("Err not set correctly")
	}
}

func TestNewNetworkError(t *testing.T) {
	underlying := errors.New("timeout")
	err := NewNetworkError("request timeout", underlying)

	if err.Code != ExitNetworkError {
		t.Errorf("Code = %d, want %d", err.Code, ExitNetworkError)
	}
	if err.Message != "request timeout" {
		t.Errorf("Message = %q, want %q", err.Message, "request timeout")
	}
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(404, "resource not found")

	if err.Code != ExitAPIError {
		t.Errorf("Code = %d, want %d", err.Code, ExitAPIError)
	}
	if err.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, 404)
	}
	if err.Message != "resource not found" {
		t.Errorf("Message = %q, want %q", err.Message, "resource not found")
	}
}

func TestErrorTypeChecks(t *testing.T) {
	configErr := NewConfigError("config error", nil)
	networkErr := NewNetworkError("network error", nil)
	apiErr := NewAPIError(500, "api error")
	otherErr := errors.New("other error")

	tests := []struct {
		name     string
		err      error
		isConfig bool
		isNet    bool
		isAPI    bool
	}{
		{"config error", configErr, true, false, false},
		{"network error", networkErr, false, true, false},
		{"api error", apiErr, false, false, true},
		{"other error", otherErr, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsConfigError(tt.err) != tt.isConfig {
				t.Errorf("IsConfigError() = %v, want %v", IsConfigError(tt.err), tt.isConfig)
			}
			if IsNetworkError(tt.err) != tt.isNet {
				t.Errorf("IsNetworkError() = %v, want %v", IsNetworkError(tt.err), tt.isNet)
			}
			if IsAPIError(tt.err) != tt.isAPI {
				t.Errorf("IsAPIError() = %v, want %v", IsAPIError(tt.err), tt.isAPI)
			}
		})
	}
}

func TestCLIError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &CLIError{
		Code: ExitNetworkError,
		Err:  underlying,
	}

	unwrapped := err.Unwrap()
	if unwrapped != underlying {
		t.Error("Unwrap() did not return the underlying error")
	}
}
