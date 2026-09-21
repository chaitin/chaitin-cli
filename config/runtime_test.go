package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeProductAppliesEnvOverrides(t *testing.T) {
	t.Setenv("SAFELINE_URL", "https://env.example.com")
	t.Setenv("SAFELINE_API_KEY", "env-key")

	cfg, err := Load(filepath.Join("..", "config.yaml.example"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	type runtimeConfig struct {
		URL    string `yaml:"url"`
		APIKey string `yaml:"api_key"`
	}

	productCfg, err := DecodeProduct[runtimeConfig](cfg, "safeline")
	if err != nil {
		t.Fatalf("DecodeProduct() error = %v", err)
	}

	if productCfg.URL != "https://env.example.com" {
		t.Fatalf("URL = %q, want env override", productCfg.URL)
	}
	if productCfg.APIKey != "env-key" {
		t.Fatalf("APIKey = %q, want env override", productCfg.APIKey)
	}
}

func TestDecodeProductUsesEnvWithoutConfigSection(t *testing.T) {
	t.Setenv("SAFELINE_URL", "https://env-only.example.com")
	t.Setenv("SAFELINE_API_KEY", "env-only-key")

	type runtimeConfig struct {
		URL    string `yaml:"url"`
		APIKey string `yaml:"api_key"`
	}

	productCfg, err := DecodeProduct[runtimeConfig](Raw{}, "safeline")
	if err != nil {
		t.Fatalf("DecodeProduct() error = %v", err)
	}

	if productCfg.URL != "https://env-only.example.com" {
		t.Fatalf("URL = %q, want env-only value", productCfg.URL)
	}
	if productCfg.APIKey != "env-only-key" {
		t.Fatalf("APIKey = %q, want env-only value", productCfg.APIKey)
	}
}

func TestDecodeProductCamelCaseProductEnvOverrides(t *testing.T) {
	t.Setenv("CLOUD_ATLAS_URL", "https://camel.example.com")
	t.Setenv("CLOUD_ATLAS_TOKEN", "camel-token")
	t.Setenv("CLOUD_ATLAS_SPACE_ID", "7")

	type cloudAtlasConfig struct {
		URL     string `yaml:"url"`
		Token   string `yaml:"token"`
		SpaceID string `yaml:"space_id"`
	}

	cfg, err := DecodeProduct[cloudAtlasConfig](Raw{}, "cloudAtlas")
	if err != nil {
		t.Fatalf("DecodeProduct() error = %v", err)
	}
	if cfg.URL != "https://camel.example.com" {
		t.Fatalf("URL = %q, want env override", cfg.URL)
	}
	if cfg.Token != "camel-token" {
		t.Fatalf("Token = %q, want env override", cfg.Token)
	}
	if cfg.SpaceID != "7" {
		t.Fatalf("SpaceID = %q, want env override", cfg.SpaceID)
	}
}

func TestLoadEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("SAFELINE_URL=https://dotenv.example.com\nSAFELINE_API_KEY=dotenv-key\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_ = os.Unsetenv("SAFELINE_URL")
	_ = os.Unsetenv("SAFELINE_API_KEY")

	if err := LoadEnvFile(envPath); err != nil {
		t.Fatalf("LoadEnvFile() error = %v", err)
	}

	if got := os.Getenv("SAFELINE_URL"); got != "https://dotenv.example.com" {
		t.Fatalf("SAFELINE_URL = %q, want dotenv value", got)
	}
	if got := os.Getenv("SAFELINE_API_KEY"); got != "dotenv-key" {
		t.Fatalf("SAFELINE_API_KEY = %q, want dotenv value", got)
	}
}
