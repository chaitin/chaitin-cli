package agentcompose

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/chaitin/chaitin-cli/config"
	"gopkg.in/yaml.v3"
)

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name, input, want string
		wantErr           bool
	}{
		{"https", "https://example.com/", "https://example.com", false},
		{"http", "http://127.0.0.1:8081", "http://127.0.0.1:8081", false},
		{"path prefix", "https://example.com/agent-compose", "https://example.com/agent-compose", false},
		{"path prefix slash", "https://example.com/agent-compose/", "https://example.com/agent-compose", false},
		{"missing scheme", "example.com", "", true},
		{"query", "https://example.com?token=secret", "", true},
		{"fragment", "https://example.com/#x", "", true},
		{"userinfo", "https://user:secret@example.com", "", true},
		{"unsupported scheme", "ftp://example.com", "", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeBaseURL(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr %t", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestJSONParentArgumentError(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"--json", "auth", "nonsense"})
	err := cmd.Execute()
	cliErr, ok := err.(*CLIError)
	if !ok || !cliErr.json || cliErr.ExitCode() != exitUsage {
		t.Fatalf("error = %#v", err)
	}
}

func TestInvalidEnumFlagsAreUsageErrors(t *testing.T) {
	for _, args := range [][]string{{"run", "ls", "--status", "garbage"}, {"run", "ls", "--source", "garbage"}, {"scheduler", "runs", "--status", "garbage"}, {"ps", "--status", "garbage"}} {
		cmd := NewCommand()
		cmd.SetArgs(args)
		state := stateFromCommand(cmd)
		state.options.URL = "https://example.com"
		state.options.Token = "token"
		err := cmd.Execute()
		coder, ok := err.(interface{ ExitCode() int })
		if !ok || coder.ExitCode() != exitUsage {
			t.Fatalf("%v error = %v", args, err)
		}
	}
}

func TestCompletedEventFieldsAreTopLevel(t *testing.T) {
	runRecord := map[string]any{"event_type": "completed", "run_id": "run-id"}
	addRunFields(runRecord, runDTO{ID: "run-id", Status: "succeeded", ExitCode: 0})
	if runRecord["status"] != "succeeded" || runRecord["id"] != "run-id" {
		t.Fatalf("run record = %#v", runRecord)
	}
	if _, nested := runRecord["run"]; nested {
		t.Fatalf("run record is nested: %#v", runRecord)
	}
	for _, key := range []string{"sandbox_id", "sandbox_short_id", "error", "started_at", "completed_at", "duration_ms"} {
		if _, present := runRecord[key]; present {
			t.Fatalf("optional run field %s is present: %#v", key, runRecord)
		}
	}
	execRecord := map[string]any{"event_type": "completed"}
	addExecFields(execRecord, execDTO{ExecID: "exec-id", ExitCode: 7, Success: false})
	if execRecord["exec_id"] != "exec-id" || execRecord["exit_code"] != int32(7) || execRecord["success"] != false {
		t.Fatalf("exec record = %#v", execRecord)
	}
	if _, nested := execRecord["result"]; nested {
		t.Fatalf("exec record is nested: %#v", execRecord)
	}
	for _, key := range []string{"run_id", "cwd", "stdout", "stderr", "output", "error"} {
		if _, present := execRecord[key]; present {
			t.Fatalf("optional Exec field %s is present: %#v", key, execRecord)
		}
	}
}

func TestSchedulerRunStableJSONFields(t *testing.T) {
	var output bytes.Buffer
	err := writeJSON(&output, schedulerRunFromProto(&agentcomposev2.SchedulerRun{RunId: "run-id", SandboxIds: []string{"sandbox-id"}}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"run_id":"run-id"`) || !strings.Contains(output.String(), `"run_short_id":"run-id"`) || !strings.Contains(output.String(), `"sandbox_ids":["sandbox-id"]`) || strings.Contains(output.String(), `"id":`) {
		t.Fatalf("output = %s", output.String())
	}
}

func TestCommandTreeMatchesRemoteBoundary(t *testing.T) {
	cmd := NewCommand()
	var help bytes.Buffer
	cmd.SetOut(&help)
	cmd.SetErr(&help)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	rootHelp := help.String()
	for _, want := range []string{"auth", "status", "project", "agent", "run", "logs", "ps", "stats", "inspect", "scheduler", "sandbox", "exec"} {
		if !strings.Contains(rootHelp, want) {
			t.Errorf("root help missing %q", want)
		}
	}
	for _, forbidden := range []string{"jupyter", "prune", "volume", "cache", "image", "workspace"} {
		if strings.Contains(strings.ToLower(rootHelp), "  "+forbidden+" ") || strings.Contains(strings.ToLower(rootHelp), "--"+forbidden) {
			t.Errorf("root help unexpectedly contains %q", forbidden)
		}
	}
	if strings.Contains(rootHelp, "--interactive") {
		t.Error("root help unexpectedly contains --interactive")
	}
	scheduler, _, err := cmd.Find([]string{"scheduler"})
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"run", "enable", "disable", "prune"} {
		for _, child := range scheduler.Commands() {
			if child.Name() == forbidden {
				t.Errorf("scheduler contains forbidden command %q", forbidden)
			}
		}
	}
	run, _, _ := cmd.Find([]string{"run"})
	for _, flag := range []string{"prompt", "command", "sandbox", "driver", "keep-running", "rm", "detach"} {
		if run.Flags().Lookup(flag) == nil {
			t.Errorf("run missing --%s", flag)
		}
	}
	for _, flag := range []string{"interactive", "tty", "jupyter"} {
		if run.Flags().Lookup(flag) != nil {
			t.Errorf("run unexpectedly has --%s", flag)
		}
	}
	execCmd, _, _ := cmd.Find([]string{"exec"})
	for _, flag := range []string{"command", "cwd", "env", "exec-timeout", "max-output-bytes"} {
		if execCmd.Flags().Lookup(flag) == nil {
			t.Errorf("exec missing --%s", flag)
		}
	}
	for _, flag := range []string{"interactive", "tty", "prompt", "run"} {
		if execCmd.Flags().Lookup(flag) != nil {
			t.Errorf("exec unexpectedly has --%s", flag)
		}
	}
}

func TestBearerTransportClonesRequestAndInjectsToken(t *testing.T) {
	var got string
	base := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Header.Get("Authorization")
		return &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})
	transport := &bearerTransport{token: "secret-value", base: base}
	request, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
	request.Header.Set("Authorization", "original")
	if _, err := transport.RoundTrip(request); err != nil {
		t.Fatal(err)
	}
	if got != "Bearer secret-value" {
		t.Fatalf("Authorization = %q", got)
	}
	if request.Header.Get("Authorization") != "original" {
		t.Fatal("original request was mutated")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestApplyRuntimeConfigFlagPrecedence(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"--url", "https://flag.example", "--project", "flag-project", "status"})
	if err := cmd.ParseFlags([]string{"--url", "https://flag.example", "--project", "flag-project"}); err != nil {
		t.Fatal(err)
	}
	var node yaml.Node
	if err := node.Encode(productConfig{URL: "https://config.example", APIToken: "token", DefaultProject: "config-project", Timeout: "9s"}); err != nil {
		t.Fatal(err)
	}
	ApplyRuntimeConfig(cmd, config.Raw{productName: node}, "config.yaml", false)
	state := stateFromCommand(cmd)
	if state.options.URL != "https://flag.example" || state.options.Project != "flag-project" {
		t.Fatalf("options = %+v", state.options)
	}
	if state.options.Token != "token" || state.options.timeoutText != "9s" {
		t.Fatalf("config not applied: %+v", state.options)
	}
}

func TestApplyRuntimeConfigAPIKeyAlias(t *testing.T) {
	cmd := NewCommand()
	var node yaml.Node
	if err := node.Encode(map[string]string{"url": "https://example.com/agent-compose", "api_key": "legacy-key"}); err != nil {
		t.Fatal(err)
	}
	ApplyRuntimeConfig(cmd, config.Raw{productName: node}, "config.yaml", false)
	state := stateFromCommand(cmd)
	if state.options.URL != "https://example.com/agent-compose" || state.options.Token != "legacy-key" || state.options.TokenSource != "config" {
		t.Fatalf("options = %+v", state.options)
	}
}

func TestApplyRuntimeConfigPrefersAPITokenOverAPIKey(t *testing.T) {
	cmd := NewCommand()
	var node yaml.Node
	if err := node.Encode(productConfig{URL: "https://example.com", APIToken: "canonical-token", APIKey: "legacy-key"}); err != nil {
		t.Fatal(err)
	}
	ApplyRuntimeConfig(cmd, config.Raw{productName: node}, "config.yaml", false)
	state := stateFromCommand(cmd)
	if state.options.Token != "canonical-token" || state.options.TokenSource != "config" {
		t.Fatalf("options = %+v", state.options)
	}
}

func TestApplyRuntimeConfigIgnoresUnrelatedAPIKeyEnv(t *testing.T) {
	t.Setenv("AGENT_COMPOSE_API_KEY", "environment-key")
	cmd := NewCommand()
	var node yaml.Node
	if err := node.Encode(productConfig{URL: "https://example.com", APIToken: "config-token"}); err != nil {
		t.Fatal(err)
	}
	ApplyRuntimeConfig(cmd, config.Raw{productName: node}, "config.yaml", false)
	state := stateFromCommand(cmd)
	if state.options.Token != "config-token" || state.options.TokenSource != "config" {
		t.Fatalf("options = %+v", state.options)
	}
}

func TestApplyRuntimeConfigEnvironmentPrecedence(t *testing.T) {
	t.Setenv("AGENT_COMPOSE_URL", "https://environment.example")
	t.Setenv("AGENT_COMPOSE_API_TOKEN", "environment-token")
	t.Setenv("AGENT_COMPOSE_DEFAULT_PROJECT", "environment-project")
	t.Setenv("AGENT_COMPOSE_TIMEOUT", "7s")
	t.Setenv("AGENT_COMPOSE_INSECURE", "true")
	cmd := NewCommand()
	var node yaml.Node
	if err := node.Encode(productConfig{URL: "https://config.example", APIToken: "config-token", DefaultProject: "config-project", Timeout: "9s"}); err != nil {
		t.Fatal(err)
	}
	ApplyRuntimeConfig(cmd, config.Raw{productName: node}, "config.yaml", false)
	state := stateFromCommand(cmd)
	if state.options.URL != "https://environment.example" || state.options.Token != "environment-token" || state.options.Project != "environment-project" || state.options.timeoutText != "7s" || !state.options.Insecure || state.options.TokenSource != "environment" {
		t.Fatalf("options = %+v", state.options)
	}
}

func TestNormalizeSandboxStatuses(t *testing.T) {
	got := normalizeSandboxStatuses([]string{"running", " stopped ", ""})
	want := []string{"RUNNING", "STOPPED"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
