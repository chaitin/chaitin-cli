package system

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chaitin/workspace-cli/products/safeline/cmd"
	"github.com/spf13/cobra"
)

// NewCommand creates the system command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "system",
		Short: "System management commands",
		Long:  "Commands for SafeLine system management.",
	}
	c.AddCommand(newLogCmd())
	c.AddCommand(newLicenseCmd())
	c.AddCommand(newMachineIDCmd())
	return c
}

func newLogCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "log",
		Short: "System log commands",
		Long:  "Commands for viewing system logs.",
	}
	c.AddCommand(newLogListCmd())
	return c
}

func newLogListCmd() *cobra.Command {
	var count, offset int

	c := &cobra.Command{
		Use:   "list",
		Short: "List system logs",
		Long: `List system logs.

Examples:
  safeline system log list
  safeline system log list --count 50 --offset 0`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			query := map[string]string{
				"scope":  "monitor:system",
				"count":  fmt.Sprintf("%d", count),
				"offset": fmt.Sprintf("%d", offset),
			}

			env, err := cl.Do("GET", "/api/FilterV2API", nil, query)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output
			var resp struct {
				Items []struct {
					ID         int    `json:"id"`
					CreateTime string `json:"create_time"`
					LogType    string `json:"log_type"`
					Content    string `json:"content"`
					Username   string `json:"username"`
					IP         string `json:"ip"`
				} `json:"items"`
			}
			if err := json.Unmarshal(env.Data, &resp); err != nil {
				return err
			}

			headers := []string{"ID", "Time", "Type", "User", "Content"}
			var rows [][]string
			for _, r := range resp.Items {
				rows = append(rows, []string{
					fmt.Sprintf("%d", r.ID),
					r.CreateTime,
					r.LogType,
					r.Username,
					truncate(r.Content, 50),
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}

	c.Flags().IntVar(&count, "count", 20, "Number of logs to return")
	c.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")

	return c
}

func newLicenseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "license",
		Short: "Get license information",
		Long: `Get SafeLine license information.

Examples:
  safeline system license`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/LicenseAPI", nil, nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Key-value output
			var result struct {
				LicenseID       string   `json:"license_id"`
				ClientID        string   `json:"client_id"`
				ClientName      string   `json:"client_name"`
				ProductVersion  string   `json:"product_version"`
				NotValidBefore  int64    `json:"not_valid_before"`
				NotValidAfter   int64    `json:"not_valid_after"`
				IsNeverExpire   bool     `json:"is_never_expire"`
				MaxMinionNum    int      `json:"max_minion_num"`
				QPSThreshold    int      `json:"qps_threshold"`
				ProductFeatures []string `json:"product_features"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}

			expires := "Never"
			if !result.IsNeverExpire {
				expires = formatTimestamp(result.NotValidAfter)
			}

			cmd.PrintKeyValue(map[string]string{
				"License ID":    result.LicenseID,
				"Client ID":     result.ClientID,
				"Client Name":   result.ClientName,
				"Product":       result.ProductVersion,
				"Valid From":    formatTimestamp(result.NotValidBefore),
				"Expires At":    expires,
				"Max Nodes":     fmt.Sprintf("%d", result.MaxMinionNum),
				"QPS Threshold": fmt.Sprintf("%d", result.QPSThreshold),
				"Features":      strings.Join(result.ProductFeatures, ", "),
			})
			return nil
		},
	}
}

func formatTimestamp(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

func newMachineIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "machine-id",
		Short: "Get machine ID",
		Long: `Get SafeLine machine ID.

Examples:
  safeline system machine-id`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/MachineIDAPI", nil, nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Key-value output
			var result struct {
				MachineID string `json:"machine_id"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}

			cmd.PrintKeyValue(map[string]string{
				"Machine ID": result.MachineID,
			})
			return nil
		},
	}
}

// truncate truncates a string to the given length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
