package agentcompose

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	agentcomposev2connect "github.com/chaitin/agent-compose/proto/agentcompose/v2/agentcomposev2connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidateEventLogTarget(t *testing.T) {
	if err := validateEventLogTarget("evt_abc123", false); err != nil {
		t.Fatalf("validateEventLogTarget(full id) returned error: %v", err)
	}
	for _, value := range []string{"evt_", "evt", "run-123", "abc"} {
		err := validateEventLogTarget(value, false)
		if err == nil {
			t.Fatalf("validateEventLogTarget(%q) = nil, want usage error", value)
		}
		cliErr, ok := err.(*CLIError)
		if !ok || cliErr.ExitCode() != exitUsage {
			t.Fatalf("validateEventLogTarget(%q) = %#v, want usage error", value, err)
		}
	}
}

// eventRunStub emulates the daemon-side event filter: known event ids return
// the seeded runs, unknown ids return NotFound.
type eventRunStub struct {
	agentcomposev2connect.UnimplementedRunServiceHandler
	mu           sync.Mutex
	runs         []*agentcomposev2.RunSummary
	notFound     bool
	truncated    bool
	listRequests []*agentcomposev2.ListRunsRequest
}

func (s *eventRunStub) ListRuns(_ context.Context, req *connect.Request[agentcomposev2.ListRunsRequest]) (*connect.Response[agentcomposev2.ListRunsResponse], error) {
	s.mu.Lock()
	s.listRequests = append(s.listRequests, req.Msg)
	notFound := s.notFound
	truncated := s.truncated
	agentName := req.Msg.GetAgentName()
	s.mu.Unlock()
	if req.Msg.GetEventId() == "" {
		return connect.NewResponse(&agentcomposev2.ListRunsResponse{}), nil
	}
	if notFound {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("event %s not found", req.Msg.GetEventId()))
	}
	runs := make([]*agentcomposev2.RunSummary, 0, len(s.runs))
	for _, run := range s.runs {
		if agentName != "" && run.GetAgentName() != agentName {
			continue
		}
		runs = append(runs, run)
	}
	return connect.NewResponse(&agentcomposev2.ListRunsResponse{Runs: runs, Total: uint32(len(runs)), EventScopeTruncated: truncated}), nil
}

func (s *eventRunStub) FollowRunLogs(_ context.Context, req *connect.Request[agentcomposev2.FollowRunLogsRequest], stream *connect.ServerStream[agentcomposev2.RunLogChunk]) error {
	for _, line := range strings.SplitAfter(s.followOutput(req.Msg.GetRunId()), "\n") {
		if line == "" {
			continue
		}
		if err := stream.Send(&agentcomposev2.RunLogChunk{Data: line, RunStatus: agentcomposev2.RunStatus_RUN_STATUS_SUCCEEDED}); err != nil {
			return err
		}
	}
	return stream.Send(&agentcomposev2.RunLogChunk{RunStatus: agentcomposev2.RunStatus_RUN_STATUS_SUCCEEDED, IsFinal: true})
}

func (s *eventRunStub) followOutput(runID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, run := range s.runs {
		if run.GetRunId() == runID {
			return runID + " output\n"
		}
	}
	return ""
}

func newEventLogsTestServer(t *testing.T, run *eventRunStub) *httptest.Server {
	return newEventLogsTestServerForHandler(t, run)
}

// newEventLogsTestServerForHandler takes the outer handler value so stub
// overrides on embedded structs are dispatched, and asserts the CLI attaches
// the Bearer token to every request including the run service calls.
func newEventLogsTestServerForHandler(t *testing.T, run agentcomposev2connect.RunServiceHandler) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	path, handler := agentcomposev2connect.NewProjectServiceHandler(&projectStub{project: fixtureProject()})
	mux.Handle(path, handler)
	path, handler = agentcomposev2connect.NewRunServiceHandler(run)
	mux.Handle(path, handler)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestExecuteLogsForEventSendsFilterAndValidatesSelectors(t *testing.T) {
	run := &eventRunStub{runs: []*agentcomposev2.RunSummary{{
		RunId: "run-event-agent", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent", SandboxId: "sandbox-1",
	}}}
	server := newEventLogsTestServer(t, run)

	stdout, stderr, err := executeCommand(t, server.URL, false, "logs", "--event", "evt_event_test")
	if err != nil {
		t.Fatalf("logs --event returned error: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "run-event-ag") || !strings.Contains(stdout, "run-event-agent output") {
		t.Fatalf("logs --event output = %q", stdout)
	}
	if len(run.listRequests) != 1 {
		t.Fatalf("ListRuns calls = %d, want 1", len(run.listRequests))
	}
	request := run.listRequests[0]
	if request.GetEventId() != "evt_event_test" {
		t.Fatalf("ListRuns event_id = %q", request.GetEventId())
	}
	if request.GetProjectId() != "project-aaaaaaaaaaaaaaaa" {
		t.Fatalf("ListRuns project id = %q", request.GetProjectId())
	}
	if request.GetLimit() != 200 || request.GetOffset() != 0 {
		t.Fatalf("ListRuns paging = limit %d offset %d, want 200/0", request.GetLimit(), request.GetOffset())
	}

	for _, testCase := range []struct {
		args    []string
		message string
	}{
		{[]string{"logs", "--event", "evt_event_test", "--run", "run-x"}, "--run and --event are mutually exclusive"},
		{[]string{"logs", "--event", "evt_event_test", "--sandbox", "sandbox-x"}, "--sandbox and --event are mutually exclusive"},
		{[]string{"logs", "agent", "--event", "evt_event_test"}, "a positional target and --event are mutually exclusive"},
		{[]string{"logs", "--event", "evt_"}, "full event id"},
	} {
		_, _, err := executeCommand(t, server.URL, false, testCase.args...)
		cliErr, ok := err.(*CLIError)
		if !ok || cliErr.ExitCode() != exitUsage {
			t.Fatalf("%v error = %#v, want usage error", testCase.args, err)
		}
		if !strings.Contains(cliErr.Message, testCase.message) {
			t.Fatalf("%v message = %q, want %q", testCase.args, cliErr.Message, testCase.message)
		}
	}
}

func TestExecuteLogsForEventResolvesAgentReference(t *testing.T) {
	run := &eventRunStub{runs: []*agentcomposev2.RunSummary{
		{RunId: "run-event-agent", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent", SandboxId: "sandbox-1"},
		{RunId: "run-event-other", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "other", SandboxId: "sandbox-2"},
	}}
	server := newEventLogsTestServer(t, run)

	stdout, stderr, err := executeCommand(t, server.URL, false, "logs", "--event", "evt_agent_ref", "--agent", "agent-bbbbbbbbbbbbbbbb")
	if err != nil {
		t.Fatalf("logs --event --agent <managed-id> returned error: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "run-event-agent output") || strings.Contains(stdout, "run-event-other output") {
		t.Fatalf("logs --event --agent <managed-id> output = %q", stdout)
	}
	request := run.listRequests[len(run.listRequests)-1]
	if request.GetAgentName() != "agent" {
		t.Fatalf("ListRuns agent_name = %q, want resolved canonical name agent", request.GetAgentName())
	}
}

func TestExecuteLogsForEventJSONOutputAndEmptyNotice(t *testing.T) {
	run := &eventRunStub{runs: []*agentcomposev2.RunSummary{{
		RunId: "run-event-agent", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent", SandboxId: "sandbox-1",
	}}}
	server := newEventLogsTestServer(t, run)

	stdout, _, err := executeCommand(t, server.URL, false, "--json", "logs", "--event", "evt_json_test")
	if err != nil {
		t.Fatalf("logs --event --json returned error: %v", err)
	}
	decoder := json.NewDecoder(strings.NewReader(stdout))
	var decoded struct {
		AgentName string `json:"agent_name"`
		RunID     string `json:"run_id"`
		IsFinal   bool   `json:"is_final"`
		Content   string `json:"content"`
	}
	var sawContent, sawFinal bool
	for decoder.More() {
		if err := decoder.Decode(&decoded); err != nil {
			t.Fatalf("decode NDJSON line: %v\n%s", err, stdout)
		}
		if decoded.RunID != "run-event-agent" {
			t.Fatalf("NDJSON run_id = %q", decoded.RunID)
		}
		if decoded.Content == "run-event-agent output\n" {
			sawContent = true
		}
		if decoded.IsFinal {
			sawFinal = true
		}
	}
	if !sawContent || !sawFinal {
		t.Fatalf("NDJSON output missing content or final chunk: %s", stdout)
	}

	empty := &eventRunStub{}
	emptyServer := newEventLogsTestServer(t, empty)
	// JSON is a record stream: an empty result writes no records, matching
	// logs --json with no targets.
	jsonOut, errOut, err := executeCommand(t, emptyServer.URL, false, "--json", "logs", "--event", "evt_empty_test")
	if err != nil {
		t.Fatalf("logs --event empty --json returned error: %v\nstderr=%s", err, errOut)
	}
	if strings.TrimSpace(jsonOut) != "" {
		t.Fatalf("logs --event empty --json output = %q, want no records", jsonOut)
	}

	textOut, textErr, err := executeCommand(t, emptyServer.URL, false, "logs", "--event", "evt_empty_test")
	if err != nil {
		t.Fatalf("logs --event empty returned error: %v", err)
	}
	if strings.TrimSpace(textOut) != "" || !strings.Contains(textErr, "No runs are associated with event evt_empty_test") {
		t.Fatalf("logs --event empty stdout/stderr = %q / %q", textOut, textErr)
	}
}

func TestExecuteLogsForEventNotFoundMapsToNotFound(t *testing.T) {
	server := newEventLogsTestServer(t, &eventRunStub{notFound: true})
	_, _, err := executeCommand(t, server.URL, false, "logs", "--event", "evt_missing")
	cliErr, ok := err.(*CLIError)
	if !ok || cliErr.ExitCode() != exitNotFound || !strings.Contains(cliErr.Message, "evt_missing") {
		t.Fatalf("not-found error = %#v, want not_found exit %d", err, exitNotFound)
	}
}

func TestExecuteLogsForEventScopeTruncatedWarns(t *testing.T) {
	server := newEventLogsTestServer(t, &eventRunStub{truncated: true})
	_, stderr, err := executeCommand(t, server.URL, false, "logs", "--event", "evt_big")
	if err != nil {
		t.Fatalf("logs --event truncated returned error: %v", err)
	}
	if !strings.Contains(stderr, "more associated events than the daemon resolves") || !strings.Contains(stderr, "evt_big") {
		t.Fatalf("logs --event truncated stderr = %q", stderr)
	}

	_, stderr, err = executeCommand(t, server.URL, false, "--json", "logs", "--event", "evt_big")
	if err != nil {
		t.Fatalf("logs --event truncated --json returned error: %v", err)
	}
	if !strings.Contains(stderr, "more associated events than the daemon resolves") {
		t.Fatalf("logs --event truncated --json stderr = %q", stderr)
	}
}

func TestExecuteLogsForEventReplaysOldestFirst(t *testing.T) {
	// The daemon returns newest first; mixed fraction precision would expose a
	// string-key sort bug, so the middle run uses a shorter fraction.
	late := mustTimestamp(t, "2026-09-21T10:00:00.15Z")
	early := mustTimestamp(t, "2026-09-21T10:00:00.1Z")
	earliest := mustTimestamp(t, "2026-09-21T10:00:00Z")
	run := &eventRunStub{runs: []*agentcomposev2.RunSummary{
		{RunId: "run-late", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent", SandboxId: "sandbox-3", StartedAt: late},
		{RunId: "run-early", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent", SandboxId: "sandbox-2", StartedAt: early},
		{RunId: "run-earliest", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent", SandboxId: "sandbox-1", StartedAt: earliest},
	}}
	server := newEventLogsTestServer(t, run)

	stdout, _, err := executeCommand(t, server.URL, false, "logs", "--event", "evt_order_test")
	if err != nil {
		t.Fatalf("logs --event returned error: %v", err)
	}
	earliestPos := strings.Index(stdout, "run-earliest output")
	earlyPos := strings.Index(stdout, "run-early output")
	latePos := strings.Index(stdout, "run-late output")
	if earliestPos == -1 || earlyPos == -1 || latePos == -1 || !(earliestPos < earlyPos && earlyPos < latePos) {
		t.Fatalf("replay order wrong: %q", stdout)
	}
}

func TestExecuteLogsForEventPagesUntilTotal(t *testing.T) {
	// A page smaller than total must advance the offset and stop at total.
	first := make([]*agentcomposev2.RunSummary, 0, 200)
	for i := 0; i < 200; i++ {
		first = append(first, &agentcomposev2.RunSummary{RunId: fmt.Sprintf("run-page1-%03d", i), ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent"})
	}
	run := &pagedEventRunStub{firstPage: first, secondPage: []*agentcomposev2.RunSummary{
		{RunId: "run-page2", ProjectId: "project-aaaaaaaaaaaaaaaa", AgentName: "agent"},
	}}
	server := newEventLogsTestServerForHandler(t, run)

	stdout, _, err := executeCommand(t, server.URL, false, "logs", "--event", "evt_paged_test")
	if err != nil {
		t.Fatalf("logs --event paged returned error: %v", err)
	}
	if !strings.Contains(stdout, "run-page2 output") {
		t.Fatalf("second page missing from output: %q", stdout)
	}
	if len(run.listRequests) != 2 {
		t.Fatalf("ListRuns calls = %d, want 2", len(run.listRequests))
	}
	if run.listRequests[1].GetOffset() != 200 {
		t.Fatalf("second request offset = %d, want 200", run.listRequests[1].GetOffset())
	}
}

// pagedEventRunStub returns a full first page and a partial second page,
// exercising the offset+total paging termination.
type pagedEventRunStub struct {
	eventRunStub
	firstPage  []*agentcomposev2.RunSummary
	secondPage []*agentcomposev2.RunSummary
	page       int
}

func (s *pagedEventRunStub) ListRuns(_ context.Context, req *connect.Request[agentcomposev2.ListRunsRequest]) (*connect.Response[agentcomposev2.ListRunsResponse], error) {
	s.mu.Lock()
	s.listRequests = append(s.listRequests, req.Msg)
	s.page++
	page := s.page
	s.mu.Unlock()
	if req.Msg.GetEventId() == "" {
		return connect.NewResponse(&agentcomposev2.ListRunsResponse{}), nil
	}
	switch page {
	case 1:
		return connect.NewResponse(&agentcomposev2.ListRunsResponse{Runs: s.firstPage, Total: uint32(len(s.firstPage) + len(s.secondPage))}), nil
	default:
		return connect.NewResponse(&agentcomposev2.ListRunsResponse{Runs: s.secondPage, Total: uint32(len(s.firstPage) + len(s.secondPage))}), nil
	}
}

func (s *pagedEventRunStub) FollowRunLogs(_ context.Context, req *connect.Request[agentcomposev2.FollowRunLogsRequest], stream *connect.ServerStream[agentcomposev2.RunLogChunk]) error {
	if err := stream.Send(&agentcomposev2.RunLogChunk{Data: req.Msg.GetRunId() + " output\n", RunStatus: agentcomposev2.RunStatus_RUN_STATUS_SUCCEEDED}); err != nil {
		return err
	}
	return stream.Send(&agentcomposev2.RunLogChunk{RunStatus: agentcomposev2.RunStatus_RUN_STATUS_SUCCEEDED, IsFinal: true})
}

func mustTimestamp(t *testing.T, value string) *timestamppb.Timestamp {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse %s: %v", value, err)
	}
	return timestamppb.New(parsed)
}
