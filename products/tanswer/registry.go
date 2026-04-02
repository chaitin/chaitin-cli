package tanswer

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed apis/*.json
var apiSpecs embed.FS

// NewCommand 创建 answer 产品的根命令，自动加载所有 JSON API 定义。
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tanswer",
		Short: "T-Answer product APIs",
	}

	cmd.PersistentFlags().String("url", "", "API URL (e.g. https://api.example.com)")
	cmd.PersistentFlags().String("api-key", "", "API Key for authentication")
	cmd.PersistentFlags().Bool("raw", false, "Output raw JSON without formatting")

	entries, err := apiSpecs.ReadDir("apis")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded API specs: %v", err))
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := apiSpecs.ReadFile(filepath.Join("apis", entry.Name()))
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
			name := strings.TrimSuffix(filepath.Base(entry.Name()), ".json")
			fmt.Printf("warning: failed to parse %s: %v\n", name, err)
			continue
		}

		parsed := parseOperations(ops)
		registerOperations(cmd, parsed)
	}

	return cmd
}
