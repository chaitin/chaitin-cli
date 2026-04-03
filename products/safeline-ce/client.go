package safelinece

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client HTTP 客户端
type Client struct {
	config     *Config
	httpClient *http.Client
	baseURL    string
}

// NewClient 创建新的 HTTP 客户端
func NewClient(cfg *Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		baseURL: strings.TrimSuffix(cfg.Endpoint, "/"),
	}
}

// Do 执行 HTTP 请求
func (c *Client) Do(ctx context.Context, method, path string, body, result interface{}) error {
	reqURL := c.buildURL(path)

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

	c.injectHeaders(req, false)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewNetworkError("request failed", err)
	}
	defer resp.Body.Close()

	return c.handleResponse(resp, result)
}

// Get 执行 GET 请求
func (c *Client) Get(ctx context.Context, path string, query url.Values, result interface{}) error {
	if len(query) > 0 {
		path = path + "?" + query.Encode()
	}
	return c.Do(ctx, "GET", path, nil, result)
}

// Post 执行 POST 请求
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	return c.Do(ctx, "POST", path, body, result)
}

// Put 执行 PUT 请求
func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	return c.Do(ctx, "PUT", path, body, result)
}

// Delete 执行 DELETE 请求
func (c *Client) Delete(ctx context.Context, path string, result interface{}) error {
	return c.Do(ctx, "DELETE", path, nil, result)
}

// UploadFile 上传文件
func (c *Client) UploadFile(ctx context.Context, path string, files map[string]string, result interface{}) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for fieldname, filepath := range files {
		file, err := os.Open(filepath)
		if err != nil {
			return NewConfigError(fmt.Sprintf("failed to open file %s", filepath), err)
		}
		defer file.Close()

		part, err := writer.CreateFormFile(fieldname, filepath)
		if err != nil {
			return fmt.Errorf("failed to create form file: %w", err)
		}

		if _, err := io.Copy(part, file); err != nil {
			return fmt.Errorf("failed to copy file content: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	reqURL := c.buildURL(path)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, &buf)
	if err != nil {
		return NewNetworkError("failed to create request", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-SLCE-API-TOKEN", c.config.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewNetworkError("request failed", err)
	}
	defer resp.Body.Close()

	return c.handleResponse(resp, result)
}

func (c *Client) buildURL(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.baseURL + path
}

func (c *Client) injectHeaders(req *http.Request, isMultipart bool) {
	req.Header.Set("X-SLCE-API-TOKEN", c.config.Token)
	if !isMultipart {
		req.Header.Set("Content-Type", "application/json")
	}
}

func (c *Client) handleResponse(resp *http.Response, result interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewNetworkError("failed to read response body", err)
	}

	if resp.StatusCode >= 400 {
		return c.handleError(resp.StatusCode, body)
	}

	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

func (c *Client) handleError(statusCode int, body []byte) error {
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err == nil && len(apiResp.Errors) > 0 {
		msgs := make([]string, len(apiResp.Errors))
		for i, e := range apiResp.Errors {
			msgs[i] = e.Message
		}
		return NewAPIError(statusCode, strings.Join(msgs, "; "))
	}

	return NewAPIError(statusCode, fmt.Sprintf("API request failed (status %d)", statusCode))
}
