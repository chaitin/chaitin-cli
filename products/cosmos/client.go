package cosmos

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

// jsonRPCRequest 是 JSON-RPC 2.0 请求体。
type jsonRPCRequest struct {
	ID      string      `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type jsonRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e jsonRPCError) Error() string {
	message := strings.TrimSpace(e.Message)
	if message == "" {
		message = "unknown error"
	}

	if e.Code != 0 {
		if len(e.Data) > 0 {
			return fmt.Sprintf("json-rpc error %d: %s: %s", e.Code, message, compactJSON(e.Data))
		}
		return fmt.Sprintf("json-rpc error %d: %s", e.Code, message)
	}
	if len(e.Data) > 0 {
		return fmt.Sprintf("json-rpc error: %s: %s", message, compactJSON(e.Data))
	}
	return fmt.Sprintf("json-rpc error: %s", message)
}

type dryRunRequestSummary struct {
	DryRun     bool              `json:"dry_run"`
	URL        string            `json:"url"`
	HTTPMethod string            `json:"http_method"`
	Headers    map[string]string `json:"headers"`
	Body       jsonRPCRequest    `json:"body"`
}

// doRequest 构造 JSON-RPC 2.0 请求并发送到 /rpc 端点。
func doRequest(cmd *cobra.Command, method string, params map[string]interface{}, raw bool) error {
	return doRequestWithParamStyle(cmd, method, params, raw, "")
}

func doRequestWithParamStyle(cmd *cobra.Command, method string, params map[string]interface{}, raw bool, paramStyle string) error {
	server := serverURL
	if server == "" {
		return fmt.Errorf("cosmos url is not configured")
	}
	server = strings.TrimRight(server, "/")

	url := server + "/pedestal/rpc"

	if params == nil {
		params = make(map[string]interface{})
	}

	var rpcParams interface{} = params
	if paramStyle == paramStyleArray {
		rpcParams = []interface{}{params}
	}

	rpcReq := jsonRPCRequest{
		ID:      uuid.New().String(),
		JSONRPC: "2.0",
		Method:  method,
		Params:  rpcParams,
	}

	data, err := json.Marshal(rpcReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	if dryRun {
		return outputDryRunRequest(cmd.OutOrStdout(), url, rpcReq, raw)
	}

	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+apiToken)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %d - %s\n", resp.StatusCode, string(respBody))
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	if err := parseJSONRPCError(respBody); err != nil {
		return err
	}

	return outputResponse(cmd.OutOrStdout(), respBody, raw)
}

func parseJSONRPCError(data []byte) error {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}

	var envelope struct {
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil
	}
	if len(envelope.Error) == 0 || bytes.Equal(bytes.TrimSpace(envelope.Error), []byte("null")) {
		return nil
	}

	var rpcErr jsonRPCError
	if err := json.Unmarshal(envelope.Error, &rpcErr); err != nil {
		return fmt.Errorf("json-rpc error: %s", string(envelope.Error))
	}
	return rpcErr
}

func compactJSON(data []byte) string {
	var compacted bytes.Buffer
	if err := json.Compact(&compacted, data); err == nil {
		return compacted.String()
	}
	return string(data)
}

func outputDryRunRequest(w io.Writer, url string, rpcReq jsonRPCRequest, raw bool) error {
	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}
	if apiToken != "" {
		headers["Authorization"] = "Bearer " + maskSecret(apiToken)
	}

	summary := dryRunRequestSummary{
		DryRun:     true,
		URL:        url,
		HTTPMethod: http.MethodPost,
		Headers:    headers,
		Body:       rpcReq,
	}

	data, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal dry-run request: %w", err)
	}

	return outputResponse(w, data, raw)
}

func maskSecret(secret string) string {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ""
	}
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}

// outputResponse 输出响应体。raw=true 时输出紧凑 JSON，否则格式化输出。
func outputResponse(w io.Writer, data []byte, raw bool) error {
	if raw || len(data) == 0 {
		_, err := fmt.Fprintln(w, string(data))
		return err
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		_, err := fmt.Fprintln(w, string(data))
		return err
	}

	_, err := fmt.Fprintln(w, pretty.String())
	return err
}
