package ddr

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chaitin/workspace-cli/config"
	"github.com/spf13/cobra"
)

//go:embed openapi.json
var openAPIFS embed.FS

var (
	runtimeCfg           Config
	runtimeConfigPath    string
	dryRun               bool
	verbose              bool
	headerApp            string
	headerDebug          string
	headerCrypt          string
	headerTimezone       string
	headerAccept         string
	headerAcceptLanguage string
	headerContentType    string
	headerIfNoneMatch    string
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ddr",
		Short: "DDR management tool",
		Long: `DDR CLI

面向数据安全运营场景的命令行工具，覆盖数据扫描与发现、数据分类分级、数据资产地图、敏感数据监测、文件解析、策略管控、外发监测、全流程追踪、身份关联、数据分析、风险识别与操作审批等核心能力。

支持行为管控相关场景，包括外发管控、落盘管控、邮件管控、代码管控、剪贴板管理、网页管控、IM 管控、网络管控与浏览器管控。

`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			applyRuntimeConfig(cmd)
		},
	}

	cmd.PersistentFlags().String("url", "", "API URL")
	cmd.PersistentFlags().String("api-key", "", "API key used as Authorization header")
	cmd.PersistentFlags().String("company-id", "", "Company ID used as X-CS-Header-Company header")
	cmd.PersistentFlags().StringP("output", "o", "table", "Output format (table|json)")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print request URL, headers, and body")
	cmd.PersistentFlags().StringVar(&headerApp, "x-cs-header-app", "qzh", "Value for X-CS-Header-App header")
	cmd.PersistentFlags().StringVar(&headerDebug, "x-cs-header-debug", "", "Value for X-CS-Header-Debug header")
	cmd.PersistentFlags().StringVar(&headerCrypt, "x-cs-header-crypt", "", "Value for X-CS-Header-Crypt header")
	cmd.PersistentFlags().StringVar(&headerTimezone, "x-cs-header-timezone", "", "Value for X-CS-Header-Timezone header")
	cmd.PersistentFlags().StringVar(&headerAccept, "accept", "application/json", "Value for Accept header")
	cmd.PersistentFlags().StringVar(&headerAcceptLanguage, "accept-language", "zh", "Value for Accept-Language header")
	cmd.PersistentFlags().StringVar(&headerContentType, "content-type", "application/json", "Value for Content-Type header")
	cmd.PersistentFlags().StringVar(&headerIfNoneMatch, "if-none-match", "", "Value for If-None-Match header")
	cmd.AddCommand(newGetAPITokenCommand())

	if err := loadDynamicCommands(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	return cmd
}

func ApplyRuntimeConfig(cmd *cobra.Command, cfg config.Raw, configPath string, isDryRun bool) {
	productCfg, err := config.DecodeProduct[Config](cfg, "ddr")
	if err != nil {
		return
	}
	runtimeCfg = productCfg
	runtimeConfigPath = configPath
	dryRun = isDryRun
}

func applyRuntimeConfig(cmd *cobra.Command) {
	if flag := cmd.Flags().Lookup("url"); flag != nil && !flag.Changed && runtimeCfg.URL != "" {
		_ = cmd.Flags().Set("url", runtimeCfg.URL)
	}
	if flag := cmd.Flags().Lookup("api-key"); flag != nil && !flag.Changed && runtimeCfg.APIKey != "" {
		_ = cmd.Flags().Set("api-key", runtimeCfg.APIKey)
	}
	if flag := cmd.Flags().Lookup("company-id"); flag != nil && !flag.Changed && runtimeCfg.CompanyID != "" {
		_ = cmd.Flags().Set("company-id", runtimeCfg.CompanyID)
	}
}

func loadDynamicCommands(cmd *cobra.Command) error {
	data, err := openAPIFS.ReadFile("openapi.json")
	if err != nil {
		return fmt.Errorf("failed to read embedded openapi.json: %w", err)
	}

	var api OpenAPI
	if err := json.Unmarshal(data, &api); err != nil {
		return fmt.Errorf("failed to parse openapi.json: %w", err)
	}

	parser := NewParser()
	commands, err := parser.GenerateCommands(&api)
	if err != nil {
		return fmt.Errorf("failed to generate commands: %w", err)
	}

	for _, command := range commands {
		cmd.AddCommand(command)
	}
	return nil
}

func getRenderer(cmd *cobra.Command) Renderer {
	format := FormatTable
	if output, _ := cmd.Flags().GetString("output"); output == "json" {
		format = FormatJSON
	}
	return NewRenderer(format, cmd.OutOrStdout())
}

func getClient(cmd *cobra.Command) *Client {
	url, _ := cmd.Flags().GetString("url")
	apiKey, _ := cmd.Flags().GetString("api-key")
	companyID, _ := cmd.Flags().GetString("company-id")
	headers := map[string]string{
		"X-CS-Header-App":      headerApp,
		"X-CS-Header-Debug":    headerDebug,
		"X-CS-Header-Crypt":    headerCrypt,
		"X-CS-Header-Timezone": headerTimezone,
		"Accept":               headerAccept,
		"Accept-Language":      headerAcceptLanguage,
		"Content-Type":         headerContentType,
		"If-None-Match":        headerIfNoneMatch,
	}
	return NewClient(&Config{
		URL:       url,
		APIKey:    apiKey,
		CompanyID: companyID,
	}, headers, verbose)
}
