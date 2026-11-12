package agentcompose

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"
)

func TestConnectErrorExitMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		kind string
		code int
	}{{"unauthenticated", connect.NewError(connect.CodeUnauthenticated, errors.New("secret detail")), "authentication_failed", 3}, {"permission", connect.NewError(connect.CodePermissionDenied, errors.New("denied")), "permission_denied", 4}, {"not found", connect.NewError(connect.CodeNotFound, errors.New("missing")), "not_found", 5}, {"invalid", connect.NewError(connect.CodeInvalidArgument, errors.New("invalid")), "invalid_argument", 2}, {"unimplemented", connect.NewError(connect.CodeUnimplemented, errors.New("procedure not implemented")), "api_unimplemented", 7}, {"version", connect.NewError(connect.CodeUnimplemented, errors.New("unimplemented")), "version_incompatible", 7}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mapped := mapConnectError(test.err, "https://example.com", true)
			cli, ok := mapped.(*CLIError)
			if !ok {
				t.Fatalf("error type = %T", mapped)
			}
			if cli.Kind != test.kind || cli.ExitCode() != test.code {
				t.Fatalf("error = %+v", cli)
			}
			var out bytes.Buffer
			if err := cli.RenderError(&out); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(out.String(), "secret detail") {
				t.Fatal("sensitive remote authentication detail leaked")
			}
		})
	}
}

func TestStatusTransportMapsGatewayWithoutResponseBody(t *testing.T) {
	for _, status := range []int{http.StatusBadGateway, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			transport := &statusTransport{base: roundTripperFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: status, Body: http.NoBody}, nil
			})}
			request, _ := http.NewRequest(http.MethodPost, "https://example.com", nil)
			_, err := transport.RoundTrip(request)
			mapped := mapConnectError(err, "https://example.com", true).(*CLIError)
			if mapped.HTTPStatus != status || mapped.ExitCode() != exitNetwork {
				t.Fatalf("mapped = %+v", mapped)
			}
		})
	}
}

func TestPartialFailureAlwaysRendersTargetArrays(t *testing.T) {
	err := newError("partial_failure", "failed", 1, true)
	err.Completed = []string{}
	err.Unattempted = []string{}
	var out bytes.Buffer
	if renderErr := err.RenderError(&out); renderErr != nil {
		t.Fatal(renderErr)
	}
	for _, field := range []string{"completed_targets", "unattempted_targets"} {
		if !strings.Contains(out.String(), field) {
			t.Errorf("JSON missing %s: %s", field, out.String())
		}
	}
}
