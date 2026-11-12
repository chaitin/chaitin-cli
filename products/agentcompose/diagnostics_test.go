package agentcompose

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/chaitin/agent-compose/proto/agentcompose/v2/agentcomposev2connect"
)

func (s *runStub) GetRun(_ context.Context, req *connect.Request[agentcomposev2.GetRunRequest]) (*connect.Response[agentcomposev2.GetRunResponse], error) {
	detail, ok := s.details[req.Msg.GetRunId()]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("Run not found"))
	}
	return connect.NewResponse(&agentcomposev2.GetRunResponse{Run: detail}), nil
}

type llmStub struct {
	agentcomposev2connect.UnimplementedLLMServiceHandler
	providers    []*agentcomposev2.LLMProvider
	defaultModel *agentcomposev2.LLMModelReference
}

func (s *llmStub) ListProviders(_ context.Context, req *connect.Request[agentcomposev2.ListProvidersRequest]) (*connect.Response[agentcomposev2.ListProvidersResponse], error) {
	start := min(int(req.Msg.GetOffset()), len(s.providers))
	end := min(start+int(req.Msg.GetLimit()), len(s.providers))
	return connect.NewResponse(&agentcomposev2.ListProvidersResponse{Providers: s.providers[start:end], Total: uint32(len(s.providers))}), nil
}

func (s *llmStub) GetDefaultModel(context.Context, *connect.Request[agentcomposev2.GetDefaultModelRequest]) (*connect.Response[agentcomposev2.GetDefaultModelResponse], error) {
	return connect.NewResponse(&agentcomposev2.GetDefaultModelResponse{Model: s.defaultModel}), nil
}

func TestParseTimeBound(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 123456789, time.UTC)
	tests := []struct {
		name, input string
		want        time.Time
		wantNil     bool
		wantErr     bool
	}{
		{name: "empty", input: " ", wantNil: true},
		{name: "rfc3339", input: "2026-10-14T08:00:00Z", want: time.Date(2026, 10, 14, 8, 0, 0, 0, time.UTC)},
		{name: "rfc3339 sub-millisecond is truncated", input: "2026-10-14T08:00:00.123456Z", want: time.Date(2026, 10, 14, 8, 0, 0, 123000000, time.UTC)},
		{name: "minutes", input: "30m", want: time.Date(2026, 10, 15, 11, 30, 0, 123000000, time.UTC)},
		{name: "hours", input: "24h", want: time.Date(2026, 10, 14, 12, 0, 0, 123000000, time.UTC)},
		{name: "days", input: "7d", want: time.Date(2026, 10, 8, 12, 0, 0, 123000000, time.UTC)},
		{name: "zero duration", input: "0s", wantErr: true},
		{name: "negative duration", input: "-1h", wantErr: true},
		{name: "fractional days", input: "1.5d", wantErr: true},
		{name: "garbage", input: "yesterday", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseTimeBound(test.input, now)
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr %t", err, test.wantErr)
			}
			if test.wantErr {
				return
			}
			if test.wantNil {
				if got != nil {
					t.Fatalf("got %v, want nil", got)
				}
				return
			}
			if !got.AsTime().Equal(test.want) {
				t.Fatalf("got %s, want %s", got.AsTime(), test.want)
			}
		})
	}
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    map[string]string
		wantErr bool
	}{
		{name: "none", input: nil, want: nil},
		{name: "pairs", input: []string{"env=prod", "team=sec"}, want: map[string]string{"env": "prod", "team": "sec"}},
		{name: "empty value", input: []string{"flag="}, want: map[string]string{"flag": ""}},
		{name: "value with equals", input: []string{"expr=a=b"}, want: map[string]string{"expr": "a=b"}},
		{name: "missing equals", input: []string{"env"}, wantErr: true},
		{name: "empty key", input: []string{"=prod"}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseLabels(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr %t", err, test.wantErr)
			}
			if len(got) != len(test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
			for key, value := range test.want {
				if got[key] != value {
					t.Fatalf("got %v, want %v", got, test.want)
				}
			}
		})
	}
}

func TestRunListForwardsDiagnosticFilters(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	run := &runStub{}
	server, _ := newTestServer(t, project, run)
	before := time.Now()
	_, _, err := executeCommand(t, server.URL, false, "--json", "run", "ls", "--started-from", "24h", "--scheduler-run", "srun-1", "--label", "env=prod", "--label", "team=sec")
	if err != nil {
		t.Fatal(err)
	}
	if len(run.listRequests) != 1 {
		t.Fatalf("requests = %+v", run.listRequests)
	}
	req := run.listRequests[0]
	if req.GetProjectId() != "project-aaaaaaaaaaaaaaaa" || req.GetSchedulerRunId() != "srun-1" || req.GetLabels()["env"] != "prod" || req.GetLabels()["team"] != "sec" {
		t.Fatalf("request = %+v", req)
	}
	from := req.GetStartedFrom().AsTime()
	if from.Nanosecond()%int(time.Millisecond) != 0 {
		t.Fatalf("started_from has sub-millisecond precision: %s", from)
	}
	if lower, upper := before.Add(-24*time.Hour-time.Second), time.Now().Add(-24*time.Hour); from.Before(lower) || from.After(upper) {
		t.Fatalf("started_from = %s, want about 24h ago", from)
	}
}

func TestRunListRejectsInvalidDiagnosticFilters(t *testing.T) {
	for _, args := range [][]string{
		{"run", "ls", "--started-from", "yesterday"},
		{"run", "ls", "--started-to", "-1h"},
		{"run", "ls", "--label", "env"},
		{"run", "ls", "--all-projects", "--project", "fixture"},
	} {
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

func TestRunListAllProjectsSkipsProjectResolution(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	run := &runStub{runs: []*agentcomposev2.RunSummary{
		{RunId: "run-aaaaaaaaaaaaaaaa", ProjectId: "project-a", ProjectName: "alpha", AgentName: "triage", Status: agentcomposev2.RunStatus_RUN_STATUS_FAILED},
		{RunId: "run-bbbbbbbbbbbbbbbb", ProjectId: "project-b", ProjectName: "beta", AgentName: "review", Status: agentcomposev2.RunStatus_RUN_STATUS_FAILED},
	}}
	server, _ := newTestServer(t, project, run)
	out, _, err := executeCommand(t, server.URL, false, "run", "ls", "--all-projects", "--status", "failed")
	if err != nil {
		t.Fatal(err)
	}
	if len(project.requests) != 0 {
		t.Fatalf("ListProjects was called %d times", len(project.requests))
	}
	if len(run.listRequests) != 1 || run.listRequests[0].GetProjectId() != "" || run.listRequests[0].GetStatus() != agentcomposev2.RunStatus_RUN_STATUS_FAILED {
		t.Fatalf("requests = %+v", run.listRequests)
	}
	if !strings.Contains(out, "PROJECT") || !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Fatalf("output = %s", out)
	}

	out, _, err = executeCommand(t, server.URL, false, "--json", "run", "ls", "--all-projects")
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatal(err)
	}
	if _, present := payload["project"]; present {
		t.Fatalf("all-projects JSON has a project field: %s", out)
	}
	if runs, _ := payload["runs"].([]any); len(runs) != 2 {
		t.Fatalf("output = %s", out)
	}
}

func TestInspectRunIncludesDiagnosticFields(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	summary := &agentcomposev2.RunSummary{RunId: "run-aaaaaaaaaaaaaaaa", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent", Status: agentcomposev2.RunStatus_RUN_STATUS_FAILED, Error: "agent exited", SchedulerId: "scheduler-1", TriggerId: "trigger-1", SchedulerRunId: "srun-1"}
	run := &runStub{
		runs: []*agentcomposev2.RunSummary{summary},
		details: map[string]*agentcomposev2.RunDetail{"run-aaaaaaaaaaaaaaaa": {
			Summary:      summary,
			ErrorStack:   "panic: boom\ngoroutine 1",
			CleanupError: "sandbox stop timed out",
			Warnings:     []string{"image pull slow"},
			Labels:       map[string]string{"env": "prod"},
			LogsPath:     "/data/runs/run-aaaa/logs",
			ArtifactsDir: "/data/runs/run-aaaa/artifacts",
		}},
	}
	server, _ := newTestServer(t, project, run, &resourceStub{})
	out, _, err := executeCommand(t, server.URL, false, "--json", "inspect", "run", "run-aaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	var got runDTO
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.ErrorStack != "panic: boom\ngoroutine 1" || got.CleanupError != "sandbox stop timed out" || len(got.Warnings) != 1 || got.Labels["env"] != "prod" || got.LogsPath == "" || got.ArtifactsDir == "" {
		t.Fatalf("detail fields missing: %s", out)
	}
	if got.SchedulerID != "scheduler-1" || got.TriggerID != "trigger-1" || got.SchedulerRunID != "srun-1" {
		t.Fatalf("summary fields missing: %s", out)
	}
}

func TestLLMListMarksDefaultModelAndOmitsHeaders(t *testing.T) {
	maxTokens := int32(4096)
	llm := &llmStub{
		providers: []*agentcomposev2.LLMProvider{{
			Id: "provider-1", Name: "openai", Protocol: "responses", Enabled: true, ApiKeySet: true,
			Models: []*agentcomposev2.LLMModelSpec{
				{Id: "gpt-a", Headers: map[string]string{"X-Secret": "header-secret"}, MaxOutputTokens: &maxTokens},
				{Id: "gpt-b"},
			},
			Capabilities: &agentcomposev2.LLMProviderCapabilities{Probes: []*agentcomposev2.LLMProtocolProbe{{Protocol: "responses", Outcome: "auth_failed", Detail: "401"}}},
		}},
		defaultModel: &agentcomposev2.LLMModelReference{ProviderId: "provider-1", ModelId: "gpt-b"},
	}
	server, _ := newTestServer(t, nil, nil, llm)
	out, _, err := executeCommand(t, server.URL, false, "llm", "ls")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "gpt-b (default)") || strings.Contains(out, "gpt-a (default)") || !strings.Contains(out, "responses=auth_failed") {
		t.Fatalf("output = %s", out)
	}
	out, _, err = executeCommand(t, server.URL, false, "--json", "llm", "ls")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "header-secret") || strings.Contains(out, "X-Secret") {
		t.Fatalf("JSON leaks model headers: %s", out)
	}
	if !strings.Contains(out, `"default_model":{"provider_id":"provider-1","model_id":"gpt-b"}`) || !strings.Contains(out, `"api_key_set":true`) {
		t.Fatalf("output = %s", out)
	}
}

func TestLLMListWithoutDefaultModel(t *testing.T) {
	server, _ := newTestServer(t, nil, nil, &llmStub{})
	out, errOut, err := executeCommand(t, server.URL, false, "--json", "llm", "ls")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "default_model") || errOut != "" {
		t.Fatalf("stdout = %s, stderr = %s", out, errOut)
	}
	_, errOut, err = executeCommand(t, server.URL, false, "llm", "ls")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut, "No default model is configured") {
		t.Fatalf("stderr = %s", errOut)
	}
}
