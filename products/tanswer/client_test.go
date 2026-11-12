package tanswer

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientAddsApiTokenHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Api-Token") != "token-123" {
			t.Fatalf("Api-Token header = %q", r.Header.Get("Api-Token"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, APIToken: "token-123", Timeout: time.Second})
	resp, err := client.DoJSON(context.Background(), http.MethodGet, "/api/openapi/test", nil, nil)
	if err != nil {
		t.Fatalf("DoJSON returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d", resp.StatusCode)
	}
}

func TestClientRejectsNonPathTargetsBeforeRequest(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, APIToken: "token-123", Timeout: time.Second})
	for _, target := range []string{
		server.URL + "/steal",
		"//example.com/steal",
		"api/openapi/test",
	} {
		_, err := client.DoJSON(context.Background(), http.MethodGet, target, nil, nil)
		if err == nil {
			t.Fatalf("DoJSON(%q) returned nil error", target)
		}
	}
	if calls != 0 {
		t.Fatalf("server received %d unexpected requests", calls)
	}
}

func TestClientCanSkipTLSVerification(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL:            server.URL,
		APIToken:           "token-123",
		Timeout:            time.Second,
		InsecureSkipVerify: true,
	})
	resp, err := client.DoJSON(context.Background(), http.MethodGet, "/rpc", nil, nil)
	if err != nil {
		t.Fatalf("DoJSON returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d", resp.StatusCode)
	}
	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", client.httpClient.Transport)
	}
	if transport.TLSClientConfig == nil || !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("TLSClientConfig = %#v", transport.TLSClientConfig)
	}
}

func TestClientKeepsTLSVerificationByDefault(t *testing.T) {
	client := NewClient(Config{BaseURL: "https://tanswer.test", Timeout: time.Second})
	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", client.httpClient.Transport)
	}
	if transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("TLSClientConfig should not skip verification: %#v", transport.TLSClientConfig)
	}
	_ = tls.VersionTLS12
}

func TestClientDownloadsUnifiedDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/download" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Api-Token") != "token-123" {
			t.Fatalf("Api-Token header = %q", r.Header.Get("Api-Token"))
		}
		if r.URL.Query().Get("id") != "AssetDownloadServer.DownloadAssetTemplate" {
			t.Fatalf("id query = %q", r.URL.Query().Get("id"))
		}
		raw, err := base64.StdEncoding.DecodeString(r.URL.Query().Get("query"))
		if err != nil {
			t.Fatalf("decode query: %v", err)
		}
		var query map[string]any
		if err := json.Unmarshal(raw, &query); err != nil {
			t.Fatalf("decode query json: %v", err)
		}
		if query["with_data"] != true || query["with_example"] != false {
			t.Fatalf("query = %#v", query)
		}
		ids := query["id_list"].([]any)
		if len(ids) != 2 || ids[0] != float64(3) || ids[1] != float64(7) {
			t.Fatalf("id_list = %#v", ids)
		}
		w.Header().Set("Content-Disposition", `attachment; filename="asset.xlsx";`)
		_, _ = io.WriteString(w, "xlsx-bytes")
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, APIToken: "token-123", Timeout: time.Second})
	file, err := client.Download(context.Background(), "AssetDownloadServer.DownloadAssetTemplate", map[string]any{
		"with_data":    true,
		"with_example": false,
		"id_list":      []int64{3, 7},
	})
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	if file.StatusCode != http.StatusOK || string(file.Body) != "xlsx-bytes" {
		t.Fatalf("file = %#v body=%q", file, string(file.Body))
	}
	if file.FileName != "asset.xlsx" {
		t.Fatalf("FileName = %q", file.FileName)
	}
}

func TestClientUploadsMultipartFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/upload" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("id") != "AssetUploadService.UploadAsset" {
			t.Fatalf("id query = %q", r.URL.Query().Get("id"))
		}
		if r.Header.Get("Api-Token") != "token-123" {
			t.Fatalf("Api-Token header = %q", r.Header.Get("Api-Token"))
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("MultipartReader: %v", err)
		}
		part, err := reader.NextPart()
		if err != nil {
			t.Fatalf("NextPart: %v", err)
		}
		if part.FormName() != "file" || part.FileName() != "assets.xlsx" {
			t.Fatalf("part name=%q filename=%q", part.FormName(), part.FileName())
		}
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatalf("read part: %v", err)
		}
		if string(body) != "xlsx-bytes" {
			t.Fatalf("part body = %q", string(body))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"result": map[string]any{
				"record_count": map[string]any{
					"record_total":  2,
					"success_total": 2,
					"fail_total":    0,
					"cover_total":   0,
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, APIToken: "token-123", Timeout: time.Second})
	resp, err := client.UploadFile(context.Background(), "AssetUploadService.UploadAsset", "assets.xlsx", strings.NewReader("xlsx-bytes"))
	if err != nil {
		t.Fatalf("UploadFile returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d", resp.StatusCode)
	}
	if !strings.Contains(string(resp.Body), "record_count") {
		t.Fatalf("Body = %s", string(resp.Body))
	}
}

func TestClientUploadsMultipartFileWithFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/upload" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("id") != "AlarmUploadService.ImportWhiteList" {
			t.Fatalf("id query = %q", r.URL.Query().Get("id"))
		}
		if r.Header.Get("Api-Token") != "token-123" {
			t.Fatalf("Api-Token header = %q", r.Header.Get("Api-Token"))
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		if got := r.MultipartForm.Value["param"]; len(got) != 1 || got[0] != `{"pattern":"import"}` {
			t.Fatalf("param form field = %#v", got)
		}
		files := r.MultipartForm.File["file"]
		if len(files) != 1 || files[0].Filename != "whitelist.xlsx" {
			t.Fatalf("file headers = %#v", files)
		}
		file, err := files[0].Open()
		if err != nil {
			t.Fatalf("open file: %v", err)
		}
		defer file.Close()
		body, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("read file: %v", err)
		}
		if string(body) != "xlsx-bytes" {
			t.Fatalf("file body = %q", string(body))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"result": map[string]any{
				"record_count": map[string]any{"record_total": 1, "success_total": 1},
			},
		})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, APIToken: "token-123", Timeout: time.Second})
	resp, err := client.UploadFileWithFields(context.Background(), "AlarmUploadService.ImportWhiteList", "whitelist.xlsx", strings.NewReader("xlsx-bytes"), map[string]string{
		"param": `{"pattern":"import"}`,
	})
	if err != nil {
		t.Fatalf("UploadFileWithFields returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d", resp.StatusCode)
	}
	if !strings.Contains(string(resp.Body), "record_count") {
		t.Fatalf("Body = %s", string(resp.Body))
	}
}
