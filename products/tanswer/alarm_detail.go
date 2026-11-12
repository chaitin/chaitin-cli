package tanswer

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type alarmDetailOptions struct {
	id string
}

type alarmDetailRPCResult struct {
	Data map[string]any `json:"data"`
}

func newAlarmDetailCommand(opts *RootOptions) *cobra.Command {
	var alarmOpts alarmDetailOptions
	cmd := &cobra.Command{
		Use:   "detail",
		Short: "查看威胁告警详情",
		Long: "查看威胁告警详情，用于从告警列表返回的 doc_id 下钻到单条告警。该命令返回产品已有详情字段，包括列表基础字段、cve_list、alert_msg 和 alarm_description。\n\n" +
			"输出：doc_id、detail。detail 中保留后端告警详情字段，如 name、severity、timestamp、attacker、victim、payload、cve_list、alert_msg、alarm_description。",
		Example: "  chaitin-cli tanswer alarm detail --id '<doc_id>'",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(alarmOpts.id) == "" {
				return writeAlarmDetailError(cmd, "MISSING_ALARM_ID", "missing alarm doc_id: set --id", false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var result alarmDetailRPCResult
			if err := client.CallRPC(cmd.Context(), "AlarmService.GetAlarm", map[string]any{"doc_id": alarmOpts.id}, &result); err != nil {
				return writeAlarmDetailError(cmd, "ALARM_DETAIL_FAILED", err.Error(), true)
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查看威胁告警详情",
				Command: "chaitin-cli tanswer alarm detail",
				Query:   map[string]any{"doc_id": alarmOpts.id},
				Data: map[string]any{
					"doc_id": alarmOpts.id,
					"detail": result.Data,
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&alarmOpts.id, "id", "", "alarm doc_id from alarm list")
	return cmd
}

func writeAlarmDetailError(cmd *cobra.Command, code string, message string, retryable bool) error {
	raw, renderErr := RenderJSON(ErrorEnvelope{
		Success: false,
		Task:    "查看威胁告警详情",
		Command: "chaitin-cli tanswer alarm detail",
		Error:   CLIError{Code: code, Message: message, Retryable: retryable},
	})
	if renderErr != nil {
		return renderErr
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}
