package agentcompose

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"

	"connectrpc.com/connect"
)

const (
	exitGeneral     = 1
	exitUsage       = 2
	exitAuth        = 3
	exitPermission  = 4
	exitNotFound    = 5
	exitNetwork     = 6
	exitUnsupported = 7
	exitCanceled    = 130
)

type CLIError struct {
	Kind           string   `json:"kind"`
	Message        string   `json:"message"`
	RPCCode        string   `json:"rpc_code,omitempty"`
	HTTPStatus     int      `json:"http_status,omitempty"`
	RemoteExitCode *int     `json:"remote_exit_code,omitempty"`
	Operation      string   `json:"operation,omitempty"`
	FailedTarget   string   `json:"failed_target,omitempty"`
	Completed      []string `json:"completed_targets,omitempty"`
	Unattempted    []string `json:"unattempted_targets,omitempty"`

	code  int
	json  bool
	cause error
}

type exitStatus struct{ code int }

func (e exitStatus) Error() string               { return fmt.Sprintf("command exited with status %d", e.code) }
func (e exitStatus) ExitCode() int               { return e.code }
func (e exitStatus) RenderError(io.Writer) error { return nil }

func (e *CLIError) Error() string { return e.Message }
func (e *CLIError) Unwrap() error { return e.cause }
func (e *CLIError) ExitCode() int { return e.code }

func (e *CLIError) RenderError(w io.Writer) error {
	if e.json {
		return json.NewEncoder(w).Encode(e)
	}
	_, err := fmt.Fprintf(w, "Error: %s (kind: %s)\n", e.Message, e.Kind)
	return err
}

func (e CLIError) MarshalJSON() ([]byte, error) {
	output := map[string]any{"kind": e.Kind, "message": e.Message}
	if e.RPCCode != "" {
		output["rpc_code"] = e.RPCCode
	}
	if e.HTTPStatus != 0 {
		output["http_status"] = e.HTTPStatus
	}
	if e.RemoteExitCode != nil {
		output["remote_exit_code"] = *e.RemoteExitCode
	}
	if e.Operation != "" {
		output["operation"] = e.Operation
	}
	if e.FailedTarget != "" {
		output["failed_target"] = e.FailedTarget
	}
	if e.Kind == "partial_failure" {
		completed := e.Completed
		if completed == nil {
			completed = []string{}
		}
		unattempted := e.Unattempted
		if unattempted == nil {
			unattempted = []string{}
		}
		output["completed_targets"] = completed
		output["unattempted_targets"] = unattempted
	}
	return json.Marshal(output)
}

func newError(kind, message string, code int, jsonOutput bool) *CLIError {
	return &CLIError{Kind: kind, Message: message, code: code, json: jsonOutput}
}

func usageError(message string, jsonOutput bool) error {
	return newError("usage", message, exitUsage, jsonOutput)
}

func unsupportedDryRun(operation string, jsonOutput bool) error {
	return newError("dry_run_unsupported", operation+" does not support --dry-run; no changes were made", exitUnsupported, jsonOutput)
}

func canceledError(jsonOutput bool) error {
	return newError("canceled", "operation canceled by user", exitCanceled, jsonOutput)
}

func mapConnectError(err error, targetURL string, jsonOutput bool) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return canceledError(jsonOutput)
	}
	var statusErr *httpStatusError
	if errors.As(err, &statusErr) {
		if statusErr.status == 502 {
			return &CLIError{Kind: "upstream_gateway_failure", Message: "upstream gateway failure: " + targetURL, HTTPStatus: 502, code: exitNetwork, json: jsonOutput, cause: err}
		}
		return &CLIError{Kind: "service_unavailable", Message: "Token API is disabled, unavailable, or temporarily unavailable: " + targetURL, HTTPStatus: statusErr.status, code: exitNetwork, json: jsonOutput, cause: err}
	}

	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		code := connectErr.Code()
		result := &CLIError{RPCCode: code.String(), json: jsonOutput, cause: err}
		switch code {
		case connect.CodeUnauthenticated:
			result.Kind, result.Message, result.code = "authentication_failed", "Token is missing or invalid; run agent-compose auth login", exitAuth
		case connect.CodePermissionDenied:
			result.Kind, result.Message, result.code = "permission_denied", "permission denied; this operation requires an admin Token", exitPermission
		case connect.CodeNotFound:
			result.Kind, result.Message, result.code = "not_found", safeRemoteMessage(connectErr, "resource not found"), exitNotFound
		case connect.CodeInvalidArgument:
			result.Kind, result.Message, result.code = "invalid_argument", safeRemoteMessage(connectErr, "invalid request"), exitUsage
		case connect.CodeFailedPrecondition:
			result.Kind, result.Message, result.code = "failed_precondition", safeRemoteMessage(connectErr, "operation precondition was not met"), exitGeneral
		case connect.CodeUnimplemented:
			if strings.Contains(strings.ToLower(connectErr.Message()), "not implemented") {
				result.Kind, result.Message, result.code = "api_unimplemented", "the requested operation is not implemented by the target API", exitUnsupported
			} else {
				result.Kind, result.Message, result.code = "version_incompatible", "target daemon, proxy, and CLI API versions are incompatible", exitUnsupported
			}
		case connect.CodeUnavailable, connect.CodeDeadlineExceeded:
			result.Kind, result.Message, result.code = "service_unavailable", "service unavailable or request timed out: "+targetURL, exitNetwork
		case connect.CodeCanceled:
			return canceledError(jsonOutput)
		default:
			result.Kind, result.Message, result.code = "rpc_error", safeRemoteMessage(connectErr, "remote request failed"), exitGeneral
		}
		return result
	}

	var urlErr *url.Error
	var netErr net.Error
	if errors.As(err, &urlErr) || errors.As(err, &netErr) {
		return &CLIError{Kind: "network_error", Message: "network or TLS error connecting to " + targetURL, code: exitNetwork, json: jsonOutput, cause: err}
	}
	return &CLIError{Kind: "execution_failed", Message: err.Error(), code: exitGeneral, json: jsonOutput, cause: err}
}

func safeRemoteMessage(err *connect.Error, fallback string) string {
	message := strings.TrimSpace(err.Message())
	if message == "" {
		return fallback
	}
	return message
}
