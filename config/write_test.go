package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetProduct(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := SetProduct(path, "ddr", struct {
		URL       string `yaml:"url"`
		APIKey    string `yaml:"api_key"`
		CompanyID string `yaml:"company_id"`
	}{
		URL:       "https://example.com",
		APIKey:    "Serval token",
		CompanyID: "corp-1",
	}); err != nil {
		t.Fatalf("SetProduct() error = %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	var got struct {
		URL       string `yaml:"url"`
		APIKey    string `yaml:"api_key"`
		CompanyID string `yaml:"company_id"`
	}
	node := cfg["ddr"]
	if err := node.Decode(&got); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if got.URL != "https://example.com" || got.APIKey != "Serval token" || got.CompanyID != "corp-1" {
		t.Fatalf("unexpected config: %+v", got)
	}
}

func TestSetProductCreatesParentDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".chaitin-cli", "config.yaml")

	if err := SetProduct(path, "ddr", struct {
		URL string `yaml:"url"`
	}{
		URL: "https://example.com",
	}); err != nil {
		t.Fatalf("SetProduct() error = %v", err)
	}

	if _, err := Load(path); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestSetProductPreservesOtherProducts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("cosmos:\n  url: https://cosmos.example\n  api_key: cosmos-token\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := SetProduct(path, "ddr", struct {
		URL    string `yaml:"url"`
		APIKey string `yaml:"api_key"`
	}{
		URL:    "https://ddr.example.com",
		APIKey: "ddr-token",
	}); err != nil {
		t.Fatalf("SetProduct() error = %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	var cosmos struct {
		URL    string `yaml:"url"`
		APIKey string `yaml:"api_key"`
	}
	cosmosNode := cfg["cosmos"]
	if err := cosmosNode.Decode(&cosmos); err != nil {
		t.Fatalf("Decode(cosmos) error = %v", err)
	}
	if cosmos.URL != "https://cosmos.example" || cosmos.APIKey != "cosmos-token" {
		t.Fatalf("cosmos config was not preserved: %+v", cosmos)
	}

	var ddr struct {
		URL    string `yaml:"url"`
		APIKey string `yaml:"api_key"`
	}
	ddrNode := cfg["ddr"]
	if err := ddrNode.Decode(&ddr); err != nil {
		t.Fatalf("Decode(ddr) error = %v", err)
	}
	if ddr.URL != "https://ddr.example.com" || ddr.APIKey != "ddr-token" {
		t.Fatalf("ddr config was not written: %+v", ddr)
	}
}

func TestSetProductWriteErrorMentionsConfigFlag(t *testing.T) {
	dir := t.TempDir()
	parent := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(parent, []byte("file"), 0o600); err != nil {
		t.Fatalf("WriteFile(parent) error = %v", err)
	}

	err := SetProduct(filepath.Join(parent, "config.yaml"), "monkeyscan", struct {
		URL string `yaml:"url"`
	}{
		URL: "https://example.com",
	})
	if err == nil {
		t.Fatal("SetProduct() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "-c/--config") {
		t.Fatalf("SetProduct() error = %q, want -c/--config hint", err)
	}
}

func TestSetProductTightensExistingPermissions(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "config")
	if err := os.Mkdir(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("other:\n  value: preserved\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o666); err != nil {
		t.Fatal(err)
	}
	if err := SetProduct(path, "agent-compose", map[string]string{"api_token": "secret"}); err != nil {
		t.Fatal(err)
	}
	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("dir mode = %o, want 700", got)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("file mode = %o, want 600", got)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg["other"]; !ok {
		t.Fatal("other product was not preserved")
	}
}

func TestSetProductRejectsSymbolicLinkTargets(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.yaml")
	if err := os.WriteFile(target, []byte("safe: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "config.yaml")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	err := SetProduct(link, "agent-compose", map[string]string{"api_token": "secret"})
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("error = %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "safe: true\n" {
		t.Fatalf("target was modified: %q", data)
	}
}

func TestSetProductRejectsSymbolicLinkDirectory(t *testing.T) {
	realDir := t.TempDir()
	parent := t.TempDir()
	linkDir := filepath.Join(parent, "config")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatal(err)
	}
	err := SetProduct(filepath.Join(linkDir, "config.yaml"), "agent-compose", map[string]string{"api_token": "secret"})
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("error = %v", err)
	}
}
