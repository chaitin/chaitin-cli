package safeline

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/chaitin/chaitin-cli/egressproxy"
	cmdpkg "github.com/chaitin/chaitin-cli/products/safeline/cmd"
	"github.com/chaitin/chaitin-cli/products/safeline/cmd/acl"
	"github.com/chaitin/chaitin-cli/products/safeline/cmd/cert"
	"github.com/chaitin/chaitin-cli/products/safeline/cmd/inspect"
	"github.com/chaitin/chaitin-cli/products/safeline/cmd/ipgroup"
	"github.com/chaitin/chaitin-cli/products/safeline/cmd/log"
	"github.com/chaitin/chaitin-cli/products/safeline/cmd/network"
	"github.com/chaitin/chaitin-cli/products/safeline/cmd/policygroup"
	"github.com/chaitin/chaitin-cli/products/safeline/cmd/policyrule"
	"github.com/chaitin/chaitin-cli/products/safeline/cmd/site"
	"github.com/chaitin/chaitin-cli/products/safeline/cmd/stats"
	"github.com/chaitin/chaitin-cli/products/safeline/cmd/system"
	"github.com/chaitin/chaitin-cli/products/safeline/pkg/auth"
	"github.com/chaitin/chaitin-cli/products/safeline/pkg/client"
	"github.com/chaitin/chaitin-cli/products/safeline/version"
	"github.com/spf13/cobra"
)

var (
	url                     string
	apiKey                  string
	indent                  bool
	insecure                bool
	dryRun                  bool
	safelineVersionOverride string
	operationModeOverride   string
	configVersion           string
	configOperationMode     string
)

type runtimeConfig struct {
	URL           string `yaml:"url"`
	APIKey        string `yaml:"api_key"`
	Version       string `yaml:"version"`
	OperationMode string `yaml:"operation_mode"`
}

// NewCommand creates the safeline product command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "safeline",
		Short: "SafeLine WAF API CLI",
		Long:  "Command-line interface for SafeLine Skyview WAF APIs.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "help" || cmd.Name() == "completion" || cmd.Name() == "version" {
				return nil
			}
			// Check parents for help/completion
			for p := cmd.Parent(); p != nil; p = p.Parent() {
				if p.Name() == "help" || p.Name() == "completion" {
					return nil
				}
			}

			if url == "" {
				url = os.Getenv("SAFELINE_URL")
			}
			if apiKey == "" {
				apiKey = os.Getenv("SAFELINE_API_KEY")
			}

			if url == "" {
				return fmt.Errorf("URL is required (use --url or set SAFELINE_URL)")
			}

			// Sync flags to cmd package for subcommands
			output := "table"
			if indent {
				output = "json"
			}
			cmdpkg.SetFlags(url, apiKey, output, insecure, dryRun, safelineVersionOverride, operationModeOverride, configVersion, configOperationMode)

			// Detect server version (non-blocking, failures are silent)
			detectServerVersion()

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&url, "url", "", "Skyview URL (or SAFELINE_URL)")
	cmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key (or SAFELINE_API_KEY)")
	cmd.PersistentFlags().BoolVar(&indent, "indent", false, "Pretty-print JSON output")
	cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	cmd.PersistentFlags().BoolVar(&insecure, "insecure", true, "Skip TLS certificate verification")
	cmd.PersistentFlags().StringVar(&safelineVersionOverride, "safeline-version", "", "SafeLine version override for capability selection, such as 23.01.014")
	cmd.PersistentFlags().StringVar(&operationModeOverride, "operation-mode", "", "SafeLine operation mode override, such as software-reverse-proxy")

	return cmd
}

// detectServerVersion attempts to detect and cache the server version.
// Failures are silent — version-aware features will use fallback instead.
func detectServerVersion() {
	cachePath := version.DefaultCachePath()

	// Try loading from cache first
	if cachePath != "" {
		entry, err := version.LoadCachedVersion(cachePath, url)
		if err == nil && entry.Version != "" {
			cmdpkg.SetServerVersion(entry.Version)
			return
		}
	}

	// Query the server
	transport := &auth.Transport{Token: apiKey}
	httpClient := &http.Client{Transport: transport}
	if insecure {
		httpClient.Transport = &auth.Transport{
			Base:  egressproxy.Wrap(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}),
			Token: apiKey,
		}
	}

	v, err := version.GetServerVersionFromAPI(url, httpClient)
	if err != nil {
		// Silent failure — version detection is best-effort
		return
	}

	cmdpkg.SetServerVersion(v)

	// Cache the result
	if cachePath != "" {
		_ = version.CacheVersion(cachePath, url, v)
	}
}

// RegisterModules adds all API module commands to the safeline root command.
func RegisterModules(cmd *cobra.Command) {
	// Register new user-friendly commands
	cmd.AddCommand(stats.NewCommand())
	cmd.AddCommand(inspect.NewCommand())
	cmd.AddCommand(cert.NewCommand())
	cmd.AddCommand(site.NewCommand())
	cmd.AddCommand(ipgroup.NewCommand())
	cmd.AddCommand(acl.NewCommand())
	cmd.AddCommand(policygroup.NewCommand())
	cmd.AddCommand(policyrule.NewCommand())
	cmd.AddCommand(system.NewCommand())
	cmd.AddCommand(log.NewCommand())
	cmd.AddCommand(network.NewCommand())

}

// NewClient creates an authenticated Skyview client from command flags.
func NewClient() *client.Client {
	transport := &auth.Transport{
		Token: apiKey,
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	if insecure {
		httpClient.Transport = egressproxy.Wrap(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		})
		// Wrap auth transport
		httpClient.Transport = &auth.Transport{
			Base:  httpClient.Transport,
			Token: apiKey,
		}
	}

	return client.New(url, httpClient)
}

// PrintJSON prints JSON data to stdout.
func PrintJSON(data []byte) error {
	if !indent {
		_, err := fmt.Fprintln(os.Stdout, string(data))
		return err
	}
	var pretty interface{}
	if err := json.Unmarshal(data, &pretty); err != nil {
		_, err := fmt.Fprintln(os.Stdout, string(data))
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(pretty)
}

// PrintEnvelope prints the data from an API response envelope.
func PrintEnvelope(env *client.Envelope) error {
	if env.IsWarning() {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", env.GetWarningText())
	}
	return PrintJSON(env.Data)
}

// ReadInput reads JSON input from file or stdin.
func ReadInput(file string) (io.Reader, error) {
	if file == "" || file == "-" {
		return os.Stdin, nil
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", file, err)
	}
	return f, nil
}

// ApplyRuntimeConfig applies configuration from config.yaml to command flags.
// If a flag is not explicitly set by the user, it uses the value from config file.
func ApplyRuntimeConfig(cmd *cobra.Command, cfg config.Raw) {
	productCfg, err := config.DecodeProduct[runtimeConfig](cfg, "safeline")
	if err != nil {
		return
	}

	if flag := cmd.Flags().Lookup("url"); flag != nil && !flag.Changed {
		url = productCfg.URL
	}

	if flag := cmd.Flags().Lookup("api-key"); flag != nil && !flag.Changed {
		apiKey = productCfg.APIKey
	}

	configVersion = productCfg.Version
	configOperationMode = productCfg.OperationMode
}
