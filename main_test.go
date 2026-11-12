package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentComposeBusyBoxAliasShowsProductHelp(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	originalArgs := os.Args
	os.Args = []string{"/tmp/agent-compose", "--help"}
	t.Cleanup(func() { os.Args = originalArgs })
	var out bytes.Buffer
	app.root.SetOut(&out)
	app.root.SetErr(&out)
	if err := app.execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Remotely operate Agent Compose") || strings.Contains(out.String(), "cloudAtlas") {
		t.Fatalf("alias help = %s", out.String())
	}
}

func TestAgentComposeConfigErrorUsesJSONAndUsageExitCode(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "invalid.yaml")
	if err := os.WriteFile(configPath, []byte("agent-compose:\n  url: [broken\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	app, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	app.root.SetArgs([]string{"--config", configPath, "agent-compose", "--json", "status"})
	err = app.execute()
	var coder interface{ ExitCode() int }
	if !errors.As(err, &coder) || coder.ExitCode() != 2 {
		t.Fatalf("error = %v", err)
	}
	var renderer interface{ RenderError(io.Writer) error }
	if !errors.As(err, &renderer) {
		t.Fatalf("error does not render itself: %T", err)
	}
	var output bytes.Buffer
	if err := renderer.RenderError(&output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"kind":"usage"`) || !strings.Contains(output.String(), `"message":`) {
		t.Fatalf("rendered error = %s", output.String())
	}
}

func TestNewAppUsesDefaultConfigPath(t *testing.T) {
	homeDir := t.TempDir()
	setTestHome(t, homeDir)

	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}

	want := filepath.Join(homeDir, defaultConfigDir, defaultConfigFile)
	if app.configPath != want {
		t.Fatalf("app.configPath = %q, want %q", app.configPath, want)
	}
}

func TestDryRunHelpMentionsRequestSummary(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}
	flag := app.root.PersistentFlags().Lookup("dry-run")
	if flag == nil {
		t.Fatalf("missing --dry-run flag")
	}
	if !strings.Contains(flag.Usage, "request summary") {
		t.Fatalf("dry-run usage = %q, want request summary mention", flag.Usage)
	}
}

func TestCosmosDryRunFromRootDoesNotRequireNetwork(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("cosmos:\n  url: https://cosmos.example\n  api_key: secret-token-value\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}

	var out strings.Builder
	var errOut strings.Builder
	app.root.SetOut(&out)
	app.root.SetErr(&errOut)
	app.root.SetArgs([]string{
		"--dry-run",
		"-c", configPath,
		"cosmos", "asset", "save-host-asset",
		"--ip", "10.0.0.1/32",
		"--name", "dry-run-asset",
		"--organization_id", "1",
	})

	if err := app.execute(); err != nil {
		t.Fatalf("execute() error = %v\nstderr:\n%s", err, errOut.String())
	}

	got := out.String()
	for _, want := range []string{
		`"dry_run": true`,
		`"url": "https://cosmos.example/pedestal/rpc"`,
		`"method": "AssetService.SaveHostAsset"`,
		`"ip": "10.0.0.1/32"`,
		`"name": "dry-run-asset"`,
		`"organization_id": 1`,
		`"Authorization": "Bearer secr...alue"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "secret-token-value") {
		t.Fatalf("dry-run output leaked token:\n%s", got)
	}
}

func TestRootHelpIncludesCodeForce(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}

	var out strings.Builder
	app.root.SetOut(&out)
	app.root.SetErr(&out)
	app.root.SetArgs([]string{"--help"})
	if err := app.root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "codeforce") {
		t.Fatalf("root help missing codeforce:\n%s", out.String())
	}
}

func TestEnsureRuntimeConfigLoaded(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom.yaml")
	if err := os.WriteFile(configPath, []byte("tanswer:\n  endpoint: https://example.com\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("Chdir() cleanup error = %v", err)
		}
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	app := &app{configPath: configPath}
	if err := app.ensureRuntimeConfigLoaded(); err != nil {
		t.Fatalf("ensureRuntimeConfigLoaded() error = %v", err)
	}

	if !app.configLoaded {
		t.Fatal("configLoaded = false, want true")
	}
	if _, ok := app.config["tanswer"]; !ok {
		t.Fatal("config missing tanswer section")
	}
}

func TestEnsureRuntimeConfigLoadedWithMissingConfigFile(t *testing.T) {
	app := &app{configPath: filepath.Join(t.TempDir(), "missing.yaml")}
	if err := app.ensureRuntimeConfigLoaded(); err != nil {
		t.Fatalf("ensureRuntimeConfigLoaded() error = %v", err)
	}

	if !app.configLoaded {
		t.Fatal("configLoaded = false, want true")
	}
	if len(app.config) != 0 {
		t.Fatalf("len(config) = %d, want 0", len(app.config))
	}
}

func TestRuntimeConfigLoadsRecognizedLocalBeforeGlobal(t *testing.T) {
	homeDir := t.TempDir()
	workDir := t.TempDir()
	setTestHome(t, homeDir)
	withWorkingDir(t, workDir)

	globalPath := filepath.Join(homeDir, defaultConfigDir, defaultConfigFile)
	if err := os.MkdirAll(filepath.Dir(globalPath), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(globalPath, []byte("monkeyscan:\n  url: https://global.example\n  api_key: global-key\ncosmos:\n  url: https://cosmos.example\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(global) error = %v", err)
	}
	if err := os.WriteFile(localConfigPath(), []byte("monkeyscan:\n  url: https://local.example\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(local) error = %v", err)
	}

	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}
	if err := app.ensureRuntimeConfigLoaded(); err != nil {
		t.Fatalf("ensureRuntimeConfigLoaded() error = %v", err)
	}

	var monkeyscanCfg struct {
		URL    string `yaml:"url"`
		APIKey string `yaml:"api_key"`
	}
	monkeyscanNode := app.config["monkeyscan"]
	if err := monkeyscanNode.Decode(&monkeyscanCfg); err != nil {
		t.Fatalf("Decode(monkeyscan) error = %v", err)
	}
	if monkeyscanCfg.URL != "https://local.example" || monkeyscanCfg.APIKey != "" {
		t.Fatalf("monkeyscan config = %+v, want local config only", monkeyscanCfg)
	}
	if _, ok := app.config["cosmos"]; ok {
		t.Fatal("global cosmos section should not be loaded when local config is recognized")
	}
}

func TestRuntimeConfigIgnoresUnrecognizedLocalConfig(t *testing.T) {
	homeDir := t.TempDir()
	workDir := t.TempDir()
	setTestHome(t, homeDir)
	withWorkingDir(t, workDir)

	globalPath := filepath.Join(homeDir, defaultConfigDir, defaultConfigFile)
	if err := os.MkdirAll(filepath.Dir(globalPath), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(globalPath, []byte("monkeyscan:\n  url: https://global.example\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(global) error = %v", err)
	}
	if err := os.WriteFile(localConfigPath(), []byte("database:\n  host: localhost\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(local) error = %v", err)
	}

	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}
	if err := app.ensureRuntimeConfigLoaded(); err != nil {
		t.Fatalf("ensureRuntimeConfigLoaded() error = %v", err)
	}

	if _, ok := app.config["database"]; ok {
		t.Fatal("unrecognized local config section should be ignored")
	}
	var monkeyscanCfg struct {
		URL string `yaml:"url"`
	}
	monkeyscanNode := app.config["monkeyscan"]
	if err := monkeyscanNode.Decode(&monkeyscanCfg); err != nil {
		t.Fatalf("Decode(monkeyscan) error = %v", err)
	}
	if monkeyscanCfg.URL != "https://global.example" {
		t.Fatalf("monkeyscan url = %q, want global", monkeyscanCfg.URL)
	}
}

func TestRuntimeConfigFallsBackToGlobalWhenLocalConfigParseFails(t *testing.T) {
	homeDir := t.TempDir()
	workDir := t.TempDir()
	setTestHome(t, homeDir)
	withWorkingDir(t, workDir)

	globalPath := filepath.Join(homeDir, defaultConfigDir, defaultConfigFile)
	if err := os.MkdirAll(filepath.Dir(globalPath), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(globalPath, []byte("monkeyscan:\n  url: https://global.example\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(global) error = %v", err)
	}
	if err := os.WriteFile(localConfigPath(), []byte("monkeyscan:\n  url: [broken\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(local) error = %v", err)
	}

	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}
	if err := app.ensureRuntimeConfigLoaded(); err != nil {
		t.Fatalf("ensureRuntimeConfigLoaded() error = %v", err)
	}

	var monkeyscanCfg struct {
		URL string `yaml:"url"`
	}
	monkeyscanNode := app.config["monkeyscan"]
	if err := monkeyscanNode.Decode(&monkeyscanCfg); err != nil {
		t.Fatalf("Decode(monkeyscan) error = %v", err)
	}
	if monkeyscanCfg.URL != "https://global.example" {
		t.Fatalf("monkeyscan url = %q, want global", monkeyscanCfg.URL)
	}
}

func TestTAnswerGlobalConfigFallbackAtCLIEntry(t *testing.T) {
	homeDir := t.TempDir()
	workDir := t.TempDir()
	setTestHome(t, homeDir)
	withWorkingDir(t, workDir)
	t.Setenv("TANSWER_URL", "")
	t.Setenv("TANSWER_API_KEY", "")
	t.Setenv("TANSWER_TIMEOUT", "")
	t.Setenv("TANSWER_INSECURE", "")

	globalPath := filepath.Join(homeDir, defaultConfigDir, defaultConfigFile)
	if err := os.MkdirAll(filepath.Dir(globalPath), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(globalPath, []byte("tanswer:\n  url: https://global-tanswer.example\n  api_key: placeholder\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(global) error = %v", err)
	}

	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}
	var out strings.Builder
	app.root.SetOut(&out)
	app.root.SetErr(&out)
	app.root.SetArgs([]string{"tanswer", "auth", "status"})
	if err := app.execute(); err != nil {
		t.Fatalf("execute() error = %v\\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), `"environment": "https://global-tanswer.example"`) || !strings.Contains(out.String(), `"token_set": true`) {
		t.Fatalf("auth status did not use global T-Answer configuration:\n%s", out.String())
	}
}

func TestWriteConfigPathUsesRecognizedLocalConfig(t *testing.T) {
	homeDir := t.TempDir()
	workDir := t.TempDir()
	setTestHome(t, homeDir)
	withWorkingDir(t, workDir)

	if err := os.WriteFile(localConfigPath(), []byte("monkeyscan:\n  url: https://local.example\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(local) error = %v", err)
	}
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}
	if got := app.writeConfigPath(); got != localConfigPath() {
		t.Fatalf("writeConfigPath() = %q, want %q", got, localConfigPath())
	}
}

func TestWriteConfigPathUsesGlobalWhenLocalMissingOrUnrecognized(t *testing.T) {
	homeDir := t.TempDir()
	workDir := t.TempDir()
	setTestHome(t, homeDir)
	withWorkingDir(t, workDir)

	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}
	if got := app.writeConfigPath(); got != defaultConfigPath() {
		t.Fatalf("writeConfigPath() without local = %q, want %q", got, defaultConfigPath())
	}

	if err := os.WriteFile(localConfigPath(), []byte("database:\n  host: localhost\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(local) error = %v", err)
	}
	if got := app.writeConfigPath(); got != defaultConfigPath() {
		t.Fatalf("writeConfigPath() with unrecognized local = %q, want %q", got, defaultConfigPath())
	}
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("Chdir() cleanup error = %v", err)
		}
	})
}

func setTestHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
}

func TestTAnswerExplicitFalseInsecureFlagOverridesEnvironment(t *testing.T) {
	homeDir := t.TempDir()
	workDir := t.TempDir()
	setTestHome(t, homeDir)
	withWorkingDir(t, workDir)
	t.Setenv("TANSWER_URL", "http://127.0.0.1:9")
	t.Setenv("TANSWER_API_KEY", "placeholder")
	t.Setenv("TANSWER_INSECURE", "true")

	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}
	var out strings.Builder
	app.root.SetOut(&out)
	app.root.SetErr(&out)
	app.root.SetArgs([]string{"--config", filepath.Join(t.TempDir(), "empty.yaml"), "tanswer", "--insecure=false", "auth", "status"})

	if err := app.execute(); err != nil {
		t.Fatalf("execute() error = %v", err)
	}

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			InsecureSkipVerify bool `json:"insecure_skip_verify"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(out.String()), &payload); err != nil {
		t.Fatalf("decode auth status: %v", err)
	}
	if !payload.Success {
		t.Fatal("auth status success = false")
	}
	if payload.Data.InsecureSkipVerify {
		t.Fatal("--insecure=false must override TANSWER_INSECURE=true")
	}
}
