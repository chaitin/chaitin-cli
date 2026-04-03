package stats

import (
	"encoding/json"
	"fmt"

	"github.com/chaitin/workspace-cli/products/safeline/cmd"
	"github.com/spf13/cobra"
)

// NewCommand creates the stats command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "stats",
		Short: "Statistics and overview commands",
		Long:  "Commands for viewing SafeLine statistics and overview data.",
	}
	c.AddCommand(newOverviewCmd())
	return c
}

// OverviewResult represents the filtered overview response.
type OverviewResult struct {
	SrcIP [][]interface{} `json:"src_ip"`
	Host  [][]interface{} `json:"host"`
	Total map[string]int  `json:"total"`
}

func newOverviewCmd() *cobra.Command {
	var duration string

	c := &cobra.Command{
		Use:   "overview",
		Short: "Get statistics overview",
		Long: `Get statistics overview for the specified duration.

Duration options:
  h    24 hours (hourly breakdown)
  d    30 days (daily breakdown)

Response fields:
  src_ip       Top source IPs with attack counts
  host         Top hosts with attack counts
  total        Total request, deny, and attack counts

Examples:
  # Get 24-hour statistics
  safeline stats overview --duration h

  # Get 30-day statistics
  safeline stats overview --duration d`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			query := map[string]string{
				"duration": duration,
				"host":     "attack",
				"src_ip":   "attack",
				"total":    "true",
			}

			env, err := cl.Do("GET", "/api/OverviewAPI", nil, query)
			if err != nil {
				return err
			}

			// Parse and filter response
			var fullResp struct {
				SrcIP [][]interface{} `json:"src_ip"`
				Host  [][]interface{} `json:"host"`
				Total map[string]int  `json:"total"`
			}
			if err := json.Unmarshal(env.Data, &fullResp); err != nil {
				return err
			}

			result := OverviewResult{
				SrcIP: fullResp.SrcIP,
				Host:  fullResp.Host,
				Total: fullResp.Total,
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintResult(c, result)
			}

			// Table output
			return printOverviewTable(result)
		},
	}

	c.Flags().StringVar(&duration, "duration", "h", "Statistics duration (h=24h, d=30d)")
	c.RegisterFlagCompletionFunc("duration", func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"h", "d"}, cobra.ShellCompDirectiveNoFileComp
	})

	return c
}

func printOverviewTable(result OverviewResult) error {
	// Print total
	cmd.PrintKeyValue(map[string]string{
		"Requests": fmt.Sprintf("%d", result.Total["request"]),
		"Attacks":  fmt.Sprintf("%d", result.Total["attack"]),
		"Deny":     fmt.Sprintf("%d", result.Total["deny"]),
	})

	// Print top source IPs
	if len(result.SrcIP) > 0 {
		fmt.Println("\nTop Source IPs:")
		headers := []string{"#", "IP", "Count"}
		var rows [][]string
		for i, ip := range result.SrcIP {
			if len(ip) >= 2 {
				rows = append(rows, []string{fmt.Sprintf("%d", i+1), fmt.Sprintf("%v", ip[0]), fmt.Sprintf("%v", ip[1])})
			}
		}
		cmd.PrintTable(headers, rows)
	}

	// Print top hosts
	if len(result.Host) > 0 {
		fmt.Println("\nTop Hosts:")
		headers := []string{"#", "Host", "Count"}
		var rows [][]string
		for i, host := range result.Host {
			if len(host) >= 2 {
				rows = append(rows, []string{fmt.Sprintf("%d", i+1), fmt.Sprintf("%v", host[0]), fmt.Sprintf("%v", host[1])})
			}
		}
		cmd.PrintTable(headers, rows)
	}

	return nil
}
