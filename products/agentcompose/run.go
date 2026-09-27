package agentcompose

import (
	"context"
	"fmt"
	"io"
	"strings"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/chaitin/agent-compose/proto/agentcompose/v2/agentcomposev2connect"
	"github.com/spf13/cobra"
)

type runDTO struct {
	ID             string `json:"id" yaml:"id"`
	ShortID        string `json:"short_id" yaml:"short_id"`
	ProjectID      string `json:"project_id" yaml:"project_id"`
	ProjectName    string `json:"project_name" yaml:"project_name"`
	AgentName      string `json:"agent_name" yaml:"agent_name"`
	Source         string `json:"source" yaml:"source"`
	Status         string `json:"status" yaml:"status"`
	SandboxID      string `json:"sandbox_id,omitempty" yaml:"sandbox_id,omitempty"`
	SandboxShortID string `json:"sandbox_short_id,omitempty" yaml:"sandbox_short_id,omitempty"`
	ExitCode       int32  `json:"exit_code" yaml:"exit_code"`
	Error          string `json:"error,omitempty" yaml:"error,omitempty"`
	StartedAt      string `json:"started_at,omitempty" yaml:"started_at,omitempty"`
	CompletedAt    string `json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
	DurationMS     int64  `json:"duration_ms,omitempty" yaml:"duration_ms,omitempty"`
	Prompt         string `json:"prompt,omitempty" yaml:"prompt,omitempty"`
	Output         string `json:"output,omitempty" yaml:"output,omitempty"`
	ResultJSON     string `json:"result_json,omitempty" yaml:"result_json,omitempty"`
	Driver         string `json:"driver,omitempty" yaml:"driver,omitempty"`
	Image          string `json:"image_ref,omitempty" yaml:"image_ref,omitempty"`
}

func runFromSummary(run *agentcomposev2.RunSummary) runDTO {
	return runDTO{ID: run.GetRunId(), ShortID: firstNonEmpty(run.GetRunShortId(), shortID(run.GetRunId())), ProjectID: run.GetProjectId(), ProjectName: run.GetProjectName(), AgentName: run.GetAgentName(), Source: enumText(run.GetSource(), "RUN_SOURCE_"), Status: enumText(run.GetStatus(), "RUN_STATUS_"), SandboxID: run.GetSandboxId(), SandboxShortID: firstNonEmpty(run.GetSandboxShortId(), shortID(run.GetSandboxId())), ExitCode: run.GetExitCode(), Error: run.GetError(), StartedAt: timestampText(run.GetStartedAt()), CompletedAt: timestampText(run.GetCompletedAt()), DurationMS: run.GetDurationMs()}
}
func runFromDetail(detail *agentcomposev2.RunDetail) runDTO {
	result := runFromSummary(detail.GetSummary())
	result.Prompt = detail.GetPrompt()
	result.Output = detail.GetOutput()
	result.ResultJSON = detail.GetResultJson()
	result.Driver = detail.GetDriver()
	result.Image = detail.GetImageRef()
	return result
}

type runStartOptions struct {
	Prompt, Command, Sandbox, Driver string
	KeepRunning, Remove, Detach      bool
}

func addRunStartFlags(cmd *cobra.Command, options *runStartOptions) {
	cmd.Flags().StringVar(&options.Prompt, "prompt", "", "Prompt text")
	cmd.Flags().StringVar(&options.Command, "command", "", "Command text")
	cmd.Flags().StringVar(&options.Sandbox, "sandbox", "", "Existing Sandbox ID")
	cmd.Flags().StringVar(&options.Driver, "driver", "", "Runtime driver override")
	cmd.Flags().BoolVar(&options.KeepRunning, "keep-running", false, "Keep Sandbox running after completion")
	cmd.Flags().BoolVar(&options.Remove, "rm", false, "Remove Sandbox after completion")
	cmd.Flags().BoolVarP(&options.Detach, "detach", "d", false, "Start and return immediately")
}

func newRunCommand(state *commandState) *cobra.Command {
	parentOptions := runStartOptions{}
	cmd := &cobra.Command{Use: "run [agent-ref]", Short: "Run an Agent or query Runs", Args: rangeArgs(0, 1, state), RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return executeRun(cmd, state, args[0], parentOptions)
	}}
	addRunStartFlags(cmd, &parentOptions)
	startOptions := runStartOptions{}
	start := &cobra.Command{Use: "start <agent-ref>", Short: "Run an Agent with an explicit disambiguation entry", Args: exactArgs(1, state), RunE: func(cmd *cobra.Command, args []string) error { return executeRun(cmd, state, args[0], startOptions) }}
	addRunStartFlags(start, &startOptions)
	cmd.AddCommand(start, newRunListCommand(state), newRunEventsCommand(state), newRunStopCommand(state))
	return cmd
}

func executeRun(cmd *cobra.Command, state *commandState, agentRef string, options runStartOptions) error {
	if state.options.DryRun {
		return unsupportedDryRun("run", state.options.JSON)
	}
	if options.KeepRunning && options.Remove {
		return usageError("--keep-running and --rm are mutually exclusive", state.options.JSON)
	}
	ctx, cancel := requestContext(cmd, state)
	project, err := resolveProject(ctx, state, state.clients().project)
	cancel()
	if err != nil {
		return err
	}
	agent, err := resolveAgent(ctx, project, agentRef, state)
	if err != nil {
		return err
	}
	if options.Sandbox != "" {
		ctx, cancel := requestContext(cmd, state)
		sandbox, sandboxErr := resolveSandbox(ctx, state, state.clients().sandbox, project.GetSummary().GetProjectId(), options.Sandbox)
		cancel()
		if sandboxErr != nil {
			return sandboxErr
		}
		options.Sandbox = sandbox.GetSandboxId()
	}
	policy := agentcomposev2.RunSandboxCleanupPolicy_RUN_SANDBOX_CLEANUP_POLICY_STOP_ON_COMPLETION
	if options.KeepRunning {
		policy = agentcomposev2.RunSandboxCleanupPolicy_RUN_SANDBOX_CLEANUP_POLICY_KEEP_RUNNING
	}
	if options.Remove {
		policy = agentcomposev2.RunSandboxCleanupPolicy_RUN_SANDBOX_CLEANUP_POLICY_REMOVE_ON_COMPLETION
	}
	req := &agentcomposev2.RunAgentRequest{ProjectId: project.GetSummary().GetProjectId(), AgentName: agent.GetAgentName(), Prompt: options.Prompt, Command: options.Command, SandboxId: options.Sandbox, Driver: options.Driver, Source: agentcomposev2.RunSource_RUN_SOURCE_MANUAL, CleanupPolicy: policy}
	client := state.clients().run
	if options.Detach {
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		resp, err := client.StartAgentRun(ctx, connect.NewRequest(&agentcomposev2.StartAgentRunRequest{Run: req}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		output := runFromSummary(resp.Msg.GetRun())
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), output)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Run: %s\nSandbox: %s\nStatus: %s\n", output.ID, output.SandboxID, output.Status)
		return nil
	}
	ctx, cancel = streamContext(cmd)
	defer cancel()
	stream, err := client.StreamAgentRun(ctx, connect.NewRequest(req))
	if err != nil {
		return mapConnectError(err, state.options.URL, state.options.JSON)
	}
	for stream.Receive() {
		event := stream.Msg()
		if state.options.JSON {
			record := map[string]any{"event_type": enumText(event.GetEventType(), "STREAM_AGENT_RUN_EVENT_TYPE_"), "run_id": event.GetRunId(), "chunk": event.GetChunk(), "stream": enumText(event.GetStream(), "STDIO_STREAM_"), "created_at": timestampText(event.GetCreatedAt()), "warnings": event.GetWarnings()}
			if event.GetRun() != nil {
				addRunFields(record, runFromSummary(event.GetRun()))
			}
			if err := writeJSON(cmd.OutOrStdout(), record); err != nil {
				return err
			}
		} else {
			target := cmd.OutOrStdout()
			if event.GetStream() == agentcomposev2.StdioStream_STDIO_STREAM_STDERR {
				target = cmd.ErrOrStderr()
			}
			if _, err := io.WriteString(target, event.GetChunk()); err != nil {
				return err
			}
		}
	}
	if err := stream.Err(); err != nil {
		return mapConnectError(err, state.options.URL, state.options.JSON)
	}
	return nil
}

func addRunFields(record map[string]any, run runDTO) {
	record["id"] = run.ID
	record["short_id"] = run.ShortID
	record["project_id"] = run.ProjectID
	record["project_name"] = run.ProjectName
	record["agent_name"] = run.AgentName
	record["source"] = run.Source
	record["status"] = run.Status
	record["exit_code"] = run.ExitCode
	if run.SandboxID != "" {
		record["sandbox_id"] = run.SandboxID
	}
	if run.SandboxShortID != "" {
		record["sandbox_short_id"] = run.SandboxShortID
	}
	if run.Error != "" {
		record["error"] = run.Error
	}
	if run.StartedAt != "" {
		record["started_at"] = run.StartedAt
	}
	if run.CompletedAt != "" {
		record["completed_at"] = run.CompletedAt
	}
	if run.DurationMS != 0 {
		record["duration_ms"] = run.DurationMS
	}
}

func resolveAgent(ctx context.Context, project *agentcomposev2.Project, ref string, state *commandState) (*agentcomposev2.ProjectAgent, error) {
	matches := make([]*agentcomposev2.ProjectAgent, 0)
	for _, agent := range project.GetAgents() {
		if agent.GetAgentName() == ref || agent.GetManagedAgentId() == ref {
			matches = append(matches, agent)
		}
	}
	if len(matches) == 0 {
		targets, supported, resolveErr := resolveResourceTargets(ctx, state, ref, agentcomposev2.ResourceKind_RESOURCE_KIND_AGENT)
		if resolveErr != nil {
			return nil, resolveErr
		}
		projectID := project.GetSummary().GetProjectId()
		for _, agent := range project.GetAgents() {
			for _, target := range targets {
				if target.GetId() == agent.GetManagedAgentId() && target.GetProjectId() == projectID {
					matches = append(matches, agent)
				}
			}
			if !supported && strings.HasPrefix(agent.GetManagedAgentId(), ref) {
				matches = append(matches, agent)
			}
		}
	}
	if len(matches) == 0 {
		return nil, newError("not_found", fmt.Sprintf("agent %q was not found in project %s", ref, project.GetSummary().GetName()), exitNotFound, state.options.JSON)
	}
	if len(matches) > 1 {
		return nil, usageError("agent reference is ambiguous", state.options.JSON)
	}
	return matches[0], nil
}

type runListOptions struct {
	offsetOptions
	Agent, Scheduler, Status, Source, StartedFrom, StartedTo, Sandbox string
}

func newRunListCommand(state *commandState) *cobra.Command {
	options := runListOptions{offsetOptions: offsetOptions{Limit: 50}}
	cmd := &cobra.Command{Use: "ls", Short: "List Runs", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error {
		if err := validateOffsetOptions(cmd, options.offsetOptions, state); err != nil {
			return err
		}
		if _, err := parseRunStatus(options.Status); err != nil {
			return usageError(err.Error(), state.options.JSON)
		}
		if _, err := parseRunSource(options.Source); err != nil {
			return usageError(err.Error(), state.options.JSON)
		}
		if _, err := parseOptionalTimestamp(options.StartedFrom); err != nil {
			return usageError("invalid --started-from timestamp", state.options.JSON)
		}
		if _, err := parseOptionalTimestamp(options.StartedTo); err != nil {
			return usageError("invalid --started-to timestamp", state.options.JSON)
		}
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		runs, hasMore, next, err := listRuns(ctx, state.clients().run, project.GetSummary().GetProjectId(), options)
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		output := make([]runDTO, 0, len(runs))
		for _, run := range runs {
			output = append(output, runFromSummary(run))
		}
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), struct {
				Project projectDTO `json:"project"`
				Runs    []runDTO   `json:"runs"`
				HasMore bool       `json:"has_more"`
				Next    uint32     `json:"next_offset,omitempty"`
			}{projectFromProto(project.GetSummary()), output, hasMore, next})
		}
		table := newTable(cmd.OutOrStdout(), "RUN\tAGENT\tSTATUS\tSOURCE\tSANDBOX\tSTARTED")
		for _, run := range output {
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\n", run.ShortID, run.AgentName, run.Status, run.Source, run.SandboxShortID, run.StartedAt)
		}
		if err := table.Flush(); err != nil {
			return err
		}
		if hasMore {
			fmt.Fprintf(cmd.ErrOrStderr(), "More Runs are available; continue with --offset %d or use --all-pages\n", next)
		}
		return nil
	}}
	addOffsetFlags(cmd, &options.offsetOptions)
	cmd.Flags().StringVar(&options.Agent, "agent", "", "Filter by Agent name")
	cmd.Flags().StringVar(&options.Scheduler, "scheduler", "", "Filter by Scheduler ID")
	cmd.Flags().StringVar(&options.Status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&options.Source, "source", "", "Filter by source")
	cmd.Flags().StringVar(&options.StartedFrom, "started-from", "", "Filter by start time lower bound")
	cmd.Flags().StringVar(&options.StartedTo, "started-to", "", "Filter by start time upper bound")
	cmd.Flags().StringVar(&options.Sandbox, "sandbox", "", "Filter by Sandbox ID")
	return cmd
}

func listRuns(ctx context.Context, client agentcomposev2connect.RunServiceClient, projectID string, options runListOptions) ([]*agentcomposev2.RunSummary, bool, uint32, error) {
	status, err := parseRunStatus(options.Status)
	if err != nil {
		return nil, false, 0, err
	}
	source, err := parseRunSource(options.Source)
	if err != nil {
		return nil, false, 0, err
	}
	startedFrom, err := parseOptionalTimestamp(options.StartedFrom)
	if err != nil {
		return nil, false, 0, err
	}
	startedTo, err := parseOptionalTimestamp(options.StartedTo)
	if err != nil {
		return nil, false, 0, err
	}
	offset := options.Offset
	runs := make([]*agentcomposev2.RunSummary, 0)
	for {
		pageSize := offsetPageSize(options.offsetOptions, len(runs))
		resp, err := client.ListRuns(ctx, connect.NewRequest(&agentcomposev2.ListRunsRequest{ProjectId: projectID, AgentName: options.Agent, SchedulerId: options.Scheduler, Status: status, Source: source, StartedFrom: startedFrom, StartedTo: startedTo, SandboxId: options.Sandbox, Offset: offset, Limit: pageSize}))
		if err != nil {
			return nil, false, 0, err
		}
		page := resp.Msg.GetRuns()
		runs = append(runs, page...)
		next := offset + uint32(len(page))
		more := next < resp.Msg.GetTotal()
		if !options.AllPages && uint32(len(runs)) >= options.Limit {
			if uint32(len(runs)) > options.Limit {
				runs = runs[:options.Limit]
			}
			if !more {
				next = 0
			}
			return runs, more, next, nil
		}
		if !more || len(page) == 0 {
			return runs, false, 0, nil
		}
		offset = next
	}
}

func parseRunStatus(raw string) (agentcomposev2.RunStatus, error) {
	switch strings.ToLower(raw) {
	case "":
		return agentcomposev2.RunStatus_RUN_STATUS_UNSPECIFIED, nil
	case "pending":
		return agentcomposev2.RunStatus_RUN_STATUS_PENDING, nil
	case "running":
		return agentcomposev2.RunStatus_RUN_STATUS_RUNNING, nil
	case "succeeded":
		return agentcomposev2.RunStatus_RUN_STATUS_SUCCEEDED, nil
	case "failed":
		return agentcomposev2.RunStatus_RUN_STATUS_FAILED, nil
	case "canceled":
		return agentcomposev2.RunStatus_RUN_STATUS_CANCELED, nil
	default:
		return 0, fmt.Errorf("invalid Run status %q", raw)
	}
}
func parseRunSource(raw string) (agentcomposev2.RunSource, error) {
	switch strings.ToLower(raw) {
	case "":
		return agentcomposev2.RunSource_RUN_SOURCE_UNSPECIFIED, nil
	case "manual":
		return agentcomposev2.RunSource_RUN_SOURCE_MANUAL, nil
	case "scheduler":
		return agentcomposev2.RunSource_RUN_SOURCE_SCHEDULER, nil
	case "api":
		return agentcomposev2.RunSource_RUN_SOURCE_API, nil
	default:
		return 0, fmt.Errorf("invalid Run source %q", raw)
	}
}

func resolveRun(ctx context.Context, state *commandState, client agentcomposev2connect.RunServiceClient, projectID, ref string) (*agentcomposev2.RunSummary, error) {
	runs, _, _, err := listRuns(ctx, client, projectID, runListOptions{offsetOptions: offsetOptions{AllPages: true, Limit: 50}})
	if err != nil {
		return nil, mapConnectError(err, state.options.URL, state.options.JSON)
	}
	matches := make([]*agentcomposev2.RunSummary, 0)
	for _, run := range runs {
		if run.GetRunId() == ref {
			matches = append(matches, run)
		}
	}
	if len(matches) == 0 {
		targets, supported, resolveErr := resolveResourceTargets(ctx, state, ref, agentcomposev2.ResourceKind_RESOURCE_KIND_RUN)
		if resolveErr != nil {
			return nil, resolveErr
		}
		for _, run := range runs {
			for _, target := range targets {
				if target.GetId() == run.GetRunId() && target.GetProjectId() == projectID {
					matches = append(matches, run)
				}
			}
			if !supported && (run.GetRunShortId() == ref || strings.HasPrefix(run.GetRunId(), ref)) {
				matches = append(matches, run)
			}
		}
	}
	if len(matches) == 0 {
		return nil, newError("not_found", fmt.Sprintf("Run %q was not found", ref), exitNotFound, state.options.JSON)
	}
	if len(matches) > 1 {
		return nil, usageError("Run reference is ambiguous", state.options.JSON)
	}
	return matches[0], nil
}

func newRunEventsCommand(state *commandState) *cobra.Command {
	options := offsetOptions{Limit: 50}
	cmd := &cobra.Command{Use: "events <run-ref>", Short: "List Run events", Args: exactArgs(1, state), RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateOffsetOptions(cmd, options, state); err != nil {
			return err
		}
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		run, err := resolveRun(ctx, state, state.clients().run, project.GetSummary().GetProjectId(), args[0])
		if err != nil {
			return err
		}
		events, history, more, next, err := listRunEvents(ctx, state.clients().run, run.GetRunId(), options)
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		if state.options.JSON {
			output := make([]runEventDTO, 0, len(events))
			for _, event := range events {
				output = append(output, runEventFromProto(event))
			}
			return writeJSON(cmd.OutOrStdout(), map[string]any{"events": output, "history_available": history, "has_more": more, "next_offset": next})
		}
		for _, event := range events {
			fmt.Fprintf(cmd.OutOrStdout(), "%d\t%s\t%s\n", event.GetSeq(), enumText(event.GetKind(), "RUN_EVENT_KIND_"), event.GetText())
		}
		return nil
	}}
	addOffsetFlags(cmd, &options)
	return cmd
}

func listRunEvents(ctx context.Context, client agentcomposev2connect.RunServiceClient, runID string, options offsetOptions) ([]*agentcomposev2.RunEvent, bool, bool, uint32, error) {
	events := make([]*agentcomposev2.RunEvent, 0)
	offset := options.Offset
	history := false
	for {
		pageSize := offsetPageSize(options, len(events))
		resp, err := client.ListRunEvents(ctx, connect.NewRequest(&agentcomposev2.ListRunEventsRequest{RunId: runID, Limit: pageSize, Offset: offset}))
		if err != nil {
			return nil, false, false, 0, err
		}
		page := resp.Msg.GetEvents()
		events = append(events, page...)
		history = history || resp.Msg.GetHistoryAvailable()
		next := offset + uint32(len(page))
		more := next < resp.Msg.GetTotal()
		if !options.AllPages && uint32(len(events)) >= options.Limit {
			if uint32(len(events)) > options.Limit {
				events = events[:options.Limit]
			}
			if !more {
				next = 0
			}
			return events, history, more, next, nil
		}
		if !more || len(page) == 0 {
			return events, history, false, 0, nil
		}
		offset = next
	}
}

func newRunStopCommand(state *commandState) *cobra.Command {
	reason := ""
	cmd := &cobra.Command{Use: "stop <run-ref>", Short: "Stop a Run", Args: exactArgs(1, state), RunE: func(cmd *cobra.Command, args []string) error {
		if state.options.DryRun {
			return unsupportedDryRun("run stop", state.options.JSON)
		}
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		run, err := resolveRun(ctx, state, state.clients().run, project.GetSummary().GetProjectId(), args[0])
		if err != nil {
			return err
		}
		resp, err := state.clients().run.StopRun(ctx, connect.NewRequest(&agentcomposev2.StopRunRequest{RunId: run.GetRunId(), Reason: reason}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		output := runFromDetail(resp.Msg.GetRun())
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), output)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Run %s stop requested\n", output.ID)
		return nil
	}}
	cmd.Flags().StringVar(&reason, "reason", "", "Stop reason")
	return cmd
}
