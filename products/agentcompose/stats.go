package agentcompose

import (
	"fmt"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/spf13/cobra"
)

type metricDTO struct {
	Value   float64 `json:"value" yaml:"value"`
	Unit    string  `json:"unit,omitempty" yaml:"unit,omitempty"`
	Status  string  `json:"status" yaml:"status"`
	Message string  `json:"message,omitempty" yaml:"message,omitempty"`
}
type statsDTO struct {
	SandboxID     string    `json:"sandbox_id" yaml:"sandbox_id"`
	Driver        string    `json:"driver" yaml:"driver"`
	SampledAt     string    `json:"sampled_at" yaml:"sampled_at"`
	CPU           metricDTO `json:"cpu_percent" yaml:"cpu_percent"`
	MemoryUsage   metricDTO `json:"memory_usage_bytes" yaml:"memory_usage_bytes"`
	MemoryLimit   metricDTO `json:"memory_limit_bytes" yaml:"memory_limit_bytes"`
	MemoryPercent metricDTO `json:"memory_percent" yaml:"memory_percent"`
	NetworkRX     metricDTO `json:"network_rx_bytes" yaml:"network_rx_bytes"`
	NetworkTX     metricDTO `json:"network_tx_bytes" yaml:"network_tx_bytes"`
	BlockRead     metricDTO `json:"block_read_bytes" yaml:"block_read_bytes"`
	BlockWrite    metricDTO `json:"block_write_bytes" yaml:"block_write_bytes"`
	Uptime        metricDTO `json:"uptime_seconds" yaml:"uptime_seconds"`
}

func metricFromProto(metric *agentcomposev2.MetricValue) metricDTO {
	if metric == nil {
		return metricDTO{Status: "unspecified"}
	}
	return metricDTO{Value: metric.GetValue(), Unit: metric.GetUnit(), Status: enumText(metric.GetStatus(), "METRIC_STATUS_"), Message: metric.GetMessage()}
}
func statsFromProto(stats *agentcomposev2.SandboxStats) statsDTO {
	return statsDTO{SandboxID: stats.GetSandboxId(), Driver: stats.GetDriver(), SampledAt: timestampText(stats.GetSampledAt()), CPU: metricFromProto(stats.GetCpuPercent()), MemoryUsage: metricFromProto(stats.GetMemoryUsageBytes()), MemoryLimit: metricFromProto(stats.GetMemoryLimitBytes()), MemoryPercent: metricFromProto(stats.GetMemoryPercent()), NetworkRX: metricFromProto(stats.GetNetworkRxBytes()), NetworkTX: metricFromProto(stats.GetNetworkTxBytes()), BlockRead: metricFromProto(stats.GetBlockReadBytes()), BlockWrite: metricFromProto(stats.GetBlockWriteBytes()), Uptime: metricFromProto(stats.GetUptimeSeconds())}
}
func newStatsCommand(state *commandState) *cobra.Command {
	return &cobra.Command{Use: "stats [sandbox-ref]", Short: "Show Sandbox resource statistics", Args: rangeArgs(0, 1, state), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		project, err := resolveProject(ctx, state, state.clients().project)
		if err != nil {
			return err
		}
		targets := make([]*agentcomposev2.Sandbox, 0)
		if len(args) == 1 {
			sandbox, err := resolveSandbox(ctx, state, state.clients().sandbox, project.GetSummary().GetProjectId(), args[0])
			if err != nil {
				return err
			}
			targets = append(targets, sandbox)
		} else {
			targets, _, _, err = listSandboxes(ctx, state.clients().sandbox, project.GetSummary().GetProjectId(), []agentcomposev2.SandboxStatus{agentcomposev2.SandboxStatus_SANDBOX_STATUS_RUNNING}, offsetOptions{AllPages: true, Limit: 50})
			if err != nil {
				return mapConnectError(err, state.options.URL, state.options.JSON)
			}
		}
		output := make([]statsDTO, 0, len(targets))
		for _, target := range targets {
			resp, err := state.clients().sandbox.GetSandboxStats(ctx, connect.NewRequest(&agentcomposev2.GetSandboxStatsRequest{SandboxId: target.GetSandboxId()}))
			if err != nil {
				return mapConnectError(err, state.options.URL, state.options.JSON)
			}
			output = append(output, statsFromProto(resp.Msg.GetStats()))
		}
		if len(args) == 1 {
			if state.options.JSON {
				return writeJSON(cmd.OutOrStdout(), output[0])
			}
			return writeYAML(cmd.OutOrStdout(), output[0])
		}
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), struct {
				Project projectDTO `json:"project"`
				Stats   []statsDTO `json:"stats"`
			}{projectFromProto(project.GetSummary()), output})
		}
		for _, stats := range output {
			fmt.Fprintf(cmd.OutOrStdout(), "%s CPU %.2f%% Memory %.2f%%\n", shortID(stats.SandboxID), stats.CPU.Value, stats.MemoryPercent.Value)
		}
		return nil
	}}
}
