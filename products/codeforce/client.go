package codeforce

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/chaitin/chaitin-cli/egressproxy"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    string
}

type responseEnvelope struct {
	Code    *json.RawMessage `json:"code"`
	Data    json.RawMessage  `json:"data"`
	Message any              `json:"message"`
	Msg     any              `json:"msg"`
	Error   any              `json:"error"`
}

func NewClient(cfg Config) *Client {
	cfg = cfg.normalized()
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
			Transport: egressproxy.Wrap(&http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.Insecure,
				},
			}),
		},
		baseURL: strings.TrimRight(cfg.URL, "/"),
	}
}

func (c *Client) DoJSON(ctx context.Context, method, path string, query url.Values, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path, query), reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	c.injectHeaders(req, body != nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.handleJSONResponse(resp, result)
}

func (c *Client) DoMultipart(ctx context.Context, method, path string, fields map[string]string, files map[string]string, result any) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, value := range fields {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if err := writer.WriteField(key, value); err != nil {
			return fmt.Errorf("write multipart field %s: %w", key, err)
		}
	}
	for field, filePath := range files {
		if strings.TrimSpace(filePath) == "" {
			continue
		}
		if err := addMultipartFile(writer, field, filePath); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path, nil), &buf)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	c.injectHeaders(req, false)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.handleJSONResponse(resp, result)
}

func (c *Client) Download(ctx context.Context, method, path string, query url.Values, body any) ([]byte, http.Header, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path, query), reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}
	c.injectHeaders(req, body != nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.Header, fmt.Errorf("API request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return data, resp.Header, nil
}

func addMultipartFile(writer *multipart.Writer, field, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file %s: %w", filePath, err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile(field, filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("create multipart file field %s: %w", field, err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy file %s: %w", filePath, err)
	}
	return nil
}

func (c *Client) buildURL(path string, query url.Values) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		if len(query) > 0 {
			return path + "?" + query.Encode()
		}
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
}

func (c *Client) injectHeaders(req *http.Request, hasJSONBody bool) {
	token := strings.TrimSpace(c.config.AccessToken)
	if token != "" {
		req.Header.Set("Authorization", bearerToken(token))
		req.Header.Set("X-API-Key", tokenWithoutBearer(token))
		req.Header.Set("Token", tokenWithoutBearer(token))
		req.Header.Set("Cookie", "port=; user_role=1; userId=1; authorized_token="+tokenWithoutBearer(token)+"; refresh_token="+tokenWithoutBearer(token))
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN")
	req.Header.Set("User-Agent", codeforceUserAgent)
	req.Header.Set("Referer", c.baseURL+"/")
	if hasJSONBody {
		req.Header.Set("Content-Type", "application/json")
	}
}

func bearerToken(token string) string {
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return token
	}
	return "Bearer " + token
}

func tokenWithoutBearer(token string) string {
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return strings.TrimSpace(token[len("bearer "):])
	}
	return token
}

func (c *Client) handleJSONResponse(resp *http.Response, result any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}

	var envelope responseEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && (envelope.Code != nil || envelope.Data != nil || envelope.Message != nil || envelope.Msg != nil || envelope.Error != nil) {
		if envelope.Code != nil && !isSuccessCode(*envelope.Code) {
			return fmt.Errorf("CodeForce error %s: %s", rawJSONString(*envelope.Code), firstNonEmpty(valueString(envelope.Message), valueString(envelope.Msg), valueString(envelope.Error)))
		}
		if result == nil || len(envelope.Data) == 0 || string(envelope.Data) == "null" {
			return nil
		}
		if err := json.Unmarshal(envelope.Data, result); err != nil {
			return fmt.Errorf("parse response data: %w", err)
		}
		return nil
	}

	if result == nil {
		return nil
	}
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	return nil
}

func isSuccessCode(raw json.RawMessage) bool {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	switch v := value.(type) {
	case float64:
		return v == 0 || v == 200
	case string:
		return v == "0" || v == "200" || strings.EqualFold(v, "success")
	case nil:
		return true
	default:
		return false
	}
}

func rawJSONString(raw json.RawMessage) string {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	return valueString(value)
}

func valueString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.RawMessage:
		return rawJSONString(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(data)
	}
}
