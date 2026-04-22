package safelinece

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/spf13/cobra"
)

//go:embed openapi.json
var openAPIFS embed.FS

// 运行时配置
var (
	runtimeCfg runtimeConfig
	dryRun     bool
)

type runtimeConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

// NewCommand 创建 safeline-ce 命令
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "safeline-ce",
		Short: "SafeLine Community Edition - WAF management tool",
		Long: `SafeLine CE CLI - SafeLine 社区版命令行管理工具

快速入门:
  # 1. 创建配置文件 config.yaml
  safeline-ce:
    url: https://your-server:9443
    api_key: your-api-key

  # 2. 查看帮助
  chaitin-cli safeline-ce --help

  # 3. 查看站点列表
  chaitin-cli safeline-ce site list

  # 4. 查看攻击日志
  chaitin-cli safeline-ce log attack list

  # 5. 查看统计概览
  chaitin-cli safeline-ce stat overview

文档: https://github.com/chaitin/safeline-ce`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// 应用运行时配置到命令 flags
			applyRuntimeConfig(cmd)
		},
	}

	// 全局 flags
	cmd.PersistentFlags().String("url", "", "API URL (e.g. https://your-server:9443)")
	cmd.PersistentFlags().String("api-key", "", "API key for authentication")
	cmd.PersistentFlags().StringP("output", "o", "table", "Output format (table|json)")
	cmd.PersistentFlags().Bool("verbose", false, "Verbose output")

	// 加载动态命令
	if err := loadDynamicCommands(cmd); err != nil {
		// 允许在没有配置时显示帮助，但打印警告
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	return cmd
}

// ApplyRuntimeConfig 应用运行时配置（供 main.go 调用）
func ApplyRuntimeConfig(cmd *cobra.Command, cfg config.Raw, isDryRun bool) {
	productCfg, err := config.DecodeProduct[runtimeConfig](cfg, "safeline-ce")
	if err != nil {
		return
	}
	runtimeCfg = productCfg
	dryRun = isDryRun
}

// applyRuntimeConfig 应用运行时配置到命令 flags
func applyRuntimeConfig(cmd *cobra.Command) {
	// 如果 flag 未设置，使用配置文件的值
	if flag := cmd.Flags().Lookup("url"); flag != nil && !flag.Changed && runtimeCfg.URL != "" {
		_ = cmd.Flags().Set("url", runtimeCfg.URL)
	}
	if flag := cmd.Flags().Lookup("api-key"); flag != nil && !flag.Changed && runtimeCfg.APIKey != "" {
		_ = cmd.Flags().Set("api-key", runtimeCfg.APIKey)
	}
}

// loadDynamicCommands 从嵌入的 OpenAPI 定义加载动态命令
func loadDynamicCommands(cmd *cobra.Command) error {
	// 从嵌入的 FS 读取 openapi.json
	data, err := openAPIFS.ReadFile("openapi.json")
	if err != nil {
		return fmt.Errorf("failed to read embedded openapi.json: %w", err)
	}

	var api OpenAPI
	if err := json.Unmarshal(data, &api); err != nil {
		return fmt.Errorf("failed to parse openapi.json: %w", err)
	}

	// 创建解析器并生成命令
	parser := NewParser()
	commands, err := parser.GenerateCommands(&api)
	if err != nil {
		return fmt.Errorf("failed to generate commands: %w", err)
	}

	// 添加动态生成的命令
	for _, c := range commands {
		cmd.AddCommand(c)
	}

	// 注册 override 命令（特殊操作）
	registerOverrides(cmd)

	return nil
}

// getRenderer 从命令获取输出格式并创建渲染器
func getRenderer(cmd *cobra.Command) Renderer {
	format := FormatTable
	if o, _ := cmd.Flags().GetString("output"); o == "json" {
		format = FormatJSON
	}
	return NewRenderer(format, os.Stdout)
}

// getClient 从命令创建客户端
func getClient(cmd *cobra.Command) *Client {
	url, _ := cmd.Flags().GetString("url")
	apiKey, _ := cmd.Flags().GetString("api-key")

	cfg := &Config{
		URL:    url,
		APIKey: apiKey,
	}

	return NewClient(cfg)
}

// registerOverrides 注册需要特殊处理的命令
func registerOverrides(cmd *cobra.Command) {
	// 查找 cert 命令并移除不需要的子命令
	for _, c := range cmd.Commands() {
		if c.Use == "cert" {
			// 移除 delete 和 upload 子命令
			var toRemove []*cobra.Command
			for _, subCmd := range c.Commands() {
				if subCmd.Use == "delete" || subCmd.Use == "upload" {
					toRemove = append(toRemove, subCmd)
				}
			}
			for _, subCmd := range toRemove {
				c.RemoveCommand(subCmd)
			}
			break
		}
	}

	// 查找或创建 stat 命令
	var statCmd *cobra.Command
	for _, c := range cmd.Commands() {
		if c.Use == "stat" {
			statCmd = c
			break
		}
	}

	if statCmd == nil {
		statCmd = &cobra.Command{
			Use:   "stat",
			Short: "数据统计",
		}
		cmd.AddCommand(statCmd)
	}

	// stat overview 命令 (聚合调用)
	overviewCmd := &cobra.Command{
		Use:   "overview",
		Short: "概览统计（聚合调用）",
		RunE:  runStatOverview,
	}
	statCmd.AddCommand(overviewCmd)
}

// runStatOverview 执行 stat overview 聚合查询
func runStatOverview(cmd *cobra.Command, args []string) error {
	client := getClient(cmd)
	ctx := cmd.Context()

	// 并发获取各项统计数据
	type statResult struct {
		name string
		data interface{}
		err  error
	}

	results := make(chan statResult, 3)

	// 获取 QPS
	go func() {
		var result interface{}
		err := client.Get(ctx, "/api/stat/qps", nil, &result)
		results <- statResult{name: "qps", data: result, err: err}
	}()

	// 获取访问统计
	go func() {
		var result interface{}
		err := client.Get(ctx, "/api/stat/advance/access", nil, &result)
		results <- statResult{name: "access", data: result, err: err}
	}()

	// 获取拦截统计
	go func() {
		var result interface{}
		err := client.Get(ctx, "/api/stat/advance/attack", nil, &result)
		results <- statResult{name: "attack", data: result, err: err}
	}()

	// 收集结果
	overview := make(map[string]interface{})
	for i := 0; i < 3; i++ {
		r := <-results
		if r.err != nil {
			overview[r.name] = map[string]string{"error": r.err.Error()}
		} else {
			overview[r.name] = r.data
		}
	}

	renderer := getRenderer(cmd)
	return renderer.Render(overview)
}

// Execute 执行命令并处理退出码
func Execute(cmd *cobra.Command) error {
	err := cmd.Execute()
	if err == nil {
		return nil
	}

	if cliErr, ok := err.(*CLIError); ok {
		os.Exit(cliErr.Code)
	}
	os.Exit(1)
	return err
}
