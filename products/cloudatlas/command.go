package cloudatlas

import (
	"fmt"
	"os"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/chaitin/chaitin-cli/products/cloudatlas/spec"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	runtimeCfg       Config
	dryRun           bool
	outputFormat     string
	runtimeInsecure  bool
	verbose          bool
	verboseSensitive bool
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloudAtlas",
		Short: "Cloud Atlas asset, exposure, risk, intelligence, and task management",
		Long: `Cloud Atlas CLI

Commands are generated from the embedded OpenAPI schema.
Authentication uses the Token HTTP header.

Config example:
  cloudAtlas:
    url: https://your-cloud-atlas.example
    token: your-token
    space_id: your-space-id`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			applyRuntimeConfig(cmd)
		},
	}

	cmd.PersistentFlags().String("url", "", "Cloud Atlas OpenAPI base URL；也可配置 cloudAtlas.url 或 CLOUD_ATLAS_URL")
	cmd.PersistentFlags().String("token", "", "Cloud Atlas API token；请求时作为 Token header 发送，也可配置 cloudAtlas.token 或 CLOUD_ATLAS_TOKEN")
	cmd.PersistentFlags().String("space-id", "", "Cloud Atlas 空间 ID；作为所有请求的 space query 默认值，也可配置 cloudAtlas.space_id；子命令 --space 可覆盖")
	cmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "输出格式；可选值: table, json")
	cmd.PersistentFlags().BoolVar(&runtimeInsecure, "insecure", true, "跳过 TLS 证书校验；测试或自签名环境使用，默认 true")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "打印请求 URL、headers 和 body；敏感值默认会被脱敏")
	cmd.PersistentFlags().BoolVar(&verboseSensitive, "verbose-sensitive", false, "verbose 时打印 token 等敏感值明文，仅在本地调试时使用")

	api, err := ParseSpec(spec.OpenAPIJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to parse embedded Cloud Atlas OpenAPI schema: %v\n", err)
		return cmd
	}
	commands, err := NewParser(api).GenerateCommands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to generate Cloud Atlas commands: %v\n", err)
		return cmd
	}
	cmd.AddCommand(commands...)
	return cmd
}

func ApplyRuntimeConfig(cmd *cobra.Command, cfg config.Raw, isDryRun bool) {
	productCfg, err := config.DecodeProduct[Config](cfg, "cloudAtlas")
	if err != nil {
		return
	}
	runtimeCfg = productCfg
	dryRun = isDryRun
}

func applyRuntimeConfig(cmd *cobra.Command) {
	if flag := lookupFlag(cmd, "url"); flag != nil && !flag.Changed && runtimeCfg.URL != "" {
		_ = setFlag(cmd, "url", runtimeCfg.URL)
	}
	if flag := lookupFlag(cmd, "token"); flag != nil && !flag.Changed && runtimeCfg.Token != "" {
		_ = setFlag(cmd, "token", runtimeCfg.Token)
	}
	if flag := lookupFlag(cmd, "space-id"); flag != nil && !flag.Changed && runtimeCfg.SpaceID != "" {
		_ = setFlag(cmd, "space-id", runtimeCfg.SpaceID)
	}
}

func lookupFlag(cmd *cobra.Command, name string) *pflag.Flag {
	if flag := cmd.Flags().Lookup(name); flag != nil {
		return flag
	}
	return cmd.PersistentFlags().Lookup(name)
}

func setFlag(cmd *cobra.Command, name, value string) error {
	if cmd.Flags().Lookup(name) != nil {
		return cmd.Flags().Set(name, value)
	}
	return cmd.PersistentFlags().Set(name, value)
}
