package cmd

import (
	"fmt"
	"github.com/chaitin/chaitin-cli/products/insight/client"
	"github.com/spf13/cobra"
)

func NewSystemCmd(getClient func(cmd *cobra.Command) *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "System information (License, Machine ID)",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "license",
		Short: "Read license info",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)
			resp, err := c.Request("GET", "/mgt/api/license", nil)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(resp))
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "machine-id",
		Short: "Get Machine ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)
			resp, err := c.Request("GET", "/mgt/api/noauth/machine_id", nil)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(resp))
			return nil
		},
	})

	return cmd
}
