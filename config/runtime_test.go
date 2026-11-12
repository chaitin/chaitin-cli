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

func TestDecodeProductSupportsCamelCaseEnvironmentPrefix(t *testing.T) {
	type runtimeConfig struct {
		URL     string `yaml:"url"`
		Token   string `yaml:"token"`
		SpaceID string `yaml:"space_id"`
	}

	t.Run("documented prefix", func(t *testing.T) {
		t.Setenv("CLOUD_ATLAS_URL", "https://documented.example.com/openapi")
		t.Setenv("CLOUD_ATLAS_TOKEN", "documented-token")
		t.Setenv("CLOUD_ATLAS_SPACE_ID", "6")

		cfg, err := DecodeProduct[runtimeConfig](Raw{}, "cloudAtlas")
		if err != nil {
			t.Fatalf("DecodeProduct() error = %v", err)
		}
		if cfg.URL != "https://documented.example.com/openapi" || cfg.Token != "documented-token" || cfg.SpaceID != "6" {
			t.Fatalf("DecodeProduct() = %#v, want documented CLOUD_ATLAS_* values", cfg)
		}
	})

	t.Run("legacy prefix remains compatible", func(t *testing.T) {
		t.Setenv("CLOUD_ATLAS_URL", "")
		t.Setenv("CLOUD_ATLAS_TOKEN", "")
		t.Setenv("CLOUD_ATLAS_SPACE_ID", "")
		t.Setenv("CLOUDATLAS_URL", "https://legacy.example.com/openapi")
		t.Setenv("CLOUDATLAS_TOKEN", "legacy-token")
		t.Setenv("CLOUDATLAS_SPACE_ID", "7")

		cfg, err := DecodeProduct[runtimeConfig](Raw{}, "cloudAtlas")
		if err != nil {
			t.Fatalf("DecodeProduct() error = %v", err)
		}
		if cfg.URL != "https://legacy.example.com/openapi" || cfg.Token != "legacy-token" || cfg.SpaceID != "7" {
			t.Fatalf("DecodeProduct() = %#v, want legacy CLOUDATLAS_* values", cfg)
		}
	})

	t.Run("documented prefix wins", func(t *testing.T) {
		t.Setenv("CLOUDATLAS_URL", "https://legacy.example.com/openapi")
		t.Setenv("CLOUD_ATLAS_URL", "https://documented.example.com/openapi")

		cfg, err := DecodeProduct[runtimeConfig](Raw{}, "cloudAtlas")
		if err != nil {
			t.Fatalf("DecodeProduct() error = %v", err)
		}
		if cfg.URL != "https://documented.example.com/openapi" {
			t.Fatalf("URL = %q, want documented CLOUD_ATLAS_URL to take precedence", cfg.URL)
		}
	})
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
