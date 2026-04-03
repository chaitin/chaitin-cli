package safelinece

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// Parser OpenAPI 解析器
type Parser struct {
	client   *Client
	renderer Renderer
}

// NewParser 创建解析器
func NewParser(client *Client, renderer Renderer) *Parser {
	return &Parser{
		client:   client,
		renderer: renderer,
	}
}

// GenerateCommands 生成 Cobra 命令
func (p *Parser) GenerateCommands(api *OpenAPI) ([]*cobra.Command, error) {
	tagCommands := make(map[string]*cobra.Command)

	for path, pathItem := range api.Paths {
		operations := []struct {
			method    string
			operation *Operation
		}{
			{"GET", pathItem.Get},
			{"POST", pathItem.Post},
			{"PUT", pathItem.Put},
			{"DELETE", pathItem.Delete},
		}

		for _, op := range operations {
			if op.operation == nil {
				continue
			}

			tag := "default"
			if len(op.operation.Tags) > 0 {
				tag = op.operation.Tags[0]
			}

			// 解析嵌套命令 (如 "log/attack" -> parent="log", child="attack")
			parentTag, childTag := parseNestedTag(tag)

			// 确保父命令存在
			if _, exists := tagCommands[parentTag]; !exists {
				tagCommands[parentTag] = &cobra.Command{
					Use:   parentTag,
					Short: fmt.Sprintf("%s commands", parentTag),
				}
			}

			// 如果是嵌套命令，确保子命令存在
			targetCmd := tagCommands[parentTag]
			if childTag != "" {
				targetCmd = getOrCreateChildCommand(tagCommands[parentTag], childTag)
			}

			cmd := p.createOperationCommand(op.method, path, op.operation, api.BasePath)
			targetCmd.AddCommand(cmd)
		}
	}

	var commands []*cobra.Command
	for _, cmd := range tagCommands {
		commands = append(commands, cmd)
	}

	return commands, nil
}

// parseNestedTag 解析嵌套 tag (如 "log/attack" -> "log", "attack")
func parseNestedTag(tag string) (parent, child string) {
	parts := strings.SplitN(tag, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

// getOrCreateChildCommand 获取或创建子命令
func getOrCreateChildCommand(parent *cobra.Command, childName string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Use == childName {
			return cmd
		}
	}

	child := &cobra.Command{
		Use:   childName,
		Short: fmt.Sprintf("%s %s commands", parent.Use, childName),
	}
	parent.AddCommand(child)
	return child
}

func (p *Parser) createOperationCommand(method, path string, op *Operation, basePath string) *cobra.Command {
	opName := operationName(method, path)

	cmd := &cobra.Command{
		Use:   opName,
		Short: op.Summary,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.executeCommand(cmd, method, path, basePath, op.Parameters)
		},
	}

	// 添加参数 flags
	for _, param := range op.Parameters {
		addFlag(cmd, param)
	}

	// 为 list 命令添加分页参数
	if opName == "list" {
		if cmd.Flags().Lookup("page") == nil {
			cmd.Flags().Int("page", 1, "Page number")
		}
		if cmd.Flags().Lookup("size") == nil {
			cmd.Flags().Int("size", 20, "Page size")
		}
	}

	return cmd
}

func (p *Parser) executeCommand(cmd *cobra.Command, method, path, basePath string, params []Parameter) error {
	ctx := context.Background()

	// 构建 URL
	apiPath := basePath + path

	// 收集参数
	query := url.Values{}
	var body map[string]interface{}
	pathParams := make(map[string]string)

	for _, param := range params {
		val, err := cmd.Flags().GetString(param.Name)
		if err != nil {
			continue
		}

		switch param.In {
		case "path":
			pathParams[param.Name] = val
		case "query":
			if val != "" {
				query.Set(param.Name, val)
			}
		case "body", "formData":
			if body == nil {
				body = make(map[string]interface{})
			}
			body[param.Name] = val
		}
	}

	// 替换路径参数
	for name, val := range pathParams {
		apiPath = strings.ReplaceAll(apiPath, "{"+name+"}", val)
	}

	// 添加分页参数
	if page, err := cmd.Flags().GetInt("page"); err == nil && page > 0 {
		query.Set("page", fmt.Sprintf("%d", page))
	}
	if size, err := cmd.Flags().GetInt("size"); err == nil && size > 0 {
		query.Set("size", fmt.Sprintf("%d", size))
	}

	// 执行请求
	var result interface{}
	var err error

	switch method {
	case "GET":
		err = p.client.Get(ctx, apiPath, query, &result)
	case "POST":
		err = p.client.Post(ctx, apiPath, body, &result)
	case "PUT":
		err = p.client.Put(ctx, apiPath, body, &result)
	case "DELETE":
		err = p.client.Delete(ctx, apiPath, &result)
	}

	if err != nil {
		return err
	}

	return p.renderer.Render(result)
}

func operationName(method, path string) string {
	// 特殊路径处理 - cert 相关
	switch {
	case strings.HasSuffix(path, "/system") && method == "GET":
		return "info"
	case strings.HasSuffix(path, "/system") && method == "PUT":
		return "update"
	case strings.HasSuffix(path, "/system/authorize") && method == "GET":
		return "get"
	case strings.HasSuffix(path, "/system/authorize") && method == "DELETE":
		return "delete"
	case strings.HasSuffix(path, "/cert") && method == "POST":
		return "upload"
	}

	// switch 相关处理
	if strings.Contains(path, "/switch") {
		// policy/switch 直接返回 switch
		if strings.Contains(path, "/policy/switch") {
			return "switch"
		}
		// skynet/rule/switch 返回 get/set
		switch method {
		case "GET":
			return "get"
		case "PUT":
			return "set"
		}
	}

	// 其他特殊路径处理
	switch {
	case strings.Contains(path, "/detail"):
		return "get"
	case strings.Contains(path, "/append"):
		return "append"
	case strings.Contains(path, "/qps"):
		return "qps"
	case strings.Contains(path, "/advance/access") && !strings.Contains(path, "/trend"):
		return "access"
	case strings.Contains(path, "/advance/attack") || strings.Contains(path, "/advance/intercept"):
		return "attack"
	case strings.Contains(path, "/trend/access"):
		return "access"
	case strings.Contains(path, "/trend/intercept"):
		return "intercept"
	case strings.Contains(path, "/global/mode"):
		if method == "GET" {
			return "get"
		}
		return "update"
	}

	// 默认根据 HTTP method 判断
	switch method {
	case "GET":
		if strings.Contains(path, "{id}") || strings.Contains(path, ":id") {
			return "get"
		}
		return "list"
	case "POST":
		return "create"
	case "PUT":
		return "update"
	case "DELETE":
		return "delete"
	}
	return strings.ToLower(method)
}

func addFlag(cmd *cobra.Command, param Parameter) {
	switch param.Type {
	case "integer":
		cmd.Flags().Int(param.Name, 0, param.Description)
	case "boolean":
		cmd.Flags().Bool(param.Name, false, param.Description)
	default:
		cmd.Flags().String(param.Name, "", param.Description)
	}

	if param.Required {
		cmd.MarkFlagRequired(param.Name)
	}
}

