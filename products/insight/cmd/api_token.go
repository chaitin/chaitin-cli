package cmd

import (
	"fmt"
	"github.com/chaitin/chaitin-cli/products/insight/client"
	"github.com/spf13/cobra"
)

func NewApiKeyCmd(getClient func(cmd *cobra.Command) *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-key",
		Short: "Manage Insight API key",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)

			count, _ := cmd.Flags().GetInt("count")
			offset, _ := cmd.Flags().GetInt("offset")

			// build query string
			query := ""
			if count > 0 || offset > 0 {
				query = fmt.Sprintf("?count=%d&offset=%d", count, offset)
			}

			resp, err := c.Request("GET", "/mgt/api/api_tokens"+query, nil)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("output")
			renderer := client.NewRenderer(format, cmd.OutOrStdout())
			return renderer.Render(resp)
		},
	}

	listCmd.Flags().Int("count", 20, "page count")
	listCmd.Flags().Int("offset", 0, "page offset")
	cmd.AddCommand(listCmd)

	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Get API token details",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)

			id, _ := cmd.Flags().GetInt("id")
			query := ""
			if id > 0 {
				query = fmt.Sprintf("?id=%d", id)
			}

			resp, err := c.Request("GET", "/mgt/api/api_token"+query, nil)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("output")
			renderer := client.NewRenderer(format, cmd.OutOrStdout())
			return renderer.Render(resp)
		},
	}
	infoCmd.Flags().Int("id", 0, "API Token ID")
	cmd.AddCommand(infoCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API token",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = getClient(cmd)

			// For POST requests, we typically read a JSON file or stdin using a standard helper.
			// Since we haven't implemented a full JSON payload flag like --payload-file yet,
			// this serves as a structural placeholder for mutating operations.
			return fmt.Errorf("create API token is not fully implemented: requires JSON payload support")
		},
	}
	cmd.AddCommand(createCmd)

	return cmd
}
