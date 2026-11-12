package agentcompose

import (
	"context"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	state := &commandState{options: runtimeOptions{timeoutText: "30s"}}
	cmd := &cobra.Command{
		Use:           productName,
		Short:         "Remotely operate Agent Compose through the Token/RBAC API",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          noArgs(state),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.SetContext(context.WithValue(context.Background(), stateKey{}, state))
	cmd.PersistentFlags().StringVar(&state.options.URL, "url", "", "Token/RBAC API base URL, including an optional path prefix")
	cmd.PersistentFlags().StringVarP(&state.options.Project, "project", "p", "", "Default project name or ID")
	cmd.PersistentFlags().StringVar(&state.options.timeoutText, "timeout", "30s", "Unary request timeout")
	cmd.PersistentFlags().BoolVar(&state.options.Insecure, "insecure", false, "Skip TLS certificate verification")
	cmd.PersistentFlags().BoolVar(&state.options.JSON, "json", false, "Output JSON or NDJSON")
	cmd.PersistentPreRunE = func(command *cobra.Command, _ []string) error { return state.prepare(command) }
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return usageError(err.Error(), state.options.JSON)
	})

	cmd.AddCommand(
		newAuthCommand(state),
		newStatusCommand(state),
		newProjectCommand(state),
		newAgentCommand(state),
		newRunCommand(state),
		newLogsCommand(state),
		newPSCommand(state, "ps"),
		newStatsCommand(state),
		newInspectCommand(state),
		newSchedulerCommand(state),
		newSandboxCommand(state),
		newExecCommand(state),
		newLLMCommand(state),
	)
	return cmd
}

func noArgs(state *commandState) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) != 0 {
			return usageError("this command accepts no arguments", state.options.JSON)
		}
		return nil
	}
}

func exactArgs(n int, state *commandState) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) != n {
			return usageError("invalid number of arguments", state.options.JSON)
		}
		return nil
	}
}

func rangeArgs(minimum, maximum int, state *commandState) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) < minimum || len(args) > maximum {
			return usageError("invalid number of arguments", state.options.JSON)
		}
		return nil
	}
}

func requestContext(cmd *cobra.Command, state *commandState) (context.Context, context.CancelFunc) {
	return context.WithTimeout(cmd.Context(), state.options.Timeout)
}

func streamContext(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	return context.WithCancel(cmd.Context())
}
