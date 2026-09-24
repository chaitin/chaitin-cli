package agentcompose

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/chaitin/agent-compose/proto/agentcompose/v2/agentcomposev2connect"
	"github.com/spf13/cobra"
)

type triggerDTO struct {
	SchedulerID      string `json:"scheduler_id" yaml:"scheduler_id"`
	SchedulerShortID string `json:"scheduler_short_id" yaml:"scheduler_short_id"`
	AgentName        string `json:"agent_name" yaml:"agent_name"`
	TriggerID        string `json:"trigger_id" yaml:"trigger_id"`
	TriggerShortID   string `json:"trigger_short_id" yaml:"trigger_short_id"`
	Name             string `json:"name" yaml:"name"`
	Kind             string `json:"trigger_kind" yaml:"trigger_kind"`
	Source           string `json:"trigger_source" yaml:"trigger_source"`
	Enabled          bool   `json:"enabled" yaml:"enabled"`
	Cron             string `json:"cron,omitempty" yaml:"cron,omitempty"`
	Interval         string `json:"interval,omitempty" yaml:"interval,omitempty"`
	Timeout          string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}
type schedulerDTO struct {
	SchedulerID      string       `json:"scheduler_id" yaml:"scheduler_id"`
	SchedulerShortID string       `json:"scheduler_short_id" yaml:"scheduler_short_id"`
	AgentName        string       `json:"agent_name" yaml:"agent_name"`
	Enabled          bool         `json:"enabled" yaml:"enabled"`
	TriggerCount     uint32       `json:"trigger_count" yaml:"trigger_count"`
	Triggers         []triggerDTO `json:"triggers" yaml:"triggers"`
}

func schedulerAndTriggers(response *agentcomposev2.GetSchedulerResponse) (schedulerDTO, []triggerDTO) {
	scheduler := response.GetScheduler()
	result := schedulerDTO{SchedulerID: scheduler.GetSchedulerId(), SchedulerShortID: shortID(scheduler.GetSchedulerId()), AgentName: scheduler.GetAgentName(), Enabled: scheduler.GetEnabled(), TriggerCount: scheduler.GetTriggerCount(), Triggers: make([]triggerDTO, 0, len(response.GetTriggers()))}
	for _, trigger := range response.GetTriggers() {
		spec := trigger.GetSpec()
		item := triggerDTO{SchedulerID: result.SchedulerID, SchedulerShortID: result.SchedulerShortID, AgentName: result.AgentName, TriggerID: trigger.GetTriggerId(), TriggerShortID: shortID(trigger.GetTriggerId()), Name: spec.GetName(), Kind: enumText(spec.GetKind(), "TRIGGER_KIND_"), Source: "declarative", Enabled: trigger.GetEnabled(), Cron: spec.GetCron(), Interval: spec.GetInterval(), Timeout: spec.GetTimeout()}
		result.Triggers = append(result.Triggers, item)
	}
	return result, result.Triggers
}

type schedulerRunDTO struct {
	ID               string   `json:"run_id" yaml:"run_id"`
	ShortID          string   `json:"run_short_id" yaml:"run_short_id"`
	ProjectID        string   `json:"project_id" yaml:"project_id"`
	AgentName        string   `json:"agent_name" yaml:"agent_name"`
	SchedulerID      string   `json:"scheduler_id" yaml:"scheduler_id"`
	SchedulerShortID string   `json:"scheduler_short_id" yaml:"scheduler_short_id"`
	TriggerID        string   `json:"trigger_id" yaml:"trigger_id"`
	TriggerShortID   string   `json:"trigger_short_id" yaml:"trigger_short_id"`
	TriggerKind      string   `json:"trigger_kind" yaml:"trigger_kind"`
	TriggerSource    string   `json:"trigger_source" yaml:"trigger_source"`
	Status           string   `json:"status" yaml:"status"`
	StartedAt        string   `json:"started_at,omitempty" yaml:"started_at,omitempty"`
	CompletedAt      string   `json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
	DurationMS       int64    `json:"duration_ms,omitempty" yaml:"duration_ms,omitempty"`
	Error            string   `json:"error,omitempty" yaml:"error,omitempty"`
	ResultJSON       string   `json:"result_json,omitempty" yaml:"result_json,omitempty"`
	PayloadJSON      string   `json:"payload_json,omitempty" yaml:"payload_json,omitempty"`
	SandboxIDs       []string `json:"sandbox_ids,omitempty" yaml:"sandbox_ids,omitempty"`
}

func schedulerRunFromProto(run *agentcomposev2.SchedulerRun) schedulerRunDTO {
	result := schedulerRunDTO{ID: run.GetRunId(), ShortID: shortID(run.GetRunId()), ProjectID: run.GetProjectId(), AgentName: run.GetAgentName(), SchedulerID: run.GetSchedulerId(), SchedulerShortID: shortID(run.GetSchedulerId()), TriggerID: run.GetTriggerId(), TriggerShortID: shortID(run.GetTriggerId()), TriggerKind: enumText(run.GetTriggerKind(), "TRIGGER_KIND_"), TriggerSource: run.GetTriggerSource(), Status: enumText(run.GetStatus(), "SCHEDULER_RUN_STATUS_"), DurationMS: run.GetDurationMs(), Error: run.GetError(), ResultJSON: run.GetResultJson(), PayloadJSON: run.GetPayloadJson(), SandboxIDs: append([]string(nil), run.GetSandboxIds()...)}
	if run.GetStartedAt() != nil {
		result.StartedAt = run.GetStartedAt().AsTime().Format(time.RFC3339)
	}
	if run.GetCompletedAt() != nil {
		result.CompletedAt = run.GetCompletedAt().AsTime().Format(time.RFC3339)
	}
	return result
}

func newSchedulerCommand(state *commandState) *cobra.Command {
	cmd := &cobra.Command{Use: "scheduler", Short: "Query and operate Schedulers", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }}
	cmd.AddCommand(newSchedulerListCommand(state), newSchedulerInspectCommand(state), newSchedulerRunsCommand(state), newSchedulerLogsCommand(state), newSchedulerInvokeCommand(state), newSchedulerTriggerCommand(state), newSchedulerStopCommand(state))
	return cmd
}

func schedulerResponse(ctx context.Context, state *commandState, client agentcomposev2connect.ProjectServiceClient, project *agentcomposev2.Project, ref string) (*agentcomposev2.GetSchedulerResponse, error) {
	matches := make([]*agentcomposev2.ProjectScheduler, 0)
	for _, scheduler := range project.GetSchedulers() {
		if scheduler.GetAgentName() == ref || scheduler.GetSchedulerId() == ref || strings.HasPrefix(scheduler.GetSchedulerId(), ref) {
			matches = append(matches, scheduler)
		}
	}
	if len(matches) == 0 {
		return nil, newError("not_found", fmt.Sprintf("Scheduler %q was not found", ref), exitNotFound, state.options.JSON)
	}
	if len(matches) > 1 {
		return nil, usageError("Scheduler reference is ambiguous", state.options.JSON)
	}
	resp, err := client.GetScheduler(ctx, connect.NewRequest(&agentcomposev2.GetSchedulerRequest{Project: projectRefID(project.GetSummary().GetProjectId()), AgentName: matches[0].GetAgentName()}))
	if err != nil {
		return nil, mapConnectError(err, state.options.URL, state.options.JSON)
	}
	return resp.Msg, nil
}
func resolveTrigger(response *agentcomposev2.GetSchedulerResponse, ref string, state *commandState) (*agentcomposev2.ResolvedTrigger, error) {
	matches := make([]*agentcomposev2.ResolvedTrigger, 0)
	for _, trigger := range response.GetTriggers() {
		if trigger.GetSpec().GetName() == ref || trigger.GetTriggerId() == ref || strings.HasPrefix(trigger.GetTriggerId(), ref) {
			matches = append(matches, trigger)
		}
	}
	if len(matches) == 0 {
		return nil, newError("not_found", fmt.Sprintf("trigger %q was not found", ref), exitNotFound, state.options.JSON)
	}
	if len(matches) > 1 {
		return nil, usageError("trigger reference is ambiguous", state.options.JSON)
	}
	return matches[0], nil
}

func newSchedulerListCommand(state *commandState) *cobra.Command {
	verbose := false
	cmd := &cobra.Command{Use: "ls [agent-ref]", Short: "List all current Scheduler triggers", Args: rangeArgs(0, 1, state), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		targets := project.GetSchedulers()
		if len(args) == 1 {
			response, err := schedulerResponse(ctx, state, state.clients().project, project, args[0])
			if err != nil {
				return err
			}
			targets = []*agentcomposev2.ProjectScheduler{response.GetScheduler()}
		}
		triggers := make([]triggerDTO, 0)
		for _, target := range targets {
			response, err := schedulerResponse(ctx, state, state.clients().project, project, target.GetSchedulerId())
			if err != nil {
				return err
			}
			_, items := schedulerAndTriggers(response)
			triggers = append(triggers, items...)
		}
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), struct {
				Project  projectDTO   `json:"project"`
				Triggers []triggerDTO `json:"triggers"`
			}{projectFromProto(project.GetSummary()), triggers})
		}
		table := newTable(cmd.OutOrStdout(), "SCHEDULER\tAGENT\tTRIGGER\tKIND\tSOURCE\tENABLED")
		for _, trigger := range triggers {
			id := trigger.SchedulerShortID
			if verbose {
				id = trigger.SchedulerID
			}
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%t\n", id, trigger.AgentName, trigger.Name, trigger.Kind, trigger.Source, trigger.Enabled)
		}
		return table.Flush()
	}}
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show full IDs")
	return cmd
}

func newSchedulerInspectCommand(state *commandState) *cobra.Command {
	scope := ""
	cmd := &cobra.Command{Use: "inspect <ref>", Short: "Inspect a Scheduler, trigger, or Scheduler Run", Args: exactArgs(1, state), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		if scope != "" {
			response, err := schedulerResponse(ctx, state, state.clients().project, project, scope)
			if err != nil {
				return err
			}
			trigger, err := resolveTrigger(response, args[0], state)
			if err != nil {
				return err
			}
			scheduler, triggers := schedulerAndTriggers(response)
			for _, item := range triggers {
				if item.TriggerID == trigger.GetTriggerId() {
					if state.options.JSON {
						return writeJSON(cmd.OutOrStdout(), item)
					}
					return writeYAML(cmd.OutOrStdout(), item)
				}
			}
			return writeYAML(cmd.OutOrStdout(), scheduler)
		}
		matches := make([]struct {
			kind  string
			value any
		}, 0, 3)
		if response, schedulerErr := schedulerResponse(ctx, state, state.clients().project, project, args[0]); schedulerErr == nil {
			scheduler, _ := schedulerAndTriggers(response)
			matches = append(matches, struct {
				kind  string
				value any
			}{"scheduler", scheduler})
		} else if !isNotFoundError(schedulerErr) {
			return schedulerErr
		}
		if triggers, triggerErr := findTriggers(ctx, state, project, args[0]); triggerErr == nil {
			matches = append(matches, struct {
				kind  string
				value any
			}{"trigger", triggers[0]})
		} else if !isNotFoundError(triggerErr) {
			return triggerErr
		}
		run, runErr := resolveSchedulerRun(ctx, state, state.clients().project, project.GetSummary().GetProjectId(), args[0])
		if runErr == nil {
			matches = append(matches, struct {
				kind  string
				value any
			}{"scheduler-run", schedulerRunFromProto(run)})
		} else if !isNotFoundError(runErr) {
			return runErr
		}
		if len(matches) == 0 {
			return newError("not_found", fmt.Sprintf("Scheduler resource %q was not found; use --scheduler for trigger lookup", args[0]), exitNotFound, state.options.JSON)
		}
		if len(matches) > 1 {
			kinds := make([]string, 0, len(matches))
			for _, match := range matches {
				kinds = append(kinds, match.kind)
			}
			return usageError("Scheduler resource reference is ambiguous across "+strings.Join(kinds, ", ")+"; use --scheduler or a full ID", state.options.JSON)
		}
		return writeInspect(cmd, state, matches[0].value)
	}}
	cmd.Flags().StringVar(&scope, "scheduler", "", "Limit trigger lookup to a Scheduler")
	return cmd
}

type schedulerRunsOptions struct {
	offsetOptions
	Trigger, Status string
}

func newSchedulerRunsCommand(state *commandState) *cobra.Command {
	options := schedulerRunsOptions{offsetOptions: offsetOptions{Limit: 50}}
	cmd := &cobra.Command{Use: "runs [scheduler-ref]", Short: "List Scheduler Runs", Args: rangeArgs(0, 1, state), RunE: func(cmd *cobra.Command, args []string) error {
		resolvedOptions := options
		if err := validateOffsetOptions(cmd, resolvedOptions.offsetOptions, state); err != nil {
			return err
		}
		if _, err := parseSchedulerStatus(resolvedOptions.Status); err != nil {
			return usageError(err.Error(), state.options.JSON)
		}
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		agentName := ""
		if len(args) == 1 {
			response, err := schedulerResponse(ctx, state, state.clients().project, project, args[0])
			if err != nil {
				return err
			}
			agentName = response.GetScheduler().GetAgentName()
			resolvedOptions.Trigger, err = normalizeSchedulerTrigger(ctx, state, project, response, resolvedOptions.Trigger)
			if err != nil {
				return err
			}
		} else if resolvedOptions.Trigger != "" {
			resolvedOptions.Trigger, err = normalizeSchedulerTrigger(ctx, state, project, nil, resolvedOptions.Trigger)
			if err != nil {
				return err
			}
		}
		runs, more, next, err := listSchedulerRuns(ctx, state.clients().project, project.GetSummary().GetProjectId(), agentName, resolvedOptions)
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		output := make([]schedulerRunDTO, 0, len(runs))
		for _, run := range runs {
			output = append(output, schedulerRunFromProto(run))
		}
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), struct {
				Project projectDTO        `json:"project"`
				Runs    []schedulerRunDTO `json:"runs"`
				HasMore bool              `json:"has_more"`
				Next    uint32            `json:"next_offset,omitempty"`
			}{projectFromProto(project.GetSummary()), output, more, next})
		}
		table := newTable(cmd.OutOrStdout(), "RUN\tAGENT\tTRIGGER\tSTATUS\tSTARTED")
		for _, run := range output {
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\n", run.ShortID, run.AgentName, run.TriggerShortID, run.Status, run.StartedAt)
		}
		return table.Flush()
	}}
	addOffsetFlags(cmd, &options.offsetOptions)
	cmd.Flags().StringVar(&options.Trigger, "trigger", "", "Filter by trigger")
	cmd.Flags().StringVar(&options.Status, "status", "", "Filter by status")
	return cmd
}
func parseSchedulerStatus(raw string) (agentcomposev2.SchedulerRunStatus, error) {
	switch strings.ToLower(raw) {
	case "":
		return agentcomposev2.SchedulerRunStatus_SCHEDULER_RUN_STATUS_UNSPECIFIED, nil
	case "running":
		return agentcomposev2.SchedulerRunStatus_SCHEDULER_RUN_STATUS_RUNNING, nil
	case "succeeded":
		return agentcomposev2.SchedulerRunStatus_SCHEDULER_RUN_STATUS_SUCCEEDED, nil
	case "failed":
		return agentcomposev2.SchedulerRunStatus_SCHEDULER_RUN_STATUS_FAILED, nil
	case "canceled":
		return agentcomposev2.SchedulerRunStatus_SCHEDULER_RUN_STATUS_CANCELED, nil
	case "skipped":
		return agentcomposev2.SchedulerRunStatus_SCHEDULER_RUN_STATUS_SKIPPED, nil
	default:
		return 0, fmt.Errorf("invalid Scheduler Run status %q", raw)
	}
}
func listSchedulerRuns(ctx context.Context, client agentcomposev2connect.ProjectServiceClient, projectID, agentName string, options schedulerRunsOptions) ([]*agentcomposev2.SchedulerRun, bool, uint32, error) {
	status, err := parseSchedulerStatus(options.Status)
	if err != nil {
		return nil, false, 0, err
	}
	offset := options.Offset
	runs := make([]*agentcomposev2.SchedulerRun, 0)
	for {
		pageSize := offsetPageSize(options.offsetOptions, len(runs))
		resp, err := client.ListSchedulerRuns(ctx, connect.NewRequest(&agentcomposev2.ListSchedulerRunsRequest{Project: projectRefID(projectID), AgentName: agentName, TriggerId: options.Trigger, Status: status, Limit: pageSize, Offset: offset}))
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
func resolveSchedulerRun(ctx context.Context, state *commandState, client agentcomposev2connect.ProjectServiceClient, projectID, ref string) (*agentcomposev2.SchedulerRun, error) {
	runs, _, _, err := listSchedulerRuns(ctx, client, projectID, "", schedulerRunsOptions{offsetOptions: offsetOptions{AllPages: true, Limit: 50}})
	if err != nil {
		return nil, mapConnectError(err, state.options.URL, state.options.JSON)
	}
	matches := make([]*agentcomposev2.SchedulerRun, 0)
	for _, run := range runs {
		if run.GetProjectId() == projectID && (run.GetRunId() == ref || strings.HasPrefix(run.GetRunId(), ref)) {
			matches = append(matches, run)
		}
	}
	if len(matches) == 0 {
		return nil, newError("not_found", fmt.Sprintf("Scheduler Run %q was not found", ref), exitNotFound, state.options.JSON)
	}
	if len(matches) > 1 {
		return nil, usageError("Scheduler Run reference is ambiguous", state.options.JSON)
	}
	resp, err := client.GetSchedulerRun(ctx, connect.NewRequest(&agentcomposev2.GetSchedulerRunRequest{Project: projectRefID(projectID), RunId: matches[0].GetRunId()}))
	if err != nil {
		return nil, mapConnectError(err, state.options.URL, state.options.JSON)
	}
	return resp.Msg.GetRun(), nil
}

func normalizeSchedulerTrigger(ctx context.Context, state *commandState, project *agentcomposev2.Project, response *agentcomposev2.GetSchedulerResponse, ref string) (string, error) {
	if ref == "" {
		return "", nil
	}
	if response != nil {
		trigger, err := resolveTrigger(response, ref, state)
		if err == nil {
			return trigger.GetTriggerId(), nil
		}
		if !isNotFoundError(err) {
			return "", err
		}
	} else {
		triggers, err := findTriggers(ctx, state, project, ref)
		if err == nil {
			return triggers[0].TriggerID, nil
		}
		if !isNotFoundError(err) {
			return "", err
		}
	}
	agentName := ""
	if response != nil {
		agentName = response.GetScheduler().GetAgentName()
	}
	runs, _, _, err := listSchedulerRuns(ctx, state.clients().project, project.GetSummary().GetProjectId(), agentName, schedulerRunsOptions{offsetOptions: offsetOptions{AllPages: true, Limit: 50}})
	if err != nil {
		return "", mapConnectError(err, state.options.URL, state.options.JSON)
	}
	for _, run := range runs {
		if run.GetTriggerId() == ref {
			return ref, nil
		}
	}
	return "", newError("not_found", fmt.Sprintf("trigger %q was not found", ref), exitNotFound, state.options.JSON)
}

type schedulerLogsOptions struct {
	offsetOptions
	Scheduler, Trigger, Run string
	Tail                    int32
}

func newSchedulerLogsCommand(state *commandState) *cobra.Command {
	options := schedulerLogsOptions{offsetOptions: offsetOptions{Limit: 50}, Tail: -1}
	cmd := &cobra.Command{Use: "logs [scheduler-run-ref]", Short: "List Scheduler events", Args: rangeArgs(0, 1, state), RunE: func(cmd *cobra.Command, args []string) error {
		resolvedOptions := options
		if err := validateOffsetOptions(cmd, resolvedOptions.offsetOptions, state); err != nil {
			return err
		}
		if resolvedOptions.Tail < -1 {
			return usageError("--tail must be -1 or greater", state.options.JSON)
		}
		if resolvedOptions.Tail >= 0 && resolvedOptions.AllPages {
			return usageError("--tail and --all-pages are mutually exclusive", state.options.JSON)
		}
		if len(args) == 1 && resolvedOptions.Run != "" {
			return usageError("a positional Scheduler Run and --run are mutually exclusive", state.options.JSON)
		}
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		if len(args) == 1 {
			run, err := resolveSchedulerRun(ctx, state, state.clients().project, project.GetSummary().GetProjectId(), args[0])
			if err != nil {
				return err
			}
			resolvedOptions.Run = run.GetRunId()
		} else if resolvedOptions.Run != "" {
			run, err := resolveSchedulerRun(ctx, state, state.clients().project, project.GetSummary().GetProjectId(), resolvedOptions.Run)
			if err != nil {
				return err
			}
			resolvedOptions.Run = run.GetRunId()
		}
		agentName := ""
		if resolvedOptions.Scheduler != "" {
			response, err := schedulerResponse(ctx, state, state.clients().project, project, resolvedOptions.Scheduler)
			if err != nil {
				return err
			}
			agentName = response.GetScheduler().GetAgentName()
			resolvedOptions.Trigger, err = normalizeSchedulerTrigger(ctx, state, project, response, resolvedOptions.Trigger)
			if err != nil {
				return err
			}
		} else if resolvedOptions.Trigger != "" {
			resolvedOptions.Trigger, err = normalizeSchedulerTrigger(ctx, state, project, nil, resolvedOptions.Trigger)
			if err != nil {
				return err
			}
		}
		if resolvedOptions.Tail == 0 {
			return writeSchedulerEvents(cmd, state, project, nil, false, 0)
		}
		if resolvedOptions.Tail > 0 {
			tailLimit := uint32(resolvedOptions.Tail)
			if !cmd.Flags().Changed("limit") || tailLimit < resolvedOptions.Limit {
				resolvedOptions.Limit = tailLimit
			}
		}
		events, more, next, err := listSchedulerEvents(ctx, state.clients().project, project.GetSummary().GetProjectId(), agentName, resolvedOptions)
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		slices.Reverse(events)
		return writeSchedulerEvents(cmd, state, project, events, more, next)
	}}
	addOffsetFlags(cmd, &options.offsetOptions)
	cmd.Flags().StringVar(&options.Scheduler, "scheduler", "", "Limit to Scheduler")
	cmd.Flags().StringVar(&options.Trigger, "trigger", "", "Limit to trigger")
	cmd.Flags().StringVar(&options.Run, "run", "", "Limit to Scheduler Run")
	cmd.Flags().Int32VarP(&options.Tail, "tail", "n", -1, "Show trailing events")
	return cmd
}

func writeSchedulerEvents(cmd *cobra.Command, state *commandState, project *agentcomposev2.Project, events []*agentcomposev2.SchedulerEvent, more bool, next uint32) error {
	if state.options.JSON {
		output := make([]schedulerEventDTO, 0, len(events))
		for _, event := range events {
			output = append(output, schedulerEventFromProto(event))
		}
		return writeJSON(cmd.OutOrStdout(), map[string]any{"project": projectFromProto(project.GetSummary()), "events": output, "has_more": more, "next_offset": next})
	}
	for _, event := range events {
		created := ""
		if event.GetCreatedAt() != nil {
			created = event.GetCreatedAt().AsTime().Format(time.RFC3339)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s %s\n", created, strings.ToUpper(event.GetLevel()), event.GetType(), event.GetMessage())
	}
	return nil
}

func listSchedulerEvents(ctx context.Context, client agentcomposev2connect.ProjectServiceClient, projectID, agentName string, options schedulerLogsOptions) ([]*agentcomposev2.SchedulerEvent, bool, uint32, error) {
	offset := options.Offset
	events := make([]*agentcomposev2.SchedulerEvent, 0)
	for {
		pageSize := offsetPageSize(options.offsetOptions, len(events))
		resp, err := client.ListProjectSchedulerEvents(ctx, connect.NewRequest(&agentcomposev2.ListProjectSchedulerEventsRequest{Project: projectRefID(projectID), AgentName: agentName, TriggerId: options.Trigger, RunId: options.Run, Limit: pageSize, Offset: offset}))
		if err != nil {
			return nil, false, 0, err
		}
		page := resp.Msg.GetEvents()
		events = append(events, page...)
		next := offset + uint32(len(page))
		more := next < resp.Msg.GetTotal()
		if !options.AllPages && uint32(len(events)) >= options.Limit {
			if uint32(len(events)) > options.Limit {
				events = events[:options.Limit]
			}
			if !more {
				next = 0
			}
			return events, more, next, nil
		}
		if !more || len(page) == 0 {
			return events, false, 0, nil
		}
		offset = next
	}
}

func newSchedulerInvokeCommand(state *commandState) *cobra.Command {
	payload := ""
	cmd := &cobra.Command{Use: "invoke <scheduler-ref>", Short: "Invoke a Scheduler script", Args: exactArgs(1, state), RunE: func(cmd *cobra.Command, args []string) error {
		if state.options.DryRun {
			return unsupportedDryRun("scheduler invoke", state.options.JSON)
		}
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		response, err := schedulerResponse(ctx, state, state.clients().project, project, args[0])
		if err != nil {
			return err
		}
		resp, err := state.clients().project.InvokeScheduler(ctx, connect.NewRequest(&agentcomposev2.InvokeSchedulerRequest{Project: projectRefID(project.GetSummary().GetProjectId()), AgentName: response.GetScheduler().GetAgentName(), PayloadJson: payload}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		output := map[string]any{"result_json": resp.Msg.GetResultJson(), "duration_ms": resp.Msg.GetDurationMs(), "warnings": resp.Msg.GetWarnings()}
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), output)
		}
		return writeYAML(cmd.OutOrStdout(), output)
	}}
	cmd.Flags().StringVar(&payload, "payload", "", "JSON payload")
	return cmd
}
func newSchedulerTriggerCommand(state *commandState) *cobra.Command {
	payload := ""
	detach := false
	cmd := &cobra.Command{Use: "trigger <scheduler-ref> <trigger-ref>", Short: "Run a Scheduler trigger", Args: exactArgs(2, state), RunE: func(cmd *cobra.Command, args []string) error {
		if state.options.DryRun {
			return unsupportedDryRun("scheduler trigger", state.options.JSON)
		}
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		response, err := schedulerResponse(ctx, state, state.clients().project, project, args[0])
		if err != nil {
			return err
		}
		trigger, err := resolveTrigger(response, args[1], state)
		if err != nil {
			return err
		}
		projectRef := projectRefID(project.GetSummary().GetProjectId())
		var run *agentcomposev2.SchedulerRun
		if detach {
			resp, err := state.clients().project.StartSchedulerRun(ctx, connect.NewRequest(&agentcomposev2.StartSchedulerRunRequest{Project: projectRef, AgentName: response.GetScheduler().GetAgentName(), TriggerId: trigger.GetTriggerId(), PayloadJson: payload}))
			if err != nil {
				return mapConnectError(err, state.options.URL, state.options.JSON)
			}
			run = resp.Msg.GetRun()
		} else {
			resp, err := state.clients().project.RunScheduler(ctx, connect.NewRequest(&agentcomposev2.RunSchedulerRequest{Project: projectRef, AgentName: response.GetScheduler().GetAgentName(), TriggerId: trigger.GetTriggerId(), PayloadJson: payload}))
			if err != nil {
				return mapConnectError(err, state.options.URL, state.options.JSON)
			}
			run = resp.Msg.GetRun()
		}
		output := schedulerRunFromProto(run)
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), output)
		}
		return writeYAML(cmd.OutOrStdout(), output)
	}}
	cmd.Flags().StringVar(&payload, "payload", "", "JSON payload")
	cmd.Flags().BoolVarP(&detach, "detach", "d", false, "Start and return immediately")
	return cmd
}
func newSchedulerStopCommand(state *commandState) *cobra.Command {
	reason := ""
	cmd := &cobra.Command{Use: "stop <scheduler-run-ref>", Short: "Stop a Scheduler Run", Args: exactArgs(1, state), RunE: func(cmd *cobra.Command, args []string) error {
		if state.options.DryRun {
			return unsupportedDryRun("scheduler stop", state.options.JSON)
		}
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		run, err := resolveSchedulerRun(ctx, state, state.clients().project, project.GetSummary().GetProjectId(), args[0])
		if err != nil {
			return err
		}
		resp, err := state.clients().project.StopSchedulerRun(ctx, connect.NewRequest(&agentcomposev2.StopSchedulerRunRequest{Project: projectRefID(project.GetSummary().GetProjectId()), RunId: run.GetRunId(), Reason: reason}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		output := schedulerRunFromProto(resp.Msg.GetRun())
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), output)
		}
		return writeYAML(cmd.OutOrStdout(), output)
	}}
	cmd.Flags().StringVar(&reason, "reason", "", "Stop reason")
	return cmd
}
