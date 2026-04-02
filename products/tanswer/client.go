package tanswer

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/chaitin/workspace-cli/config"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

var (
	serverURL string
	apiToken  string
)

type runtimeConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

// jsonRPCRequest 是 JSON-RPC 2.0 请求体。
type jsonRPCRequest struct {
	ID      string                 `json:"id"`
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// doRequest 构造 JSON-RPC 2.0 请求并发送到 /rpc 端点。
func doRequest(cmd *cobra.Command, method string, params map[string]interface{}, raw bool) error {
	server := serverURL
	if server == "" {
		return fmt.Errorf("tanswer url is not configured")
	}
	server = strings.TrimRight(server, "/")

	url := server + "/rpc"

	if params == nil {
		params = make(map[string]interface{})
	}

	rpcReq := jsonRPCRequest{
		ID:      uuid.New().String(),
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(rpcReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if apiToken != "" {
		req.Header.Set("API-Token", apiToken)
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

	return outputResponse(cmd.OutOrStdout(), respBody, raw)
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

func ApplyRuntimeConfig(cmd *cobra.Command, cfg config.Raw) {
	productCfg, err := config.DecodeProduct[runtimeConfig](cfg, "tanswer")
	if err != nil {
		return
	}

	if flag := cmd.Flags().Lookup("url"); flag != nil && flag.Changed {
		value, flagErr := cmd.Flags().GetString("url")
		if flagErr == nil {
			serverURL = value
		}
	} else {
		serverURL = productCfg.URL
	}

	if flag := cmd.Flags().Lookup("api-key"); flag != nil && flag.Changed {
		value, flagErr := cmd.Flags().GetString("api-key")
		if flagErr == nil {
			apiToken = value
		}
	} else {
		apiToken = productCfg.APIKey
	}
}
