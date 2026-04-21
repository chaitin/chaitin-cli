package ddr

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	config     *Config
	headers    map[string]string
	httpClient *http.Client
	baseURL    string
	verbose    bool
}

func NewClient(cfg *Config, headers map[string]string, verbose bool) *Client {
	return &Client{
		config:  cfg,
		headers: headers,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		baseURL: strings.TrimSuffix(cfg.URL, "/"),
		verbose: verbose,
	}
}

func (c *Client) Do(ctx context.Context, method, path string, query url.Values, headers map[string]string, body, result interface{}) error {
	reqURL := c.buildURL(path)
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return NewNetworkError("failed to create request", err)
	}

	c.injectHeaders(req, headers, body != nil)
	if c.verbose {
		logRequest(req, body)
	}

	if dryRun {
		return renderDryRun(req, body, c.verbose)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewNetworkError("request failed", err)
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

func (c *Client) injectHeaders(req *http.Request, headers map[string]string, hasBody bool) {
	authInfo := c.config.APIKey
	if !strings.HasPrefix(authInfo, "Serval ") && !strings.HasPrefix(authInfo, "Bearer ") {
		authInfo = "Serval " + authInfo
	}
	req.Header.Set("Authorization", authInfo)
	req.Header.Set("X-CS-Header-Company", c.config.CompanyID)
	for key, value := range c.headers {
		if value == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	for key, value := range headers {
		if value == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	if hasBody && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
}

func (c *Client) handleResponse(resp *http.Response, result interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewNetworkError("failed to read response body", err)
	}

	if resp.StatusCode >= 400 {
		return NewAPIError(resp.StatusCode, fmt.Sprintf("API request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body))))
	}

	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

func renderDryRun(req *http.Request, body interface{}, alreadyLogged bool) error {
	if !alreadyLogged {
		logRequest(req, body)
	}
	return nil
}

func logRequest(req *http.Request, body interface{}) {
	fmt.Fprintf(os.Stderr, "URL: %s %s\n", req.Method, req.URL.String())
	if len(req.Header) > 0 {
		data, err := json.MarshalIndent(req.Header, "", "  ")
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
