package agentcompose

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/chaitin/agent-compose/proto/agentcompose/v2/agentcomposev2connect"
	"github.com/spf13/cobra"
)

type offsetOptions struct {
	Limit    uint32
	Offset   uint32
	AllPages bool
}

func addOffsetFlags(cmd *cobra.Command, options *offsetOptions) {
	cmd.Flags().Uint32Var(&options.Limit, "limit", 50, "Maximum total results")
	cmd.Flags().Uint32Var(&options.Offset, "offset", 0, "Result offset")
	cmd.Flags().BoolVar(&options.AllPages, "all-pages", false, "Read all remaining pages")
}

func validateOffsetOptions(cmd *cobra.Command, options offsetOptions, state *commandState) error {
	if options.AllPages && cmd.Flags().Changed("limit") {
		return usageError("--limit and --all-pages are mutually exclusive", state.options.JSON)
	}
	if !options.AllPages && options.Limit == 0 {
		return usageError("--limit must be greater than zero", state.options.JSON)
	}
	return nil
}

func newProjectCommand(state *commandState) *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Query projects", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }}
	options := offsetOptions{Limit: 50}
	verbose := false
	list := &cobra.Command{Use: "ls", Short: "List projects", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error {
		if err := validateOffsetOptions(cmd, options, state); err != nil {
			return err
		}
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		projects, total, hasMore, next, err := listProjects(ctx, state.clients().project, options)
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		output := make([]projectDTO, 0, len(projects))
		for _, project := range projects {
			output = append(output, projectFromProto(project))
		}
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), struct {
				Projects []projectDTO `json:"projects"`
				Total    uint32       `json:"total_count"`
				HasMore  bool         `json:"has_more"`
				Next     uint32       `json:"next_offset,omitempty"`
			}{output, total, hasMore, next})
		}
		table := newTable(cmd.OutOrStdout(), "ID\tNAME\tREVISION\tAGENTS\tSCHEDULERS")
		for _, project := range output {
			id := project.ShortID
			if verbose {
				id = project.ID
			}
			fmt.Fprintf(table, "%s\t%s\t%d\t%d\t%d\n", id, project.Name, project.CurrentRevision, project.AgentCount, project.SchedulerCount)
		}
		if err := table.Flush(); err != nil {
			return err
		}
		if hasMore {
			fmt.Fprintf(cmd.ErrOrStderr(), "More projects are available; continue with --offset %d or use --all-pages\n", next)
		}
		return nil
	}}
	addOffsetFlags(list, &options)
	list.Flags().BoolVar(&verbose, "verbose", false, "Show full IDs")
	cmd.AddCommand(list)
	return cmd
}

func listProjects(ctx context.Context, client agentcomposev2connect.ProjectServiceClient, options offsetOptions) ([]*agentcomposev2.ProjectSummary, uint32, bool, uint32, error) {
	offset := options.Offset
	projects := make([]*agentcomposev2.ProjectSummary, 0)
	var total uint32
	for {
		pageSize := offsetPageSize(options, len(projects))
		resp, err := client.ListProjects(ctx, connect.NewRequest(&agentcomposev2.ListProjectsRequest{Offset: offset, Limit: pageSize}))
		if err != nil {
			return nil, 0, false, 0, err
		}
		total = resp.Msg.GetTotal()
		page := resp.Msg.GetProjects()
		projects = append(projects, page...)
		next := offset + uint32(len(page))
		more := next < total
		if !options.AllPages && uint32(len(projects)) >= options.Limit {
			if uint32(len(projects)) > options.Limit {
				projects = projects[:options.Limit]
			}
			if !more {
				next = 0
			}
			return projects, total, more, next, nil
		}
		if !more || len(page) == 0 {
			return projects, total, false, 0, nil
		}
		offset = next
	}
}

func offsetPageSize(options offsetOptions, collected int) uint32 {
	if options.AllPages {
		return 100
	}
	remaining := options.Limit - uint32(collected)
	if remaining > 100 {
		return 100
	}
	return remaining
}

func projectRefID(id string) *agentcomposev2.ProjectRef {
	return &agentcomposev2.ProjectRef{Selector: &agentcomposev2.ProjectRef_ProjectId{ProjectId: id}}
}

func resolveProject(ctx context.Context, state *commandState, client agentcomposev2connect.ProjectServiceClient) (*agentcomposev2.Project, error) {
	ref := strings.TrimSpace(state.options.Project)
	if ref == "" {
		return nil, usageError("project is required; use --project or AGENT_COMPOSE_DEFAULT_PROJECT", state.options.JSON)
	}
	projects, _, _, _, err := listProjects(ctx, client, offsetOptions{AllPages: true, Limit: 50})
	if err != nil {
		return nil, mapConnectError(err, state.options.URL, state.options.JSON)
	}
	matches := make([]*agentcomposev2.ProjectSummary, 0)
	for _, project := range projects {
		if project.GetProjectId() == ref || project.GetName() == ref {
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
		return nil, newError("not_found", fmt.Sprintf("project %q was not found", ref), exitNotFound, state.options.JSON)
	}
	if len(matches) > 1 {
		names := make([]string, 0, len(matches))
		for _, match := range matches {
			names = append(names, match.GetName()+" ("+shortID(match.GetProjectId())+")")
		}
		return nil, usageError("project reference is ambiguous: "+strings.Join(names, ", "), state.options.JSON)
	}
	resp, err := client.GetProject(ctx, connect.NewRequest(&agentcomposev2.GetProjectRequest{Project: projectRefID(matches[0].GetProjectId()), IncludeSpec: true}))
	if err != nil {
		return nil, mapConnectError(err, state.options.URL, state.options.JSON)
	}
	return resp.Msg.GetProject(), nil
}

func newAgentCommand(state *commandState) *cobra.Command {
	cmd := &cobra.Command{Use: "agent", Short: "Query project agents", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }}
	list := &cobra.Command{Use: "ls", Short: "List project agents", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		agents := make([]agentDTO, 0, len(project.GetAgents()))
		for _, agent := range project.GetAgents() {
			agents = append(agents, agentFromProto(agent))
		}
		p := projectFromProto(project.GetSummary())
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), struct {
				Project projectDTO `json:"project"`
				Agents  []agentDTO `json:"agents"`
			}{p, agents})
		}
		table := newTable(cmd.OutOrStdout(), "ID\tNAME\tPROVIDER\tMODEL\tSTATUS")
		for _, agent := range agents {
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\n", agent.ShortID, agent.Name, agent.Provider, agent.Model, agent.Availability)
		}
		return table.Flush()
	}}
	cmd.AddCommand(list)
	return cmd
}
