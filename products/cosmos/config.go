package cosmos

import (
	"github.com/chaitin/chaitin-cli/config"
	"github.com/spf13/cobra"
)

var (
	serverURL string
	apiToken  string
	dryRun    bool
)

type runtimeConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

// ApplyRuntimeConfig 从 config.Raw 和命令行 flags 解析 cosmos 的运行时配置。
// 优先级：命令行 flags > 环境变量 > config.yaml。
func ApplyRuntimeConfig(cmd *cobra.Command, cfg config.Raw, isDryRun bool) {
	productCfg, err := config.DecodeProduct[runtimeConfig](cfg, "cosmos")
	if err != nil {
		return
	}
	dryRun = isDryRun

	if flag := cmd.Flags().Lookup("url"); flag != nil && flag.Changed {
		value, flagErr := cmd.Flags().GetString("url")
		if flagErr == nil {
			serverURL = value
		}
	} else {
		serverURL = productCfg.URL
	}

	if flag := cmd.Flags().Lookup("api-key"); flag != nil && flag.Changed {
		value, flagErr := cmd.Flags().GetString("api-key")
		if flagErr == nil {
			apiToken = value
		}
	} else {
		apiToken = productCfg.APIKey
	}
}
