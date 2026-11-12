package cmd

import (
	"fmt"
	"github.com/chaitin/chaitin-cli/products/insight/client"
	"github.com/spf13/cobra"
)

func NewRawCmd(getClient func(cmd *cobra.Command) *client.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "raw [method] [path]",
		Short: "Send a raw HTTP request to Insight API",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := args[0]
			path := args[1]
			c := getClient(cmd)

			resp, err := c.Request(method, path, nil)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(resp))
			return nil
		},
	}
}
