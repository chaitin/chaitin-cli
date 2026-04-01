package safelinece

import (
	"context"
	"fmt"
	"net/url"
	"os"
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

			if _, exists := tagCommands[tag]; !exists {
				tagCommands[tag] = &cobra.Command{
					Use:   tag,
					Short: fmt.Sprintf("%s commands", tag),
				}
			}

			cmd := p.createOperationCommand(op.method, path, op.operation, api.BasePath)
			tagCommands[tag].AddCommand(cmd)
		}
	}

	var commands []*cobra.Command
	for _, cmd := range tagCommands {
		commands = append(commands, cmd)
	}

	return commands, nil
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
	switch method {
	case "GET":
		if strings.Contains(path, "{id}") || strings.Contains(path, "{") {
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

func newRenderer(cmd *cobra.Command) Renderer {
	format := FormatTable
	if o, _ := cmd.Flags().GetString("output"); o == "json" {
		format = FormatJSON
	}
	return NewRenderer(format, os.Stdout)
}
