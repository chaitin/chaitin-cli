package cloudatlas

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func TestDisabledOperationsAreHidden(t *testing.T) {
	cmd := NewCommand()

	for _, path := range [][]string{
		{"intelligence", "monitor-rule", "async-download"},
		{"risk", "admin-panel", "async-download"},
		{"risk", "high-risk-app", "async-download"},
		{"risk", "high-risk-port", "async-download"},
		{"intelligence", "monitor-rule", "download-status"},
		{"risk", "admin-panel", "download-status"},
		{"risk", "high-risk-app", "download-status"},
		{"risk", "high-risk-port", "download-status"},
	} {
		if found := findCommand(cmd, path); found != nil {
			t.Errorf("command %q should be hidden while the endpoint returns 500", strings.Join(path, " "))
		}
	}

	// The sibling synchronous download keeps working and must stay available.
	for _, path := range [][]string{
		{"intelligence", "monitor-rule", "sync-download"},
		{"risk", "admin-panel", "sync-download"},
		{"risk", "high-risk-app", "sync-download"},
		{"risk", "high-risk-port", "sync-download"},
	} {
		if found := findCommand(cmd, path); found == nil {
			t.Errorf("command %q should be available", strings.Join(path, " "))
		}
	}
}

func findCommand(root *cobra.Command, path []string) *cobra.Command {
	current := root
	for _, name := range path {
		var next *cobra.Command
		for _, candidate := range current.Commands() {
			if candidate.Name() == name {
				next = candidate
				break
			}
		}
		if next == nil {
			return nil
		}
		current = next
	}
	return current
}
func TestNewCommandHelpShowsGeneratedCommands(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"seed", "enterprise", "--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, want := range []string{"list", "get", "update", "patch", "create", "batch-delete", "set-confidence", "set-enabled"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("help missing %q:\n%s", want, out.String())
		}
	}
}

func TestListSendsQueryAndTokenHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/v1/asset/ip" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("TOKEN"); got != "secret-token" {
			t.Fatalf("TOKEN = %q", got)
		}
		if got := r.URL.Query().Get("space"); got != "1" {
			t.Fatalf("space = %q", got)
		}
		if got := r.URL.Query().Get("status"); got != "valid" {
			t.Fatalf("status = %q", got)
		}
		writeJSON(t, w, map[string]any{"code": 200, "message": "", "data": map[string]any{"current": 1, "size": 20, "total": 1, "items": []any{map[string]any{"id": 1, "ip": "1.1.1.1"}}}})
	}))
	defer server.Close()

	ApplyRuntimeConfig(nil, rawConfig(Config{URL: server.URL, Token: "secret-token"}), false)
	cmd := NewCommand()
	cmd.SetArgs([]string{"asset", "ip", "list", "--space", "1", "--status", "valid"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "1.1.1.1") || !strings.Contains(out.String(), "total=1") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestListUsesConfiguredSpaceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("space"); got != "42" {
			t.Fatalf("space = %q", got)
		}
		writeJSON(t, w, map[string]any{"code": 200, "message": "", "data": map[string]any{"current": 1, "size": 20, "total": 0, "items": []any{}}})
	}))
	defer server.Close()

	ApplyRuntimeConfig(nil, rawConfig(Config{URL: server.URL, Token: "secret-token", SpaceID: "42"}), false)
	cmd := NewCommand()
	cmd.SetArgs([]string{"asset", "ip", "list", "--status", "valid"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestSpaceIsRequiredForEveryCommand(t *testing.T) {
	ApplyRuntimeConfig(nil, rawConfig(Config{URL: "https://example.com", Token: "token"}), false)
	cmd := NewCommand()
	cmd.SetArgs([]string{"seed", "enterprise", "list", "--size", "1"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--space is required") {
		t.Fatalf("Execute() error = %v, want --space requirement", err)
	}
}

func TestPathParameterAndRequiredQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/task/asset/schedule" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("space"); got != "1" {
			t.Fatalf("space = %q", got)
		}
		writeJSON(t, w, map[string]any{"code": 200, "message": "", "data": []any{}})
	}))
	defer server.Close()

	ApplyRuntimeConfig(nil, rawConfig(Config{URL: server.URL, Token: "token"}), false)
	cmd := NewCommand()
	cmd.SetArgs([]string{"task", "asset", "schedule", "list", "--space", "1"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestFlowListRequiresTaskID(t *testing.T) {
	ApplyRuntimeConfig(nil, rawConfig(Config{URL: "https://example.com", Token: "token"}), false)
	cmd := NewCommand()
	cmd.SetArgs([]string{"task", "asset", "flow", "list", "--space", "1"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--task-id is required") {
		t.Fatalf("Execute() error = %v, want --task-id requirement", err)
	}
}

func TestTruncateHelpDocumentsResourceTypes(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"space", "truncate", "--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if help := out.String(); !strings.Contains(help, "attack, seed, risk, drps, asset") {
		t.Fatalf("truncate help missing resource types:\n%s", help)
	}
}

func TestCreateTaskRequiresConfirmation(t *testing.T) {
	ApplyRuntimeConfig(nil, rawConfig(Config{URL: "https://example.com", Token: "token"}), false)
	cmd := NewCommand()
	cmd.SetArgs([]string{"strategy", "vulnerability", "create-task", "--space", "1"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("Execute() error = %v, want --yes requirement", err)
	}
}

func TestBodyFileAndConfirmation(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"seed", "enterprise", "batch-delete", "--body", `{"pk":[1]}`})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("Execute() error = %v, want --yes requirement", err)
	}
}

func TestDryRunRedactsToken(t *testing.T) {
	ApplyRuntimeConfig(nil, rawConfig(Config{URL: "https://example.com", Token: "secret-token"}), true)
	cmd := NewCommand()
	cmd.SetArgs([]string{"seed", "enterprise", "create", "--space", "1", "--body", `{"name":["demo"],"level":1,"equity":100,"status":"存续"}`})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	oldVerbose := verbose
	verbose = true
	t.Cleanup(func() { verbose = oldVerbose })

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	combined := out.String() + errOut.String()
	if strings.Contains(combined, "secret-token") {
		t.Fatalf("dry-run leaked token:\n%s", combined)
	}
	if !strings.Contains(combined, "/v1/seed/enterprise/batch-create") {
		t.Fatalf("dry-run missing path:\n%s", combined)
	}
}

func TestGeneratedHelpExplainsRequiredParameterValues(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"asset", "ip", "list", "--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	help := out.String()
	for _, want := range []string{
		"--status",
		"类型: string",
		"--page",
		"类型: integer",
		"默认值: 1",
		"可由 --space-id 或 cloudAtlas.space_id 提供默认值",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q:\n%s", want, help)
		}
	}
}

func TestParameterUsageSupportsApifoxEnumFormats(t *testing.T) {
	tests := []struct {
		name   string
		schema *Schema
		want   string
	}{
		{
			name: "enumDescriptions",
			schema: &Schema{
				Type: "string",
				Enum: []any{"valid", "await"},
				XApifox: ApifoxSchema{EnumDescriptions: map[string]string{
					"valid": "已确认",
					"await": "待确认",
				}},
			},
			want: "valid=已确认, await=待确认",
		},
		{
			name: "x-apifox-enum",
			schema: &Schema{
				Type: "string",
				Enum: []any{"ignored", "invalid"},
				XApifoxEnum: []ApifoxEnumItem{
					{Value: "ignored", Description: "已排除"},
					{Value: "invalid", Description: "历史记录"},
				},
			},
			want: "ignored=已排除, invalid=历史记录",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := formatParameterUsage(Parameter{Name: "status", Description: "状态", Schema: tt.schema}, true)
			if !strings.Contains(usage, tt.want) {
				t.Fatalf("usage = %q, want %q", usage, tt.want)
			}
		})
	}
}

func TestRequestBodyHelpExplainsFields(t *testing.T) {
	body := &RequestBody{Content: map[string]MediaType{
		"application/json": {Schema: &Schema{
			Type:     "object",
			Required: []string{"status"},
			Properties: map[string]Schema{
				"status": {
					Type:        "string",
					Description: "状态",
					Enum:        []any{"enabled", "disabled"},
					XApifoxEnum: []ApifoxEnumItem{{Value: "enabled", Description: "启用"}, {Value: "disabled", Description: "停用"}},
				},
				"name": {
					Type:        "string",
					Description: "名称",
					Example:     "demo",
				},
			},
		}},
	}}

	help := NewParser(&OpenAPI{}).formatRequestBodyHelp(body)
	for _, want := range []string{"status: 状态；必填；类型: string；可选值: enabled=启用, disabled=停用", "name: 名称；类型: string；示例: demo"} {
		if !strings.Contains(help, want) {
			t.Fatalf("body help missing %q:\n%s", want, help)
		}
	}
}

func rawConfig(value Config) config.Raw {
	var node yaml.Node
	if err := node.Encode(value); err != nil {
		panic(err)
	}
	return config.Raw{"cloudAtlas": node}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
}
