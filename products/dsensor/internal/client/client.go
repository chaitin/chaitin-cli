package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles HTTP communication with the D-Sensor API.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// New creates a new API client.
func New(baseURL, apiKey string, insecure bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

// APIResponse wraps the three response patterns used by D-Sensor.
type APIResponse struct {
	Data  json.RawMessage `json:"data"`
	Err   string          `json:"err,omitempty"`
	Msg   string          `json:"msg,omitempty"`
	Total *int            `json:"total,omitempty"`
}

// Do sends a POST request to the given path with the JSON body.
func (c *Client) Do(path string, body []byte) (*http.Response, []byte, error) {
	if c.BaseURL == "" {
		return nil, nil, fmt.Errorf("未配置 URL，请使用 --url 或设置 DSENSOR_URL 环境变量")
	}

	url := c.BaseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("API-Token", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return resp, respBody, fmt.Errorf("API 返回错误 (状态码 %d): %s", resp.StatusCode, string(respBody))
	}

	return resp, respBody, nil
}

// DoAndParse sends a request and parses the response into the API wrapper.
func (c *Client) DoAndParse(path string, body []byte) (*http.Response, *APIResponse, error) {
	resp, respBody, err := c.Do(path, body)
	if err != nil {
		return resp, nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return resp, nil, fmt.Errorf("解析响应 JSON 失败: %w\n原始响应: %s", err, string(respBody))
	}

	if apiResp.Err != "" {
		return resp, &apiResp, fmt.Errorf("API 业务错误: %s (msg: %s)", apiResp.Err, apiResp.Msg)
	}

	return resp, &apiResp, nil
}
