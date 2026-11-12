package cmd

import (
	"net/url"

	"github.com/chaitin/chaitin-cli/products/insight/client"
	"github.com/spf13/cobra"
)

func NewResultCmd(getClient func(cmd *cobra.Command) *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "result",
		Short: "Task execution results and vulnerabilities",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List task results",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)

			taskId, _ := cmd.Flags().GetString("task-id")
			executionId, _ := cmd.Flags().GetString("execution-id")

			query := url.Values{}
			query.Set("id", taskId)
			if executionId != "" {
				query.Set("execution_id", executionId)
			}

			resp, err := c.Request("GET", "/exposure/api/result?"+query.Encode(), nil)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("output")
			renderer := client.NewRenderer(format, cmd.OutOrStdout())
			return renderer.Render(resp)
		},
	}
	listCmd.Flags().String("task-id", "", "filter by task id")
	listCmd.Flags().String("execution-id", "", "filter by execution id")
	listCmd.MarkFlagRequired("task-id")
	cmd.AddCommand(listCmd)

	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: "Comparison of risks for a task execution",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)

			execId, _ := cmd.Flags().GetString("exec-id")

			query := ""
			if execId != "" {
				query = "?exec_id=" + execId
			}

			resp, err := c.Request("GET", "/exposure/api/result/comparison"+query, nil)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("output")
			renderer := client.NewRenderer(format, cmd.OutOrStdout())
			return renderer.Render(resp)
		},
	}
	diffCmd.Flags().String("exec-id", "", "execution id to compare")
	diffCmd.MarkFlagRequired("exec-id")
	cmd.AddCommand(diffCmd)

	return cmd
}
