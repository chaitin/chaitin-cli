package xray

import (
	"github.com/chaitin/workspace-cli/products/xray/cli"
	"github.com/spf13/cobra"
)

func NewCommand() (*cobra.Command, error) {
	return cli.MakeCommand()
}
