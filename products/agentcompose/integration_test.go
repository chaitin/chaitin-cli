package agentcompose

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/chaitin/agent-compose/proto/agentcompose/v2/agentcomposev2connect"
	healthv1 "github.com/chaitin/agent-compose/proto/health/v1"
	"github.com/chaitin/agent-compose/proto/health/v1/healthv1connect"
	"github.com/chaitin/chaitin-cli/config"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/yaml.v3"
)

type projectStub struct {
	agentcomposev2connect.UnimplementedProjectServiceHandler
	mu                     sync.Mutex
	requests               []*agentcomposev2.ListProjectsRequest
	project                *agentcomposev2.Project
	projects               []*agentcomposev2.ProjectSummary
	schedulerRuns          []*agentcomposev2.SchedulerRun
	schedulerEvents        []*agentcomposev2.SchedulerEvent
	schedulerRunRequests   []*agentcomposev2.ListSchedulerRunsRequest
	schedulerEventRequests []*agentcomposev2.ListProjectSchedulerEventsRequest
	schedulerResponses     map[string]*agentcomposev2.GetSchedulerResponse
}

func (s *projectStub) ListSchedulerRuns(_ context.Context, req *connect.Request[agentcomposev2.ListSchedulerRunsRequest]) (*connect.Response[agentcomposev2.ListSchedulerRunsResponse], error) {
	s.schedulerRunRequests = append(s.schedulerRunRequests, proto.Clone(req.Msg).(*agentcomposev2.ListSchedulerRunsRequest))
	start := int(req.Msg.GetOffset())
	if start >= len(s.schedulerRuns) {
		return connect.NewResponse(&agentcomposev2.ListSchedulerRunsResponse{Total: uint32(len(s.schedulerRuns))}), nil
	}
	end := start + int(req.Msg.GetLimit())
	if end > len(s.schedulerRuns) {
		end = len(s.schedulerRuns)
	}
	return connect.NewResponse(&agentcomposev2.ListSchedulerRunsResponse{Runs: s.schedulerRuns[start:end], Total: uint32(len(s.schedulerRuns))}), nil
}

func (s *projectStub) ListProjectSchedulerEvents(_ context.Context, req *connect.Request[agentcomposev2.ListProjectSchedulerEventsRequest]) (*connect.Response[agentcomposev2.ListProjectSchedulerEventsResponse], error) {
	s.schedulerEventRequests = append(s.schedulerEventRequests, proto.Clone(req.Msg).(*agentcomposev2.ListProjectSchedulerEventsRequest))
	start := int(req.Msg.GetOffset())
	if start >= len(s.schedulerEvents) {
		return connect.NewResponse(&agentcomposev2.ListProjectSchedulerEventsResponse{Total: uint32(len(s.schedulerEvents))}), nil
	}
	end := start + int(req.Msg.GetLimit())
	if end > len(s.schedulerEvents) {
		end = len(s.schedulerEvents)
	}
	return connect.NewResponse(&agentcomposev2.ListProjectSchedulerEventsResponse{Events: s.schedulerEvents[start:end], Total: uint32(len(s.schedulerEvents))}), nil
}

func (s *projectStub) GetSchedulerRun(_ context.Context, req *connect.Request[agentcomposev2.GetSchedulerRunRequest]) (*connect.Response[agentcomposev2.GetSchedulerRunResponse], error) {
	for _, run := range s.schedulerRuns {
		if run.GetRunId() == req.Msg.GetRunId() {
			return connect.NewResponse(&agentcomposev2.GetSchedulerRunResponse{Run: run}), nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("Scheduler Run not found"))
}

func (s *projectStub) GetScheduler(_ context.Context, req *connect.Request[agentcomposev2.GetSchedulerRequest]) (*connect.Response[agentcomposev2.GetSchedulerResponse], error) {
	response, ok := s.schedulerResponses[req.Msg.GetAgentName()]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("Scheduler not found"))
	}
	return connect.NewResponse(response), nil
}

func (s *projectStub) ListProjects(_ context.Context, req *connect.Request[agentcomposev2.ListProjectsRequest]) (*connect.Response[agentcomposev2.ListProjectsResponse], error) {
	s.mu.Lock()
	s.requests = append(s.requests, req.Msg)
	s.mu.Unlock()
	all := s.projects
	if len(all) == 0 {
		all = []*agentcomposev2.ProjectSummary{s.project.GetSummary()}
	}
	start := int(req.Msg.GetOffset())
	if start >= len(all) {
		return connect.NewResponse(&agentcomposev2.ListProjectsResponse{Total: uint32(len(all))}), nil
	}
	end := start + int(req.Msg.GetLimit())
	if end > len(all) {
		end = len(all)
	}
	return connect.NewResponse(&agentcomposev2.ListProjectsResponse{Projects: all[start:end], Total: uint32(len(all))}), nil
}
func (s *projectStub) GetProject(context.Context, *connect.Request[agentcomposev2.GetProjectRequest]) (*connect.Response[agentcomposev2.GetProjectResponse], error) {
	return connect.NewResponse(&agentcomposev2.GetProjectResponse{Project: s.project}), nil
}

type runStub struct {
	agentcomposev2connect.UnimplementedRunServiceHandler
	mu            sync.Mutex
	runs          []*agentcomposev2.RunSummary
	listRequests  []*agentcomposev2.ListRunsRequest
	followCalls   int
	chunks        int
	blockFollow   bool
	followStarted chan struct{}
}

type sandboxStub struct {
	agentcomposev2connect.UnimplementedSandboxServiceHandler
	mu           sync.Mutex
	sandboxes    []*agentcomposev2.Sandbox
	listRequests []*agentcomposev2.ListSandboxesRequest
	statsCalls   []string
	history      *agentcomposev2.ListSandboxHistoryResponse
	historyCalls []string
	stopCalls    []string
	failStop     string
}

type resourceStub struct {
	agentcomposev2connect.UnimplementedResourceServiceHandler
	requests []*agentcomposev2.ResolveResourceIDRequest
	targets  map[string][]*agentcomposev2.ResourceTarget
}

func (s *resourceStub) ResolveID(_ context.Context, req *connect.Request[agentcomposev2.ResolveResourceIDRequest]) (*connect.Response[agentcomposev2.ResolveResourceIDResponse], error) {
	s.requests = append(s.requests, proto.Clone(req.Msg).(*agentcomposev2.ResolveResourceIDRequest))
	if targets, ok := s.targets[req.Msg.GetId()]; ok {
		return connect.NewResponse(&agentcomposev2.ResolveResourceIDResponse{Targets: targets}), nil
	}
	for _, kind := range req.Msg.GetKinds() {
		if kind == agentcomposev2.ResourceKind_RESOURCE_KIND_PROJECT && strings.HasPrefix("project-aaaaaaaaaaaaaaaa", req.Msg.GetId()) {
			return connect.NewResponse(&agentcomposev2.ResolveResourceIDResponse{Targets: []*agentcomposev2.ResourceTarget{{Kind: kind, Id: "project-aaaaaaaaaaaaaaaa", ProjectId: "project-aaaaaaaaaaaaaaaa"}}}), nil
		}
	}
	return connect.NewResponse(&agentcomposev2.ResolveResourceIDResponse{}), nil
}

func (s *sandboxStub) ListSandboxHistory(_ context.Context, req *connect.Request[agentcomposev2.ListSandboxHistoryRequest]) (*connect.Response[agentcomposev2.ListSandboxHistoryResponse], error) {
	s.mu.Lock()
	s.historyCalls = append(s.historyCalls, req.Msg.GetSandboxId())
	s.mu.Unlock()
	return connect.NewResponse(s.history), nil
}

func (s *sandboxStub) ListSandboxes(_ context.Context, req *connect.Request[agentcomposev2.ListSandboxesRequest]) (*connect.Response[agentcomposev2.ListSandboxesResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listRequests = append(s.listRequests, proto.Clone(req.Msg).(*agentcomposev2.ListSandboxesRequest))
	start := int(req.Msg.GetOffset())
	if start >= len(s.sandboxes) {
		return connect.NewResponse(&agentcomposev2.ListSandboxesResponse{Total: uint32(len(s.sandboxes))}), nil
	}
	end := start + int(req.Msg.GetLimit())
	if end > len(s.sandboxes) {
		end = len(s.sandboxes)
	}
	return connect.NewResponse(&agentcomposev2.ListSandboxesResponse{Sandboxes: s.sandboxes[start:end], Total: uint32(len(s.sandboxes))}), nil
}

func (s *sandboxStub) GetSandboxStats(_ context.Context, req *connect.Request[agentcomposev2.GetSandboxStatsRequest]) (*connect.Response[agentcomposev2.GetSandboxStatsResponse], error) {
	s.mu.Lock()
	s.statsCalls = append(s.statsCalls, req.Msg.GetSandboxId())
	s.mu.Unlock()
	return connect.NewResponse(&agentcomposev2.GetSandboxStatsResponse{Stats: &agentcomposev2.SandboxStats{SandboxId: req.Msg.GetSandboxId()}}), nil
}

func (s *sandboxStub) GetSandbox(_ context.Context, req *connect.Request[agentcomposev2.GetSandboxRequest]) (*connect.Response[agentcomposev2.GetSandboxResponse], error) {
	for _, sandbox := range s.sandboxes {
		if sandbox.GetSandboxId() == req.Msg.GetSandboxId() {
			return connect.NewResponse(&agentcomposev2.GetSandboxResponse{Sandbox: sandbox}), nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("Sandbox not found"))
}

func (s *sandboxStub) StopSandbox(_ context.Context, req *connect.Request[agentcomposev2.StopSandboxRequest]) (*connect.Response[agentcomposev2.StopSandboxResponse], error) {
	s.stopCalls = append(s.stopCalls, req.Msg.GetSandboxId())
	if req.Msg.GetSandboxId() == s.failStop {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("cannot stop"))
	}
	return connect.NewResponse(&agentcomposev2.StopSandboxResponse{}), nil
}

func (s *runStub) ListRuns(_ context.Context, req *connect.Request[agentcomposev2.ListRunsRequest]) (*connect.Response[agentcomposev2.ListRunsResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listRequests = append(s.listRequests, proto.Clone(req.Msg).(*agentcomposev2.ListRunsRequest))
	start := int(req.Msg.GetOffset())
	if start >= len(s.runs) {
		return connect.NewResponse(&agentcomposev2.ListRunsResponse{}), nil
	}
	end := start + int(req.Msg.GetLimit())
	if end > len(s.runs) {
		end = len(s.runs)
	}
	return connect.NewResponse(&agentcomposev2.ListRunsResponse{Runs: s.runs[start:end], Total: uint32(len(s.runs))}), nil
}
func (s *runStub) FollowRunLogs(ctx context.Context, _ *connect.Request[agentcomposev2.FollowRunLogsRequest], stream *connect.ServerStream[agentcomposev2.RunLogChunk]) error {
	s.mu.Lock()
	s.followCalls++
	chunks := s.chunks
	block := s.blockFollow
	started := s.followStarted
	s.mu.Unlock()
	if started != nil {
		close(started)
	}
	if block {
		<-ctx.Done()
		return ctx.Err()
	}
	for i := range chunks {
		if err := stream.Send(&agentcomposev2.RunLogChunk{Data: fmt.Sprintf("line-%d\n", i), Offset: uint64(i), RunStatus: agentcomposev2.RunStatus_RUN_STATUS_RUNNING}); err != nil {
			return err
		}
	}
	return nil
}

type healthStub struct {
	healthv1connect.UnimplementedHealthServiceHandler
}

type validatingHealthStub struct {
	healthv1connect.UnimplementedHealthServiceHandler
	want string
}

func (s validatingHealthStub) Status(_ context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[healthv1.HealthStatusResponse], error) {
	if req.Header().Get("Authorization") != "Bearer "+s.want {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid Token"))
	}
	return connect.NewResponse(&healthv1.HealthStatusResponse{BuildVersion: "test"}), nil
}

func (healthStub) Status(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[healthv1.HealthStatusResponse], error) {
	return connect.NewResponse(&healthv1.HealthStatusResponse{BuildVersion: "test"}), nil
}

func newTestServer(t *testing.T, project *projectStub, run *runStub, extras ...any) (*httptest.Server, *int) {
	t.Helper()
	mux := http.NewServeMux()
	if project != nil {
		path, handler := agentcomposev2connect.NewProjectServiceHandler(project)
		mux.Handle(path, handler)
	}
	if run != nil {
		path, handler := agentcomposev2connect.NewRunServiceHandler(run)
		mux.Handle(path, handler)
	}
	for _, extra := range extras {
		switch handler := extra.(type) {
		case *sandboxStub:
			path, connectHandler := agentcomposev2connect.NewSandboxServiceHandler(handler)
			mux.Handle(path, connectHandler)
		case *resourceStub:
			path, connectHandler := agentcomposev2connect.NewResourceServiceHandler(handler)
			mux.Handle(path, connectHandler)
		}
	}
	path, handler := healthv1connect.NewHealthServiceHandler(healthStub{})
	mux.Handle(path, handler)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	return server, &requests
}

func fixtureProject() *agentcomposev2.Project {
	return &agentcomposev2.Project{Summary: &agentcomposev2.ProjectSummary{ProjectId: "project-aaaaaaaaaaaaaaaa", Name: "fixture", AgentCount: 1}, Agents: []*agentcomposev2.ProjectAgent{{ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent", ManagedAgentId: "agent-bbbbbbbbbbbbbbbb"}}}
}
func executeCommand(t *testing.T, serverURL string, dryRun bool, args ...string) (string, string, error) {
	t.Helper()
	cmd := NewCommand()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	var node yaml.Node
	if err := node.Encode(productConfig{URL: serverURL, APIToken: "test-token", DefaultProject: "fixture", Timeout: "5s"}); err != nil {
		t.Fatal(err)
	}
	ApplyRuntimeConfig(cmd, config.Raw{productName: node}, t.TempDir()+"/config.yaml", dryRun)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

func TestNestedProjectCommandAndBearerToken(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	server, _ := newTestServer(t, project, nil)
	out, _, err := executeCommand(t, server.URL, false, "--json", "project", "ls", "--all-pages")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"projects"`) || strings.Contains(out, "test-token") {
		t.Fatalf("output = %s", out)
	}
}

func TestRunListProbesAndNeverFetchesLogs(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	run := &runStub{runs: []*agentcomposev2.RunSummary{{RunId: "run-1", ProjectId: "project-aaaaaaaaaaaaaaaa"}, {RunId: "run-2", ProjectId: "project-aaaaaaaaaaaaaaaa"}}}
	server, _ := newTestServer(t, project, run)
	out, _, err := executeCommand(t, server.URL, false, "--json", "run", "ls", "--limit", "1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"has_more":true`) {
		t.Fatalf("output = %s", out)
	}
	if run.followCalls != 0 {
		t.Fatalf("FollowRunLogs calls = %d", run.followCalls)
	}
	if len(run.listRequests) != 1 || run.listRequests[0].GetLimit() != 1 {
		t.Fatalf("requests = %+v", run.listRequests)
	}
}

func TestRunListLimitOverOnePageKeepsContinuationOffset(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	runs := make([]*agentcomposev2.RunSummary, 201)
	for i := range runs {
		runs[i] = &agentcomposev2.RunSummary{RunId: fmt.Sprintf("run-%03d", i), ProjectId: "project-aaaaaaaaaaaaaaaa"}
	}
	run := &runStub{runs: runs}
	server, _ := newTestServer(t, project, run)
	out, _, err := executeCommand(t, server.URL, false, "--json", "run", "ls", "--limit", "150")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"next_offset":150`) || !strings.Contains(out, `"has_more":true`) {
		t.Fatalf("output = %s", out)
	}
	if len(run.listRequests) != 2 || run.listRequests[0].GetLimit() != 100 || run.listRequests[1].GetLimit() != 50 || run.listRequests[1].GetOffset() != 100 {
		t.Fatalf("requests = %#v", run.listRequests)
	}
}

func TestProjectListLimitOverOnePageKeepsContinuationOffset(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	for i := range 201 {
		project.projects = append(project.projects, &agentcomposev2.ProjectSummary{ProjectId: fmt.Sprintf("project-%03d", i), Name: fmt.Sprintf("project-%03d", i)})
	}
	server, _ := newTestServer(t, project, nil)
	out, _, err := executeCommand(t, server.URL, false, "--json", "project", "ls", "--limit", "150")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"next_offset":150`) || !strings.Contains(out, `"has_more":true`) {
		t.Fatalf("output = %s", out)
	}
	if len(project.requests) != 2 || project.requests[0].GetLimit() != 100 || project.requests[1].GetLimit() != 50 || project.requests[1].GetOffset() != 100 {
		t.Fatalf("requests = %#v", project.requests)
	}
}

func TestSandboxListFiltersIgnoredServerFieldsAndPreservesCursor(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	sandbox := &sandboxStub{sandboxes: []*agentcomposev2.Sandbox{
		{SandboxId: "wrong-project", ProjectId: "project-other", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_RUNNING},
		{SandboxId: "stopped", ProjectId: "project-aaaaaaaaaaaaaaaa", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_STOPPED},
		{SandboxId: "running-1", ProjectId: "project-aaaaaaaaaaaaaaaa", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_RUNNING},
		{SandboxId: "running-2", ProjectId: "project-aaaaaaaaaaaaaaaa", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_RUNNING},
	}}
	server, _ := newTestServer(t, project, &runStub{}, sandbox)
	out, _, err := executeCommand(t, server.URL, false, "--json", "ps", "--limit", "1")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"sandbox_id":"running-1"`, `"has_more":true`, `"next_offset":3`} {
		if !strings.Contains(out, want) {
			t.Fatalf("output = %s, want %s", out, want)
		}
	}
	if strings.Contains(out, "wrong-project") || strings.Contains(out, "stopped") || strings.Contains(out, "running-2") {
		t.Fatalf("output contains filtered Sandbox: %s", out)
	}
	out, _, err = executeCommand(t, server.URL, false, "--json", "ps", "--limit", "1", "--offset", "3")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"sandbox_id":"running-2"`) || !strings.Contains(out, `"has_more":false`) {
		t.Fatalf("continued output = %s", out)
	}
}

func TestSandboxListAssociatesSchedulerRun(t *testing.T) {
	project := &projectStub{
		project:       fixtureProject(),
		schedulerRuns: []*agentcomposev2.SchedulerRun{{RunId: "scheduler-run-123456", SandboxIds: []string{"scheduler-sandbox"}}},
	}
	sandbox := &sandboxStub{sandboxes: []*agentcomposev2.Sandbox{{SandboxId: "scheduler-sandbox", ProjectId: "project-aaaaaaaaaaaaaaaa", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_RUNNING}}}
	server, _ := newTestServer(t, project, &runStub{}, sandbox)
	out, _, err := executeCommand(t, server.URL, false, "--json", "ps")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"run_id":"scheduler-run-123456"`) {
		t.Fatalf("output = %s", out)
	}
}

func TestSchedulerLogsTailReadsNewestEventsAndOutputsChronologically(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	for i := range 100 {
		project.schedulerEvents = append(project.schedulerEvents, &agentcomposev2.SchedulerEvent{Id: fmt.Sprintf("newest-%03d", i), Message: fmt.Sprintf("message-%03d", i)})
	}
	server, _ := newTestServer(t, project, nil)
	out, _, err := executeCommand(t, server.URL, false, "--json", "scheduler", "logs", "--tail", "10")
	if err != nil {
		t.Fatal(err)
	}
	if len(project.schedulerEventRequests) != 1 || project.schedulerEventRequests[0].GetLimit() != 10 {
		t.Fatalf("requests = %#v", project.schedulerEventRequests)
	}
	if strings.Index(out, "newest-009") > strings.Index(out, "newest-000") || strings.Contains(out, "newest-010") {
		t.Fatalf("output is not the newest 10 events in chronological order: %s", out)
	}
}

func TestSchedulerLogsResolvesRunShortIDAndRejectsDuplicateSelectors(t *testing.T) {
	project := &projectStub{
		project:         fixtureProject(),
		schedulerRuns:   []*agentcomposev2.SchedulerRun{{RunId: "scheduler-run-complete", ProjectId: "project-aaaaaaaaaaaaaaaa"}},
		schedulerEvents: []*agentcomposev2.SchedulerEvent{{Id: "event"}},
	}
	server, _ := newTestServer(t, project, nil)
	if _, _, err := executeCommand(t, server.URL, false, "--json", "scheduler", "logs", "--run", "scheduler-run"); err != nil {
		t.Fatal(err)
	}
	last := project.schedulerEventRequests[len(project.schedulerEventRequests)-1]
	if last.GetRunId() != "scheduler-run-complete" {
		t.Fatalf("run id = %q", last.GetRunId())
	}
	_, _, err := executeCommand(t, server.URL, false, "scheduler", "logs", "scheduler-run", "--run", "scheduler-run-complete")
	coder, ok := err.(interface{ ExitCode() int })
	if !ok || coder.ExitCode() != exitUsage {
		t.Fatalf("duplicate selector error = %v", err)
	}
}

func TestSchedulerRunsAcceptsExactHistoricalTrigger(t *testing.T) {
	project := &projectStub{
		project:       fixtureProject(),
		schedulerRuns: []*agentcomposev2.SchedulerRun{{RunId: "historical-run", ProjectId: "project-aaaaaaaaaaaaaaaa", TriggerId: "deleted-trigger-full-id"}},
	}
	server, _ := newTestServer(t, project, nil)
	if _, _, err := executeCommand(t, server.URL, false, "--json", "scheduler", "runs", "--trigger", "deleted-trigger-full-id"); err != nil {
		t.Fatal(err)
	}
	last := project.schedulerRunRequests[len(project.schedulerRunRequests)-1]
	if last.GetTriggerId() != "deleted-trigger-full-id" {
		t.Fatalf("trigger id = %q", last.GetTriggerId())
	}
}

func TestSchedulerTriggerNameAmbiguityAcrossSchedulers(t *testing.T) {
	fixture := fixtureProject()
	fixture.Schedulers = []*agentcomposev2.ProjectScheduler{{SchedulerId: "scheduler-a", AgentName: "agent-a"}, {SchedulerId: "scheduler-b", AgentName: "agent-b"}}
	project := &projectStub{
		project: fixture,
		schedulerResponses: map[string]*agentcomposev2.GetSchedulerResponse{
			"agent-a": {Scheduler: fixture.Schedulers[0], Triggers: []*agentcomposev2.ResolvedTrigger{{TriggerId: "trigger-a", Spec: &agentcomposev2.TriggerSpec{Name: "daily"}}}},
			"agent-b": {Scheduler: fixture.Schedulers[1], Triggers: []*agentcomposev2.ResolvedTrigger{{TriggerId: "trigger-b", Spec: &agentcomposev2.TriggerSpec{Name: "daily"}}}},
		},
	}
	server, _ := newTestServer(t, project, nil)
	_, _, err := executeCommand(t, server.URL, false, "scheduler", "runs", "--trigger", "daily")
	coder, ok := err.(interface{ ExitCode() int })
	if !ok || coder.ExitCode() != exitUsage {
		t.Fatalf("error = %v", err)
	}
}

func TestStatsFiltersProjectAndRunningStatusClientSide(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	sandbox := &sandboxStub{sandboxes: []*agentcomposev2.Sandbox{
		{SandboxId: "wrong-project", ProjectId: "project-other", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_RUNNING},
		{SandboxId: "stopped", ProjectId: "project-aaaaaaaaaaaaaaaa", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_STOPPED},
		{SandboxId: "running", ProjectId: "project-aaaaaaaaaaaaaaaa", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_RUNNING},
	}}
	server, _ := newTestServer(t, project, nil, sandbox)
	if _, _, err := executeCommand(t, server.URL, false, "--json", "stats"); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(sandbox.statsCalls, ","); got != "running" {
		t.Fatalf("stats calls = %q, want running", got)
	}
	if len(sandbox.listRequests) != 1 || sandbox.listRequests[0].GetLimit() != 100 {
		t.Fatalf("Sandbox full scan requests = %#v", sandbox.listRequests)
	}
}

func TestResolveSandboxRejectsOtherProjects(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	sandbox := &sandboxStub{sandboxes: []*agentcomposev2.Sandbox{
		{SandboxId: "sandbox-other", ProjectId: "project-other", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_RUNNING},
		{SandboxId: "sandbox-local", ProjectId: "project-aaaaaaaaaaaaaaaa", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_RUNNING},
	}}
	server, _ := newTestServer(t, project, nil, sandbox)
	_, _, err := executeCommand(t, server.URL, false, "--json", "stats", "sandbox-other")
	coder, ok := err.(interface{ ExitCode() int })
	if !ok || coder.ExitCode() != exitNotFound {
		t.Fatalf("error = %v", err)
	}
}

func TestGenericInspectProjectDoesNotRequireDefaultProject(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	server, _ := newTestServer(t, project, &runStub{}, &sandboxStub{})
	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--json", "inspect", "fixture"})
	var node yaml.Node
	if err := node.Encode(productConfig{URL: server.URL, APIToken: "test-token", Timeout: "5s"}); err != nil {
		t.Fatal(err)
	}
	ApplyRuntimeConfig(cmd, config.Raw{productName: node}, t.TempDir()+"/config.yaml", false)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"name":"fixture"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestProjectShortIDUsesResolveID(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	resource := &resourceStub{}
	server, _ := newTestServer(t, project, &runStub{}, &sandboxStub{}, resource)
	out, _, err := executeCommand(t, server.URL, false, "--json", "inspect", "project-aaa")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"name":"fixture"`) || len(resource.requests) == 0 || resource.requests[0].GetKinds()[0] != agentcomposev2.ResourceKind_RESOURCE_KIND_PROJECT {
		t.Fatalf("output = %s, requests = %#v", out, resource.requests)
	}
}

func TestPositionalLogsTargetsUseResolveIDForAllResourceKinds(t *testing.T) {
	project := fixtureProject()
	run := &runStub{runs: []*agentcomposev2.RunSummary{{RunId: "run-full", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent", SandboxId: "sandbox-full"}}}
	sandbox := &sandboxStub{sandboxes: []*agentcomposev2.Sandbox{{SandboxId: "sandbox-full", ProjectId: "project-aaaaaaaaaaaaaaaa"}}}
	resource := &resourceStub{targets: map[string][]*agentcomposev2.ResourceTarget{
		"encoded-project": {{Kind: agentcomposev2.ResourceKind_RESOURCE_KIND_PROJECT, Id: "project-aaaaaaaaaaaaaaaa", ProjectId: "project-aaaaaaaaaaaaaaaa"}},
		"encoded-agent":   {{Kind: agentcomposev2.ResourceKind_RESOURCE_KIND_AGENT, Id: "agent-bbbbbbbbbbbbbbbb", ProjectId: "project-aaaaaaaaaaaaaaaa"}},
		"encoded-run":     {{Kind: agentcomposev2.ResourceKind_RESOURCE_KIND_RUN, Id: "run-full", ProjectId: "project-aaaaaaaaaaaaaaaa"}},
		"encoded-sandbox": {{Kind: agentcomposev2.ResourceKind_RESOURCE_KIND_SANDBOX, Id: "sandbox-full", ProjectId: "project-aaaaaaaaaaaaaaaa"}},
	}}
	server, _ := newTestServer(t, &projectStub{project: project}, run, sandbox, resource)
	cmd := NewCommand()
	var node yaml.Node
	if err := node.Encode(productConfig{URL: server.URL, APIToken: "test-token", DefaultProject: "fixture", Timeout: "5s"}); err != nil {
		t.Fatal(err)
	}
	ApplyRuntimeConfig(cmd, config.Raw{productName: node}, t.TempDir()+"/config.yaml", false)
	state := stateFromCommand(cmd)
	for _, test := range []struct{ ref, kind, value string }{{"encoded-project", "project", "project-aaaaaaaaaaaaaaaa"}, {"encoded-agent", "agent", "agent"}, {"encoded-run", "run", "run-full"}, {"encoded-sandbox", "sandbox", "sandbox-full"}} {
		kind, value, err := resolveLogsTarget(context.Background(), state, project, test.ref)
		if err != nil || kind != test.kind || value != test.value {
			t.Fatalf("%s = %s/%s, %v", test.ref, kind, value, err)
		}
	}
}

func TestLogsJSONStreamsNDJSON(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	run := &runStub{runs: []*agentcomposev2.RunSummary{{RunId: "run-1234567890123456", RunShortId: "run-123", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent"}}, chunks: 200}
	server, _ := newTestServer(t, project, run)
	out, _, err := executeCommand(t, server.URL, false, "--json", "logs", "--run", "run-123")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 200 {
		t.Fatalf("NDJSON lines = %d, want 200", len(lines))
	}
	if run.followCalls != 1 {
		t.Fatalf("FollowRunLogs calls = %d", run.followCalls)
	}
}

func TestLogsCancellationReturns130(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	run := &runStub{
		runs:          []*agentcomposev2.RunSummary{{RunId: "run-cancel", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent"}},
		blockFollow:   true,
		followStarted: make(chan struct{}),
	}
	server, _ := newTestServer(t, project, run)
	cmd := NewCommand()
	cmd.SetArgs([]string{"logs", "--run", "run-cancel"})
	var node yaml.Node
	if err := node.Encode(productConfig{URL: server.URL, APIToken: "test-token", DefaultProject: "fixture", Timeout: "5s"}); err != nil {
		t.Fatal(err)
	}
	ApplyRuntimeConfig(cmd, config.Raw{productName: node}, t.TempDir()+"/config.yaml", false)
	ctx, cancel := context.WithCancel(context.Background())
	cmd.SetContext(ctx)
	done := make(chan error, 1)
	go func() { done <- cmd.Execute() }()
	<-run.followStarted
	cancel()
	err := <-done
	coder, ok := err.(interface{ ExitCode() int })
	if !ok || coder.ExitCode() != exitCanceled {
		t.Fatalf("error = %v", err)
	}
}

func TestLogsFallsBackToSandboxHistoryWithoutRun(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	sandbox := &sandboxStub{
		sandboxes: []*agentcomposev2.Sandbox{{SandboxId: "sandbox-history", ProjectId: "project-aaaaaaaaaaaaaaaa", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_STOPPED}},
		history: &agentcomposev2.ListSandboxHistoryResponse{
			Cells:  []*agentcomposev2.SandboxHistoryCell{{Output: "first\nsecond\n"}},
			Events: []*agentcomposev2.SandboxHistoryEvent{{Id: "event-1", Type: "completed", Level: "info"}},
		},
	}
	server, _ := newTestServer(t, project, &runStub{}, sandbox)
	out, _, err := executeCommand(t, server.URL, false, "--json", "logs", "--sandbox", "sandbox-history", "--tail", "1")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"sandbox_id":"sandbox-history"`, `"output":"second\n"`, `"events":[{"id":"event-1"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("output = %s, want %s", out, want)
		}
	}
	if len(sandbox.historyCalls) != 1 || sandbox.historyCalls[0] != "sandbox-history" {
		t.Fatalf("history calls = %#v", sandbox.historyCalls)
	}
	if !strings.HasSuffix(strings.TrimSpace(out), "}") || strings.Count(strings.TrimSpace(out), "\n") != 0 {
		t.Fatalf("history JSON is not one NDJSON record: %q", out)
	}
}

func TestDryRunWriteCommandsMakeNoRequests(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	run := &runStub{}
	server, requests := newTestServer(t, project, run)
	tests := [][]string{{"run", "agent"}, {"run", "stop", "run"}, {"scheduler", "invoke", "agent"}, {"scheduler", "trigger", "agent", "trigger"}, {"scheduler", "stop", "run"}, {"sandbox", "stop", "sandbox"}, {"sandbox", "resume", "sandbox"}, {"sandbox", "rm", "sandbox"}, {"exec", "sandbox", "--", "true"}}
	for _, args := range tests {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			before := *requests
			_, _, err := executeCommand(t, server.URL, true, args...)
			coder, ok := err.(interface{ ExitCode() int })
			if !ok || coder.ExitCode() != exitUnsupported {
				t.Fatalf("error = %v", err)
			}
			if *requests != before {
				t.Fatalf("network requests = %d, want %d", *requests, before)
			}
		})
	}
}

func TestDryRunWriteCommandsNeedNeitherURLNorToken(t *testing.T) {
	for _, args := range [][]string{{"run", "agent"}, {"run", "stop", "run"}, {"scheduler", "invoke", "agent"}, {"scheduler", "trigger", "agent", "trigger"}, {"scheduler", "stop", "run"}, {"sandbox", "stop", "sandbox"}, {"sandbox", "resume", "sandbox"}, {"sandbox", "rm", "sandbox"}, {"exec", "sandbox", "--", "true"}, {"auth", "login", "--token-stdin"}, {"auth", "logout"}} {
		cmd := NewCommand()
		cmd.SetArgs(args)
		ApplyRuntimeConfig(cmd, config.Raw{}, t.TempDir()+"/config.yaml", true)
		err := cmd.Execute()
		coder, ok := err.(interface{ ExitCode() int })
		if !ok || coder.ExitCode() != exitUnsupported {
			t.Fatalf("%v error = %v", args, err)
		}
	}
}

func TestSandboxBatchFailureReportsCompletedAndUnattemptedTargets(t *testing.T) {
	project := &projectStub{project: fixtureProject()}
	sandbox := &sandboxStub{
		sandboxes: []*agentcomposev2.Sandbox{
			{SandboxId: "sandbox-one", ProjectId: "project-aaaaaaaaaaaaaaaa"},
			{SandboxId: "sandbox-two", ProjectId: "project-aaaaaaaaaaaaaaaa"},
			{SandboxId: "sandbox-three", ProjectId: "project-aaaaaaaaaaaaaaaa"},
		},
		failStop: "sandbox-two",
	}
	server, _ := newTestServer(t, project, nil, sandbox)
	_, _, err := executeCommand(t, server.URL, false, "--json", "sandbox", "stop", "sandbox-one", "sandbox-two", "sandbox-three")
	cliErr, ok := err.(*CLIError)
	if !ok || cliErr.Kind != "partial_failure" || strings.Join(cliErr.Completed, ",") != "sandbox-one" || strings.Join(cliErr.Unattempted, ",") != "sandbox-three" {
		t.Fatalf("error = %#v", err)
	}
	if got := strings.Join(sandbox.stopCalls, ","); got != "sandbox-one,sandbox-two" {
		t.Fatalf("stop calls = %q", got)
	}
}

func TestMissingTokenFailsBeforeNetwork(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected request") }))
	defer server.Close()
	cmd := NewCommand()
	cmd.SetArgs([]string{"--url", server.URL, "status"})
	err := cmd.Execute()
	coder, ok := err.(interface{ ExitCode() int })
	if !ok || coder.ExitCode() != exitAuth {
		t.Fatalf("error = %v", err)
	}
}

func TestAuthLoginValidatesBeforeSavingAndHidesToken(t *testing.T) {
	mux := http.NewServeMux()
	path, handler := healthv1connect.NewHealthServiceHandler(validatingHealthStub{want: "new-secret-token"})
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	defer server.Close()
	configPath := t.TempDir() + "/nested/config.yaml"
	cmd := NewCommand()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetIn(strings.NewReader("new-secret-token\n"))
	cmd.SetArgs([]string{"--url", server.URL, "auth", "login", "--token-stdin"})
	ApplyRuntimeConfig(cmd, config.Raw{}, configPath, false)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "new-secret-token") || strings.Contains(errOut.String(), "new-secret-token") {
		t.Fatal("Token leaked in output")
	}
	stored, err := loadStoredConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if stored.APIToken != "new-secret-token" || stored.URL != server.URL {
		t.Fatalf("stored = %+v", stored)
	}
}

func TestAuthLoginInvalidTokenDoesNotWrite(t *testing.T) {
	mux := http.NewServeMux()
	path, handler := healthv1connect.NewHealthServiceHandler(validatingHealthStub{want: "valid"})
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	defer server.Close()
	configPath := t.TempDir() + "/config.yaml"
	cmd := NewCommand()
	cmd.SetIn(strings.NewReader("invalid\n"))
	cmd.SetArgs([]string{"--url", server.URL, "auth", "login", "--token-stdin"})
	ApplyRuntimeConfig(cmd, config.Raw{}, configPath, false)
	err := cmd.Execute()
	coder, ok := err.(interface{ ExitCode() int })
	if !ok || coder.ExitCode() != exitAuth {
		t.Fatalf("error = %v", err)
	}
	if _, statErr := os.Stat(configPath); !os.IsNotExist(statErr) {
		t.Fatalf("config was written: %v", statErr)
	}
}

func TestAuthLogoutOnlyClearsStoredTokenAndWarnsForEnvironment(t *testing.T) {
	t.Setenv("AGENT_COMPOSE_API_TOKEN", "environment-token")
	configPath := t.TempDir() + "/config.yaml"
	if err := config.SetProduct(configPath, productName, productConfig{URL: "https://example.com", APIToken: "stored-token", DefaultProject: "fixture", Timeout: "30s"}); err != nil {
		t.Fatal(err)
	}
	raw, err := config.Load(configPath)
	if err != nil {
		t.Fatal(err)
	}
	cmd := NewCommand()
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"auth", "logout"})
	ApplyRuntimeConfig(cmd, raw, configPath, false)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	stored, err := loadStoredConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if stored.APIToken != "" || stored.DefaultProject != "fixture" {
		t.Fatalf("stored = %+v", stored)
	}
	if !strings.Contains(errOut.String(), "AGENT_COMPOSE_API_TOKEN") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestAuthStatusReportsEnvironmentTokenSource(t *testing.T) {
	t.Setenv("AGENT_COMPOSE_API_TOKEN", "test-token")
	server, _ := newTestServer(t, nil, nil)
	out, _, err := executeCommand(t, server.URL, false, "--json", "auth", "status")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"token_source":"environment"`) || strings.Contains(out, "test-token") {
		t.Fatalf("output = %s", out)
	}
}
