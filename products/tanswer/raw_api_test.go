package tanswer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRawAPIHelpExplainsFallback(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "api", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"Open API 兜底调用", "METHOD", "PATH", "--query", "--body", "--preview", "--confirm", "--dry-run 不适用于", "<known-unwrapped-openapi-path>"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "OpsService.GetBaseInfo") {
		t.Fatalf("raw API help must not bypass the system status semantic command:\n%s", text)
	}
}

func TestRawAPIPotentiallyMutatingRequestReturnsPreviewWithoutSending(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		t.Fatalf("unconfirmed raw API request must not be sent")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer", "--url", server.URL, "--api-key", "token-123",
		"api", "POST", "/rpc",
		"--body", `{"method":"AssetService.Create"}`,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if called {
		t.Fatal("unconfirmed raw API request was sent")
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	data := env.Data.(map[string]any)
	if data["requires_confirmation"] != true || data["confirmed"] != false || data["confirmation_token"] != "CONFIRM_TANSWER_RAW_API_WRITE" {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestRawAPIPotentiallyMutatingRequestRequiresExactConfirm(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	for _, confirm := range []string{"wrong", "CONFIRM_TANSWER_RAW_API_WRITE"} {
		var out bytes.Buffer
		cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
		cmd.SetArgs([]string{
			"tanswer", "--url", server.URL, "--api-key", "token-123",
			"api", "POST", "/rpc", "--confirm", confirm,
		})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("confirm %q: Execute returned error: %v", confirm, err)
		}
	}
	if calls != 1 {
		t.Fatalf("server calls = %d, want 1", calls)
	}
}

func TestRawAPIPotentiallyMutatingRequestRejectsWrongConfirmStructurally(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		t.Fatal("wrong raw API confirmation token must not send a request")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer", "--url", server.URL, "--api-key", "token-123",
		"api", "POST", "/rpc", "--confirm", "wrong",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if called {
		t.Fatal("wrong raw API confirmation token sent a request")
	}

	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode structured confirmation error: %v\n%s", err, out.String())
	}
	if env.Success || env.Error.Code != "RAW_API_CONFIRMATION_REQUIRED" || env.Error.Retryable {
		t.Fatalf("confirmation error = %#v", env)
	}
}

func TestRawAPIGetWithQueryOutputsStatusAndRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/api/example" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("count") != "10" || r.URL.Query().Get("offset") != "0" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer", "--url", server.URL, "--api-key", "token-123",
		"api", "GET", "/api/example",
		"--query", `{"count":10,"offset":0}`,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Success != true {
		t.Fatalf("success = %v", env.Success)
	}
	data := env.Data.(map[string]any)
	if data["status_code"] != float64(http.StatusAccepted) {
		t.Fatalf("status_code = %v", data["status_code"])
	}
	raw := data["raw"].(map[string]any)
	if raw["ok"] != true {
		t.Fatalf("raw = %#v", raw)
	}
}

func TestRawAPIPostBodyFileOutputsNon2xxStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["method"] != "OpsService.GetBaseInfo" {
			t.Fatalf("body = %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	bodyFile := t.TempDir() + "/request.json"
	if err := os.WriteFile(bodyFile, []byte(`{"method":"OpsService.GetBaseInfo"}`), 0o600); err != nil {
		t.Fatalf("write body file: %v", err)
	}

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer", "--url", server.URL, "--api-key", "token-123",
		"api", "POST", "/rpc",
		"--body", "@" + bodyFile,
		"--confirm", rawAPIWriteConfirmToken,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Success != false {
		t.Fatalf("success = %v", env.Success)
	}
	data := env.Data.(map[string]any)
	if data["status_code"] != float64(http.StatusBadRequest) {
		t.Fatalf("status_code = %v", data["status_code"])
	}
	raw := data["raw"].(map[string]any)
	if raw["error"] != "bad request" {
		t.Fatalf("raw = %#v", raw)
	}
}

func TestRawAPINonJSONResponseOutputsRawString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer", "--url", server.URL, "--api-key", "token-123",
		"api", "GET", "/api/text-error",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Success != false {
		t.Fatalf("success = %v", env.Success)
	}
	data := env.Data.(map[string]any)
	if data["status_code"] != float64(http.StatusBadGateway) {
		t.Fatalf("status_code = %v", data["status_code"])
	}
	if data["raw"] != "bad gateway" {
		t.Fatalf("raw = %#v", data["raw"])
	}
}

func TestRawAPIRejectsExternalURLPathBeforeRequest(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		t.Fatalf("external URL should not be requested")
	}))
	defer server.Close()

	for _, path := range []string{server.URL + "/steal", "//example.com/steal"} {
		var out bytes.Buffer
		cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
		cmd.SetArgs([]string{
			"tanswer", "--url", server.URL, "--api-key", "token-123",
			"api", "GET", path,
		})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("path %q: Execute returned error: %v", path, err)
		}
		var env ErrorEnvelope
		if err := json.Unmarshal(out.Bytes(), &env); err != nil {
			t.Fatalf("path %q: decode structured error: %v\\n%s", path, err, out.String())
		}
		if env.Success || env.Error.Code != "INVALID_API_PATH" || env.Error.Retryable {
			t.Fatalf("path %q: error = %#v", path, env)
		}
	}
	if called {
		t.Fatalf("external request was sent")
	}
}

func TestRawAPIRejectsRelativePath(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer", "--url", "https://tanswer.test", "--api-key", "token-123",
		"api", "GET", "api/example",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode structured error: %v\\n%s", err, out.String())
	}
	if env.Success || env.Error.Code != "INVALID_API_PATH" || env.Error.Retryable {
		t.Fatalf("error = %#v", env)
	}
}

func TestRawAPIInvalidInputReturnsStructuredErrorWithoutSending(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		t.Fatal("invalid raw API input must not send a request")
	}))
	defer server.Close()

	tests := []struct {
		name string
		args []string
		code string
	}{
		{
			name: "full URL path",
			args: []string{"GET", server.URL + "/steal"},
			code: "INVALID_API_PATH",
		},
		{
			name: "relative path",
			args: []string{"GET", "api/example"},
			code: "INVALID_API_PATH",
		},
		{
			name: "invalid query JSON",
			args: []string{"GET", "/api/example", "--query", "{"},
			code: "INVALID_API_QUERY",
		},
		{
			name: "invalid body JSON",
			args: []string{"POST", "/api/example", "--body", "{"},
			code: "INVALID_API_BODY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
			cmd.SetArgs(append([]string{
				"tanswer", "--url", server.URL, "--api-key", "token-123", "api",
			}, tt.args...))

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			var env ErrorEnvelope
			if err := json.Unmarshal(out.Bytes(), &env); err != nil {
				t.Fatalf("decode structured input error: %v\\n%s", err, out.String())
			}
			if env.Success || env.Error.Code != tt.code || env.Error.Retryable {
				t.Fatalf("input error = %#v", env)
			}
		})
	}
	if called {
		t.Fatal("invalid raw API input sent a request")
	}
}
