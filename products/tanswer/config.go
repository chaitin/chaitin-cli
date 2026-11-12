package tanswer

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/spf13/cobra"
)

type ConfigOptions struct {
	Address               string
	Token                 string
	Timeout               string
	InsecureSkipVerify    bool
	InsecureSkipVerifySet bool
	AllowMissingToken     bool
}

type Config struct {
	BaseURL            string
	APIToken           string
	Timeout            time.Duration
	InsecureSkipVerify bool
}

type runtimeConfig struct {
	URL      string `yaml:"url"`
	APIKey   string `yaml:"api_key"`
	Timeout  string `yaml:"timeout"`
	Insecure bool   `yaml:"insecure"`
}

func ApplyRuntimeConfig(cmd *cobra.Command, raw config.Raw) {
	productCfg, err := config.DecodeProduct[runtimeConfig](raw, "tanswer")
	if err != nil {
		return
	}
	applyFlagString(cmd, "url", productCfg.URL)
	applyFlagString(cmd, "api-key", productCfg.APIKey)
	applyFlagString(cmd, "timeout", productCfg.Timeout)
	if productCfg.Insecure {
		applyFlagString(cmd, "insecure", "true")
	}
}

func LoadConfig(opts ConfigOptions) (Config, error) {
	addr := firstNonEmpty(opts.Address, os.Getenv("TANSWER_URL"))
	token := firstNonEmpty(opts.Token, os.Getenv("TANSWER_API_KEY"))
	timeoutText := firstNonEmpty(opts.Timeout, os.Getenv("TANSWER_TIMEOUT"), "30s")
	insecureSkipVerify := opts.InsecureSkipVerify
	if !opts.InsecureSkipVerifySet && !insecureSkipVerify {
		insecureSkipVerify = truthy(os.Getenv("TANSWER_INSECURE"))
	}

	addr = strings.TrimRight(strings.TrimSpace(addr), "/")
	if addr == "" {
		return Config{}, errors.New("missing Quanxi address: set --url or TANSWER_URL")
	}
	if token == "" && !opts.AllowMissingToken {
		return Config{}, errors.New("missing OpenAPI token: set --api-key or TANSWER_API_KEY")
	}
	timeout, err := time.ParseDuration(timeoutText)
	if err != nil {
		return Config{}, err
	}
	return Config{
		BaseURL:            addr,
		APIToken:           token,
		Timeout:            timeout,
		InsecureSkipVerify: insecureSkipVerify,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "t", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func applyFlagString(cmd *cobra.Command, name, value string) {
	if value == "" {
		return
	}
	if flag := cmd.Flags().Lookup(name); flag != nil && !flag.Changed {
		_ = cmd.Flags().Set(name, value)
		return
	}
	if flag := cmd.PersistentFlags().Lookup(name); flag != nil && !flag.Changed {
		_ = cmd.PersistentFlags().Set(name, value)
	}
}
