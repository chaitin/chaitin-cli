package insight

import (
	"fmt"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/chaitin/chaitin-cli/products/insight/client"
	"github.com/chaitin/chaitin-cli/products/insight/cmd"
	"github.com/chaitin/chaitin-cli/products/insight/models"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	runtimeCfg      models.Config
	runtimeInsecure bool
	verbose         bool
	dryRun          bool
)

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "insight",
		Short: "Insight API CLI",
		Long: `Insight risk operations management platform CLI

Authentication uses the API key.

Config example:
  insight:
    url: https://your-insight.example.com
    api_key: your-api-key

Environment variables:
  INSIGHT_URL
  INSIGHT_API_KEY

Common examples:
  chaitin-cli insight task list
  chaitin-cli insight api-token info`,
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			applyRuntimeConfig(c)
			if runtimeCfg.URL == "" {
				return fmt.Errorf("URL is required (use --url or configure insight.url / INSIGHT_URL)")
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().String("url", "", "Insight API URL")
	rootCmd.PersistentFlags().String("api-key", "", "Insight API key")
	rootCmd.PersistentFlags().StringP("output", "o", "json", "Output format (table|json)")
	rootCmd.PersistentFlags().BoolVar(&runtimeInsecure, "insecure", true, "Skip TLS certificate verification")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print request URL, headers, and body")

	// Helper function to lazily initialize the client when a command actually runs
	getClientFn := func(c *cobra.Command) *client.Client {
		return client.NewClient(runtimeCfg, runtimeInsecure, verbose, dryRun, c.OutOrStdout(), c.ErrOrStderr())
	}

	// Register subcommands from the cmd package
	rootCmd.AddCommand(cmd.NewApiKeyCmd(getClientFn))
	rootCmd.AddCommand(cmd.NewTaskCmd(getClientFn))
	rootCmd.AddCommand(cmd.NewSystemCmd(getClientFn))
	rootCmd.AddCommand(cmd.NewResultCmd(getClientFn))
	rootCmd.AddCommand(cmd.NewSnapshotCmd(getClientFn))
	rootCmd.AddCommand(cmd.NewRawCmd(getClientFn))
	rootCmd.AddCommand(cmd.NewVulnCmd(getClientFn))
	rootCmd.AddCommand(cmd.NewAssetCmd(getClientFn))
	rootCmd.AddCommand(cmd.NewOrderCmd(getClientFn))

	return rootCmd
}

func ApplyRuntimeConfig(c *cobra.Command, cfg config.Raw, isDryRun bool) {
	productCfg, err := config.DecodeProduct[models.Config](cfg, "insight")
	if err != nil {
		return
	}
	runtimeCfg = productCfg
	dryRun = isDryRun
}

func applyRuntimeConfig(c *cobra.Command) {
	if flag := lookupFlag(c, "url"); flag != nil && !flag.Changed && runtimeCfg.URL != "" {
		_ = setFlag(c, "url", runtimeCfg.URL)
	}
	if flag := lookupFlag(c, "api-key"); flag != nil && !flag.Changed && runtimeCfg.APIKey != "" {
		_ = setFlag(c, "api-key", runtimeCfg.APIKey)
	}
	if flag := lookupFlag(c, "url"); flag != nil {
		runtimeCfg.URL = flag.Value.String()
	}
	if flag := lookupFlag(c, "api-key"); flag != nil {
		runtimeCfg.APIKey = flag.Value.String()
	}
}

func lookupFlag(c *cobra.Command, name string) *pflag.Flag {
	if flag := c.Flags().Lookup(name); flag != nil {
		return flag
	}
	if flag := c.PersistentFlags().Lookup(name); flag != nil {
		return flag
	}
	return c.InheritedFlags().Lookup(name)
}

func setFlag(c *cobra.Command, name, value string) error {
	if c.Flags().Lookup(name) != nil {
		return c.Flags().Set(name, value)
	}
	if c.PersistentFlags().Lookup(name) != nil {
		return c.PersistentFlags().Set(name, value)
	}
	return c.InheritedFlags().Set(name, value)
}
