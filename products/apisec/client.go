package apisec

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/chaitin/chaitin-cli/egressproxy"
	"net/textproto"
	"net/url"
	"os"
	"strings"
	"time"
)

var dryRun bool

type dryRunResult struct{}

func (dryRunResult) Error() string { return "dry run" }

type Client struct {
	config     *Config
	httpClient *http.Client
	baseURL    string
	verbose    bool
}

type responseEnvelope struct {
	Err  any             `json:"err"`
	Data json.RawMessage `json:"data"`
	Msg  any             `json:"msg"`
}

func NewClient(cfg *Config, verbose bool) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: egressproxy.Wrap(&http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}),
		},
		baseURL: strings.TrimSuffix(cfg.URL, "/"),
		verbose: verbose,
	}
}

func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any, result any) error {
	reqURL := c.buildURL(path)
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	c.injectHeaders(req, body != nil)

	if c.verbose || dryRun {
		logRequest(req, body)
	}
	if dryRun {
		return dryRunResult{}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.handleResponse(resp, result)
}

func (c *Client) buildURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.baseURL + path
}

func (c *Client) injectHeaders(req *http.Request, hasBody bool) {
	if c.config.APIToken != "" {
		req.Header.Set("API-TOKEN", c.config.APIToken)
	}
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
}

func (c *Client) handleResponse(resp *http.Response, result any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if len(body) == 0 || result == nil {
		return nil
	}

	var envelope responseEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && (envelope.Data != nil || envelope.Err != nil || envelope.Msg != nil) {
		if hasAPIError(envelope.Err) {
			return fmt.Errorf("APISec error %s: %s", valueString(envelope.Err), valueString(envelope.Msg))
		}
		if envelope.Data == nil {
			return nil
		}
		return json.Unmarshal(envelope.Data, result)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	return nil
}

func hasAPIError(value any) bool {
	switch v := value.(type) {
	case nil:
		return false
	case string:
		return v != ""
	default:
		return true
	}
}

func valueString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(data)
	}
}

func logRequest(req *http.Request, body any) {
	fmt.Fprintf(os.Stderr, "URL: %s %s\n", req.Method, req.URL.String())
	if len(req.Header) > 0 {
		headers := req.Header.Clone()
		if !verboseSensitive {
			maskHeader(headers, "API-TOKEN")
		}
		data, err := json.MarshalIndent(headers, "", "  ")
		if err == nil {
			fmt.Fprintf(os.Stderr, "Headers:\n%s\n", string(data))
		}
	}
	if body != nil {
		data, err := json.MarshalIndent(body, "", "  ")
		if err == nil {
			fmt.Fprintf(os.Stderr, "Body:\n%s\n", string(data))
			return
		}
		fmt.Fprintf(os.Stderr, "Body: %v\n", body)
	}
}

func maskHeader(headers http.Header, name string) {
	name = textproto.CanonicalMIMEHeaderKey(name)
	values, ok := headers[name]
	if !ok {
		return
	}
	masked := make([]string, 0, len(values))
	for _, value := range values {
		masked = append(masked, maskSecret(value))
	}
	headers[name] = masked
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "..." + value[len(value)-4:]
}
