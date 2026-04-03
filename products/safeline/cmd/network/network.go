package network

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaitin/workspace-cli/products/safeline/cmd"
	"github.com/spf13/cobra"
)

// NewCommand creates the network command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "network",
		Short: "Network configuration commands",
		Long:  "Commands for viewing SafeLine network configuration (hardware mode only).",
	}
	c.AddCommand(newWorkgroupCmd())
	c.AddCommand(newInterfaceCmd())
	c.AddCommand(newGatewayCmd())
	c.AddCommand(newRouteCmd())
	return c
}

func newWorkgroupCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "workgroup",
		Short:   "Workgroup commands",
		Long:    "Commands for viewing workgroups (network interface groups).",
		Aliases: []string{"wg"},
	}
	c.AddCommand(newWorkgroupListCmd())
	c.AddCommand(newWorkgroupGetCmd())
	return c
}

func newWorkgroupListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List workgroups",
		Long: `List all workgroups.

Note: This command only works in hardware deployment modes.
In software mode, it will return an error.

Examples:
  safeline network workgroup list`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := checkHardwareMode(); err != nil {
				return err
			}

			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/WorkGroupAPI", nil, nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output
			var results []struct {
				Name    string `json:"name"`
				Comment string `json:"comment"`
				Mode    string `json:"mode"`
				Type    string `json:"type"`
				Links   []struct {
					Name   string `json:"name"`
					Direct string `json:"direct"`
				} `json:"links"`
			}
			if err := json.Unmarshal(env.Data, &results); err != nil {
				return err
			}

			headers := []string{"Name", "Comment", "Mode", "Type", "Links"}
			var rows [][]string
			for _, r := range results {
				var linkNames []string
				for _, link := range r.Links {
					linkNames = append(linkNames, link.Name)
				}
				rows = append(rows, []string{
					r.Name,
					r.Comment,
					r.Mode,
					r.Type,
					strings.Join(linkNames, ", "),
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}
}

func newWorkgroupGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get workgroup details",
		Long: `Get workgroup details by name.

Note: This command only works in hardware deployment modes.

Examples:
  safeline network workgroup get "group1"`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := checkHardwareMode(); err != nil {
				return err
			}

			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/WorkGroupAPI", nil, map[string]string{"name": args[0]})
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// API returns array
			var results []struct {
				Name    string `json:"name"`
				Comment string `json:"comment"`
				Mode    string `json:"mode"`
				Type    string `json:"type"`
				Links   []struct {
					Name   string `json:"name"`
					Direct string `json:"direct"`
				} `json:"links"`
			}
			if err := json.Unmarshal(env.Data, &results); err != nil {
				return err
			}

			if len(results) == 0 {
				return fmt.Errorf("workgroup %q not found", args[0])
			}

			result := results[0]
			var linkDetails []string
			for _, link := range result.Links {
				linkDetails = append(linkDetails, fmt.Sprintf("%s(%s)", link.Name, link.Direct))
			}

			cmd.PrintKeyValue(map[string]string{
				"Name":    result.Name,
				"Comment": result.Comment,
				"Mode":    result.Mode,
				"Type":    result.Type,
				"Links":   strings.Join(linkDetails, ", "),
			})
			return nil
		},
	}
}

func newInterfaceCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "interface",
		Short:   "Network interface commands",
		Long:    "Commands for viewing network interfaces.",
		Aliases: []string{"if"},
	}
	c.AddCommand(newInterfaceListCmd())
	c.AddCommand(newInterfaceIPCmd())
	return c
}

func newInterfaceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List network interfaces",
		Long: `List all network interfaces.

Note: This command only works in hardware deployment modes.

Examples:
  safeline network interface list`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := checkHardwareMode(); err != nil {
				return err
			}

			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/LinkAPI", nil, nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output
			var results []struct {
				Name    string `json:"name"`
				Type    string `json:"type"`
				Comment string `json:"comment"`
				Attr    struct {
					State string `json:"state"`
				} `json:"attr"`
			}
			if err := json.Unmarshal(env.Data, &results); err != nil {
				return err
			}

			headers := []string{"Name", "Type", "State", "Comment"}
			var rows [][]string
			for _, r := range results {
				rows = append(rows, []string{
					r.Name,
					r.Type,
					r.Attr.State,
					r.Comment,
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}
}

func newInterfaceIPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ip <name>",
		Short: "Get interface IP information",
		Long: `Get IP information for a network interface.

Note: This command only works in hardware deployment modes.

Examples:
  safeline network interface ip eth0`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := checkHardwareMode(); err != nil {
				return err
			}

			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/LinkIPAPI", nil, map[string]string{"name": args[0]})
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// API returns array of addresses
			var addrs []struct {
				Addr string `json:"addr"`
				Mask string `json:"mask"`
				Mode string `json:"mode"`
			}
			if err := json.Unmarshal(env.Data, &addrs); err != nil {
				return err
			}

			var ipv4Addrs, ipv6Addrs []string
			for _, addr := range addrs {
				fullAddr := fmt.Sprintf("%s/%s", addr.Addr, addr.Mask)
				if strings.Contains(addr.Addr, ":") {
					ipv6Addrs = append(ipv6Addrs, fullAddr)
				} else {
					ipv4Addrs = append(ipv4Addrs, fullAddr)
				}
			}

			cmd.PrintKeyValue(map[string]string{
				"Interface": args[0],
				"IPv4":      strings.Join(ipv4Addrs, ", "),
				"IPv6":      strings.Join(ipv6Addrs, ", "),
			})
			return nil
		},
	}
}

// OperationMode represents the deployment mode.
type OperationMode string

const (
	ModeSoftwareReverseProxy        OperationMode = "software_reverse_proxy"
	ModeHardwareReverseProxy        OperationMode = "hardware_reverse_proxy"
	ModeHardwareTransparentProxy    OperationMode = "hardware_transparent_proxy"
	ModeHardwarePortMirroring       OperationMode = "hardware_port_mirroring"
	ModeHardwareTransparentBridging OperationMode = "hardware_transparent_bridging"
	ModeHardwareTrafficDetection    OperationMode = "hardware_traffic_detection"
)

// checkHardwareMode checks if the current deployment mode supports network commands.
func checkHardwareMode() error {
	cl := cmd.NewClient()

	env, err := cl.Do("GET", "/api/ServerControlledConfigAPI", nil, map[string]string{"type": "operation_mode"})
	if err != nil {
		return fmt.Errorf("failed to get operation mode: %w", err)
	}

	// API returns array like ["Software Reverse Proxy"]
	var modes []string
	if err := json.Unmarshal(env.Data, &modes); err != nil {
		return fmt.Errorf("failed to parse operation mode: %w", err)
	}

	if len(modes) > 0 && modes[0] == "Software Reverse Proxy" {
		return fmt.Errorf("该功能在软件模式下不支持")
	}

	return nil
}

func newGatewayCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "gateway",
		Short:   "Default gateway commands",
		Long:    "Commands for viewing default gateway configuration.",
		Aliases: []string{"gw"},
	}
	c.AddCommand(newGatewayGetCmd())
	return c
}

func newGatewayGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get default gateway",
		Long: `Get the current default gateway configuration.

Examples:
  safeline network gateway get`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := checkHardwareMode(); err != nil {
				return err
			}

			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/DefaultGatewayAPI", nil, nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Key-value output
			var result struct {
				IPv4Gateway string `json:"ipv4_gateway"`
				IPv6Gateway string `json:"ipv6_gateway"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}

			cmd.PrintKeyValue(map[string]string{
				"IPv4 Gateway": result.IPv4Gateway,
				"IPv6 Gateway": result.IPv6Gateway,
			})
			return nil
		},
	}
}

func newRouteCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "route",
		Short:   "Static route commands",
		Long:    "Commands for viewing static routes.",
		Aliases: []string{"sr"},
	}
	c.AddCommand(newRouteListCmd())
	return c
}

func newRouteListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List static routes",
		Long: `List all static routes.

Examples:
  safeline network route list`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := checkHardwareMode(); err != nil {
				return err
			}

			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/StaticRouteAPI", nil, nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output
			var results []struct {
				ID      string `json:"id"`
				Addr    string `json:"addr"`
				Mask    string `json:"mask"`
				Gateway string `json:"gateway"`
			}
			if err := json.Unmarshal(env.Data, &results); err != nil {
				return err
			}

			headers := []string{"ID", "Address", "Mask", "Gateway"}
			var rows [][]string
			for _, r := range results {
				rows = append(rows, []string{
					r.ID,
					r.Addr,
					r.Mask,
					r.Gateway,
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}
}
