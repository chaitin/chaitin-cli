package agentcompose

import (
	"fmt"

	"connectrpc.com/connect"
	healthv1 "github.com/chaitin/agent-compose/proto/health/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

func newStatusCommand(state *commandState) *cobra.Command {
	return &cobra.Command{Use: "status", Short: "Show remote daemon status", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		resp, err := state.clients().health.Status(ctx, connect.NewRequest(&emptypb.Empty{}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), healthDTO(resp.Msg))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Status: healthy\nVersion: %s\nUptime: %ds\nStarted: %s\n", firstNonEmpty(resp.Msg.GetBuildVersion(), resp.Msg.GetVersion(), "unknown"), resp.Msg.GetUptimeSeconds(), timestampText(resp.Msg.GetStartedAt()))
		return nil
	}}
}

func healthDTO(status *healthv1.HealthStatusResponse) map[string]any {
	return map[string]any{"status": "healthy", "version": firstNonEmpty(status.GetBuildVersion(), status.GetVersion()), "current_time": timestampText(status.GetCurrentTime()), "started_at": timestampText(status.GetStartedAt()), "uptime_seconds": status.GetUptimeSeconds(), "go_version": status.GetGoVersion(), "num_goroutines": status.GetNumGoroutines()}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
