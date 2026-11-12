package monkeyscan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type client struct {
	baseURL    string
	key        string
	httpClient *http.Client
}

func newClient(baseURL, key string) *client {
	return &client{
		baseURL: strings.TrimRight(baseURL, "/"),
		key:     strings.TrimSpace(key),
		httpClient: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (c *client) Status(ctx context.Context) (*statusResponse, error) {
	var out statusResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/v1/cli/status", nil, &out); err != nil {
		return nil, err
	}
	if out.Ready && !out.ReviewReady {
		out.ReviewReady = true
	}
	return &out, nil
}

func (c *client) CreateDiffReview(ctx context.Context, req diffReviewRequest) (*diffReviewResponse, error) {
	var out diffReviewResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/cli/reviews/diff", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) ReviewDetail(ctx context.Context, runID string) (*reviewDetail, error) {
	var out reviewDetail
	if err := c.doJSON(ctx, http.MethodGet, "/api/v1/cli/reviews/"+url.PathEscape(runID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) CreateRepoScan(ctx context.Context, req scanCreateRequest) (*scanCreateResponse, error) {
	var out scanCreateResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/cli/scans?type=repo", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) CreateArchiveScan(ctx context.Context, filePath, taskName string) (*scanCreateResponse, error) {
	return c.CreateArchiveScanWithName(ctx, filePath, taskName, filepath.Base(filePath))
}

func (c *client) CreateArchiveScanWithName(ctx context.Context, filePath, taskName, fileName string) (*scanCreateResponse, error) {
	reader, writer := io.Pipe()
	form := multipart.NewWriter(writer)
	go writeArchiveScanMultipart(writer, form, filePath, taskName, fileName)
	var out scanCreateResponse
	if err := c.do(ctx, http.MethodPost, "/api/v1/cli/scans?type=archive", reader, form.FormDataContentType(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func writeArchiveScanMultipart(pipe *io.PipeWriter, form *multipart.Writer, filePath, taskName, fileName string) {
	var err error
	defer func() {
		if err != nil {
			_ = pipe.CloseWithError(err)
			return
		}
		_ = pipe.Close()
	}()
	if strings.TrimSpace(taskName) != "" {
		if err = form.WriteField("task_name", taskName); err != nil {
			err = fmt.Errorf("写入任务名称失败: %w", err)
			return
		}
	}
	if strings.TrimSpace(fileName) != "" {
		if err = form.WriteField("file_name", fileName); err != nil {
			err = fmt.Errorf("写入文件名称失败: %w", err)
			return
		}
	}
	file, openErr := os.Open(filePath)
	if openErr != nil {
		err = fmt.Errorf("打开上传文件失败: %w", openErr)
		return
	}
	defer file.Close()
	part, createErr := form.CreateFormFile("file", filepath.Base(filePath))
	if createErr != nil {
		err = fmt.Errorf("创建上传表单失败: %w", createErr)
		return
	}
	if _, err = io.Copy(part, file); err != nil {
		err = fmt.Errorf("写入上传文件失败: %w", err)
		return
	}
	if err = form.Close(); err != nil {
		err = fmt.Errorf("关闭上传表单失败: %w", err)
	}
}

func (c *client) ListScans(ctx context.Context) (*scanListResponse, error) {
	var out scanListResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/v1/cli/scans", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) ScanResult(ctx context.Context, taskGroupID string, full bool) (*scanResultResponse, error) {
	q := url.Values{}
	if strings.TrimSpace(taskGroupID) != "" {
		q.Set("task_group_id", strings.TrimSpace(taskGroupID))
	}
	if full {
		q.Set("full", strconv.FormatBool(full))
	}
	path := "/api/v1/cli/scans/result"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out scanResultResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	var reqBody io.Reader
	contentType := ""
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("编码请求失败: %w", err)
		}
		reqBody = bytes.NewReader(data)
		contentType = "application/json"
	}
	return c.do(ctx, method, path, reqBody, contentType, out)
}

func (c *client) do(ctx context.Context, method, path string, body io.Reader, contentType string, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求 MonkeyScan 服务失败: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("MonkeyScan API 返回错误 (%d): %s", resp.StatusCode, responseErrorMessage(data))
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	return nil
}

func responseErrorMessage(data []byte) string {
	var payload struct {
		Error   any `json:"error"`
		Message any `json:"message"`
		Msg     any `json:"msg"`
	}
	if err := json.Unmarshal(data, &payload); err == nil {
		for _, v := range []any{payload.Error, payload.Message, payload.Msg} {
			if text := strings.TrimSpace(fmt.Sprint(v)); text != "" && text != "<nil>" {
				return text
			}
		}
	}
	return strings.TrimSpace(string(data))
}

func waitReviewResult(ctx context.Context, client *client, runID string) (*reviewDetail, error) {
	deadline := time.NewTimer(pollTimeout)
	defer deadline.Stop()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		detail, err := client.ReviewDetail(ctx, runID)
		if err != nil {
			return nil, err
		}
		if isTerminalReviewStatus(detail.Run.Status, detail.Run.TaskGroupStatus) {
			return detail, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline.C:
			return nil, fmt.Errorf("等待 Review 结果超时")
		case <-ticker.C:
		}
	}
}

func isTerminalReviewStatus(runStatus, groupStatus string) bool {
	return isSuccessfulReviewStatus(runStatus, groupStatus) || isFailedReviewStatus(runStatus, groupStatus)
}

func isSuccessfulReviewStatus(runStatus, groupStatus string) bool {
	if strings.EqualFold(runStatus, "completed") || strings.EqualFold(runStatus, "success") || strings.EqualFold(groupStatus, "success") {
		return true
	}
	return false
}

func isFailedReviewStatus(runStatus, groupStatus string) bool {
	switch strings.ToLower(runStatus) {
	case "failed", "skipped":
		return true
	}
	if strings.EqualFold(groupStatus, "failed") {
		return true
	}
	return false
}

func waitScanResult(ctx context.Context, client *client, taskGroupID string, full bool) (*scanResultResponse, error) {
	deadline := time.NewTimer(pollTimeout)
	defer deadline.Stop()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		result, err := client.ScanResult(ctx, taskGroupID, full)
		if err != nil {
			return nil, err
		}
		if isTerminalScanStatus(result.Task.Status) {
			return result, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline.C:
			return nil, fmt.Errorf("等待扫描结果超时")
		case <-ticker.C:
		}
	}
}

func isTerminalScanStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "success", "failed", "partial_success":
		return true
	default:
		return false
	}
}
