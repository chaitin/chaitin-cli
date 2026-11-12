package tanswer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type Client struct {
	cfg        Config
	httpClient *http.Client
}

type HTTPResponse struct {
	StatusCode int             `json:"status_code"`
	Header     http.Header     `json:"-"`
	Body       json.RawMessage `json:"body"`
}

type DownloadedFile struct {
	StatusCode int
	Header     http.Header
	FileName   string
	Body       []byte
}

type UploadedFile struct {
	StatusCode int             `json:"status_code"`
	Header     http.Header     `json:"-"`
	Body       json.RawMessage `json:"body"`
}

func NewClient(cfg Config) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // Explicit CLI opt-in to bypass certificate validation.
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
	}
}

func (c *Client) DoJSON(ctx context.Context, method, path string, query map[string]string, body any) (*HTTPResponse, error) {
	target, err := c.url(path, query)
	if err != nil {
		return nil, err
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return nil, err
	}
	if c.cfg.APIToken != "" {
		req.Header.Set("Api-Token", c.cfg.APIToken)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &HTTPResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: json.RawMessage(raw)}, nil
}

func (c *Client) Download(ctx context.Context, methodID string, query any) (*DownloadedFile, error) {
	rawQuery := []byte("{}")
	if query != nil {
		raw, err := json.Marshal(query)
		if err != nil {
			return nil, err
		}
		rawQuery = raw
	}
	target, err := c.url("/api/download", map[string]string{
		"id":    methodID,
		"query": base64.StdEncoding.EncodeToString(rawQuery),
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	if c.cfg.APIToken != "" {
		req.Header.Set("Api-Token", c.cfg.APIToken)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &DownloadedFile{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		FileName:   filenameFromContentDisposition(resp.Header.Get("Content-Disposition")),
		Body:       body,
	}, nil
}

func (c *Client) UploadFile(ctx context.Context, methodID string, fileName string, reader io.Reader) (*UploadedFile, error) {
	return c.UploadFileWithFields(ctx, methodID, fileName, reader, nil)
}

func (c *Client) UploadFileWithFields(ctx context.Context, methodID string, fileName string, reader io.Reader, fields map[string]string) (*UploadedFile, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, err
		}
	}
	part, err := writer.CreateFormFile("file", filepath.Base(fileName))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, reader); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	target, err := c.url("/api/upload", map[string]string{"id": methodID})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, &body)
	if err != nil {
		return nil, err
	}
	if c.cfg.APIToken != "" {
		req.Header.Set("Api-Token", c.cfg.APIToken)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &UploadedFile{StatusCode: resp.StatusCode, Header: resp.Header, Body: json.RawMessage(raw)}, nil
}

func filenameFromContentDisposition(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(value)
	if err != nil {
		return ""
	}
	return params["filename"]
}

func (c *Client) url(path string, query map[string]string) (string, error) {
	path = strings.TrimSpace(path)
	parsedPath, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	if parsedPath.IsAbs() || parsedPath.Host != "" {
		return "", fmt.Errorf("api path must be under the configured TAnswer URL, not a full URL")
	}
	if !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("api path must start with /")
	}
	u, err := url.Parse(c.cfg.BaseURL)
	if err != nil {
		return "", err
	}
	u.Path = strings.TrimRight(u.Path, "/") + parsedPath.Path
	u.RawPath = ""
	u.RawQuery = parsedPath.RawQuery
	u.Fragment = ""
	values := u.Query()
	for key, value := range query {
		values.Set(key, value)
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}
