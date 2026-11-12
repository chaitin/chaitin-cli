package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/chaitin/chaitin-cli/products/insight/client"
	"github.com/chaitin/chaitin-cli/products/insight/models"
	"github.com/spf13/cobra"
)

func TestSnapshotAssetMapsTaskIDToInsightIDQuery(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		fmt.Fprint(w, `{"code":0,"data":{"list":[],"total":0,"affect_version_count":0}}`)
	}))
	t.Cleanup(server.Close)

	var out, errOut bytes.Buffer
	snapshotCmd := NewSnapshotCmd(func(cmd *cobra.Command) *client.Client {
		return client.NewClient(models.Config{
			URL:    server.URL,
			APIKey: "test-token",
		}, true, false, false, &out, &errOut)
	})
	snapshotCmd.SetOut(&out)
	snapshotCmd.SetErr(&errOut)
	snapshotCmd.SetArgs([]string{"asset", "--task-id", "123", "--execution-id", "456"})

	if err := snapshotCmd.Execute(); err != nil {
		t.Fatalf("execute snapshot asset: %v\nstderr: %s", err, errOut.String())
	}

	if gotPath != "/exposure/api/snapshot/asset" {
		t.Fatalf("path = %q, want /exposure/api/snapshot/asset", gotPath)
	}
	if gotQuery.Get("id") != "123" {
		t.Fatalf("id query = %q, want task id 123", gotQuery.Get("id"))
	}
	if gotQuery.Get("task_id") != "" {
		t.Fatalf("task_id query = %q, want empty", gotQuery.Get("task_id"))
	}
	if gotQuery.Get("execution_id") != "456" {
		t.Fatalf("execution_id query = %q, want 456", gotQuery.Get("execution_id"))
	}
}
