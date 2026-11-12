package tanswer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const rawAPIWriteConfirmToken = "CONFIRM_TANSWER_RAW_API_WRITE"

func newRawAPICommand(opts *RootOptions) *cobra.Command {
	var queryText string
	var bodyText string
	var preview bool
	var confirm string
	cmd := &cobra.Command{
		Use:   "api <METHOD> <PATH>",
		Short: "调用用户已知且已授权的 Open API",
		Long:  "Open API 兜底调用。用于访问语义快捷层未覆盖、但用户已知且已授权的全悉 Open API。该命令不提供专属参数说明或专属摘要，只结构化返回状态码和原始响应。\n\nGET 和 HEAD 可直接执行；其他方法可能改变状态，默认仅返回请求预览且不会发送请求。审阅 preview 并取得用户对本次请求的明确确认后，才可使用 --confirm CONFIRM_TANSWER_RAW_API_WRITE 执行。确认 token 是技术校验，不等同于用户授权。根级 --dry-run 不适用于 tanswer。",
		Example: "  chaitin-cli tanswer api POST '/<known-unwrapped-openapi-path>' --body '{\"<known_parameter>\":\"<value>\"}' --preview\n" +
			"  chaitin-cli tanswer api GET /api/openapi/rpc/openapi.json",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := validateRawAPIPath(args[1])
			if err != nil {
				return writeRawAPIInputError(cmd, "INVALID_API_PATH", err)
			}
			cfg, err := LoadConfig(ConfigOptions{
				Address:            opts.Address,
				Token:              opts.Token,
				Timeout:            opts.Timeout,
				InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet,
			})
			if err != nil {
				return err
			}
			query, err := parseQueryJSON(queryText)
			if err != nil {
				return writeRawAPIInputError(cmd, "INVALID_API_QUERY", err)
			}
			body, err := parseBodyJSON(bodyText)
			if err != nil {
				return writeRawAPIInputError(cmd, "INVALID_API_BODY", err)
			}
			method := strings.ToUpper(strings.TrimSpace(args[0]))
			if rawAPIRequiresConfirmation(method) {
				request := map[string]any{"method": method, "path": path, "query": query, "body": body}
				writePreview := BuildWritePreview(WriteOperationSpec{
					Task:          "Open API 兜底写操作预览",
					Command:       "chaitin-cli tanswer api",
					OperationType: "tanswer_raw_api_request",
					RiskLevel:     "write_high",
					Target:        map[string]any{"method": method, "path": path},
					ChangeSummary: request,
					Impact:        map[string]any{"request_will_be_sent": true},
					RiskWarnings: []string{
						"Raw Open API requests do not have product-specific change summaries and may change T-Answer state.",
						"Confirm the HTTP method, path, query, and body are the intended authorized request before executing.",
					},
					ConfirmToken: rawAPIWriteConfirmToken,
				})
				if preview || strings.TrimSpace(confirm) == "" {
					return writeRawAPIResult(cmd, true, "Open API 兜底写操作预览", request, writePreview)
				}
				if err := ValidateWriteConfirmation(confirm, rawAPIWriteConfirmToken); err != nil {
					return writeRawAPIConfirmationError(cmd, err)
				}
			}

			client := NewClient(cfg)
			resp, err := client.DoJSON(cmd.Context(), method, path, query, body)
			if err != nil {
				return err
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: resp.StatusCode >= 200 && resp.StatusCode < 300,
				Task:    "Open API 兜底调用",
				Command: "chaitin-cli tanswer api",
				Data: map[string]any{
					"status_code": resp.StatusCode,
					"raw":         rawAPIResponseValue(resp.Body),
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&queryText, "query", "", "query parameters as JSON object")
	cmd.Flags().StringVar(&bodyText, "body", "", "request body as JSON object or @file path")
	cmd.Flags().BoolVar(&preview, "preview", false, "return request preview without sending a potentially mutating request")
	cmd.Flags().StringVar(&confirm, "confirm", "", "exact confirmation token required for potentially mutating requests")
	return cmd
}

func rawAPIRequiresConfirmation(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET", "HEAD":
		return false
	default:
		return true
	}
}

func writeRawAPIResult(cmd *cobra.Command, success bool, task string, query any, data any) error {
	raw, err := RenderJSON(SuccessEnvelope{
		Success: success,
		Task:    task,
		Command: "chaitin-cli tanswer api",
		Query:   query,
		Data:    data,
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func writeRawAPIConfirmationError(cmd *cobra.Command, err error) error {
	raw, renderErr := RenderJSON(ErrorEnvelope{
		Success: false,
		Task:    "Open API 兜底写操作",
		Command: "chaitin-cli tanswer api",
		Error:   CLIError{Code: "RAW_API_CONFIRMATION_REQUIRED", Message: err.Error(), Retryable: false},
	})
	if renderErr != nil {
		return renderErr
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func writeRawAPIInputError(cmd *cobra.Command, code string, err error) error {
	raw, renderErr := RenderJSON(ErrorEnvelope{
		Success: false,
		Task:    "Open API 兜底调用",
		Command: "chaitin-cli tanswer api",
		Error:   CLIError{Code: code, Message: err.Error(), Retryable: false},
	})
	if renderErr != nil {
		return renderErr
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func validateRawAPIPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	parsed, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() || parsed.Host != "" {
		return "", fmt.Errorf("api path must be a path under the configured T-Answer URL, not a full URL")
	}
	if !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("api path must start with /")
	}
	return path, nil
}

func rawAPIResponseValue(body []byte) any {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return ""
	}
	if json.Valid(trimmed) {
		return json.RawMessage(trimmed)
	}
	return string(body)
}

func parseQueryJSON(text string) (map[string]string, error) {
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}
	raw := map[string]any{}
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return nil, err
	}
	out := map[string]string{}
	for key, value := range raw {
		out[key] = fmt.Sprint(value)
	}
	return out, nil
}

func parseBodyJSON(text string) (any, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	if strings.HasPrefix(text, "@") {
		raw, err := os.ReadFile(strings.TrimPrefix(text, "@"))
		if err != nil {
			return nil, err
		}
		text = string(raw)
	}
	var body any
	if err := json.Unmarshal([]byte(text), &body); err != nil {
		return nil, err
	}
	return body, nil
}
