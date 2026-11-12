package tanswer

import (
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"
)

type RootOptions struct {
	Address               string
	Token                 string
	Timeout               string
	InsecureSkipVerify    bool
	InsecureSkipVerifySet bool
	Out                   io.Writer
	ErrOut                io.Writer
}

// NewCommand creates the T-Answer product command for chaitin-cli.
func NewCommand() *cobra.Command {
	opts := RootOptions{}
	return newTAnswerCommand(&opts, true)
}

// NewRootCommand is kept for product-level tests that exercise the full
// chaitin-cli tanswer command path.
func NewRootCommand(opts RootOptions) *cobra.Command {
	if opts.Out == nil {
		opts.Out = io.Discard
	}
	if opts.ErrOut == nil {
		opts.ErrOut = io.Discard
	}

	root := &cobra.Command{
		Use:           "chaitin-cli",
		Short:         "Chaitin command line tools",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(opts.Out)
	root.SetErr(opts.ErrOut)

	root.AddCommand(newTAnswerCommand(&opts, true))
	return root
}

func newTAnswerCommand(opts *RootOptions, registerFlags bool) *cobra.Command {
	if opts == nil {
		opts = &RootOptions{}
	}

	cmd := &cobra.Command{
		Use:   "tanswer",
		Short: "全悉安全运营 CLI（运行 \"chaitin-cli tanswer --help\" 开始）",
		Long: "全悉面向人类操作者和 AI Agent 的语义命令入口。\n\n" +
			"开始使用：通过 --url/--api-key/--timeout/--insecure、TANSWER_URL/TANSWER_API_KEY/TANSWER_TIMEOUT/TANSWER_INSECURE，或 config.yaml 的 tanswer.url/tanswer.api_key/tanswer.timeout/tanswer.insecure 配置连接；随后运行 \"chaitin-cli tanswer auth check\"。\n" +
			"发现命令：先运行本命令或领域命令的 --help；需要机器可读的完整参数、输出和风险契约时运行 \"chaitin-cli tanswer manifest\"。\n" +
			"操作规则：优先使用 semantic commands；仅在目标能力未被语义命令覆盖且调用者已知并获授权时使用 api fallback。受保护写操作和非 GET/HEAD 的 api fallback 都必须先使用 --preview，核对目标、影响和风险；AI Agent 必须等待用户对本次变更或请求的明确确认后，才能使用命令 help 或 manifest 指定的 --confirm token。确认 token 是技术校验，不等同于用户授权。根级 --dry-run 不适用于 tanswer。",
		Example: `  chaitin-cli tanswer auth check --url 'https://<全悉 Web 端 IP>' --api-key '<全悉 OpenAPI Token>'
  chaitin-cli tanswer alarm --help
  chaitin-cli tanswer manifest`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			syncOptionsFromFlags(cmd, opts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	if registerFlags {
		cmd.PersistentFlags().StringVar(&opts.Address, "url", "", "T-Answer console URL, or TANSWER_URL")
		cmd.PersistentFlags().StringVar(&opts.Token, "api-key", "", "OpenAPI Token, or TANSWER_API_KEY")
		cmd.PersistentFlags().StringVar(&opts.Timeout, "timeout", "", "request timeout, or TANSWER_TIMEOUT; default 30s")
		cmd.PersistentFlags().BoolVar(&opts.InsecureSkipVerify, "insecure", false, "skip TLS certificate verification, or TANSWER_INSECURE")
	}

	cmd.AddCommand(newAuthCommand(opts))
	cmd.AddCommand(newRawAPICommand(opts))
	cmd.AddCommand(newAlarmCommand(opts))
	cmd.AddCommand(newFileAlarmCommand(opts))
	cmd.AddCommand(newAssetCommand(opts))
	cmd.AddCommand(newSystemCommand(opts))
	cmd.AddCommand(newMetadataCommand(opts))
	cmd.AddCommand(newPolicyCommand(opts))
	cmd.AddCommand(newResponseCommand(opts))
	cmd.AddCommand(newManifestCommand(opts))
	return cmd
}

func syncOptionsFromFlags(cmd *cobra.Command, opts *RootOptions) {
	if cmd == nil || opts == nil {
		return
	}
	if value := inheritedFlagValue(cmd, "url"); value != "" {
		opts.Address = value
	}
	if value := inheritedFlagValue(cmd, "api-key"); value != "" {
		opts.Token = value
	}
	if value := inheritedFlagValue(cmd, "timeout"); value != "" {
		opts.Timeout = value
	}
	if flag := cmd.Flags().Lookup("insecure"); flag != nil && flag.Changed {
		if insecure, err := strconv.ParseBool(flag.Value.String()); err == nil {
			opts.InsecureSkipVerify = insecure
			opts.InsecureSkipVerifySet = true
		}
	}
}

func inheritedFlagValue(cmd *cobra.Command, name string) string {
	if flag := cmd.Flags().Lookup(name); flag != nil {
		return flag.Value.String()
	}
	if flag := cmd.InheritedFlags().Lookup(name); flag != nil {
		return flag.Value.String()
	}
	return ""
}

func Execute(opts RootOptions) error {
	cmd := NewRootCommand(opts)
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("chaitin-cli failed: %w", err)
	}
	return nil
}
