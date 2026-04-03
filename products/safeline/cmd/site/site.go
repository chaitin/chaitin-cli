package site

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaitin/workspace-cli/products/safeline/cmd"
	"github.com/chaitin/workspace-cli/products/safeline/pkg/client"
	"github.com/spf13/cobra"
)

// OperationMode represents the deployment mode.
type OperationMode string

const (
	ModeSoftwareReverseProxy        OperationMode = "Software Reverse Proxy"
	ModeHardwareReverseProxy        OperationMode = "Hardware Reverse Proxy"
	ModeSoftwareClusterReverseProxy OperationMode = "Software Cluster Reverse Proxy"
	ModeSoftwarePortMirroring       OperationMode = "Software Port Mirroring"
	ModeHardwareTransparentProxy    OperationMode = "Hardware Transparent Proxy"
	ModeHardwareTransparentBridging OperationMode = "Hardware Transparent Bridging"
	ModeHardwarePortMirroring       OperationMode = "Hardware Port Mirroring"
	ModeHardwareTrafficDetection    OperationMode = "Hardware Traffic Detection"
	ModeHardwareRouterProxy         OperationMode = "Hardware Router Proxy"
)

// Port handles both int (Software Reverse Proxy) and string (Hardware modes) port formats.
type Port struct {
	RawPort interface{} `json:"-"` // int or string
	SSL     bool        `json:"ssl"`
	HTTP2   bool        `json:"http2"`
	SNI     bool        `json:"sni"`
	NonHTTP bool        `json:"non_http"`
}

// UnmarshalJSON handles both int and string port formats.
func (p *Port) UnmarshalJSON(data []byte) error {
	// Define a temporary struct with port as json.RawMessage
	var temp struct {
		Port    json.RawMessage `json:"port"`
		SSL     bool            `json:"ssl"`
		HTTP2   bool            `json:"http2"`
		SNI     bool            `json:"sni"`
		NonHTTP bool            `json:"non_http"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	p.SSL = temp.SSL
	p.HTTP2 = temp.HTTP2
	p.SNI = temp.SNI
	p.NonHTTP = temp.NonHTTP

	// Try to parse port as int first
	var portInt int
	if err := json.Unmarshal(temp.Port, &portInt); err == nil {
		p.RawPort = portInt
		return nil
	}

	// Try to parse as string
	var portStr string
	if err := json.Unmarshal(temp.Port, &portStr); err != nil {
		return err
	}
	p.RawPort = portStr
	return nil
}

// String returns the formatted port string.
func (p Port) String() string {
	var portStr string
	switch v := p.RawPort.(type) {
	case int:
		portStr = fmt.Sprintf("%d", v)
	case string:
		portStr = v
	default:
		portStr = fmt.Sprintf("%v", p.RawPort)
	}
	if p.SSL {
		portStr += "(SSL)"
	}
	return portStr
}

// Ports is a slice of Port with helper methods.
type Ports []Port

// Strings returns formatted port strings.
func (ps Ports) Strings() []string {
	var result []string
	for _, p := range ps {
		result = append(result, p.String())
	}
	return result
}

// NewCommand creates the site command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "site",
		Short: "Site management commands",
		Long:  "Commands for managing SafeLine sites (websites).",
	}
	c.AddCommand(newListCmd())
	c.AddCommand(newGetCmd())
	c.AddCommand(newUpdateCmd())
	c.AddCommand(newEnableCmd())
	c.AddCommand(newDisableCmd())
	return c
}

// serverConfig represents the response from ServerControlledConfigAPI
type serverConfig struct {
	OperationMode []string `json:"operation_mode"`
}

// getOperationMode detects the current deployment mode.
func getOperationMode(cl *client.Client) (OperationMode, error) {
	env, err := cl.Do("GET", "/api/ServerControlledConfigAPI", nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get operation mode: %w", err)
	}

	var config serverConfig
	if err := json.Unmarshal(env.Data, &config); err != nil {
		return "", fmt.Errorf("failed to parse operation mode: %w", err)
	}

	if len(config.OperationMode) == 0 {
		return ModeSoftwareReverseProxy, nil
	}

	return OperationMode(config.OperationMode[0]), nil
}

// getAPIPath returns the appropriate API path based on deployment mode.
func getAPIPath(mode OperationMode) string {
	switch mode {
	case ModeSoftwareReverseProxy:
		return "/api/SoftwareReverseProxyWebsiteAPI"
	case ModeHardwareReverseProxy:
		return "/api/HardwareReverseProxyWebsiteAPI"
	case ModeSoftwareClusterReverseProxy:
		return "/api/SoftwareClusterReverseProxyWebsiteAPI"
	case ModeSoftwarePortMirroring:
		return "/api/SoftwarePortMirroringWebsiteAPI"
	case ModeHardwareTransparentProxy:
		return "/api/HardwareTransparentProxyWebsiteAPI"
	case ModeHardwareTransparentBridging:
		return "/api/HardwareTransparentBridgingWebsiteAPI"
	case ModeHardwarePortMirroring:
		return "/api/HardwarePortMirroringWebsiteAPI"
	case ModeHardwareTrafficDetection:
		return "/api/HardwareTrafficDetectionWebsiteAPI"
	case ModeHardwareRouterProxy:
		return "/api/HardwareReverseProxyWebsiteAPI" // Use HardwareReverseProxy API
	default:
		return "/api/SoftwareReverseProxyWebsiteAPI"
	}
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all sites",
		Long: `List all sites.

The command automatically detects the deployment mode and uses the appropriate API.

Examples:
  safeline site list
  safeline site list --indent`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			// Detect deployment mode
			mode, err := getOperationMode(cl)
			if err != nil {
				return err
			}

			apiPath := getAPIPath(mode)
			env, err := cl.Do("GET", apiPath, nil, nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output
			var results []struct {
				ID          int      `json:"id"`
				Name        string   `json:"name"`
				IsEnabled   bool     `json:"is_enabled"`
				ServerNames []string `json:"server_names"`
				IP          []string `json:"ip"`
				Ports       Ports    `json:"ports"`
				URLPaths    []struct {
					Op      string `json:"op"`
					URLPath string `json:"url_path"`
				} `json:"url_paths"`
			}
			if err := json.Unmarshal(env.Data, &results); err != nil {
				return err
			}

			headers := []string{"ID", "Name", "Enabled", "Domains", "IP", "Ports", "URL Paths"}
			var rows [][]string
			for _, r := range results {
				enabled := "No"
				if r.IsEnabled {
					enabled = "Yes"
				}

				// Format URL paths
				var urlPaths []string
				for _, up := range r.URLPaths {
					urlPaths = append(urlPaths, fmt.Sprintf("%s %s", up.Op, up.URLPath))
				}

				rows = append(rows, []string{
					fmt.Sprintf("%d", r.ID),
					r.Name,
					enabled,
					strings.Join(r.ServerNames, ", "),
					strings.Join(r.IP, ", "),
					strings.Join(r.Ports.Strings(), ", "),
					strings.Join(urlPaths, ", "),
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get site details",
		Long: `Get site details by ID.

The command automatically detects the deployment mode and uses the appropriate API.

Examples:
  safeline site get 1
  safeline site get 1 --indent`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			mode, err := getOperationMode(cl)
			if err != nil {
				return err
			}

			apiPath := getAPIPath(mode)
			env, err := cl.Do("GET", apiPath, nil, map[string]string{"id": args[0]})
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Key-value output for single site
			var result struct {
				ID          int      `json:"id"`
				Name        string   `json:"name"`
				IsEnabled   bool     `json:"is_enabled"`
				ServerNames []string `json:"server_names"`
				IP          []string `json:"ip"`
				Ports       Ports    `json:"ports"`
				URLPaths    []struct {
					Op      string `json:"op"`
					URLPath string `json:"url_path"`
				} `json:"url_paths"`
				PolicyGroup *int   `json:"policy_group"`
				Comment     string `json:"remark"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}

			enabled := "No"
			if result.IsEnabled {
				enabled = "Yes"
			}

			// Format URL paths
			var urlPaths []string
			for _, up := range result.URLPaths {
				urlPaths = append(urlPaths, fmt.Sprintf("%s %s", up.Op, up.URLPath))
			}

			// Policy group
			pgStr := "None"
			if result.PolicyGroup != nil && *result.PolicyGroup > 0 {
				pgStr = fmt.Sprintf("%d", *result.PolicyGroup)
			}

			cmd.PrintKeyValue(map[string]string{
				"ID":           fmt.Sprintf("%d", result.ID),
				"Name":         result.Name,
				"Enabled":      enabled,
				"Domains":      strings.Join(result.ServerNames, ", "),
				"IP":           strings.Join(result.IP, ", "),
				"Ports":        strings.Join(result.Ports.Strings(), ", "),
				"URL Paths":    strings.Join(urlPaths, ", "),
				"Policy Group": pgStr,
				"Comment":      result.Comment,
			})
			return nil
		},
	}
}

func newUpdateCmd() *cobra.Command {
	var policyGroup int

	c := &cobra.Command{
		Use:   "update <id>",
		Short: "Update site policy group",
		Long: `Update the policy group associated with a site.

This command only allows modifying the policy_group field.

Examples:
  safeline site update 1 --policy-group 3
  safeline site update 1 --policy-group 0  # Set to none`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			mode, err := getOperationMode(cl)
			if err != nil {
				return err
			}

			apiPath := getAPIPath(mode)

			// Get current site data first
			env, err := cl.Do("GET", apiPath, nil, map[string]string{"id": args[0]})
			if err != nil {
				return err
			}

			// Modify policy_group in the raw JSON data
			var siteData map[string]interface{}
			if err := json.Unmarshal(env.Data, &siteData); err != nil {
				return err
			}

			// Update policy_group
			if policyGroup == 0 {
				siteData["policy_group"] = nil
			} else {
				siteData["policy_group"] = policyGroup
			}

			// Marshal back to JSON
			body, err := json.Marshal(siteData)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] PUT %s\n", apiPath)
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			env, err = cl.Do("PUT", apiPath, strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			var result struct {
				ID          int    `json:"id"`
				Name        string `json:"name"`
				PolicyGroup *int   `json:"policy_group"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}

			pgStr := "None"
			if result.PolicyGroup != nil && *result.PolicyGroup > 0 {
				pgStr = fmt.Sprintf("%d", *result.PolicyGroup)
			}
			cmd.PrintKeyValue(map[string]string{
				"ID":           fmt.Sprintf("%d", result.ID),
				"Name":         result.Name,
				"Policy Group": pgStr,
			})
			return nil
		},
	}

	c.Flags().IntVar(&policyGroup, "policy-group", 0, "Policy group ID (0 for none)")

	return c
}

func newEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <id>",
		Short: "Enable a site",
		Long: `Enable a site by ID.

Examples:
  safeline site enable 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			body := fmt.Sprintf(`{"id__in": [%s], "action": "enable"}`, args[0])

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] PUT /api/EnableDisableWebsiteAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", body)
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("PUT", "/api/EnableDisableWebsiteAPI", strings.NewReader(body), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("Site %s enabled successfully\n", args[0])
			return nil
		},
	}
}

func newDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <id>",
		Short: "Disable a site",
		Long: `Disable a site by ID.

Examples:
  safeline site disable 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			body := fmt.Sprintf(`{"id__in": [%s], "action": "disable"}`, args[0])

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] PUT /api/EnableDisableWebsiteAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", body)
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("PUT", "/api/EnableDisableWebsiteAPI", strings.NewReader(body), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("Site %s disabled successfully\n", args[0])
			return nil
		},
	}
}
