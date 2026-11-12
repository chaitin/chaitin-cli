package cmd

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/chaitin/chaitin-cli/products/insight/client"
	"github.com/spf13/cobra"
)

func addCommonVulnFlags(cmd *cobra.Command) {
	// Paging and bools
	cmd.Flags().Int64("count", 20, "Number of items to return")
	cmd.Flags().Uint64("offset", 0, "Number of items to skip")
	cmd.Flags().Bool("rel-asset", true, "Include related asset info")

	// Special structural params
	cmd.Flags().String("name", "", "Filter by vulnerability name (uses ilike)")

	// Array/Slice parameters
	cmd.Flags().StringSlice("vuln-level", nil, "Filter by vulnerability level")
	cmd.Flags().StringSlice("find-by", nil, "Filter by find by")

	// Status and types that use [{"oper":"=","target":value}]
	cmd.Flags().Int("rel-workflow-status", -1, "Filter by related workflow status (-1 to ignore)")
	cmd.Flags().Int("rel-asset-state", -1, "Filter by related asset state (-1 to ignore)")
	cmd.Flags().Int("vuln-status", -1, "Filter by vulnerability status (-1 to ignore)")
	cmd.Flags().Int("check-type", -1, "Filter by check type (-1 to ignore)")

	// Standard Ints
	cmd.Flags().Int("organization-id", -1, "Filter by organization ID (-1 to ignore)")
	cmd.Flags().Int("merge-num", -1, "Filter by merge number (-1 to ignore)")
	cmd.Flags().Int("scanner-vuln-weight", -1, "Filter by scanner vulnerability weight (-1 to ignore)")

	// Standard Strings
	cmd.Flags().String("vuln-tag-ids", "", "Filter by vulnerability tag IDs")
	cmd.Flags().String("rel-asset-name", "", "Filter by related asset name")
	cmd.Flags().String("rel-asset-business", "", "Filter by related asset business")
	cmd.Flags().String("rel-asset-scope", "", "Filter by related asset scope")
	cmd.Flags().String("fix-remarks", "", "Filter by fix remarks")
	cmd.Flags().String("tag-name", "", "Filter by tag name")
	cmd.Flags().String("customize-tag-name", "", "Filter by customize tag name")

	// Time filters (expecting 123-456 strings, uses [{"oper":"in","target":value}])
	cmd.Flags().String("vuln-update-time", "", "Filter by vulnerability update time")
	cmd.Flags().String("dispose-fix-time", "", "Filter by dispose fix time")
	cmd.Flags().String("vuln-last-time", "", "Filter by last found time range (e.g. 1780588800-1780675200)")
	cmd.Flags().String("vuln-first-time", "", "Filter by first found time range (e.g. 1780588800-1780675200)")
	cmd.Flags().String("plan-fix-time", "", "Filter by plan fix time")
}

func addIpVulnFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("vuln-vpt-priority", nil, "Filter by VPT priority (e.g. 1,2,3)")
	cmd.Flags().String("vuln-ip", "", "Filter by vulnerability IP (uses 'like' operator)")
	cmd.Flags().String("protocol", "", "Filter by protocol")
	cmd.Flags().Int("port", -1, "Filter by port (-1 to ignore)")
}

func addWebVulnFlags(cmd *cobra.Command) {
	cmd.Flags().Int("manager-id", -1, "Filter by manager ID (-1 to ignore)")
	cmd.Flags().String("vuln-url", "", "Filter by vulnerability URL")
	cmd.Flags().String("vuln-url-domain", "", "Filter by URL domain")
	cmd.Flags().Int("vuln-url-port", -1, "Filter by URL port (-1 to ignore)")
	cmd.Flags().String("vuln-url-scheme", "", "Filter by URL scheme")
}

func buildCommonQueryParams(cmd *cobra.Command) map[string]interface{} {
	count, _ := cmd.Flags().GetInt64("count")
	offset, _ := cmd.Flags().GetUint64("offset")
	relAsset, _ := cmd.Flags().GetBool("rel-asset")

	params := map[string]interface{}{
		"count":     count,
		"offset":    offset,
		"rel_asset": relAsset,
	}

	if v, _ := cmd.Flags().GetString("name"); v != "" {
		params["name"] = []map[string]interface{}{{"oper": "ilike", "target": v}}
	}

	if flags := []string{"vuln-level", "find-by"}; len(flags) > 0 {
		for _, f := range flags {
			if v, _ := cmd.Flags().GetStringSlice(f); len(v) > 0 {
				var intSlice []int
				isInts := true
				for _, val := range v {
					if i, err := strconv.Atoi(val); err == nil {
						intSlice = append(intSlice, i)
					} else {
						isInts = false
					}
				}
				key := strings.ReplaceAll(f, "-", "_")
				if isInts && len(intSlice) > 0 {
					params[key] = intSlice
				} else {
					params[key] = v
				}
			}
		}
	}

	intFlags := []string{"rel-workflow-status", "rel-asset-state", "vuln-status", "check-type", "organization-id", "merge-num", "scanner-vuln-weight"}
	for _, f := range intFlags {
		if v, _ := cmd.Flags().GetInt(f); v != -1 {
			key := strings.ReplaceAll(f, "-", "_")
			if f == "rel-workflow-status" || f == "vuln-status" || f == "rel-asset-state" || f == "check-type" {
				params[key] = []map[string]interface{}{{"oper": "=", "target": v}}
			} else {
				params[key] = v
			}
		}
	}

	strFlags := []string{"vuln-tag-ids", "rel-asset-name", "rel-asset-business", "rel-asset-scope", "fix-remarks", "tag-name", "customize-tag-name"}
	for _, f := range strFlags {
		if v, _ := cmd.Flags().GetString(f); v != "" {
			key := strings.ReplaceAll(f, "-", "_")
			params[key] = v
		}
	}

	timeFlags := []string{"vuln-update-time", "dispose-fix-time", "vuln-last-time", "vuln-first-time", "plan-fix-time"}
	for _, f := range timeFlags {
		if v, _ := cmd.Flags().GetString(f); v != "" {
			key := strings.ReplaceAll(f, "-", "_")
			params[key] = []map[string]string{{"oper": "in", "target": v}}
		}
	}

	return params
}

func buildIpQueryParams(cmd *cobra.Command, params map[string]interface{}) {
	if v, _ := cmd.Flags().GetStringSlice("vuln-vpt-priority"); len(v) > 0 {
		var intSlice []int
		isInts := true
		for _, val := range v {
			if i, err := strconv.Atoi(val); err == nil {
				intSlice = append(intSlice, i)
			} else {
				isInts = false
			}
		}
		if isInts && len(intSlice) > 0 {
			params["vuln_vpt_priority"] = intSlice
		} else {
			params["vuln_vpt_priority"] = v
		}
	}

	if v, _ := cmd.Flags().GetString("vuln-ip"); v != "" {
		params["vuln_ip"] = []map[string]string{{"oper": "like", "target": v}}
	}
	if v, _ := cmd.Flags().GetString("protocol"); v != "" {
		params["protocol"] = v
	}
	if v, _ := cmd.Flags().GetInt("port"); v != -1 {
		params["port"] = v
	}
}

func buildWebQueryParams(cmd *cobra.Command, params map[string]interface{}) {
	if v, _ := cmd.Flags().GetInt("manager-id"); v != -1 {
		params["manager_id"] = v
	}
	if v, _ := cmd.Flags().GetString("vuln-url"); v != "" {
		params["vuln_url"] = v
	}
	if v, _ := cmd.Flags().GetString("vuln-url-domain"); v != "" {
		params["vuln_url_domain"] = v
	}
	if v, _ := cmd.Flags().GetInt("vuln-url-port"); v != -1 {
		params["vuln_url_port"] = v
	}
	if v, _ := cmd.Flags().GetString("vuln-url-scheme"); v != "" {
		params["vuln_url_scheme"] = v
	}
}

func NewVulnCmd(getClient func(cmd *cobra.Command) *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vuln",
		Short: "Manage Vulnerabilities",
	}

	ipCmd := &cobra.Command{
		Use:   "ip",
		Short: "List IP vulnerabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)
			params := buildCommonQueryParams(cmd)
			buildIpQueryParams(cmd, params)

			payload := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      "0",
				"method":  "ScanVulnIpService.SearchScanVulnIpList",
				"params":  params,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				return err
			}

			resp, err := c.Request("POST", "/pedestal/rpc", bytes.NewReader(payloadBytes))
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("output")
			renderer := client.NewRenderer(format, cmd.OutOrStdout())
			return renderer.Render(resp)
		},
	}
	addCommonVulnFlags(ipCmd)
	addIpVulnFlags(ipCmd)
	cmd.AddCommand(ipCmd)

	webCmd := &cobra.Command{
		Use:   "web",
		Short: "List Web vulnerabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)
			params := buildCommonQueryParams(cmd)
			buildWebQueryParams(cmd, params)

			payload := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      "0",
				"method":  "ScanVulnIpService.SearchScanVulnWebList",
				"params":  params,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				return err
			}

			resp, err := c.Request("POST", "/pedestal/rpc", bytes.NewReader(payloadBytes))
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("output")
			renderer := client.NewRenderer(format, cmd.OutOrStdout())
			return renderer.Render(resp)
		},
	}
	addCommonVulnFlags(webCmd)
	addWebVulnFlags(webCmd)
	cmd.AddCommand(webCmd)

	return cmd
}
