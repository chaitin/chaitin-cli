package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chaitin/chaitin-cli/products/insight/client"
	"github.com/chaitin/chaitin-cli/products/insight/models"
	"github.com/spf13/cobra"
)

func TestTaskExecutionCommandsUseExecutionIDQuery(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "execution detail command",
			args: []string{"execution", "--id", "456"},
		},
		{
			name: "legacy status alias",
			args: []string{"status", "--exec-id", "456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotURI string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotURI = r.URL.RequestURI()
				fmt.Fprint(w, `{"code":0,"data":{"id":456,"task_id":123,"status":2}}`)
			}))
			t.Cleanup(server.Close)

			var out, errOut bytes.Buffer
			taskCmd := NewTaskCmd(func(cmd *cobra.Command) *client.Client {
				return client.NewClient(models.Config{
					URL:    server.URL,
					APIKey: "test-token",
				}, true, false, false, &out, &errOut)
			})
			taskCmd.SetOut(&out)
			taskCmd.SetErr(&errOut)
			taskCmd.SetArgs(tt.args)

			if err := taskCmd.Execute(); err != nil {
				t.Fatalf("execute task command: %v\nstderr: %s", err, errOut.String())
			}

			if gotMethod != http.MethodGet {
				t.Fatalf("method = %q, want %q", gotMethod, http.MethodGet)
			}
			if gotURI != "/exposure/api/task/execution?id=456" {
				t.Fatalf("request URI = %q, want /exposure/api/task/execution?id=456", gotURI)
			}
		})
	}
}
