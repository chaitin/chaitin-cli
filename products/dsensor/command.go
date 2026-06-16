package dsensor

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/chaitin/chaitin-cli/products/dsensor/internal/cache"
	"github.com/chaitin/chaitin-cli/products/dsensor/internal/client"
	"github.com/chaitin/chaitin-cli/products/dsensor/internal/output"
	"github.com/chaitin/chaitin-cli/products/dsensor/internal/parser"
	"github.com/chaitin/chaitin-cli/products/dsensor/internal/spec"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Config struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

var (
	runtimeCfg    Config
	dryRun        bool
	urlFlag       string
	apiKeyFlag    string
	refreshCache  bool
	outputFormat  string
	insecureTLS   bool
	verbose       bool
	specHashOnRun = cache.SpecHash(spec.SpecJSON)
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dsensor",
		Short: "D-Sensor security monitoring platform management tool",
		Long: `D-Sensor CLI

D-Sensor is a security monitoring and honeypot platform.
Commands are generated from the embedded OpenAPI schema and call D-Sensor APIs through POST requests.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			applyRuntimeConfig(cmd)
		},
	}

	cmd.PersistentFlags().StringVar(&urlFlag, "url", "", "D-Sensor API URL")
	cmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "API token sent as API-Token header")
	cmd.PersistentFlags().BoolVar(&refreshCache, "refresh-cache", false, "Force refresh local server version cache")
	cmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table|json)")
	cmd.PersistentFlags().BoolVar(&insecureTLS, "insecure", false, "Skip TLS certificate verification")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print request path and body")

	s, err := parser.ParseSpec(spec.SpecJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to parse embedded D-Sensor OpenAPI schema: %v\n", err)
		return cmd
	}

	cmds, err := parser.FlattenCommands(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to generate D-Sensor commands: %v\n", err)
		return cmd
	}

	tree := parser.BuildCommandTree(cmds)
	cmd.AddCommand(tree.Commands()...)

	return cmd
}

func ApplyRuntimeConfig(cmd *cobra.Command, cfg config.Raw, isDryRun bool) {
	productCfg, err := config.DecodeProduct[Config](cfg, "dsensor")
	if err != nil {
		return
	}
	runtimeCfg = productCfg
	dryRun = isDryRun
}

func applyRuntimeConfig(cmd *cobra.Command) {
	if flag := lookupFlag(cmd, "url"); flag != nil && !flag.Changed && runtimeCfg.URL != "" {
		_ = setFlag(cmd, "url", runtimeCfg.URL)
	}
	if flag := lookupFlag(cmd, "api-key"); flag != nil && !flag.Changed && runtimeCfg.APIKey != "" {
		_ = setFlag(cmd, "api-key", runtimeCfg.APIKey)
	}

	setupRunner()
}

func lookupFlag(cmd *cobra.Command, name string) *pflag.Flag {
	if flag := cmd.Flags().Lookup(name); flag != nil {
		return flag
	}
	return cmd.PersistentFlags().Lookup(name)
}

func setFlag(cmd *cobra.Command, name, value string) error {
	if cmd.Flags().Lookup(name) != nil {
		return cmd.Flags().Set(name, value)
	}
	return cmd.PersistentFlags().Set(name, value)
}

func setupRunner() {
	if urlFlag == "" && !dryRun {
		parser.DefaultRunner = nil
		return
	}

	if !dryRun {
		checkCache(urlFlag, apiKeyFlag)
	}

	apiClient := client.New(urlFlag, apiKeyFlag, insecureTLS)
	parser.DefaultRunner = func(cmd parser.CommandSpec, body []byte) ([]byte, error) {
		if dryRun {
			return renderDryRun(cmd, body)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "POST %s\n%s\n", cmd.Path, string(body))
		}
		_, apiResp, err := apiClient.DoAndParse(cmd.Path, body)
		if err != nil {
			return nil, err
		}
		out, fmtErr := output.Format(apiResp, outputFormat)
		if fmtErr != nil {
			return nil, fmtErr
		}
		return []byte(out), nil
	}
}

func renderDryRun(cmd parser.CommandSpec, body []byte) ([]byte, error) {
	payload := map[string]any{
		"method": cmd.Method,
		"path":   cmd.Path,
		"body":   json.RawMessage(body),
	}
	return json.MarshalIndent(payload, "", "  ")
}

func checkCache(url, apiKey string) {
	cacheDir := cache.DefaultDir()

	if refreshCache {
		fmt.Fprintln(os.Stderr, "INFO: force refresh cache")
		return
	}

	cached, _ := cache.Load(cacheDir)
	if cached != nil && cached.SpecHash == specHashOnRun {
		apiClient := client.New(url, apiKey, insecureTLS)
		newState, valid := cache.CheckVersion(cached, apiClient)
		if valid && newState.ServerVersion != "" {
			fmt.Fprintf(os.Stderr, "INFO: server version: %s\n", newState.ServerVersion)
		}
		_ = cache.Save(cacheDir, newState)
		return
	}

	state := &cache.State{SpecHash: specHashOnRun}
	_ = cache.Save(cacheDir, state)
}
