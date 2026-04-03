package safelinece

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Exit codes
const (
	ExitSuccess      = 0
	ExitAPIError     = 1
	ExitNetworkError = 2
	ExitConfigError  = 3
)

// Config 配置结构
type Config struct {
	Endpoint string `yaml:"endpoint"`
	Token    string `yaml:"token"`
}

// LoadConfig 从文件和环境变量加载配置
func LoadConfig(path string) (*Config, error) {
	// 读取 YAML 文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析到 map 再提取（因为配置文件有顶层 key）
	var rawConfig map[string]Config
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg, ok := rawConfig["safeline-ce"]
	if !ok {
		return nil, fmt.Errorf("missing 'safeline-ce' section in config")
	}

	// 应用环境变量覆盖
	cfg.ApplyEnvOverrides()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ApplyEnvOverrides 应用环境变量覆盖
func (c *Config) ApplyEnvOverrides() {
	if v := os.Getenv("SAFELINE_CE_ENDPOINT"); v != "" {
		c.Endpoint = v
	}
	if v := os.Getenv("SAFELINE_CE_TOKEN"); v != "" {
		c.Token = v
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if c.Token == "" {
		return fmt.Errorf("token is required")
	}
	return nil
}

// APIResponse 通用 API 响应
type APIResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []APIError      `json:"errors,omitempty"`
}

// APIError API 错误
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OpenAPI OpenAPI 2.0 文档结构
type OpenAPI struct {
	Swagger  string              `json:"swagger"`
	BasePath string              `json:"basePath"`
	Paths    map[string]PathItem `json:"paths"`
	XCLI     *CLIExtensions      `json:"x-cli,omitempty"`
}

// CLIExtensions CLI 扩展字段
type CLIExtensions struct {
	Tags     map[string]string `json:"tags,omitempty"`
	Children map[string]string `json:"children,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
}

// PathItem 路径项
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

// Operation 操作定义
type Operation struct {
	Tags         []string    `json:"tags"`
	Summary      string      `json:"summary"`
	Parameters   []Parameter `json:"parameters"`
	XCLISummary  string      `json:"x-cli-summary,omitempty"`
}

// Parameter 参数定义
type Parameter struct {
	Name            string `json:"name"`
	In              string `json:"in"` // path, query, body, formData
	Required        bool   `json:"required"`
	Type            string `json:"type"`
	Description     string `json:"description"`
	Schema          *Schema `json:"schema,omitempty"`
	XCLIDescription string `json:"x-cli-description,omitempty"`
}

// Schema 模式定义
type Schema struct {
	Ref string `json:"$ref,omitempty"`
}

// DefaultConfigPath 返回默认配置文件路径
func DefaultConfigPath() string {
	// 优先使用当前目录的 config.yaml
	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml"
	}
	// 其次使用产品目录下的 config.yaml
	return filepath.Join("products", "safeline-ce", "config.yaml")
}
