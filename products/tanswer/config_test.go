package tanswer

import (
	"testing"
	"time"
)

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("TANSWER_URL", "https://tanswer.test/")
	t.Setenv("TANSWER_API_KEY", "token-123")
	t.Setenv("TANSWER_TIMEOUT", "15s")
	t.Setenv("TANSWER_INSECURE", "true")

	cfg, err := LoadConfig(ConfigOptions{})
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.BaseURL != "https://tanswer.test" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.APIToken != "token-123" {
		t.Fatalf("APIToken = %q", cfg.APIToken)
	}
	if cfg.Timeout != 15*time.Second {
		t.Fatalf("Timeout = %s", cfg.Timeout)
	}
	if !cfg.InsecureSkipVerify {
		t.Fatal("InsecureSkipVerify should be true")
	}
}

func TestLoadConfigRequiresAddress(t *testing.T) {
	t.Setenv("TANSWER_API_KEY", "token-123")

	_, err := LoadConfig(ConfigOptions{})
	if err == nil {
		t.Fatal("expected missing address error")
	}
}

func TestLoadConfigRequiresTokenExceptStatus(t *testing.T) {
	t.Setenv("TANSWER_URL", "https://tanswer.test")

	_, err := LoadConfig(ConfigOptions{})
	if err == nil {
		t.Fatal("expected missing token error")
	}

	cfg, err := LoadConfig(ConfigOptions{AllowMissingToken: true})
	if err != nil {
		t.Fatalf("LoadConfig with AllowMissingToken returned error: %v", err)
	}
	if cfg.APIToken != "" {
		t.Fatalf("APIToken = %q", cfg.APIToken)
	}
}

func TestLoadConfigInsecureSkipVerifyOption(t *testing.T) {
	t.Setenv("TANSWER_URL", "https://tanswer.test")
	t.Setenv("TANSWER_API_KEY", "token-123")

	cfg, err := LoadConfig(ConfigOptions{InsecureSkipVerify: true})
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if !cfg.InsecureSkipVerify {
		t.Fatal("InsecureSkipVerify should be true")
	}
}
