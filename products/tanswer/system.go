package tanswer

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newSystemCommand(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "系统基础确认语义命令",
		Long:  "系统基础确认语义命令。用于只读查看版本、License、节点状态和系统自检摘要；不进入系统高级配置或破坏性操作。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newSystemStatusCommand(opts))
	return cmd
}

func newSystemStatusCommand(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "查看系统基础状态",
		Long: "查看系统基础状态，用于快速确认全悉环境是否健康。该命令返回产品版本、部署版本、License 状态、节点列表、节点运行状态和系统自检摘要。\n\n" +
			"输出：version、license、health。health 中包含整体状态、节点总数、在线节点数、更新时间和 nodes。",
		Example: "  chaitin-cli tanswer system status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			data, err := fetchSystemStatus(cmd.Context(), client)
			if err != nil {
				return writeSystemError(cmd, "SYSTEM_STATUS_FAILED", err.Error(), true)
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查看系统基础状态",
				Command: "chaitin-cli tanswer system status",
				Data:    data,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	return cmd
}

func fetchSystemStatus(ctx context.Context, client *Client) (map[string]any, error) {
	baseInfo := map[string]any{}
	if err := client.CallRPC(ctx, "OpsService.GetBaseInfo", map[string]any{}, &baseInfo); err != nil {
		return nil, err
	}
	license := map[string]any{}
	if err := client.CallRPC(ctx, "HeraLicenseService.GetLicense", map[string]any{}, &license); err != nil {
		return nil, err
	}
	health := map[string]any{}
	if err := client.CallRPC(ctx, "OpsService.GetSystemStatusResult", map[string]any{}, &health); err != nil {
		return nil, err
	}
	return map[string]any{
		"version": summarizeSystemVersion(baseInfo),
		"license": summarizeSystemLicense(license),
		"health":  summarizeSystemHealth(health),
	}, nil
}

func summarizeSystemVersion(in map[string]any) map[string]any {
	return map[string]any{
		"product_name":    firstPresent(in, "ProductName", "product_name"),
		"product_mode":    firstPresent(in, "ProductMode", "product_mode"),
		"product_version": firstPresent(in, "ProductVersion", "product_version"),
		"version":         firstPresent(in, "Version", "version"),
		"install_time":    firstPresent(in, "InstallTime", "install_time"),
		"hardware_model":  firstPresent(in, "HardwareModel", "hardware_model"),
	}
}

func summarizeSystemLicense(in map[string]any) map[string]any {
	return map[string]any{
		"company":              firstPresent(in, "company", "Company"),
		"permanent":            firstPresent(in, "permanent", "Permanent"),
		"valid_expired":        firstPresent(in, "valid_expired", "ValidExpired"),
		"valid_expired_soon":   firstPresent(in, "valid_expired_soon", "ValidExpiredSoon"),
		"upgrade_expired":      firstPresent(in, "upgrade_expired", "UpgradeExpired"),
		"upgrade_expired_soon": firstPresent(in, "upgrade_expired_soon", "UpgradeExpiredSoon"),
		"product_version":      firstPresent(in, "product_version", "ProductVersion"),
		"license_type":         firstPresent(in, "license_type", "LicenseType"),
		"max_node_count":       firstPresent(in, "max_node_count", "MaxNodeCount"),
		"traffic_limit":        firstPresent(in, "traffic_limit", "TrafficLimit"),
		"not_valid_after":      firstPresent(in, "not_valid_after", "NotValidAfter"),
	}
}

func summarizeSystemHealth(in map[string]any) map[string]any {
	return map[string]any{
		"status":       firstPresent(in, "status", "Status"),
		"update_time":  firstPresent(in, "update_time", "UpdateTime"),
		"online_count": firstPresent(in, "online_count", "OnlineCount"),
		"total_count":  firstPresent(in, "total_count", "TotalCount"),
		"nodes":        firstPresent(in, "nodes", "Nodes"),
	}
}

func firstPresent(in map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := in[key]; ok {
			return value
		}
	}
	return nil
}

func writeSystemError(cmd *cobra.Command, code string, message string, retryable bool) error {
	raw, renderErr := RenderJSON(ErrorEnvelope{
		Success: false,
		Task:    "查看系统基础状态",
		Command: "chaitin-cli tanswer system status",
		Error:   CLIError{Code: code, Message: message, Retryable: retryable},
	})
	if renderErr != nil {
		return renderErr
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}
