package agentcompose

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/chaitin/agent-compose/proto/agentcompose/v2/agentcomposev2connect"
	"github.com/spf13/cobra"
)

func newInspectCommand(state *commandState) *cobra.Command {
	return &cobra.Command{Use: "inspect [resource-type] <ref>", Short: "Inspect a Project, Agent, Run, Sandbox, Scheduler, trigger, or Scheduler Run", Args: rangeArgs(1, 2, state), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		if len(args) == 2 {
			return inspectTyped(ctx, cmd, state, args[0], args[1])
		}
		return inspectGeneric(ctx, cmd, state, args[0])
	}}
}
func writeInspect(cmd *cobra.Command, state *commandState, value any) error {
	if state.options.JSON {
		return writeJSON(cmd.OutOrStdout(), value)
	}
	return writeYAML(cmd.OutOrStdout(), value)
}
func inspectTyped(ctx context.Context, cmd *cobra.Command, state *commandState, kind, ref string) error {
	kind = strings.ToLower(kind)
	if kind == "project" {
		summary, err := resolveProjectSummary(ctx, state, state.clients().project, ref)
		if err != nil {
			return err
		}
		resp, err := state.clients().project.GetProject(ctx, connect.NewRequest(&agentcomposev2.GetProjectRequest{Project: projectRefID(summary.GetProjectId()), IncludeSpec: true}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		return writeInspect(cmd, state, projectFromProto(resp.Msg.GetProject().GetSummary()))
	}
	project, err := resolveProject(ctx, state, state.clients().project)
	if err != nil {
		return err
	}
	projectID := project.GetSummary().GetProjectId()
	switch kind {
	case "agent":
		agent, err := resolveAgent(ctx, project, ref, state)
		if err != nil {
			return err
		}
		return writeInspect(cmd, state, agentFromProto(agent))
	case "run":
		run, err := resolveRun(ctx, state, state.clients().run, projectID, ref)
		if err != nil {
			return err
		}
		resp, err := state.clients().run.GetRun(ctx, connect.NewRequest(&agentcomposev2.GetRunRequest{ProjectId: projectID, RunId: run.GetRunId()}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		return writeInspect(cmd, state, runFromDetail(resp.Msg.GetRun()))
	case "sandbox":
		sandbox, err := resolveSandbox(ctx, state, state.clients().sandbox, projectID, ref)
		if err != nil {
			return err
		}
		resp, err := state.clients().sandbox.GetSandbox(ctx, connect.NewRequest(&agentcomposev2.GetSandboxRequest{SandboxId: sandbox.GetSandboxId()}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		return writeInspect(cmd, state, sandboxFromProto(resp.Msg.GetSandbox(), nil))
	case "scheduler":
		response, err := schedulerResponse(ctx, state, state.clients().project, project, ref)
		if err != nil {
			return err
		}
		scheduler, _ := schedulerAndTriggers(response)
		return writeInspect(cmd, state, scheduler)
	case "trigger":
		triggers, err := findTriggers(ctx, state, project, ref)
		if err != nil {
			return err
		}
		return writeInspect(cmd, state, triggers[0])
	case "scheduler-run":
		run, err := resolveSchedulerRun(ctx, state, state.clients().project, projectID, ref)
		if err != nil {
			return err
		}
		return writeInspect(cmd, state, schedulerRunFromProto(run))
	default:
		return usageError("unknown resource type; expected project, agent, run, sandbox, scheduler, trigger, or scheduler-run", state.options.JSON)
	}
}
func resolveProjectSummary(ctx context.Context, state *commandState, client agentcomposev2connect.ProjectServiceClient, ref string) (*agentcomposev2.ProjectSummary, error) {
	projects, _, _, _, err := listProjects(ctx, client, offsetOptions{AllPages: true, Limit: 50})
	if err != nil {
		return nil, mapConnectError(err, state.options.URL, state.options.JSON)
	}
	matches := make([]*agentcomposev2.ProjectSummary, 0)
	for _, project := range projects {
		if project.GetName() == ref || project.GetProjectId() == ref {
			matches = append(matches, project)
		}
	}
	if len(matches) == 0 {
		targets, supported, resolveErr := resolveResourceTargets(ctx, state, ref, agentcomposev2.ResourceKind_RESOURCE_KIND_PROJECT)
		if resolveErr != nil {
			return nil, resolveErr
		}
		for _, project := range projects {
			for _, target := range targets {
				if target.GetId() == project.GetProjectId() {
					matches = append(matches, project)
				}
			}
			if !supported && strings.HasPrefix(project.GetProjectId(), ref) {
				matches = append(matches, project)
			}
		}
	}
	if len(matches) == 0 {
		return nil, newError("not_found", fmt.Sprintf("Project %q was not found", ref), exitNotFound, state.options.JSON)
	}
	if len(matches) > 1 {
		return nil, usageError("Project reference is ambiguous", state.options.JSON)
	}
	return matches[0], nil
}
func inspectGeneric(ctx context.Context, cmd *cobra.Command, state *commandState, ref string) error {
	matches := make([]struct {
		kind  string
		value any
	}, 0)
	var matchedProject *agentcomposev2.ProjectSummary
	if projectSummary, err := resolveProjectSummary(ctx, state, state.clients().project, ref); err == nil {
		matchedProject = projectSummary
		matches = append(matches, struct {
			kind  string
			value any
		}{"project", projectFromProto(projectSummary)})
	} else if !isNotFoundError(err) {
		return err
	}
	var project *agentcomposev2.Project
	if strings.TrimSpace(state.options.Project) == "" && matchedProject != nil {
		resp, err := state.clients().project.GetProject(ctx, connect.NewRequest(&agentcomposev2.GetProjectRequest{Project: projectRefID(matchedProject.GetProjectId()), IncludeSpec: true}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		project = resp.Msg.GetProject()
	} else {
		var err error
		project, err = resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
	}
	projectID := project.GetSummary().GetProjectId()
	if agent, err := resolveAgent(ctx, project, ref, state); err == nil {
		matches = append(matches, struct {
			kind  string
			value any
		}{"agent", agentFromProto(agent)})
	} else if !isNotFoundError(err) {
		return err
	}
	if run, err := resolveRun(ctx, state, state.clients().run, projectID, ref); err == nil {
		resp, getErr := state.clients().run.GetRun(ctx, connect.NewRequest(&agentcomposev2.GetRunRequest{ProjectId: projectID, RunId: run.GetRunId()}))
		if getErr != nil {
			return mapConnectError(getErr, state.options.URL, state.options.JSON)
		}
		matches = append(matches, struct {
			kind  string
			value any
		}{"run", runFromDetail(resp.Msg.GetRun())})
	} else if !isNotFoundError(err) {
		return err
	}
	if sandbox, err := resolveSandbox(ctx, state, state.clients().sandbox, projectID, ref); err == nil {
		resp, getErr := state.clients().sandbox.GetSandbox(ctx, connect.NewRequest(&agentcomposev2.GetSandboxRequest{SandboxId: sandbox.GetSandboxId()}))
		if getErr != nil {
			return mapConnectError(getErr, state.options.URL, state.options.JSON)
		}
		matches = append(matches, struct {
			kind  string
			value any
		}{"sandbox", sandboxFromProto(resp.Msg.GetSandbox(), nil)})
	} else if !isNotFoundError(err) {
		return err
	}
	if response, err := schedulerResponse(ctx, state, state.clients().project, project, ref); err == nil {
		scheduler, _ := schedulerAndTriggers(response)
		matches = append(matches, struct {
			kind  string
			value any
		}{"scheduler", scheduler})
	} else if !isNotFoundError(err) {
		return err
	}
	if triggers, err := findTriggers(ctx, state, project, ref); err == nil {
		for _, trigger := range triggers {
			matches = append(matches, struct {
				kind  string
				value any
			}{"trigger", trigger})
		}
	} else if !isNotFoundError(err) {
		return err
	}
	if run, err := resolveSchedulerRun(ctx, state, state.clients().project, projectID, ref); err == nil {
		matches = append(matches, struct {
			kind  string
			value any
		}{"scheduler-run", schedulerRunFromProto(run)})
	} else if !isNotFoundError(err) {
		return err
	}
	if len(matches) == 0 {
		return newError("not_found", fmt.Sprintf("resource %q was not found", ref), exitNotFound, state.options.JSON)
	}
	if len(matches) > 1 {
		kinds := make([]string, 0, len(matches))
		for _, match := range matches {
			kinds = append(kinds, match.kind)
		}
		return usageError("resource reference is ambiguous across "+strings.Join(kinds, ", ")+"; use inspect <resource-type> <ref>", state.options.JSON)
	}
	return writeInspect(cmd, state, matches[0].value)
}

func isNotFoundError(err error) bool {
	var cliErr *CLIError
	return errors.As(err, &cliErr) && cliErr.Kind == "not_found"
}

func findTriggers(ctx context.Context, state *commandState, project *agentcomposev2.Project, ref string) ([]triggerDTO, error) {
	matches := make([]triggerDTO, 0)
	for _, scheduler := range project.GetSchedulers() {
		response, err := schedulerResponse(ctx, state, state.clients().project, project, scheduler.GetSchedulerId())
		if err != nil {
			return nil, err
		}
		_, triggers := schedulerAndTriggers(response)
		for _, trigger := range triggers {
			if trigger.Name == ref || trigger.TriggerID == ref || strings.HasPrefix(trigger.TriggerID, ref) {
				matches = append(matches, trigger)
			}
		}
	}
	if len(matches) == 0 {
		return nil, newError("not_found", fmt.Sprintf("trigger %q was not found", ref), exitNotFound, state.options.JSON)
	}
	if len(matches) > 1 {
		return nil, usageError("trigger reference is ambiguous; use scheduler inspect <ref> --scheduler <scheduler-ref>", state.options.JSON)
	}
	return matches, nil
}
