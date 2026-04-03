package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAppUsesDefaultConfigPath(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}

	if app.configPath != defaultConfigPathFromCWD() {
		t.Fatalf("app.configPath = %q, want %q", app.configPath, defaultConfigPathFromCWD())
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
