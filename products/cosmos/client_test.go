package cosmos

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDoRequestDryRunDoesNotSendAndMasksToken(t *testing.T) {
	oldServerURL := serverURL
	oldAPIToken := apiToken
	oldDryRun := dryRun
	t.Cleanup(func() {
		serverURL = oldServerURL
		apiToken = oldAPIToken
		dryRun = oldDryRun
	})

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
		t.Fatal("dry-run should not send HTTP request")
	}))
	defer server.Close()

	serverURL = server.URL
	apiToken = "super-secret-token-value"
	dryRun = true

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := doRequest(cmd, "AssetService.SaveHostAsset", map[string]interface{}{
		"name": "asset-1",
	}, false)
	if err != nil {
		t.Fatalf("doRequest() error = %v", err)
	}
	if called {
		t.Fatal("dry-run sent HTTP request")
	}

	got := out.String()
	for _, want := range []string{
		`"dry_run": true`,
		`"http_method": "POST"`,
		`"url": "` + server.URL + `/pedestal/rpc"`,
		`"method": "AssetService.SaveHostAsset"`,
		`"name": "asset-1"`,
		`"Authorization": "Bearer supe...alue"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, apiToken) {
		t.Fatalf("dry-run output leaked token:\n%s", got)
	}
}

func TestDoRequestReturnsJSONRPCError(t *testing.T) {
	oldServerURL := serverURL
	oldAPIToken := apiToken
	oldDryRun := dryRun
	t.Cleanup(func() {
		serverURL = oldServerURL
		apiToken = oldAPIToken
		dryRun = oldDryRun
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1","error":{"code":-32602,"message":"invalid params","data":{"field":"created_at"}}}`))
	}))
	defer server.Close()

	serverURL = server.URL
	apiToken = ""
	dryRun = false

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := doRequest(cmd, "AlarmService.GetAlarmList", map[string]interface{}{
		"count": 1,
	}, false)
	if err == nil {
		t.Fatal("doRequest() error = nil, want JSON-RPC error")
	}
	for _, want := range []string{"json-rpc error -32602", "invalid params", `"field":"created_at"`} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err.Error(), want)
		}
	}
	if out.Len() != 0 {
		t.Fatalf("doRequest() wrote stdout for JSON-RPC error:\n%s", out.String())
	}
}

func TestMaskSecret(t *testing.T) {
	if got := maskSecret("12345678"); got != "****" {
		t.Fatalf("maskSecret short = %q, want ****", got)
	}
	if got := maskSecret("123456789"); got != "1234...6789" {
		t.Fatalf("maskSecret long = %q, want 1234...6789", got)
	}
}
