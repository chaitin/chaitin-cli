package cmd

import (
	"fmt"
	"github.com/chaitin/chaitin-cli/products/insight/client"
	"github.com/spf13/cobra"
	"net/url"
)

func NewOrderCmd(getClient func(cmd *cobra.Command) *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order",
		Short: "Manage Workflow orders",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all workflow orders",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)

			page, _ := cmd.Flags().GetInt("page")
			size, _ := cmd.Flags().GetInt("size")

			q := url.Values{}
			q.Add("page", fmt.Sprintf("%d", page))
			q.Add("size", fmt.Sprintf("%d", size))

			if name, _ := cmd.Flags().GetString("name"); name != "" {
				q.Add("name", name)
			}
			if status, _ := cmd.Flags().GetInt("status"); status != 0 {
				q.Add("status", fmt.Sprintf("%d", status))
			}
			if isTimeout, _ := cmd.Flags().GetBool("is-timeout"); cmd.Flags().Changed("is-timeout") {
				q.Add("is_timeout", fmt.Sprintf("%t", isTimeout))
			}

			path := "/workflow/api/orders/all?" + q.Encode()

			resp, err := c.Request("GET", path, nil)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("output")
			renderer := client.NewRenderer(format, cmd.OutOrStdout())
			return renderer.Render(resp)
		},
	}

	listCmd.Flags().Int("page", 1, "Page number")
	listCmd.Flags().Int("size", 20, "Number of items per page")
	listCmd.Flags().String("name", "", "Filter by order name")
	listCmd.Flags().Int("status", 0, "Filter by status")
	listCmd.Flags().Bool("is-timeout", false, "Filter by timeout status")

	cmd.AddCommand(listCmd)

	return cmd
}
