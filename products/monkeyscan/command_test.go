package monkeyscan

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chaitin/chaitin-cli/config"
	"gopkg.in/yaml.v3"
)

func TestResolveKeyPrefersEnvironment(t *testing.T) {
	oldCfg := runtimeCfg
	t.Cleanup(func() { runtimeCfg = oldCfg })
	runtimeCfg = Config{APIKey: "config-key"}
	t.Setenv(envAPIKeyName, "env-key")

	key, source := resolveKey()
	if key != "env-key" || source != keySourceEnv {
		t.Fatalf("resolveKey() = %q, %q; want env-key, environment", key, source)
	}
}

func TestApplyRuntimeConfigUsesURLDefaultAndEnvOverride(t *testing.T) {
	oldCfg := runtimeCfg
	t.Cleanup(func() { runtimeCfg = oldCfg })
	t.Setenv(envURLName, "https://env.example.com/")

	raw := rawMonkeyscanConfig(t, Config{URL: "https://config.example.com", APIKey: "config-key"})
	ApplyRuntimeConfig(nil, raw, "config.yaml", false)
	applyRuntimeConfig()

	if runtimeCfg.URL != "https://env.example.com" {
		t.Fatalf("runtimeCfg.URL = %q, want env override", runtimeCfg.URL)
	}
}

func TestValidateReviewScope(t *testing.T) {
	tests := []struct {
		name    string
		scope   reviewScope
		wantErr string
	}{
		{name: "missing", wantErr: "--type 或 --all"},
		{name: "all with type", scope: reviewScope{All: true, Type: "staged"}, wantErr: "--all 不能和 --type"},
		{name: "committed missing base", scope: reviewScope{Type: "committed"}, wantErr: "必须且只能指定"},
		{name: "committed both base", scope: reviewScope{Type: "committed", Base: "main", BaseCommit: "abc"}, wantErr: "必须且只能指定"},
		{name: "staged with base", scope: reviewScope{Type: "staged", Base: "main"}, wantErr: "仅支持 --type committed"},
		{name: "staged", scope: reviewScope{Type: "staged"}},
		{name: "all", scope: reviewScope{All: true}},
		{name: "committed base", scope: reviewScope{Type: "committed", Base: "main"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateReviewScope(tt.scope)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("validateReviewScope() error = %v", err)
			}
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("validateReviewScope() error = %v, want contains %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestValidateScanOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    scanOptions
		wantErr string
	}{
		{name: "default path"},
		{name: "path", opts: scanOptions{Path: "."}},
		{name: "file", opts: scanOptions{File: "repo.zip"}},
		{name: "repo branch", opts: scanOptions{Repo: "https://github.com/acme/repo", Branch: "main"}},
		{name: "multiple sources", opts: scanOptions{Path: ".", File: "repo.zip"}, wantErr: "只能指定一个"},
		{name: "branch without repo", opts: scanOptions{Branch: "main"}, wantErr: "--branch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScanOptions(tt.opts)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("validateScanOptions() error = %v", err)
			}
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("validateScanOptions() error = %v, want contains %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestClientHandlesStatusAndAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("Authorization = %q", got)
		}
		switch r.URL.Path {
		case "/api/v1/cli/status":
			_ = json.NewEncoder(w).Encode(statusResponse{Authenticated: true, Account: accountInfo{Name: "alice"}, ReviewReady: true})
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing"})
		}
	}))
	defer server.Close()

	client := newClient(server.URL, "secret")
	status, err := client.Status(t.Context())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.Authenticated || status.Account.Name != "alice" || !status.ReviewReady {
		t.Fatalf("Status() = %+v", status)
	}
	_, err = client.ReviewDetail(t.Context(), "nope")
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("ReviewDetail() error = %v, want missing", err)
	}
}

func TestClientStatusFallsBackToReadyCompatibilityField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"authenticated": true,
			"ready":         true,
		})
	}))
	defer server.Close()

	status, err := newClient(server.URL, "secret").Status(t.Context())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.Ready || !status.ReviewReady {
		t.Fatalf("Status() = %+v, want ready and review_ready true", status)
	}
}

func TestClientScanEndpoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("Authorization = %q", got)
		}
		switch r.URL.Path {
		case "/api/v1/cli/scans":
			switch r.Method {
			case http.MethodPost:
				if r.URL.Query().Get("type") != "repo" {
					t.Fatalf("type query = %q", r.URL.Query().Get("type"))
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Fatalf("Content-Type = %q", r.Header.Get("Content-Type"))
				}
				_ = json.NewEncoder(w).Encode(scanCreateResponse{TaskGroupID: "group-1", Status: "running"})
			case http.MethodGet:
				_ = json.NewEncoder(w).Encode(scanListResponse{Items: []scanListItem{{TaskGroupID: "group-1"}}})
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		case "/api/v1/cli/scans/result":
			if r.URL.Query().Get("task_group_id") != "group-1" || r.URL.Query().Get("full") != "true" {
				t.Fatalf("query = %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(scanResultResponse{Task: scanListItem{TaskGroupID: "group-1", Status: "success"}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := newClient(server.URL, "secret")
	created, err := client.CreateRepoScan(t.Context(), scanCreateRequest{RepoURL: "https://github.com/acme/repo"})
	if err != nil {
		t.Fatalf("CreateRepoScan() error = %v", err)
	}
	if created.TaskGroupID != "group-1" {
		t.Fatalf("CreateRepoScan() = %+v", created)
	}
	list, err := client.ListScans(t.Context())
	if err != nil {
		t.Fatalf("ListScans() error = %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].TaskGroupID != "group-1" {
		t.Fatalf("ListScans() = %+v", list)
	}
	result, err := client.ScanResult(t.Context(), "group-1", true)
	if err != nil {
		t.Fatalf("ScanResult() error = %v", err)
	}
	if result.Task.Status != "success" {
		t.Fatalf("ScanResult() = %+v", result)
	}
}

func TestClientArchiveScanEndpointSendsTypeAndDisplayNames(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "monkeyscan-temp.zip")
	if err := os.WriteFile(archive, []byte("archive-content"), 0o600); err != nil {
		t.Fatalf("write archive: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cli/scans" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "archive" {
			t.Fatalf("type query = %q", r.URL.Query().Get("type"))
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Fatalf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(1024); err != nil {
			t.Fatalf("ParseMultipartForm() error = %v", err)
		}
		if got := r.FormValue("task_name"); got != "source-dir" {
			t.Fatalf("task_name = %q", got)
		}
		if got := r.FormValue("file_name"); got != "source-dir" {
			t.Fatalf("file_name = %q", got)
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("FormFile() error = %v", err)
		}
		defer file.Close()
		var uploaded bytes.Buffer
		if _, err := uploaded.ReadFrom(file); err != nil {
			t.Fatalf("read uploaded file: %v", err)
		}
		if uploaded.String() != "archive-content" {
			t.Fatalf("uploaded file = %q", uploaded.String())
		}
		_ = json.NewEncoder(w).Encode(scanCreateResponse{TaskGroupID: "group-2", Status: "running"})
	}))
	defer server.Close()

	created, err := newClient(server.URL, "secret").CreateArchiveScanWithName(t.Context(), archive, "source-dir", "source-dir")
	if err != nil {
		t.Fatalf("CreateArchiveScanWithName() error = %v", err)
	}
	if created.TaskGroupID != "group-2" {
		t.Fatalf("CreateArchiveScanWithName() = %+v", created)
	}
}

func TestRunScanFileReturnsArchiveUploadError(t *testing.T) {
	oldCfg, oldDryRun := runtimeCfg, dryRun
	t.Cleanup(func() {
		runtimeCfg = oldCfg
		dryRun = oldDryRun
	})
	dryRun = false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cli/scans" || r.URL.Query().Get("type") != "archive" {
			t.Fatalf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "user is not activated"})
	}))
	defer server.Close()
	runtimeCfg = Config{URL: server.URL, APIKey: "key"}

	sourceDir := t.TempDir()
	writeFile(t, filepath.Join(sourceDir, "README.md"), "hello\n")
	archive := filepath.Join(t.TempDir(), "source.zip")
	if err := zipDirectory(sourceDir, archive); err != nil {
		t.Fatalf("zipDirectory() error = %v", err)
	}

	cmd := NewCommand()
	cmd.SetContext(context.Background())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := runScan(cmd, scanOptions{File: archive})
	if err == nil || !strings.Contains(err.Error(), "user is not activated") {
		t.Fatalf("runScan() error = %v, want user is not activated", err)
	}
	if strings.Contains(out.String(), "MonkeyScan 扫描任务已创建") {
		t.Fatalf("runScan printed created message on error:\n%s", out.String())
	}
}

func TestValidateArchiveFileBodySizeRejectsOver100MB(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "oversized.zip")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	if err := file.Truncate(maxArchiveFileBody + 1); err != nil {
		_ = file.Close()
		t.Fatalf("truncate archive: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}

	err = validateArchiveFileBodySize(archive)
	if err == nil || !strings.Contains(err.Error(), "超过 100M 限制") {
		t.Fatalf("validateArchiveFileBodySize() error = %v, want 100M limit", err)
	}
}

func TestRenderScanResultMarkdownShowsTruncatedNotice(t *testing.T) {
	got := renderScanResultMarkdown(&scanResultResponse{
		Task:      scanListItem{TaskGroupID: "group-1", Status: "success"},
		Total:     5001,
		Items:     []scanDefect{{ID: "defect-1", Severity: "critical"}},
		Full:      true,
		Truncated: true,
	})
	if !strings.Contains(got, "仅展示前 1 条") {
		t.Fatalf("renderScanResultMarkdown() missing truncated notice:\n%s", got)
	}
}

func TestScanCommandsRejectUnexpectedArgs(t *testing.T) {
	tests := [][]string{
		{"scan", "./repo"},
		{"scan", "list", "extra"},
	}
	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd := NewCommand()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(args)
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), "unknown command") && !strings.Contains(err.Error(), "accepts 0 arg") {
				t.Fatalf("Execute(%v) error = %v, output = %s", args, err, out.String())
			}
		})
	}
}

func TestReviewStatusRecognizesSuccessRunStatus(t *testing.T) {
	if !isTerminalReviewStatus("success", "") {
		t.Fatal("success run status should be terminal")
	}
	if !isSuccessfulReviewStatus("success", "") {
		t.Fatal("success run status should be successful")
	}
}

func TestAuthSetKeyAndClearPersistConfig(t *testing.T) {
	oldCfg, oldConfigPath, oldReadSecret := runtimeCfg, runtimeConfigPath, readSecret
	t.Cleanup(func() {
		runtimeCfg = oldCfg
		runtimeConfigPath = oldConfigPath
		readSecret = oldReadSecret
	})
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	runtimeCfg = Config{URL: "https://example.com"}
	runtimeConfigPath = configPath
	readSecret = func() (string, error) { return " saved-key\n", nil }

	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"auth", "set-key"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("set-key Execute() error = %v\n%s", err, out.String())
	}
	raw, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	saved, err := config.DecodeProduct[Config](raw, productName)
	if err != nil {
		t.Fatalf("DecodeProduct() error = %v", err)
	}
	if saved.APIKey != "saved-key" || saved.URL != "https://example.com" {
		t.Fatalf("saved config = %+v", saved)
	}

	cmd = NewCommand()
	out.Reset()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"auth", "clear"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("clear Execute() error = %v\n%s", err, out.String())
	}
	raw, err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() after clear error = %v", err)
	}
	cleared, err := config.DecodeProduct[Config](raw, productName)
	if err != nil {
		t.Fatalf("DecodeProduct() after clear error = %v", err)
	}
	if cleared.APIKey != "" || cleared.URL != "https://example.com" {
		t.Fatalf("cleared config = %+v", cleared)
	}
}

func TestWriteReviewMarkdownIncludesFallbackDiff(t *testing.T) {
	path := filepath.Join(t.TempDir(), "review.md")
	detail := &reviewDetail{
		Run: reviewRun{ID: "run-1", RepositoryName: "acme/repo", Status: "completed"},
		Findings: []reviewFinding{{
			Title:          "问题",
			Severity:       "high",
			Recommendation: "修复",
			Location:       findingLocation{Path: "main.go", StartLine: 1, EndLine: 2},
		}},
	}
	if err := writeReviewMarkdown(path, detail); err != nil {
		t.Fatalf("writeReviewMarkdown() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(data)
	for _, want := range []string{"# MonkeyScan Review 结果", "推荐修复 Diff", "无", "main.go:1-2"} {
		if !strings.Contains(got, want) {
			t.Fatalf("review.md missing %q:\n%s", want, got)
		}
	}
}

func TestWriteReviewMarkdownPrefersSuggestedDiffAndSanitizesError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "review.md")
	detail := &reviewDetail{
		Run: reviewRun{
			ID:             "run-1",
			RepositoryName: "acme/repo",
			Status:         "failed",
			ErrorMessage:   "failed with msk_test_SAMPLE at http://203.0.113.10:8080/internal",
		},
		Findings: []reviewFinding{{
			Title:           "问题",
			SuggestedDiff:   "diff --git a/a b/a\n+secure\n",
			RecommendedDiff: "legacy diff",
		}},
	}
	if err := writeReviewMarkdown(path, detail); err != nil {
		t.Fatalf("writeReviewMarkdown() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(data)
	for _, want := range []string{"diff --git a/a b/a", "[redacted-secret]", "[redacted-url]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("review.md missing %q:\n%s", want, got)
		}
	}
	for _, leaked := range []string{"legacy diff", "msk_test_SAMPLE", "203.0.113.10"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("review.md leaked %q:\n%s", leaked, got)
		}
	}
}

func TestBuildReviewRequestUsesSafeCommandPath(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"chaitin-cli", "monkeyscan", "review", "--api-key", "secret-token-value", "--url", "http://203.0.113.10"}

	cmd := NewCommand()
	reviewCmd, _, err := cmd.Find([]string{"review"})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	req := buildReviewRequest(reviewCmd, &diffSnapshot{Scope: "uncommitted"}, "run-1")
	if req.Command != "monkeyscan review" {
		t.Fatalf("Command = %q, want command path only", req.Command)
	}
	for _, leaked := range []string{"secret-token-value", "203.0.113.10", "--api-key"} {
		if strings.Contains(req.Command, leaked) {
			t.Fatalf("Command leaked %q: %q", leaked, req.Command)
		}
	}
}

func TestRunReviewDryRunDoesNotRequireKeyOrHTTP(t *testing.T) {
	oldCfg, oldDryRun := runtimeCfg, dryRun
	t.Cleanup(func() {
		runtimeCfg = oldCfg
		dryRun = oldDryRun
	})
	dryRun = true
	runtimeCfg = Config{URL: "https://example.com"}

	repo := initGitRepo(t)
	writeFile(t, filepath.Join(repo, "README.md"), "hello\nchanged\n")

	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"review", "--type", "uncommitted"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "dry-run 未上传代码") {
		t.Fatalf("dry-run output unexpected:\n%s", out.String())
	}
	if _, err := os.Stat(filepath.Join(repo, reviewsDir)); !os.IsNotExist(err) {
		t.Fatalf("reviews dir exists after dry-run: %v", err)
	}
}

func TestRunReviewCreatesResultFiles(t *testing.T) {
	oldCfg, oldDryRun, oldPollInterval, oldPollTimeout := runtimeCfg, dryRun, pollInterval, pollTimeout
	t.Cleanup(func() {
		runtimeCfg = oldCfg
		dryRun = oldDryRun
		pollInterval = oldPollInterval
		pollTimeout = oldPollTimeout
	})
	pollInterval = time.Millisecond
	pollTimeout = time.Second

	var created bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer key" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/api/v1/cli/status":
			_ = json.NewEncoder(w).Encode(statusResponse{Authenticated: true, ReviewReady: true})
		case "/api/v1/cli/reviews/diff":
			created = true
			var req diffReviewRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if req.GitRemoteURL == "" || req.Diff == "" || len(req.Files) == 0 {
				t.Fatalf("request missing fields: %+v", req)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(diffReviewResponse{RunID: "run-1", TaskGroupID: "group-1", Status: "created"})
		case "/api/v1/cli/reviews/run-1":
			_ = json.NewEncoder(w).Encode(reviewDetail{
				Run:      reviewRun{ID: "run-1", RepositoryName: "acme/repo", Status: "completed", TaskGroupStatus: "success"},
				Findings: []reviewFinding{{Title: "问题", RecommendedDiff: "diff --git a/a b/a"}},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	runtimeCfg = Config{URL: server.URL, APIKey: "key"}

	repo := initGitRepo(t)
	writeFile(t, filepath.Join(repo, "README.md"), "hello\nchanged\n")

	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"review", "--type", "uncommitted"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v\n%s", err, out.String())
	}
	if !created {
		t.Fatal("CreateDiffReview was not called")
	}
	if !strings.Contains(out.String(), "Review 任务已提交") || !strings.Contains(out.String(), "结果文件") {
		t.Fatalf("output unexpected:\n%s", out.String())
	}
	matches, err := filepath.Glob(filepath.Join(repo, reviewsDir, "*", "review.md"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("review.md matches = %v, err = %v", matches, err)
	}
	statusMatches, _ := filepath.Glob(filepath.Join(repo, reviewsDir, "*", "status.json"))
	if len(statusMatches) != 1 {
		t.Fatalf("status.json matches = %v", statusMatches)
	}
}

func TestRunReviewFailedResultReturnsErrorAndSanitizesStatus(t *testing.T) {
	oldCfg, oldDryRun, oldPollInterval, oldPollTimeout := runtimeCfg, dryRun, pollInterval, pollTimeout
	t.Cleanup(func() {
		runtimeCfg = oldCfg
		dryRun = oldDryRun
		pollInterval = oldPollInterval
		pollTimeout = oldPollTimeout
	})
	pollInterval = time.Millisecond
	pollTimeout = time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cli/status":
			_ = json.NewEncoder(w).Encode(statusResponse{Authenticated: true, ReviewReady: true})
		case "/api/v1/cli/reviews/diff":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(diffReviewResponse{RunID: "run-1", TaskGroupID: "group-1", Status: "created"})
		case "/api/v1/cli/reviews/run-1":
			_ = json.NewEncoder(w).Encode(reviewDetail{
				Run: reviewRun{
					ID:              "run-1",
					RepositoryName:  "acme/repo",
					Status:          "failed",
					TaskGroupStatus: "failed",
					ErrorMessage:    "failed with token=secret at http://203.0.113.10:8080",
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	runtimeCfg = Config{URL: server.URL, APIKey: "key"}

	repo := initGitRepo(t)
	writeFile(t, filepath.Join(repo, "README.md"), "hello\nchanged\n")

	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"review", "--type", "uncommitted"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "Review 任务失败") {
		t.Fatalf("Execute() error = %v, want Review failure", err)
	}
	statusMatches, _ := filepath.Glob(filepath.Join(repo, reviewsDir, "*", "status.json"))
	if len(statusMatches) != 1 {
		t.Fatalf("status.json matches = %v", statusMatches)
	}
	data, err := os.ReadFile(statusMatches[0])
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(data)
	if strings.Contains(got, "token=secret") || strings.Contains(got, "203.0.113.10") {
		t.Fatalf("status.json leaked raw error:\n%s", got)
	}
	if !strings.Contains(got, "[redacted-secret]") || !strings.Contains(got, "[redacted-url]") {
		t.Fatalf("status.json missing redaction:\n%s", got)
	}
}

func TestCollectDiffIncludesUntrackedFiles(t *testing.T) {
	repo := initGitRepo(t)
	writeFile(t, filepath.Join(repo, "new file.txt"), "new\ncontent\n")
	if err := os.MkdirAll(filepath.Join(repo, ".monkeyscan", "reviews", "run-1"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeFile(t, filepath.Join(repo, ".monkeyscan", "reviews", "run-1", "review.md"), "local result\n")

	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	snapshot, err := collectDiff(t.Context(), reviewScope{Type: "uncommitted"})
	if err != nil {
		t.Fatalf("collectDiff() error = %v", err)
	}
	if !strings.Contains(snapshot.Diff, "diff --git a/new file.txt b/new file.txt") {
		t.Fatalf("diff missing untracked file:\n%s", snapshot.Diff)
	}
	if strings.Contains(snapshot.Diff, ".monkeyscan") {
		t.Fatalf("diff includes local review files:\n%s", snapshot.Diff)
	}
	if len(snapshot.Files) != 1 {
		t.Fatalf("files = %+v, want one untracked file", snapshot.Files)
	}
	file := snapshot.Files[0]
	if file.Path != "new file.txt" || file.Status != "added" || file.Additions != 2 || file.Patch == "" {
		t.Fatalf("untracked file metadata = %+v", file)
	}
}

func TestCollectDiffWithoutOriginRemoteContinues(t *testing.T) {
	repo := initGitRepo(t)
	runGit(t, repo, "remote", "remove", "origin")
	writeFile(t, filepath.Join(repo, "README.md"), "hello\nchanged\n")

	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	snapshot, err := collectDiff(t.Context(), reviewScope{Type: "uncommitted"})
	if err != nil {
		t.Fatalf("collectDiff() error = %v", err)
	}
	if snapshot.RemoteURL != "" {
		t.Fatalf("RemoteURL = %q, want empty without remotes", snapshot.RemoteURL)
	}
	if strings.TrimSpace(snapshot.Diff) == "" {
		t.Fatal("diff is empty")
	}
}

func rawMonkeyscanConfig(t *testing.T, value Config) config.Raw {
	t.Helper()
	data, err := yaml.Marshal(map[string]Config{productName: value})
	if err != nil {
		t.Fatalf("Marshal YAML error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	raw, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	return raw
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	runGit(t, dir, "remote", "add", "origin", "https://github.com/acme/repo.git")
	writeFile(t, filepath.Join(dir, "README.md"), "hello\n")
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "init")
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	out, err := gitOutput(t.Context(), dir, args...)
	if err != nil {
		t.Fatalf("git %v error = %v\n%s", args, err, out)
	}
}
