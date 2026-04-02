package xray

import (
	"github.com/chaitin/workspace-cli/config"
	"github.com/chaitin/workspace-cli/products/xray/cli"
	"github.com/spf13/cobra"
)

func NewCommand() (*cobra.Command, error) {
	return cli.MakeCommand()
}

func ApplyRuntimeConfig(cmd *cobra.Command, cfg config.Raw, dryRun bool) {
	_ = cmd
	cli.SetRuntimeConfig(cfg, dryRun)
}
