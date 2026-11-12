package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/chaitin/chaitin-cli/products/insight/models"
)

// Client is the HTTP client for Insight APIs
type Client struct {
	cfg      models.Config
	insecure bool
	verbose  bool
	dryRun   bool
	out      io.Writer
	errOut   io.Writer
}

func NewClient(cfg models.Config, insecure, verbose, dryRun bool, out, errOut io.Writer) *Client {
	return &Client{
		cfg:      cfg,
		insecure: insecure,
		verbose:  verbose,
		dryRun:   dryRun,
		out:      out,
		errOut:   errOut,
	}
}

func (c *Client) Request(method, path string, body io.Reader) ([]byte, error) {
	// Only parse the base URL
	u, err := url.Parse(c.cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	// Parse the incoming path to safely merge any query parameters
	parsedPath, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid request path: %v", err)
	}

	// Safely merge path and raw queries so they don't get double encoded
	u.Path, err = url.JoinPath(u.Path, parsedPath.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to join paths: %v", err)
	}
	u.RawQuery = parsedPath.RawQuery

	if c.dryRun {
		fmt.Fprintf(c.errOut, "[DRY-RUN] Would send %s %s\n", method, u.String())
		return []byte(`{"status": "dry-run ok"}`), nil
	}

	if c.verbose {
		fmt.Fprintf(c.errOut, "> %s %s\n", method, u.String())
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
		// Insight's GetUserMiddleware specifically requires a Cookie header to fetch user info
		req.Header.Set("Cookie", "jwt="+c.cfg.APIKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// NOTE: In production, configure insecure transport if needed.
	transport := &http.Transport{}
	if c.insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if c.verbose {
		fmt.Fprintf(c.errOut, "< HTTP %s\n", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Try to parse standard Insight common response structure
	// Format is typically {"code": ..., "msg": "...", "data": ...}
	var apiErr struct {
		Code *int   `json:"code"`
		Msg  string `json:"msg"`
	}
	hasAPIErrorShape := json.Unmarshal(bodyBytes, &apiErr) == nil && apiErr.Code != nil

	if resp.StatusCode >= 400 {
		if hasAPIErrorShape && apiErr.Msg != "" {
			return nil, fmt.Errorf("insight API error: %s (HTTP %d, Code %d)", apiErr.Msg, resp.StatusCode, *apiErr.Code)
		}
		// Fallback to raw body if not standard JSON error
		return nil, fmt.Errorf("insight API HTTP %d error: %s", resp.StatusCode, string(bodyBytes))
	}

	if hasAPIErrorShape && *apiErr.Code != 0 {
		msg := apiErr.Msg
		if msg == "" {
			msg = "request failed"
		}
		return nil, fmt.Errorf("insight API error: %s (Code %d)", msg, *apiErr.Code)
	}

	return bodyBytes, nil
}
