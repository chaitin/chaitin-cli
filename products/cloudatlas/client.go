package cloudatlas

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	cfg        Config
	httpClient *http.Client
	baseURL    string
	verbose    bool
}

type dryRunResult struct{}

func (dryRunResult) Error() string { return "dry run" }

func NewClient(cfg Config, insecure bool, verbose bool) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
			},
		},
		baseURL: strings.TrimRight(cfg.URL, "/"),
		verbose: verbose,
	}
}

func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any) (json.RawMessage, error) {
	if c.baseURL == "" && !dryRun {
		return nil, fmt.Errorf("cloudAtlas URL is required (use --url or configure cloudAtlas.url / CLOUD_ATLAS_URL)")
	}

	reqURL := c.buildURL(path)
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.injectHeaders(req, body != nil)

	if c.verbose || dryRun {
		logRequest(req, body)
	}
	if dryRun {
		payload := map[string]any{
			"method": method,
			"url":    reqURL,
		}
		if body != nil {
			payload["body"] = body
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		return data, dryRunResult{}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return handleResponse(resp)
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
	if c.cfg.Token != "" {
		req.Header.Set("Token", c.cfg.Token)
	}
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
}

func handleResponse(resp *http.Response) (json.RawMessage, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if len(body) == 0 {
		return nil, nil
	}
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") || bytes.HasPrefix(bytes.TrimSpace(body), []byte("<!doctype html>")) || bytes.HasPrefix(bytes.TrimSpace(body), []byte("<html")) {
		return nil, fmt.Errorf("API returned HTML instead of JSON; check that --url points to the API base path")
	}

	var envelope APIEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && (envelope.Code != 0 || envelope.Message != "" || envelope.Data != nil) {
		if envelope.Code != 0 && envelope.Code != 200 {
			return nil, fmt.Errorf("Cloud Atlas error %d: %s", envelope.Code, envelope.Message)
		}
		return envelope.Data, nil
	}

	return body, nil
}

func logRequest(req *http.Request, body any) {
	fmt.Fprintf(os.Stderr, "URL: %s %s\n", req.Method, req.URL.String())
	if len(req.Header) > 0 {
		headers := req.Header.Clone()
		if !verboseSensitive {
			maskHeader(headers, "Token")
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
