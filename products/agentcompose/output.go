package agentcompose

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

func timestampText(value *timestamppb.Timestamp) string {
	if value == nil {
		return ""
	}
	return value.AsTime().Format("2006-01-02T15:04:05.999999999Z07:00")
}

func parseOptionalTimestamp(raw string) (*timestamppb.Timestamp, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, raw)
	}
	if err != nil {
		return nil, err
	}
	return timestamppb.New(parsed), nil
}

type runEventDTO struct {
	ID          string `json:"id"`
	RunID       string `json:"run_id"`
	Seq         uint64 `json:"seq"`
	Kind        string `json:"kind"`
	Text        string `json:"text,omitempty"`
	Agent       string `json:"agent,omitempty"`
	Name        string `json:"name,omitempty"`
	PayloadJSON string `json:"payload_json,omitempty"`
	Success     bool   `json:"success"`
	ExitCode    int32  `json:"exit_code"`
	StopReason  string `json:"stop_reason,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

func runEventFromProto(event *agentcomposev2.RunEvent) runEventDTO {
	return runEventDTO{ID: event.GetId(), RunID: event.GetRunId(), Seq: event.GetSeq(), Kind: enumText(event.GetKind(), "RUN_EVENT_KIND_"), Text: event.GetText(), Agent: event.GetAgent(), Name: event.GetName(), PayloadJSON: event.GetPayloadJson(), Success: event.GetSuccess(), ExitCode: event.GetExitCode(), StopReason: event.GetStopReason(), CreatedAt: timestampText(event.GetCreatedAt())}
}

type schedulerEventDTO struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Level       string `json:"level"`
	Message     string `json:"message,omitempty"`
	PayloadJSON string `json:"payload_json,omitempty"`
	RunID       string `json:"run_id,omitempty"`
	TriggerID   string `json:"trigger_id,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	AgentName   string `json:"agent_name,omitempty"`
	SchedulerID string `json:"scheduler_id,omitempty"`
	SandboxID   string `json:"sandbox_id,omitempty"`
}

func schedulerEventFromProto(event *agentcomposev2.SchedulerEvent) schedulerEventDTO {
	return schedulerEventDTO{ID: event.GetId(), Type: event.GetType(), Level: event.GetLevel(), Message: event.GetMessage(), PayloadJSON: event.GetPayloadJson(), RunID: event.GetRunId(), TriggerID: event.GetTriggerId(), CreatedAt: timestampText(event.GetCreatedAt()), AgentName: event.GetAgentName(), SchedulerID: event.GetSchedulerId(), SandboxID: event.GetLinkedSandboxId()}
}

type sandboxHistoryCellDTO struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Source        string `json:"source,omitempty"`
	Stdout        string `json:"stdout,omitempty"`
	Stderr        string `json:"stderr,omitempty"`
	Output        string `json:"output,omitempty"`
	ExitCode      int32  `json:"exit_code"`
	Success       bool   `json:"success"`
	Running       bool   `json:"running"`
	CreatedAt     string `json:"created_at,omitempty"`
	Agent         string `json:"agent,omitempty"`
	AgentThreadID string `json:"agent_thread_id,omitempty"`
	StopReason    string `json:"stop_reason,omitempty"`
}

func sandboxHistoryCellFromProto(cell *agentcomposev2.SandboxHistoryCell) sandboxHistoryCellDTO {
	return sandboxHistoryCellDTO{ID: cell.GetId(), Type: cell.GetType(), Source: cell.GetSource(), Stdout: cell.GetStdout(), Stderr: cell.GetStderr(), Output: cell.GetOutput(), ExitCode: cell.GetExitCode(), Success: cell.GetSuccess(), Running: cell.GetRunning(), CreatedAt: timestampText(cell.GetCreatedAt()), Agent: cell.GetAgent(), AgentThreadID: cell.GetAgentThreadId(), StopReason: cell.GetStopReason()}
}

type sandboxHistoryEventDTO struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Level     string `json:"level"`
	Message   string `json:"message,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

func sandboxHistoryEventFromProto(event *agentcomposev2.SandboxHistoryEvent) sandboxHistoryEventDTO {
	return sandboxHistoryEventDTO{ID: event.GetId(), Type: event.GetType(), Level: event.GetLevel(), Message: event.GetMessage(), CreatedAt: timestampText(event.GetCreatedAt())}
}

func writeJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

func writeYAML(out io.Writer, value any) error {
	data, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}

func shortID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func enumText(value fmt.Stringer, prefixes ...string) string {
	text := value.String()
	for _, prefix := range prefixes {
		text = strings.TrimPrefix(text, prefix)
	}
	return strings.ToLower(text)
}

type projectDTO struct {
	ID              string `json:"id" yaml:"id"`
	ShortID         string `json:"short_id" yaml:"short_id"`
	Name            string `json:"name" yaml:"name"`
	CurrentRevision uint64 `json:"current_revision" yaml:"current_revision"`
	SpecHash        string `json:"spec_hash" yaml:"spec_hash"`
	AgentCount      uint32 `json:"agent_count" yaml:"agent_count"`
	SchedulerCount  uint32 `json:"scheduler_count" yaml:"scheduler_count"`
	RunningRunCount uint32 `json:"running_run_count,omitempty" yaml:"running_run_count,omitempty"`
	CreatedAt       string `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

func projectFromProto(project *agentcomposev2.ProjectSummary) projectDTO {
	return projectDTO{ID: project.GetProjectId(), ShortID: shortID(project.GetProjectId()), Name: project.GetName(), CurrentRevision: project.GetCurrentRevision(), SpecHash: project.GetSpecHash(), AgentCount: project.GetAgentCount(), SchedulerCount: project.GetSchedulerCount(), RunningRunCount: project.GetRunningRunCount(), CreatedAt: timestampText(project.GetCreatedAt()), UpdatedAt: timestampText(project.GetUpdatedAt())}
}

type agentDTO struct {
	ID               string `json:"id" yaml:"id"`
	ShortID          string `json:"short_id" yaml:"short_id"`
	Name             string `json:"name" yaml:"name"`
	Provider         string `json:"provider,omitempty" yaml:"provider,omitempty"`
	Model            string `json:"model,omitempty" yaml:"model,omitempty"`
	Image            string `json:"image,omitempty" yaml:"image,omitempty"`
	Driver           string `json:"driver,omitempty" yaml:"driver,omitempty"`
	SchedulerEnabled bool   `json:"scheduler_enabled" yaml:"scheduler_enabled"`
	Enabled          bool   `json:"enabled" yaml:"enabled"`
	Availability     string `json:"availability" yaml:"availability"`
	Health           string `json:"health" yaml:"health"`
}

func agentFromProto(agent *agentcomposev2.ProjectAgent) agentDTO {
	return agentDTO{ID: agent.GetManagedAgentId(), ShortID: shortID(agent.GetManagedAgentId()), Name: agent.GetAgentName(), Provider: agent.GetProvider(), Model: agent.GetModel(), Image: agent.GetImage(), Driver: agent.GetDriver(), SchedulerEnabled: agent.GetSchedulerEnabled(), Enabled: agent.GetEnabled(), Availability: enumText(agent.GetAvailability(), "PROJECT_AGENT_AVAILABILITY_"), Health: enumText(agent.GetHealth(), "PROJECT_AGENT_HEALTH_")}
}

func newTable(out io.Writer, header string) *tabwriter.Writer {
	table := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, header)
	return table
}
