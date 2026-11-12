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
		req.Header.Set("TOKEN", c.cfg.Token)
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
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") || bytes.HasPrefix(bytes.TrimSpace(body), []byte("<!doctype html>")) || bytes.HasPrefix(bytes.TrimSpace(body), []byte("<html")) {
		errorPrefix := "API returned HTML instead of JSON"
		if resp.StatusCode >= http.StatusBadRequest {
			errorPrefix = fmt.Sprintf("API request failed (status %d): returned HTML instead of JSON", resp.StatusCode)
		}
		if requestURL := responseRequestURL(resp); requestURL != nil && isVersionedAPIPath(requestURL.Path, "") {
			return nil, fmt.Errorf("%s; --url may point to the Cloud Atlas console root; try --url %s", errorPrefix, suggestedOpenAPIBaseURL(requestURL))
		}
		return nil, fmt.Errorf("%s; check that --url points to the API base path", errorPrefix)
	}
	if resp.StatusCode >= 400 {
		if resp.StatusCode == http.StatusUnauthorized {
			if requestURL := responseRequestURL(resp); requestURL != nil && isVersionedAPIPath(requestURL.Path, "/api") {
				return nil, fmt.Errorf("API request failed (status %d); --url may point to the browser-session API; try --url %s for TOKEN authentication", resp.StatusCode, suggestedOpenAPIBaseURL(requestURL))
			}
		}
		return nil, fmt.Errorf("API request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if len(body) == 0 {
		return nil, nil
	}

	var envelope APIEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && (envelope.Code != 0 || envelope.Message != "" || envelope.Data != nil) {
		if envelope.Code != 0 && envelope.Code != 200 {
			if envelope.Code == http.StatusUnauthorized {
				if requestURL := responseRequestURL(resp); requestURL != nil && isVersionedAPIPath(requestURL.Path, "/api") {
					return nil, fmt.Errorf("Cloud Atlas error %d: %s; --url may point to the browser-session API; try --url %s for TOKEN authentication", envelope.Code, envelope.Message, suggestedOpenAPIBaseURL(requestURL))
				}
			}
			return nil, fmt.Errorf("Cloud Atlas error %d: %s", envelope.Code, envelope.Message)
		}
		return envelope.Data, nil
	}

	return body, nil
}

func responseRequestURL(resp *http.Response) *url.URL {
	if resp == nil || resp.Request == nil {
		return nil
	}
	return resp.Request.URL
}

func isVersionedAPIPath(path, prefix string) bool {
	for _, version := range []string{"v1", "v2"} {
		base := prefix + "/" + version
		if path == base || strings.HasPrefix(path, base+"/") {
			return true
		}
	}
	return false
}

func suggestedOpenAPIBaseURL(requestURL *url.URL) string {
	suggestion := &url.URL{
		Scheme: requestURL.Scheme,
		Host:   requestURL.Host,
		Path:   "/openapi",
	}
	return suggestion.String()
}

func logRequest(req *http.Request, body any) {
	fmt.Fprintf(os.Stderr, "URL: %s %s\n", req.Method, req.URL.String())
	if len(req.Header) > 0 {
		headers := req.Header.Clone()
		if !verboseSensitive {
			maskHeader(headers, "TOKEN")
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
