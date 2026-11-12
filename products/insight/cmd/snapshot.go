package cmd

import (
	"net/url"

	"github.com/chaitin/chaitin-cli/products/insight/client"
	"github.com/spf13/cobra"
)

func NewSnapshotCmd(getClient func(cmd *cobra.Command) *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Asset snapshot management",
	}

	assetCmd := &cobra.Command{
		Use:   "asset",
		Short: "Get managed asset snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)

			taskId, _ := cmd.Flags().GetString("task-id")
			executionId, _ := cmd.Flags().GetString("execution-id")

			query := url.Values{}
			query.Set("id", taskId)
			if executionId != "" {
				query.Set("execution_id", executionId)
			}

			resp, err := c.Request("GET", "/exposure/api/snapshot/asset?"+query.Encode(), nil)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("output")
			renderer := client.NewRenderer(format, cmd.OutOrStdout())
			return renderer.Render(resp)
		},
	}
	assetCmd.Flags().String("task-id", "", "task id to query")
	assetCmd.Flags().String("execution-id", "", "filter by execution id")
	assetCmd.MarkFlagRequired("task-id")
	cmd.AddCommand(assetCmd)

	return cmd
}
