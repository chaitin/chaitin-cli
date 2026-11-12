package cmd

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/chaitin/chaitin-cli/products/insight/client"
	"github.com/spf13/cobra"
)

func addCommonAssetFlags(cmd *cobra.Command) {
	cmd.Flags().Int64("count", 20, "Number of items to return")
	cmd.Flags().Uint64("offset", 0, "Number of items to skip")

	// Arrays
	cmd.Flags().StringSlice("business-id", nil, "Filter by business ID")
	cmd.Flags().StringSlice("extra-owner-ids", nil, "Filter by extra owner IDs")

	// Ints
	cmd.Flags().Int("asset-state", -1, "Filter by asset state (-1 to ignore)")
	cmd.Flags().Int("online-status", -1, "Filter by online status (-1 to ignore)")
	cmd.Flags().Int("owner-id", -1, "Filter by owner ID (-1 to ignore)")
	cmd.Flags().Int("scope-id", -1, "Filter by scope ID (-1 to ignore)")
	cmd.Flags().Int("port", -1, "Filter by port (-1 to ignore)")

	// Bools
	cmd.Flags().String("auto-update", "", "Filter by auto update (true/false)")
	cmd.Flags().String("external", "", "Filter by external status (true/false)")

	// Strings
	cmd.Flags().String("source", "", "Filter by source")

	// Times
	cmd.Flags().String("find-at", "", "Filter by find time range (e.g. 1780588800-1780675200)")
	cmd.Flags().String("last-survival-at", "", "Filter by last survival time range")
	cmd.Flags().String("updated-at", "", "Filter by updated time range")
}

func addIpAssetFlags(cmd *cobra.Command) {
	// Arrays
	cmd.Flags().StringSlice("asset-risk", nil, "Filter by asset risk (e.g. high,low)")
	cmd.Flags().StringSlice("ip", nil, "Filter by IP address list")

	// Bools
	cmd.Flags().String("installed-host-agent", "", "Filter by installed host agent (true/false)")

	// Strings
	cmd.Flags().String("hostname", "", "Filter by hostname")
	cmd.Flags().String("mac", "", "Filter by MAC address")
	cmd.Flags().String("os-name", "", "Filter by OS name")
	cmd.Flags().String("protocol", "", "Filter by protocol")
	cmd.Flags().String("service-name", "", "Filter by service name")
}

func addWebAssetFlags(cmd *cobra.Command) {
	// Arrays
	cmd.Flags().StringSlice("site-url", nil, "Filter by site URL list")
	cmd.Flags().StringSlice("name", nil, "Filter by name list")

	// Strings
	cmd.Flags().String("app-name", "", "Filter by app name")
	cmd.Flags().String("container", "", "Filter by container")
	cmd.Flags().String("scheme", "", "Filter by scheme")
}

func buildCommonAssetFilter(cmd *cobra.Command, filter map[string]interface{}) {
	// Arrays
	if flags := []string{"business-id", "extra-owner-ids"}; len(flags) > 0 {
		for _, f := range flags {
			if v, _ := cmd.Flags().GetStringSlice(f); len(v) > 0 {
				var intSlice []int
				var strSlice []string
				isInts := true
				for _, val := range v {
					if i, err := strconv.Atoi(val); err == nil {
						intSlice = append(intSlice, i)
					} else {
						isInts = false
						strSlice = append(strSlice, val)
					}
				}
				key := strings.ReplaceAll(f, "-", "_")
				if isInts && len(intSlice) > 0 {
					filter[key] = intSlice
				} else {
					filter[key] = v
				}
			}
		}
	}

	// Ints
	intFlags := []string{"asset-state", "online-status", "owner-id", "scope-id", "port"}
	for _, f := range intFlags {
		if v, _ := cmd.Flags().GetInt(f); v != -1 {
			key := strings.ReplaceAll(f, "-", "_")
			filter[key] = v
		}
	}

	// Bools
	boolFlags := []string{"auto-update", "external"}
	for _, f := range boolFlags {
		if v, _ := cmd.Flags().GetString(f); v != "" {
			key := strings.ReplaceAll(f, "-", "_")
			if strings.ToLower(v) == "true" || v == "1" {
				filter[key] = true
			} else if strings.ToLower(v) == "false" || v == "0" {
				filter[key] = false
			}
		}
	}

	// Strings
	strFlags := []string{"source"}
	for _, f := range strFlags {
		if v, _ := cmd.Flags().GetString(f); v != "" {
			key := strings.ReplaceAll(f, "-", "_")
			filter[key] = v
		}
	}

	// Times (using oper: in)
	timeFlags := []string{"find-at", "last-survival-at", "updated-at"}
	for _, f := range timeFlags {
		if v, _ := cmd.Flags().GetString(f); v != "" {
			key := strings.ReplaceAll(f, "-", "_")
			filter[key] = []map[string]string{{"oper": "in", "target": v}}
		}
	}
}

func buildIpAssetFilter(cmd *cobra.Command, filter map[string]interface{}) {
	// Arrays
	if flags := []string{"asset-risk", "ip"}; len(flags) > 0 {
		for _, f := range flags {
			if v, _ := cmd.Flags().GetStringSlice(f); len(v) > 0 {
				var intSlice []int
				var strSlice []string
				isInts := true
				for _, val := range v {
					if i, err := strconv.Atoi(val); err == nil {
						intSlice = append(intSlice, i)
					} else {
						isInts = false
						strSlice = append(strSlice, val)
					}
				}
				key := strings.ReplaceAll(f, "-", "_")
				if isInts && len(intSlice) > 0 {
					filter[key] = intSlice
				} else {
					filter[key] = v
				}
			}
		}
	}

	// Bools
	if v, _ := cmd.Flags().GetString("installed-host-agent"); v != "" {
		if strings.ToLower(v) == "true" || v == "1" {
			filter["installed_host_agent"] = true
		} else if strings.ToLower(v) == "false" || v == "0" {
			filter["installed_host_agent"] = false
		}
	}

	// Strings
	strFlags := []string{"hostname", "mac", "os-name", "protocol", "service-name"}
	for _, f := range strFlags {
		if v, _ := cmd.Flags().GetString(f); v != "" {
			key := strings.ReplaceAll(f, "-", "_")
			filter[key] = v
		}
	}
}

func buildWebAssetFilter(cmd *cobra.Command, filter map[string]interface{}) {
	// Arrays
	if flags := []string{"site-url", "name"}; len(flags) > 0 {
		for _, f := range flags {
			if v, _ := cmd.Flags().GetStringSlice(f); len(v) > 0 {
				key := strings.ReplaceAll(f, "-", "_")
				filter[key] = v
			}
		}
	}

	// Strings
	strFlags := []string{"app-name", "container", "scheme"}
	for _, f := range strFlags {
		if v, _ := cmd.Flags().GetString(f); v != "" {
			key := strings.ReplaceAll(f, "-", "_")
			filter[key] = v
		}
	}
}

func NewAssetCmd(getClient func(cmd *cobra.Command) *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset",
		Short: "Manage Assets",
	}

	ipCmd := &cobra.Command{
		Use:   "ip",
		Short: "List IP assets",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)
			count, _ := cmd.Flags().GetInt64("count")
			offset, _ := cmd.Flags().GetUint64("offset")

			filter := map[string]interface{}{}
			buildCommonAssetFilter(cmd, filter)
			buildIpAssetFilter(cmd, filter)

			payload := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      "0",
				"method":  "AssetMgrService.IpAssetList",
				"params": map[string]interface{}{
					"count":  count,
					"offset": offset,
					"filter": filter,
				},
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
	addCommonAssetFlags(ipCmd)
	addIpAssetFlags(ipCmd)
	cmd.AddCommand(ipCmd)

	webCmd := &cobra.Command{
		Use:   "web",
		Short: "List Web assets",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)
			count, _ := cmd.Flags().GetInt64("count")
			offset, _ := cmd.Flags().GetUint64("offset")

			filter := map[string]interface{}{}
			buildCommonAssetFilter(cmd, filter)
			buildWebAssetFilter(cmd, filter)

			payload := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      "0",
				"method":  "AssetMgrService.WebsiteAssetList",
				"params": map[string]interface{}{
					"count":  count,
					"offset": offset,
					"filter": filter,
				},
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
	addCommonAssetFlags(webCmd)
	addWebAssetFlags(webCmd)
	cmd.AddCommand(webCmd)

	softwareCmd := &cobra.Command{
		Use:   "software",
		Short: "List Software assets overview",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)
			count, _ := cmd.Flags().GetInt64("count")
			offset, _ := cmd.Flags().GetUint64("offset")

			// Assuming software asset lists might use similar common filters
			filter := map[string]interface{}{}
			buildCommonAssetFilter(cmd, filter)

			payload := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      "0",
				"method":  "AssetMgrService.SoftwareAssetOverviewList",
				"params": map[string]interface{}{
					"count":  count,
					"offset": offset,
					"filter": filter,
				},
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
	addCommonAssetFlags(softwareCmd)
	cmd.AddCommand(softwareCmd)

	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "List Asset Tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)
			count, _ := cmd.Flags().GetInt64("count")
			offset, _ := cmd.Flags().GetUint64("offset")

			filter := map[string]interface{}{}

			payload := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      "0",
				"method":  "AssetMgrService.AssetTagList",
				"params": map[string]interface{}{
					"count":  count,
					"offset": offset,
					"filter": filter,
				},
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
	cmd.AddCommand(tagCmd)

	businessCmd := &cobra.Command{
		Use:   "business",
		Short: "List Asset Business Systems",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := getClient(cmd)
			count, _ := cmd.Flags().GetInt64("count")
			offset, _ := cmd.Flags().GetUint64("offset")

			filter := map[string]interface{}{}

			payload := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      "0",
				"method":  "AssetMgrService.AssetBusinessList",
				"params": map[string]interface{}{
					"count":  count,
					"offset": offset,
					"filter": filter,
				},
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
	cmd.AddCommand(businessCmd)

	return cmd
}
