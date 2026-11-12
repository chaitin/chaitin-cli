package site

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	safelinecmd "github.com/chaitin/chaitin-cli/products/safeline/cmd"
	safelineruntime "github.com/chaitin/chaitin-cli/products/safeline/runtime"
	"github.com/spf13/cobra"
)

type healthOptions struct {
	Type     string
	Host     string
	Port     int
	Path     string
	Method   string
	Expect   []string
	Interval int
	Timeout  int
	Fall     int
	Rise     int
	Yes      bool
	Check    bool
}

type healthSummary struct {
	OK         bool             `json:"ok"`
	Operation  string           `json:"operation"`
	Endpoint   string           `json:"endpoint"`
	SiteID     any              `json:"site_id,omitempty"`
	SiteName   string           `json:"site_name,omitempty"`
	SiteStatus string           `json:"site_status"`
	Config     map[string]any   `json:"config"`
	Backends   []backendSummary `json:"backends"`
}

type backendSummary struct {
	Address  string `json:"address"`
	Protocol string `json:"protocol,omitempty"`
	Status   string `json:"status"`
}

func newHealthCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "health",
		Short: "Manage SafeLine site health checks",
		Long: `Manage SafeLine site backend health checks.

Health-check configuration is supported for Software Reverse Proxy, Hardware
Reverse Proxy and Hardware Router Proxy modes. Other modes can be inspected
when the server returns health fields, but cannot be configured by this CLI.`,
	}
	c.AddCommand(newHealthGetCmd())
	c.AddCommand(newHealthEnableCmd())
	c.AddCommand(newHealthDisableCmd())
	return c
}

func newHealthGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show site health-check config and status",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx, err := resolveSiteContext()
			if err != nil {
				return err
			}
			siteData, err := getSiteData(ctx.Endpoint, args[0])
			if err != nil {
				return err
			}
			summary, err := buildHealthSummary(ctx.Endpoint, siteData)
			if err != nil {
				return err
			}
			return safelinecmd.PrintResult(c, summary)
		},
	}
}

func newHealthEnableCmd() *cobra.Command {
	opts := healthOptions{
		Type:     "http",
		Port:     0,
		Path:     "/",
		Method:   "GET",
		Expect:   []string{"http_2xx", "http_3xx"},
		Interval: 30000,
		Timeout:  1000,
		Fall:     5,
		Rise:     2,
	}
	c := &cobra.Command{
		Use:   "enable <id>",
		Short: "Enable site backend health check",
		Long: `Enable site backend health check.

Common scenarios:

Layer 7 HTTP health check:
  Validates that the backend web service responds on a path with an expected
  HTTP status class.

  chaitin-cli safeline site health enable <id> \
    --type http --path /healthz --method GET \
    --expect http_2xx --expect http_3xx --yes

  Related options:
    --type http
    --path PATH
    --method GET|HEAD
    --host HOST
    --port PORT       0 means use the backend server port
    --expect CLASS    http_2xx, http_3xx, http_4xx, http_5xx

Layer 4 TCP health check:
  Validates that the backend address and port are reachable. It does not send
  HTTP requests and does not use --path, --method, --host, or --expect.

  chaitin-cli safeline site health enable <id> --type tcp --yes

TLS handshake health check:
  Validates that the backend can complete a TLS handshake. It does not check
  an HTTP path or HTTP status code.

  chaitin-cli safeline site health enable <id> --type ssl_hello --yes

Timing options for all types:
  --interval MS       health-check interval, minimum 1000
  --timeout MS        response timeout, 1000 to 300000
  --fall N            failures before marking unhealthy, 1 to 10
  --rise N            successes before marking healthy, 1 to 10`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx, err := resolveSiteContext()
			if err != nil {
				return err
			}
			if err := validateHealthMode(ctx.OperationMode); err != nil {
				return err
			}
			cfg, err := buildHealthCheckConfig(opts)
			if err != nil {
				return err
			}
			return updateSiteHealth(c, ctx.Endpoint, args[0], cfg, opts.Yes, opts.Check, "site.health.enable")
		},
	}
	addHealthEnableFlags(c, &opts)
	return c
}

func newHealthDisableCmd() *cobra.Command {
	opts := healthOptions{}
	c := &cobra.Command{
		Use:   "disable <id>",
		Short: "Disable site backend health check",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx, err := resolveSiteContext()
			if err != nil {
				return err
			}
			if err := validateHealthMode(ctx.OperationMode); err != nil {
				return err
			}
			return updateSiteHealth(c, ctx.Endpoint, args[0], map[string]any{"is_enabled": false}, opts.Yes, opts.Check, "site.health.disable")
		},
	}
	c.Flags().BoolVar(&opts.Yes, "yes", false, "Confirm write operation")
	c.Flags().BoolVar(&opts.Check, "check", false, "Print request without writing")
	return c
}

func addHealthEnableFlags(c *cobra.Command, opts *healthOptions) {
	c.Flags().StringVar(&opts.Type, "type", opts.Type, "Health-check type: http, tcp, ssl_hello")
	c.Flags().StringVar(&opts.Host, "host", "", "HTTP health-check Host header")
	c.Flags().IntVar(&opts.Port, "port", opts.Port, "HTTP health-check port, 0 means backend port")
	c.Flags().StringVar(&opts.Path, "path", opts.Path, "HTTP health-check path")
	c.Flags().StringVar(&opts.Method, "method", opts.Method, "HTTP health-check method: GET or HEAD")
	c.Flags().StringSliceVar(&opts.Expect, "expect", opts.Expect, "HTTP expected alive status class, repeatable or comma-separated: http_2xx, http_3xx, http_4xx, http_5xx")
	c.Flags().IntVar(&opts.Interval, "interval", opts.Interval, "Health-check interval in milliseconds")
	c.Flags().IntVar(&opts.Timeout, "timeout", opts.Timeout, "Health-check timeout in milliseconds")
	c.Flags().IntVar(&opts.Fall, "fall", opts.Fall, "Unhealthy threshold")
	c.Flags().IntVar(&opts.Rise, "rise", opts.Rise, "Healthy threshold")
	c.Flags().BoolVar(&opts.Yes, "yes", false, "Confirm write operation")
	c.Flags().BoolVar(&opts.Check, "check", false, "Print request without writing")
}

func resolveSiteContext() (safelineruntime.Context, error) {
	return safelineruntime.ResolveContext(safelinecmd.NewClient(), safelineruntime.Options{
		VersionOverride:       safelinecmd.VersionOverride,
		OperationModeOverride: safelinecmd.OperationModeOverride,
		ConfigVersion:         safelinecmd.ConfigVersion,
		ConfigOperationMode:   safelinecmd.ConfigOperationMode,
	})
}

func validateHealthMode(mode safelineruntime.OperationMode) error {
	switch mode {
	case safelineruntime.ModeSoftwareReverseProxy, safelineruntime.ModeHardwareReverseProxy, safelineruntime.ModeHardwareRouterProxy:
		return nil
	default:
		return fmt.Errorf("operation mode %q does not support site health-check configuration", mode)
	}
}

func buildHealthCheckConfig(opts healthOptions) (map[string]any, error) {
	checkType := strings.ToLower(strings.TrimSpace(opts.Type))
	if checkType == "" {
		checkType = "http"
	}
	if err := validateHealthTimings(opts); err != nil {
		return nil, err
	}
	cfg := map[string]any{
		"is_enabled": true,
		"check_type": checkType,
		"interval":   opts.Interval,
		"timeout":    opts.Timeout,
		"fall":       opts.Fall,
		"rise":       opts.Rise,
	}
	switch checkType {
	case "http":
		if len(opts.Expect) == 0 {
			return nil, fmt.Errorf("http health check requires at least one --expect")
		}
		method := strings.ToUpper(strings.TrimSpace(opts.Method))
		if method == "" {
			method = "GET"
		}
		if method != "GET" && method != "HEAD" {
			return nil, fmt.Errorf("--method must be GET or HEAD")
		}
		if opts.Port < 0 || opts.Port > 65535 {
			return nil, fmt.Errorf("--port must be between 0 and 65535")
		}
		path := opts.Path
		if path == "" {
			path = "/"
		}
		cfg["host"] = opts.Host
		cfg["port"] = opts.Port
		cfg["path"] = path
		cfg["method"] = method
		cfg["check_http_expect_alive"] = opts.Expect
	case "tcp", "ssl_hello":
	default:
		return nil, fmt.Errorf("--type must be one of http, tcp, ssl_hello")
	}
	return cfg, nil
}

func validateHealthTimings(opts healthOptions) error {
	if opts.Interval < 1000 {
		return fmt.Errorf("--interval must be at least 1000")
	}
	if opts.Timeout < 1000 || opts.Timeout > 300000 {
		return fmt.Errorf("--timeout must be between 1000 and 300000")
	}
	if opts.Fall < 1 || opts.Fall > 10 {
		return fmt.Errorf("--fall must be between 1 and 10")
	}
	if opts.Rise < 1 || opts.Rise > 10 {
		return fmt.Errorf("--rise must be between 1 and 10")
	}
	return nil
}

func getSiteData(endpoint, rawID string) (map[string]any, error) {
	if _, err := parsePositiveID(rawID); err != nil {
		return nil, err
	}
	env, err := safelinecmd.NewClient().Do("GET", endpoint, nil, map[string]string{"id": rawID})
	if err != nil {
		return nil, err
	}
	var siteData map[string]any
	if err := json.Unmarshal(env.Data, &siteData); err != nil {
		return nil, fmt.Errorf("parse site response: %w", err)
	}
	return siteData, nil
}

func updateSiteHealth(c *cobra.Command, endpoint, rawID string, healthConfig map[string]any, yes, check bool, operation string) error {
	siteData, err := getSiteData(endpoint, rawID)
	if err != nil {
		return err
	}
	updated, err := applyHealthConfig(siteData, healthConfig)
	if err != nil {
		return err
	}
	preview := map[string]any{"ok": true, "operation": operation + ".check", "data": map[string]any{"endpoint": endpoint, "payload": updated}}
	if check || safelinecmd.IsDryRun() {
		return safelinecmd.PrintResult(c, preview)
	}
	if !yes {
		return fmt.Errorf("site health update is a write operation; re-run with --yes or use --check to preview")
	}
	body, err := json.Marshal(updated)
	if err != nil {
		return fmt.Errorf("marshal site health payload: %w", err)
	}
	env, err := safelinecmd.NewClient().Do("PUT", endpoint, bytes.NewReader(body), nil)
	if err != nil {
		return err
	}
	return safelinecmd.PrintResult(c, map[string]any{"ok": true, "operation": operation, "data": map[string]any{"endpoint": endpoint, "response": env.Data}})
}

func applyHealthConfig(siteData map[string]any, healthConfig map[string]any) (map[string]any, error) {
	backend, ok := siteData["backend_config"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("site does not include backend_config")
	}
	if backend["type"] != "proxy" {
		return nil, fmt.Errorf("site health checks are supported only proxy backend sites")
	}
	backend["health_check_config"] = healthConfig
	siteData["backend_config"] = backend
	return siteData, nil
}

func buildHealthSummary(endpoint string, siteData map[string]any) (healthSummary, error) {
	backend, ok := siteData["backend_config"].(map[string]any)
	if !ok {
		return healthSummary{}, fmt.Errorf("site does not include backend_config")
	}
	config, _ := backend["health_check_config"].(map[string]any)
	summary := healthSummary{
		OK:         true,
		Operation:  "site.health.get",
		Endpoint:   endpoint,
		SiteID:     siteData["id"],
		SiteName:   stringValue(siteData["name"]),
		SiteStatus: stringValue(siteData["health_check_status"]),
		Config:     config,
		Backends:   backendHealthSummaries(backend["servers"]),
	}
	return summary, nil
}

func backendHealthSummaries(raw any) []backendSummary {
	servers, ok := raw.([]any)
	if !ok {
		return nil
	}
	backends := make([]backendSummary, 0, len(servers))
	for _, rawServer := range servers {
		server, ok := rawServer.(map[string]any)
		if !ok {
			continue
		}
		host := stringValue(server["host"])
		port := portString(server["port"])
		address := host
		if port != "" {
			address = host + ":" + port
		}
		backends = append(backends, backendSummary{
			Address:  address,
			Protocol: stringValue(server["protocol"]),
			Status:   stringValue(server["health_check_status"]),
		})
	}
	return backends
}

func parsePositiveID(rawID string) (int, error) {
	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("site id must be a positive integer")
	}
	return id, nil
}

func stringValue(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func portString(v any) string {
	switch p := v.(type) {
	case float64:
		return strconv.Itoa(int(p))
	case int:
		return strconv.Itoa(p)
	case string:
		return p
	default:
		return stringValue(v)
	}
}
