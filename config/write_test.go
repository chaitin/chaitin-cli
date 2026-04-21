package config

import (
	"path/filepath"
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
