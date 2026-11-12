package tanswer

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newAuthCommand(opts *RootOptions) *cobra.Command {
	auth := &cobra.Command{
		Use:   "auth",
		Short: "Authentication and connection self-checks",
		Long:  "Authentication and connection self-checks for Quanxi. Use status to inspect the configured target and check to validate the OpenAPI Token.",
	}
	auth.AddCommand(newAuthStatusCommand(opts))
	auth.AddCommand(newAuthCheckCommand(opts))
	return auth
}

func newAuthStatusCommand(opts *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "查看当前连接状态",
		Long:  "查看当前连接状态。返回当前配置的全悉环境地址、是否配置 OpenAPI Token、请求超时时间。环境地址可通过 --url 或 TANSWER_URL 提供，Token 可通过 --api-key 或 TANSWER_API_KEY 提供。",
		Example: "  chaitin-cli tanswer auth status --url 'https://<全悉 Web 端 IP>'\n" +
			"  TANSWER_URL='https://<全悉 Web 端 IP>' chaitin-cli tanswer auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(ConfigOptions{
				Address:            opts.Address,
				Token:              opts.Token,
				Timeout:            opts.Timeout,
				InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet,
				AllowMissingToken: true,
			})
			if err != nil {
				return err
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查看当前连接状态",
				Command: "chaitin-cli tanswer auth status",
				Data: map[string]any{
					"environment":          cfg.BaseURL,
					"token_set":            cfg.APIToken != "",
					"timeout":              cfg.Timeout.String(),
					"insecure_skip_verify": cfg.InsecureSkipVerify,
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
}

func newAuthCheckCommand(opts *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "check",
		Short:   "校验 OpenAPI Token",
		Long:    "校验 OpenAPI Token 是否可用于当前全悉环境。该命令先读取基础环境信息，再调用轻量 TokenAuth 只读接口验证 Token 链路，不新增 CLI 专属 Token，不绕过现有权限、频率限制和 IP 访问策略。",
		Example: "  chaitin-cli tanswer auth check --url 'https://<全悉 Web 端 IP>' --api-key \"<全悉 OpenAPI Token>\"",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(ConfigOptions{
				Address:            opts.Address,
				Token:              opts.Token,
				Timeout:            opts.Timeout,
				InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet,
			})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			baseInfo := map[string]any{}
			if err := client.CallRPC(context.Background(), "OpsService.GetBaseInfo", map[string]any{}, &baseInfo); err != nil {
				return writeTokenCheckError(cmd, err)
			}
			tokenCheck := map[string]any{}
			if err := client.CallRPC(context.Background(), "AssetService.SearchTags", map[string]any{}, &tokenCheck); err != nil {
				return writeTokenCheckError(cmd, err)
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "校验 OpenAPI Token",
				Command: "chaitin-cli tanswer auth check",
				Data: map[string]any{
					"token_available": true,
					"base_info":       baseInfo,
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
}

func writeTokenCheckError(cmd *cobra.Command, err error) error {
	raw, renderErr := RenderJSON(ErrorEnvelope{
		Success: false,
		Task:    "校验 OpenAPI Token",
		Command: "chaitin-cli tanswer auth check",
		Error: CLIError{
			Code:      "TOKEN_CHECK_FAILED",
			Message:   err.Error(),
			Retryable: false,
		},
	})
	if renderErr != nil {
		return renderErr
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}
