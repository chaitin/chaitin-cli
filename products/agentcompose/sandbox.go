package agentcompose

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/chaitin/agent-compose/proto/agentcompose/v2/agentcomposev2connect"
	"github.com/spf13/cobra"
)

type sandboxDTO struct {
	SandboxID      string `json:"sandbox_id" yaml:"sandbox_id"`
	SandboxShortID string `json:"sandbox_short_id" yaml:"sandbox_short_id"`
	Agent          string `json:"agent" yaml:"agent"`
	Status         string `json:"status" yaml:"status"`
	RunID          string `json:"run_id,omitempty" yaml:"run_id,omitempty"`
	RunShortID     string `json:"run_short_id,omitempty" yaml:"run_short_id,omitempty"`
	Driver         string `json:"driver,omitempty" yaml:"driver,omitempty"`
	Image          string `json:"image,omitempty" yaml:"image,omitempty"`
	CreatedAt      string `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

func sandboxFromProto(sandbox *agentcomposev2.Sandbox, run *agentcomposev2.RunSummary) sandboxDTO {
	result := sandboxDTO{SandboxID: sandbox.GetSandboxId(), SandboxShortID: shortID(sandbox.GetSandboxId()), Agent: sandbox.GetAgentName(), Status: enumText(sandbox.GetStatus(), "SANDBOX_STATUS_"), Driver: sandbox.GetDriver(), Image: sandbox.GetImage()}
	if sandbox.GetCreatedAt() != nil {
		result.CreatedAt = sandbox.GetCreatedAt().AsTime().Format(time.RFC3339)
	}
	if sandbox.GetUpdatedAt() != nil {
		result.UpdatedAt = sandbox.GetUpdatedAt().AsTime().Format(time.RFC3339)
	}
	if run != nil {
		result.RunID = run.GetRunId()
		result.RunShortID = firstNonEmpty(run.GetRunShortId(), shortID(run.GetRunId()))
	}
	return result
}

type sandboxListOptions struct {
	offsetOptions
	All     bool
	Status  string
	Verbose bool
}

func newPSCommand(state *commandState, use string) *cobra.Command {
	options := sandboxListOptions{offsetOptions: offsetOptions{Limit: 50}}
	cmd := &cobra.Command{Use: use, Short: "List project Sandboxes", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error { return executeSandboxList(cmd, state, options) }}
	addOffsetFlags(cmd, &options.offsetOptions)
	cmd.Flags().BoolVarP(&options.All, "all", "a", false, "Include all Sandbox states")
	cmd.Flags().StringVar(&options.Status, "status", "", "Filter by comma-separated status")
	cmd.Flags().BoolVar(&options.Verbose, "verbose", false, "Show full IDs")
	return cmd
}

func executeSandboxList(cmd *cobra.Command, state *commandState, options sandboxListOptions) error {
	if err := validateOffsetOptions(cmd, options.offsetOptions, state); err != nil {
		return err
	}
	statuses := []string{"RUNNING"}
	if options.All {
		statuses = nil
	}
	if options.Status != "" {
		statuses = normalizeSandboxStatuses(strings.Split(options.Status, ","))
		if len(statuses) == 0 {
			return usageError("--status requires at least one Sandbox status", state.options.JSON)
		}
		for _, status := range statuses {
			if !validSandboxStatus(status) {
				return usageError("invalid Sandbox status "+strings.ToLower(status), state.options.JSON)
			}
		}
	}
	ctx, cancel := requestContext(cmd, state)
	defer cancel()
	project, err := resolveProject(ctx, state, state.clients().project)
	if err != nil {
		return err
	}
	statusFilters, err := parseSandboxStatuses(statuses)
	if err != nil {
		return usageError(err.Error(), state.options.JSON)
	}
	sandboxes, more, next, err := listSandboxes(ctx, state.clients().sandbox, project.GetSummary().GetProjectId(), statusFilters, options.offsetOptions)
	if err != nil {
		return mapConnectError(err, state.options.URL, state.options.JSON)
	}
	runs, _, _, err := listRuns(ctx, state.clients().run, project.GetSummary().GetProjectId(), runListOptions{offsetOptions: offsetOptions{AllPages: true, Limit: 50}})
	if err != nil {
		return mapConnectError(err, state.options.URL, state.options.JSON)
	}
	runBySandbox := map[string]*agentcomposev2.RunSummary{}
	for _, run := range runs {
		if run.GetSandboxId() != "" {
			if _, ok := runBySandbox[run.GetSandboxId()]; !ok {
				runBySandbox[run.GetSandboxId()] = run
			}
		}
	}
	needsSchedulerRuns := false
	for _, sandbox := range sandboxes {
		if _, ok := runBySandbox[sandbox.GetSandboxId()]; !ok {
			needsSchedulerRuns = true
			break
		}
	}
	if needsSchedulerRuns {
		schedulerRuns, _, _, schedulerErr := listSchedulerRuns(ctx, state.clients().project, project.GetSummary().GetProjectId(), "", schedulerRunsOptions{offsetOptions: offsetOptions{AllPages: true, Limit: 50}})
		if schedulerErr != nil {
			return mapConnectError(schedulerErr, state.options.URL, state.options.JSON)
		}
		for _, run := range schedulerRuns {
			for _, sandboxID := range run.GetSandboxIds() {
				if _, ok := runBySandbox[sandboxID]; !ok {
					runBySandbox[sandboxID] = &agentcomposev2.RunSummary{RunId: run.GetRunId(), RunShortId: shortID(run.GetRunId()), SandboxId: sandboxID}
				}
			}
		}
	}
	output := make([]sandboxDTO, 0, len(sandboxes))
	for _, sandbox := range sandboxes {
		output = append(output, sandboxFromProto(sandbox, runBySandbox[sandbox.GetSandboxId()]))
	}
	if state.options.JSON {
		return writeJSON(cmd.OutOrStdout(), struct {
			Project   projectDTO   `json:"project"`
			Sandboxes []sandboxDTO `json:"sandboxes"`
			HasMore   bool         `json:"has_more"`
			Next      uint32       `json:"next_offset,omitempty"`
		}{projectFromProto(project.GetSummary()), output, more, next})
	}
	table := newTable(cmd.OutOrStdout(), "SANDBOX\tAGENT\tSTATUS\tRUN\tDRIVER\tCREATED")
	for _, sandbox := range output {
		id := sandbox.SandboxShortID
		if options.Verbose {
			id = sandbox.SandboxID
		}
		fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\n", id, sandbox.Agent, sandbox.Status, sandbox.RunShortID, sandbox.Driver, sandbox.CreatedAt)
	}
	if err := table.Flush(); err != nil {
		return err
	}
	if more {
		fmt.Fprintf(cmd.ErrOrStderr(), "More Sandboxes are available; continue with --offset %d or use --all-pages\n", next)
	}
	return nil
}

func validSandboxStatus(status string) bool {
	switch status {
	case "PENDING", "RUNNING", "STOPPED", "FAILED", "DELETING":
		return true
	default:
		return false
	}
}

func normalizeSandboxStatuses(statuses []string) []string {
	result := make([]string, 0, len(statuses))
	for _, status := range statuses {
		if normalized := strings.ToUpper(strings.TrimSpace(status)); normalized != "" {
			result = append(result, normalized)
		}
	}
	return result
}

func parseSandboxStatuses(statuses []string) ([]agentcomposev2.SandboxStatus, error) {
	if len(statuses) == 0 {
		return nil, nil
	}
	result := make([]agentcomposev2.SandboxStatus, 0, len(statuses))
	for _, status := range statuses {
		parsed, err := parseSandboxStatus(status)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
}

func parseSandboxStatus(raw string) (agentcomposev2.SandboxStatus, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "PENDING":
		return agentcomposev2.SandboxStatus_SANDBOX_STATUS_PENDING, nil
	case "RUNNING":
		return agentcomposev2.SandboxStatus_SANDBOX_STATUS_RUNNING, nil
	case "STOPPED":
		return agentcomposev2.SandboxStatus_SANDBOX_STATUS_STOPPED, nil
	case "FAILED":
		return agentcomposev2.SandboxStatus_SANDBOX_STATUS_FAILED, nil
	case "DELETING":
		return agentcomposev2.SandboxStatus_SANDBOX_STATUS_DELETING, nil
	default:
		return 0, fmt.Errorf("invalid Sandbox status %q", raw)
	}
}

func listSandboxes(ctx context.Context, client agentcomposev2connect.SandboxServiceClient, projectID string, status []agentcomposev2.SandboxStatus, options offsetOptions) ([]*agentcomposev2.Sandbox, bool, uint32, error) {
	offset := options.Offset
	sandboxes := make([]*agentcomposev2.Sandbox, 0)
	statusSet := make(map[agentcomposev2.SandboxStatus]struct{}, len(status))
	for _, value := range status {
		statusSet[value] = struct{}{}
	}
	for {
		pageSize := offsetPageSize(options, len(sandboxes))
		resp, err := client.ListSandboxes(ctx, connect.NewRequest(&agentcomposev2.ListSandboxesRequest{ProjectId: projectID, Status: status, Offset: offset, Limit: pageSize}))
		if err != nil {
			return nil, false, 0, err
		}
		page := make([]*agentcomposev2.Sandbox, 0, len(resp.Msg.GetSandboxes()))
		for _, sandbox := range resp.Msg.GetSandboxes() {
			if sandbox.GetProjectId() != projectID {
				continue
			}
			if len(statusSet) > 0 {
				if _, ok := statusSet[sandbox.GetStatus()]; !ok {
					continue
				}
			}
			page = append(page, sandbox)
		}
		sandboxes = append(sandboxes, page...)
		next := offset + uint32(len(resp.Msg.GetSandboxes()))
		more := next < resp.Msg.GetTotal()
		if !options.AllPages && uint32(len(sandboxes)) >= options.Limit {
			if uint32(len(sandboxes)) > options.Limit {
				sandboxes = sandboxes[:options.Limit]
			}
			if !more {
				next = 0
			}
			return sandboxes, more, next, nil
		}
		if !more || len(resp.Msg.GetSandboxes()) == 0 {
			return sandboxes, false, 0, nil
		}
		offset = next
	}
}

func resolveSandbox(ctx context.Context, state *commandState, client agentcomposev2connect.SandboxServiceClient, projectID, ref string) (*agentcomposev2.Sandbox, error) {
	sandboxes, _, _, err := listSandboxes(ctx, client, projectID, nil, offsetOptions{AllPages: true, Limit: 50})
	if err != nil {
		return nil, mapConnectError(err, state.options.URL, state.options.JSON)
	}
	matches := make([]*agentcomposev2.Sandbox, 0)
	for _, sandbox := range sandboxes {
		if sandbox.GetSandboxId() == ref {
			matches = append(matches, sandbox)
		}
	}
	if len(matches) == 0 {
		targets, supported, resolveErr := resolveResourceTargets(ctx, state, ref, agentcomposev2.ResourceKind_RESOURCE_KIND_SANDBOX)
		if resolveErr != nil {
			return nil, resolveErr
		}
		for _, sandbox := range sandboxes {
			for _, target := range targets {
				if target.GetId() == sandbox.GetSandboxId() && target.GetProjectId() == projectID {
					matches = append(matches, sandbox)
				}
			}
			if !supported && strings.HasPrefix(sandbox.GetSandboxId(), ref) {
				matches = append(matches, sandbox)
			}
		}
	}
	if len(matches) == 0 {
		return nil, newError("not_found", fmt.Sprintf("Sandbox %q was not found", ref), exitNotFound, state.options.JSON)
	}
	if len(matches) > 1 {
		return nil, usageError("Sandbox reference is ambiguous", state.options.JSON)
	}
	return matches[0], nil
}

func newSandboxCommand(state *commandState) *cobra.Command {
	cmd := &cobra.Command{Use: "sandbox", Short: "Manage Sandboxes", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }}
	cmd.AddCommand(newPSCommand(state, "ls"), newSandboxHistoryCommand(state), newSandboxActionCommand(state, "stop"), newSandboxActionCommand(state, "resume"), newSandboxRemoveCommand(state))
	return cmd
}

func newSandboxHistoryCommand(state *commandState) *cobra.Command {
	return &cobra.Command{Use: "history <sandbox-ref>", Short: "Read the complete Sandbox history", Args: exactArgs(1, state), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		sandbox, err := resolveSandbox(ctx, state, state.clients().sandbox, project.GetSummary().GetProjectId(), args[0])
		if err != nil {
			return err
		}
		if !state.options.JSON {
			fmt.Fprintln(cmd.ErrOrStderr(), "Warning: Sandbox history is an unpaginated full response")
		}
		resp, err := state.clients().sandbox.ListSandboxHistory(ctx, connect.NewRequest(&agentcomposev2.ListSandboxHistoryRequest{SandboxId: sandbox.GetSandboxId()}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		cells := make([]sandboxHistoryCellDTO, 0, len(resp.Msg.GetCells()))
		for _, cell := range resp.Msg.GetCells() {
			cells = append(cells, sandboxHistoryCellFromProto(cell))
		}
		events := make([]sandboxHistoryEventDTO, 0, len(resp.Msg.GetEvents()))
		for _, event := range resp.Msg.GetEvents() {
			events = append(events, sandboxHistoryEventFromProto(event))
		}
		output := struct {
			Cells  []sandboxHistoryCellDTO  `json:"cells"`
			Events []sandboxHistoryEventDTO `json:"events"`
			Legacy bool                     `json:"legacy_history"`
		}{cells, events, resp.Msg.GetLegacyHistory()}
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), output)
		}
		return writeYAML(cmd.OutOrStdout(), output)
	}}
}

func newSandboxActionCommand(state *commandState, action string) *cobra.Command {
	return &cobra.Command{Use: action + " <sandbox-ref...>", Short: strings.Title(action) + " one or more Sandboxes", Args: func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return usageError("at least one Sandbox is required", state.options.JSON)
		}
		return nil
	}, RunE: func(cmd *cobra.Command, args []string) error {
		return executeSandboxAction(cmd, state, action, args, false)
	}}
}
func newSandboxRemoveCommand(state *commandState) *cobra.Command {
	force := false
	cmd := &cobra.Command{Use: "rm <sandbox-ref...>", Short: "Remove one or more Sandboxes", Args: func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return usageError("at least one Sandbox is required", state.options.JSON)
		}
		return nil
	}, RunE: func(cmd *cobra.Command, args []string) error {
		return executeSandboxAction(cmd, state, "remove", args, force)
	}}
	cmd.Flags().BoolVar(&force, "force", false, "Force removal of running Sandboxes")
	return cmd
}

func executeSandboxAction(cmd *cobra.Command, state *commandState, action string, refs []string, force bool) error {
	if state.options.DryRun {
		return unsupportedDryRun("sandbox "+action, state.options.JSON)
	}
	ctx, cancel := requestContext(cmd, state)
	defer cancel()
	project, err := resolveProject(ctx, state, state.clients().project)
	if err != nil {
		return err
	}
	completed := make([]string, 0)
	for index, ref := range refs {
		sandbox, resolveErr := resolveSandbox(ctx, state, state.clients().sandbox, project.GetSummary().GetProjectId(), ref)
		if resolveErr == nil {
			switch action {
			case "stop":
				_, resolveErr = state.clients().sandbox.StopSandbox(ctx, connect.NewRequest(&agentcomposev2.StopSandboxRequest{SandboxId: sandbox.GetSandboxId()}))
			case "resume":
				_, resolveErr = state.clients().sandbox.ResumeSandbox(ctx, connect.NewRequest(&agentcomposev2.ResumeSandboxRequest{SandboxId: sandbox.GetSandboxId()}))
			case "remove":
				_, resolveErr = state.clients().sandbox.RemoveSandbox(ctx, connect.NewRequest(&agentcomposev2.RemoveSandboxRequest{SandboxId: sandbox.GetSandboxId(), Force: force}))
			}
		}
		if resolveErr != nil {
			mapped := mapConnectError(resolveErr, state.options.URL, state.options.JSON)
			partial := newError("partial_failure", fmt.Sprintf("sandbox %s failed for %q: %v", action, ref, mapped), exitGeneral, state.options.JSON)
			partial.Operation = action
			partial.FailedTarget = ref
			partial.Completed = completed
			partial.Unattempted = append([]string(nil), refs[index+1:]...)
			return partial
		}
		completed = append(completed, ref)
		if !state.options.JSON {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", strings.Title(action), ref)
		}
	}
	if state.options.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]any{"operation": action, "completed_targets": completed})
	}
	return nil
}
