package agentcompose

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/spf13/cobra"
)

type logsOptions struct {
	Agent, Run, Sandbox, Event string
	Follow, Timestamp          bool
	Tail                       int32
}

func newLogsCommand(state *commandState) *cobra.Command {
	options := logsOptions{Tail: -1}
	cmd := &cobra.Command{Use: "logs [agent-or-id]", Short: "Stream Run logs", Args: rangeArgs(0, 1, state), RunE: func(cmd *cobra.Command, args []string) error { return executeLogs(cmd, state, args, options) }}
	cmd.Flags().StringVar(&options.Agent, "agent", "", "Filter by Agent")
	cmd.Flags().StringVar(&options.Run, "run", "", "Select one Run")
	cmd.Flags().StringVar(&options.Sandbox, "sandbox", "", "Filter by Sandbox")
	cmd.Flags().StringVar(&options.Event, "event", "", "Filter by event-bus event id (evt_...)")
	cmd.Flags().BoolVar(&options.Follow, "follow", false, "Continue following logs")
	cmd.Flags().Int32VarP(&options.Tail, "tail", "n", -1, "Number of trailing lines; -1 means all")
	cmd.Flags().BoolVarP(&options.Timestamp, "timestamp", "t", false, "Show timestamps")
	return cmd
}

func executeLogs(cmd *cobra.Command, state *commandState, args []string, options logsOptions) error {
	if options.Tail < -1 {
		return usageError("--tail must be -1 or greater", state.options.JSON)
	}
	if len(args) > 0 && options.Agent != "" {
		return usageError("a positional target and --agent are mutually exclusive", state.options.JSON)
	}
	options.Run = strings.TrimSpace(options.Run)
	options.Sandbox = strings.TrimSpace(options.Sandbox)
	options.Event = strings.TrimSpace(options.Event)
	if options.Run != "" && options.Event != "" {
		return usageError("--run and --event are mutually exclusive", state.options.JSON)
	}
	if options.Sandbox != "" && options.Event != "" {
		return usageError("--sandbox and --event are mutually exclusive", state.options.JSON)
	}
	if len(args) > 0 && options.Event != "" {
		return usageError("a positional target and --event are mutually exclusive", state.options.JSON)
	}
	if options.Event != "" {
		if err := validateEventLogTarget(options.Event, state.options.JSON); err != nil {
			return err
		}
	}
	ctx, cancel := requestContext(cmd, state)
	project, err := resolveProject(ctx, state, state.clients().project)
	if err != nil {
		cancel()
		return err
	}
	projectID := project.GetSummary().GetProjectId()
	if options.Event != "" {
		if options.Agent != "" {
			agent, agentErr := resolveAgent(ctx, project, options.Agent, state)
			if agentErr != nil {
				cancel()
				return agentErr
			}
			options.Agent = agent.GetAgentName()
		}
		// executeLogsForEvent uses cmd.Context() because the request context
		// only bounds the resolution RPCs above, not the log streams.
		cancel()
		return executeLogsForEvent(cmd, state, projectID, options)
	}
	client := state.clients().run
	targets := make([]*agentcomposev2.RunSummary, 0)
	var explicitRun *agentcomposev2.RunSummary
	var selectedSandbox *agentcomposev2.Sandbox
	if options.Run != "" {
		explicitRun, err = resolveRun(ctx, state, client, projectID, options.Run)
		if err != nil {
			cancel()
			return err
		}
	}
	if options.Sandbox != "" {
		sandbox, sandboxErr := resolveSandbox(ctx, state, state.clients().sandbox, projectID, options.Sandbox)
		if sandboxErr != nil {
			cancel()
			return sandboxErr
		}
		selectedSandbox = sandbox
		options.Sandbox = sandbox.GetSandboxId()
	}
	if options.Agent != "" {
		agent, agentErr := resolveAgent(ctx, project, options.Agent, state)
		if agentErr != nil {
			cancel()
			return agentErr
		}
		options.Agent = agent.GetAgentName()
	}
	if len(args) > 0 {
		kind, value, resolveErr := resolveLogsTarget(ctx, state, project, args[0])
		if resolveErr != nil {
			cancel()
			return resolveErr
		}
		switch kind {
		case "agent":
			options.Agent = value
		case "project":
		case "run":
			if options.Run != "" || options.Sandbox != "" {
				cancel()
				return usageError("a Run positional target cannot be combined with --run or --sandbox", state.options.JSON)
			}
			explicitRun, err = resolveRun(ctx, state, client, projectID, value)
			if err != nil {
				cancel()
				return err
			}
		case "sandbox":
			if options.Run != "" || options.Sandbox != "" {
				cancel()
				return usageError("a Sandbox positional target cannot be combined with --run or --sandbox", state.options.JSON)
			}
			selectedSandbox, err = resolveSandbox(ctx, state, state.clients().sandbox, projectID, value)
			if err != nil {
				cancel()
				return err
			}
			options.Sandbox = selectedSandbox.GetSandboxId()
		}
	}
	if explicitRun != nil {
		targets = append(targets, explicitRun)
	}
	if len(targets) == 0 {
		limit := uint32(20)
		runs, _, _, listErr := listRuns(ctx, client, projectID, runListOptions{offsetOptions: offsetOptions{Limit: limit}, Agent: options.Agent, Sandbox: options.Sandbox})
		if listErr != nil {
			cancel()
			return mapConnectError(listErr, state.options.URL, state.options.JSON)
		}
		targets = runs
	}
	if len(targets) == 0 && selectedSandbox != nil {
		err := writeSandboxHistoryLogs(ctx, cmd, state, selectedSandbox.GetSandboxId(), options)
		cancel()
		return err
	}
	cancel()
	for _, run := range targets {
		if err := followOneRun(cmd, state, projectID, run, options); err != nil {
			return err
		}
	}
	return nil
}

func writeSandboxHistoryLogs(ctx context.Context, cmd *cobra.Command, state *commandState, sandboxID string, options logsOptions) error {
	if options.Follow {
		return newError("unsupported", "logs --follow is not supported for Sandbox history", exitUnsupported, state.options.JSON)
	}
	resp, err := state.clients().sandbox.ListSandboxHistory(ctx, connect.NewRequest(&agentcomposev2.ListSandboxHistoryRequest{SandboxId: sandboxID}))
	if err != nil {
		return mapConnectError(err, state.options.URL, state.options.JSON)
	}
	var output strings.Builder
	for _, cell := range resp.Msg.GetCells() {
		text := firstNonEmpty(cell.GetOutput(), cell.GetStdout(), cell.GetStderr())
		if text == "" {
			continue
		}
		output.WriteString(text)
		if !strings.HasSuffix(text, "\n") {
			output.WriteByte('\n')
		}
	}
	text := tailText(output.String(), options.Tail)
	if state.options.JSON {
		events := make([]sandboxHistoryEventDTO, 0, len(resp.Msg.GetEvents()))
		for _, event := range resp.Msg.GetEvents() {
			events = append(events, sandboxHistoryEventFromProto(event))
		}
		return writeJSON(cmd.OutOrStdout(), struct {
			SandboxID string                   `json:"sandbox_id"`
			Output    string                   `json:"output"`
			Events    []sandboxHistoryEventDTO `json:"events,omitempty"`
		}{sandboxID, text, events})
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(), text)
	return err
}

func tailText(value string, lines int32) string {
	if lines < 0 || value == "" {
		return value
	}
	if lines == 0 {
		return ""
	}
	parts := strings.Split(strings.TrimSuffix(value, "\n"), "\n")
	if len(parts) > int(lines) {
		parts = parts[len(parts)-int(lines):]
	}
	return strings.Join(parts, "\n") + "\n"
}

func resolveLogsTarget(ctx context.Context, state *commandState, project *agentcomposev2.Project, ref string) (string, string, error) {
	type match struct{ kind, value string }
	matches := make([]match, 0)
	seen := make(map[match]struct{})
	addMatch := func(kind, value string) {
		item := match{kind, value}
		if _, ok := seen[item]; ok {
			return
		}
		seen[item] = struct{}{}
		matches = append(matches, item)
	}
	summary := project.GetSummary()
	if summary.GetName() == ref || summary.GetProjectId() == ref {
		addMatch("project", summary.GetProjectId())
	}
	for _, agent := range project.GetAgents() {
		if agent.GetAgentName() == ref || agent.GetManagedAgentId() == ref {
			addMatch("agent", agent.GetAgentName())
		}
	}
	runs, _, _, err := listRuns(ctx, state.clients().run, summary.GetProjectId(), runListOptions{offsetOptions: offsetOptions{AllPages: true, Limit: 50}})
	if err != nil {
		return "", "", mapConnectError(err, state.options.URL, state.options.JSON)
	}
	for _, run := range runs {
		if run.GetRunId() == ref {
			addMatch("run", run.GetRunId())
		}
	}
	sandboxes, _, _, err := listSandboxes(ctx, state.clients().sandbox, summary.GetProjectId(), nil, offsetOptions{AllPages: true, Limit: 50})
	if err != nil {
		return "", "", mapConnectError(err, state.options.URL, state.options.JSON)
	}
	for _, sandbox := range sandboxes {
		if sandbox.GetSandboxId() == ref {
			addMatch("sandbox", sandbox.GetSandboxId())
		}
	}
	targets, supported, err := resolveResourceTargets(ctx, state, ref,
		agentcomposev2.ResourceKind_RESOURCE_KIND_PROJECT,
		agentcomposev2.ResourceKind_RESOURCE_KIND_AGENT,
		agentcomposev2.ResourceKind_RESOURCE_KIND_RUN,
		agentcomposev2.ResourceKind_RESOURCE_KIND_SANDBOX,
	)
	if err != nil {
		return "", "", err
	}
	for _, target := range targets {
		if target.GetProjectId() != summary.GetProjectId() && target.GetId() != summary.GetProjectId() {
			continue
		}
		switch target.GetKind() {
		case agentcomposev2.ResourceKind_RESOURCE_KIND_PROJECT:
			if target.GetId() == summary.GetProjectId() {
				addMatch("project", summary.GetProjectId())
			}
		case agentcomposev2.ResourceKind_RESOURCE_KIND_AGENT:
			for _, agent := range project.GetAgents() {
				if target.GetId() == agent.GetManagedAgentId() {
					addMatch("agent", agent.GetAgentName())
				}
			}
		case agentcomposev2.ResourceKind_RESOURCE_KIND_RUN:
			for _, run := range runs {
				if target.GetId() == run.GetRunId() {
					addMatch("run", run.GetRunId())
				}
			}
		case agentcomposev2.ResourceKind_RESOURCE_KIND_SANDBOX:
			for _, sandbox := range sandboxes {
				if target.GetId() == sandbox.GetSandboxId() {
					addMatch("sandbox", sandbox.GetSandboxId())
				}
			}
		}
	}
	if !supported {
		if strings.HasPrefix(summary.GetProjectId(), ref) {
			addMatch("project", summary.GetProjectId())
		}
		for _, agent := range project.GetAgents() {
			if strings.HasPrefix(agent.GetManagedAgentId(), ref) {
				addMatch("agent", agent.GetAgentName())
			}
		}
		for _, run := range runs {
			if run.GetRunShortId() == ref || strings.HasPrefix(run.GetRunId(), ref) {
				addMatch("run", run.GetRunId())
			}
		}
		for _, sandbox := range sandboxes {
			if strings.HasPrefix(sandbox.GetSandboxId(), ref) {
				addMatch("sandbox", sandbox.GetSandboxId())
			}
		}
	}
	if len(matches) == 0 {
		return "", "", newError("not_found", fmt.Sprintf("log target %q was not found", ref), exitNotFound, state.options.JSON)
	}
	if len(matches) > 1 {
		kinds := make([]string, 0, len(matches))
		for _, item := range matches {
			kinds = append(kinds, item.kind)
		}
		return "", "", usageError("log target is ambiguous across "+strings.Join(kinds, ", ")+"; use --agent, --run, --sandbox, or a full ID", state.options.JSON)
	}
	return matches[0].kind, matches[0].value, nil
}

func followOneRun(cmd *cobra.Command, state *commandState, projectID string, run *agentcomposev2.RunSummary, options logsOptions) error {
	ctx, cancel := streamContext(cmd)
	defer cancel()
	request := &agentcomposev2.FollowRunLogsRequest{ProjectId: projectID, RunId: run.GetRunId(), Follow: options.Follow, IncludeMetadata: true}
	if options.Tail >= 0 {
		request.TailSet = true
		request.TailLines = uint32(options.Tail)
	}
	stream, err := state.clients().run.FollowRunLogs(ctx, connect.NewRequest(request))
	if err != nil {
		return mapConnectError(err, state.options.URL, state.options.JSON)
	}
	for stream.Receive() {
		chunk := stream.Msg()
		if state.options.JSON {
			record := map[string]any{"agent_name": run.GetAgentName(), "run_id": run.GetRunId(), "run_short_id": firstNonEmpty(run.GetRunShortId(), shortID(run.GetRunId())), "time": timestampText(chunk.GetCreatedAt()), "prompt": chunk.GetPrompt(), "content": chunk.GetData(), "offset": chunk.GetOffset(), "is_final": chunk.GetIsFinal(), "run_status": enumText(chunk.GetRunStatus(), "RUN_STATUS_")}
			if err := writeJSON(cmd.OutOrStdout(), record); err != nil {
				return err
			}
		} else {
			prefix := ""
			if len(run.GetRunId()) > 0 {
				prefix = firstNonEmpty(run.GetRunShortId(), shortID(run.GetRunId())) + " | "
			}
			if options.Timestamp {
				if created := timestampText(chunk.GetCreatedAt()); created != "" {
					prefix = created + " " + prefix
				}
			}
			for _, line := range strings.SplitAfter(chunk.GetData(), "\n") {
				if line != "" {
					fmt.Fprint(cmd.OutOrStdout(), prefix+line)
				}
			}
		}
	}
	if err := stream.Err(); err != nil {
		return mapConnectError(err, state.options.URL, state.options.JSON)
	}
	return nil
}
