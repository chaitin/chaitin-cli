package veinmind

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/chaitin/chaitin-cli/egressproxy"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
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
			Transport: egressproxy.Wrap(&http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}),
			Jar: mustCookieJar(),
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

	if dryRun {
		return renderDryRun(req, body, c.httpClient.Jar)
	}

	if err := c.authenticate(ctx); err != nil {
		return err
	}
	if c.verbose {
		logRequest(req, body, c.httpClient.Jar)
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
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
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
}

func mustCookieJar() http.CookieJar {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return jar
}

func (c *Client) authenticate(ctx context.Context) error {
	if c.config.APIKey == "" {
		return nil
	}

	apiSecret, apiKey, err := splitAPIToken(c.config.APIKey)
	if err != nil {
		return err
	}

	nowTime := time.Now().Unix()
	payload := map[string]interface{}{
		"key":      apiKey,
		"sign":     generateTokenSign(apiKey, apiSecret, nowTime),
		"now_time": nowTime,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL("/auth/token_login"), bytes.NewReader(data))
	if err != nil {
		return NewNetworkError("failed to create auth request", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewNetworkError("auth request failed", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewNetworkError("failed to read auth response body", err)
	}
	if resp.StatusCode >= 300 {
		return NewAPIError(resp.StatusCode, fmt.Sprintf("auth request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body))))
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.Contains(contentType, "application/json") {
		return NewAPIError(resp.StatusCode, fmt.Sprintf("auth response is not JSON (status %d, content-type %q): %s",
			resp.StatusCode,
			resp.Header.Get("Content-Type"),
			responseSnippet(body),
		))
	}

	var authResp APIResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return NewAPIError(resp.StatusCode, fmt.Sprintf("failed to parse auth JSON response (status %d): %v; body: %s",
			resp.StatusCode,
			err,
			responseSnippet(body),
		))
	}
	if c.httpClient.Jar == nil || len(c.httpClient.Jar.Cookies(req.URL)) == 0 {
		return NewAPIError(resp.StatusCode, fmt.Sprintf("auth did not establish session: code=%d msg=%q", authResp.Code, authResp.Msg))
	}

	return nil
}

func splitAPIToken(token string) (string, string, error) {
	if len(token) != 64 {
		return "", "", NewConfigError("veinmind api-key must be the complete 64-character API token", nil)
	}
	return token[:32], token[32:], nil
}

func generateTokenSign(key, secret string, nowTime int64) string {
	index := nowTime % 16
	var data string
	if index == 0 {
		data = "CT-VM-" + strconv.FormatInt(nowTime, 10) + key
	} else {
		data = key[:index] + "CT-VM-" + strconv.FormatInt(nowTime, 10) + key[index:]
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
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
			return fmt.Errorf("failed to parse JSON response (status %d, content-type %q): %w; body: %s",
				resp.StatusCode,
				resp.Header.Get("Content-Type"),
				err,
				responseSnippet(body),
			)
		}
	}

	return nil
}

func responseSnippet(body []byte) string {
	const limit = 300

	snippet := strings.TrimSpace(string(body))
	if len(snippet) > limit {
		snippet = snippet[:limit] + "..."
	}
	return snippet
}

func renderDryRun(req *http.Request, body interface{}, jar http.CookieJar) error {
	logRequest(req, body, jar)
	return nil
}

func logRequest(req *http.Request, body interface{}, jar http.CookieJar) {
	fmt.Fprintf(os.Stderr, "URL: %s %s\n", req.Method, req.URL.String())
	headers := req.Header.Clone()
	if jar != nil {
		if cookies := jar.Cookies(req.URL); len(cookies) > 0 {
			var cookieValues []string
			for _, cookie := range cookies {
				cookieValues = append(cookieValues, cookie.Name+"="+cookie.Value)
			}
			headers.Set("Cookie", strings.Join(cookieValues, "; "))
		}
	}
	if len(headers) > 0 {
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
