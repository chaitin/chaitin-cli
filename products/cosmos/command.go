package cosmos

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed apis/*.json
var apiSpecs embed.FS

// NewCommand 创建 cosmos（万象/AISOC）产品的根命令，自动加载所有 JSON API 定义。
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cosmos",
		Short: "Cosmos (万象/AISOC) product APIs",
		Long:  "Cosmos (万象/AISOC) CLI — 告警分析研判、日志查询、情报分析、IP 封禁、资产管理、漏洞扫描等安全运营能力。",
	}

	cmd.PersistentFlags().String("url", "", "Cosmos API URL (e.g. https://cosmos.example.com)")
	cmd.PersistentFlags().String("api-key", "", "JWT Bearer Token for authentication")
	cmd.PersistentFlags().Bool("raw", false, "Output raw JSON without formatting")

	entries, err := apiSpecs.ReadDir("apis")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded API specs: %v", err))
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := apiSpecs.ReadFile(path.Join("apis", entry.Name()))
		if err != nil {
			fmt.Printf("warning: failed to read %s: %v\n", entry.Name(), err)
			continue
		}

		// 跳过空文件
		content := strings.TrimSpace(string(data))
		if content == "" || content == "[]" {
			continue
		}

		var ops []APIOperation
		if err := json.Unmarshal(data, &ops); err != nil {
			name := strings.TrimSuffix(path.Base(entry.Name()), ".json")
			fmt.Printf("warning: failed to parse %s: %v\n", name, err)
			continue
		}

		parsed := parseOperations(ops)
		registerOperations(cmd, parsed)
	}

	return cmd
}
