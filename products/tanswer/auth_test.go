package tanswer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthStatusHelpMentionsPurpose(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "auth", "status", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查看当前连接状态", "TANSWER_URL", "OpenAPI Token"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAuthCheckHelpMentionsTokenValidation(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "auth", "check", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(out.String(), "校验 OpenAPI Token") {
		t.Fatalf("help missing token validation purpose:\n%s", out.String())
	}
}

func TestAuthCheckReadsBaseInfoThenValidatesTokenThroughSearchTags(t *testing.T) {
	called := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Api-Token") != "token-123" {
			t.Fatalf("Api-Token header = %q", r.Header.Get("Api-Token"))
		}

		var request rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode rpc request: %v", err)
		}
		called = append(called, request.Method)

		result := json.RawMessage(`{}`)
		if request.Method == "OpsService.GetBaseInfo" {
			result = json.RawMessage(`{"ProductName":"全悉"}`)
		}
		if request.Method != "OpsService.GetBaseInfo" && request.Method != "AssetService.SearchTags" {
			t.Fatalf("unexpected RPC method %q", request.Method)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{JSONRPC: "2.0", ID: request.ID, Result: result})
	}))
	defer server.Close()

	var output bytes.Buffer
	command := NewRootCommand(RootOptions{Out: &output, ErrOut: &output})
	command.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "auth", "check"})
	if err := command.Execute(); err != nil {
		t.Fatalf("auth check returned error: %v", err)
	}
	if len(called) != 2 || called[0] != "OpsService.GetBaseInfo" || called[1] != "AssetService.SearchTags" {
		t.Fatalf("RPC call order = %#v", called)
	}

	var result SuccessEnvelope
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode auth check output: %v; output=%s", err, output.String())
	}
	if !result.Success {
		t.Fatalf("auth check success = false: %s", output.String())
	}
}
