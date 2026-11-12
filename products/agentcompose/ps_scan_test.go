package agentcompose

import (
	"encoding/json"
	"testing"

	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
)

func psSandbox(id string) *agentcomposev2.Sandbox {
	return &agentcomposev2.Sandbox{SandboxId: id, ProjectId: "project-aaaaaaaaaaaaaaaa", Status: agentcomposev2.SandboxStatus_SANDBOX_STATUS_RUNNING}
}

func psRunIDs(t *testing.T, out string) map[string]string {
	t.Helper()
	var payload struct {
		Sandboxes []sandboxDTO `json:"sandboxes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("output = %s: %v", out, err)
	}
	runIDs := make(map[string]string, len(payload.Sandboxes))
	for _, sandbox := range payload.Sandboxes {
		runIDs[sandbox.SandboxID] = sandbox.RunID
	}
	return runIDs
}

func TestPSLooksUpSchedulerRunBySandbox(t *testing.T) {
	project := &projectStub{project: fixtureProject(), schedulerRuns: []*agentcomposev2.SchedulerRun{
		{RunId: "scheduler-run-1", SandboxIds: []string{"target"}},
		{RunId: "scheduler-run-2", SandboxIds: []string{"other"}},
	}}
	sandbox := &sandboxStub{sandboxes: []*agentcomposev2.Sandbox{psSandbox("target")}}
	server, _ := newTestServer(t, project, &runStub{}, sandbox)
	out, _, err := executeCommand(t, server.URL, false, "--json", "ps")
	if err != nil {
		t.Fatal(err)
	}
	if got := psRunIDs(t, out)["target"]; got != "scheduler-run-1" {
		t.Fatalf("target run_id = %q", got)
	}
	if len(project.batchRunRequests) != 1 {
		t.Fatalf("BatchGetLatestSchedulerRuns calls = %d, want 1", len(project.batchRunRequests))
	}
	request := project.batchRunRequests[0]
	if ids := request.GetSandboxIds(); len(ids) != 1 || ids[0] != "target" {
		t.Fatalf("batch sandbox ids = %v, want [target]", ids)
	}
	if got := request.GetProject().GetProjectId(); got != "project-aaaaaaaaaaaaaaaa" {
		t.Fatalf("batch project = %q", got)
	}
	if len(project.schedulerRunRequests) != 0 {
		t.Fatalf("ListSchedulerRuns calls = %d, want none", len(project.schedulerRunRequests))
	}
}

func TestPSLeavesSandboxWithoutSchedulerRunUnset(t *testing.T) {
	project := &projectStub{project: fixtureProject(), schedulerRuns: []*agentcomposev2.SchedulerRun{
		{RunId: "scheduler-run-1", SandboxIds: []string{"other"}},
	}}
	sandbox := &sandboxStub{sandboxes: []*agentcomposev2.Sandbox{psSandbox("orphan")}}
	server, _ := newTestServer(t, project, &runStub{}, sandbox)
	out, _, err := executeCommand(t, server.URL, false, "--json", "ps")
	if err != nil {
		t.Fatal(err)
	}
	if got := psRunIDs(t, out)["orphan"]; got != "" {
		t.Fatalf("orphan run_id = %q, want none", got)
	}
	if len(project.batchRunRequests) != 1 {
		t.Fatalf("BatchGetLatestSchedulerRuns calls = %d, want 1", len(project.batchRunRequests))
	}
}

func TestPSSkipsSchedulerRunLookupWhenRunsCoverEverySandbox(t *testing.T) {
	project := &projectStub{project: fixtureProject(), schedulerRuns: []*agentcomposev2.SchedulerRun{
		{RunId: "scheduler-run-1", SandboxIds: []string{"covered"}},
	}}
	run := &runStub{runs: []*agentcomposev2.RunSummary{{RunId: "run-1", ProjectId: "project-aaaaaaaaaaaaaaaa", SandboxId: "covered"}}}
	sandbox := &sandboxStub{sandboxes: []*agentcomposev2.Sandbox{psSandbox("covered")}}
	server, _ := newTestServer(t, project, run, sandbox)
	out, _, err := executeCommand(t, server.URL, false, "--json", "ps")
	if err != nil {
		t.Fatal(err)
	}
	if got := psRunIDs(t, out)["covered"]; got != "run-1" {
		t.Fatalf("covered run_id = %q", got)
	}
	if len(project.batchRunRequests) != 0 {
		t.Fatalf("BatchGetLatestSchedulerRuns calls = %d, want 0", len(project.batchRunRequests))
	}
	if len(project.schedulerRunRequests) != 0 {
		t.Fatalf("ListSchedulerRuns calls = %d, want 0", len(project.schedulerRunRequests))
	}
}
