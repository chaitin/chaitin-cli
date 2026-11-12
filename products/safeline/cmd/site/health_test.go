package site

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	safelineruntime "github.com/chaitin/chaitin-cli/products/safeline/runtime"
)

func TestHealthCommandIsRegistered(t *testing.T) {
	root := NewCommand()
	health, _, err := root.Find([]string{"health"})
	if err != nil {
		t.Fatalf("find health command: %v", err)
	}
	if health == nil || health.Use != "health" {
		t.Fatalf("health command not registered: %+v", health)
	}
	enable, _, err := root.Find([]string{"health", "enable"})
	if err != nil {
		t.Fatalf("find health enable command: %v", err)
	}
	for _, name := range []string{"type", "expect", "interval", "timeout", "fall", "rise", "yes", "check"} {
		if enable.Flags().Lookup(name) == nil {
			t.Fatalf("health enable missing --%s flag", name)
		}
	}
	if disable, _, err := root.Find([]string{"health", "disable"}); err != nil || disable.Flags().Lookup("yes") == nil || disable.Flags().Lookup("check") == nil {
		t.Fatalf("health disable missing command or flags: cmd=%+v err=%v", disable, err)
	}
}

func TestHealthEnableExpectFlagOverridesDefault(t *testing.T) {
	cmd := newHealthEnableCmd()
	if err := cmd.ParseFlags([]string{"--expect", "http_4xx"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	got, err := cmd.Flags().GetStringSlice("expect")
	if err != nil {
		t.Fatalf("get expect flag: %v", err)
	}
	if len(got) != 1 || got[0] != "http_4xx" {
		t.Fatalf("--expect should override defaults, got %+v", got)
	}
}

func TestHealthHelpDescribesScenarios(t *testing.T) {
	cmd := newHealthCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"enable", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute help: %v", err)
	}
	help := out.String()
	for _, want := range []string{
		"Layer 7 HTTP health check",
		"Layer 4 TCP health check",
		"TLS handshake health check",
		"--type http",
		"--type tcp",
		"--type ssl_hello",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q:\n%s", want, help)
		}
	}
}

func TestHealthSupportedModes(t *testing.T) {
	supported := []safelineruntime.OperationMode{
		safelineruntime.ModeSoftwareReverseProxy,
		safelineruntime.ModeHardwareReverseProxy,
		safelineruntime.ModeHardwareRouterProxy,
	}
	for _, mode := range supported {
		t.Run(string(mode), func(t *testing.T) {
			if err := validateHealthMode(mode); err != nil {
				t.Fatalf("validateHealthMode(%q): %v", mode, err)
			}
		})
	}

	unsupported := []safelineruntime.OperationMode{
		safelineruntime.ModeSoftwareClusterReverseProxy,
		safelineruntime.ModeSoftwarePortMirroring,
		safelineruntime.ModeHardwareTransparentProxy,
		safelineruntime.ModeHardwareTransparentBridging,
		safelineruntime.ModeHardwarePortMirroring,
		safelineruntime.ModeHardwareTrafficDetection,
	}
	for _, mode := range unsupported {
		t.Run(string(mode), func(t *testing.T) {
			err := validateHealthMode(mode)
			if err == nil || !strings.Contains(err.Error(), "does not support site health-check configuration") {
				t.Fatalf("expected unsupported mode error, got %v", err)
			}
		})
	}
}

func TestBuildHealthCheckConfigHTTP(t *testing.T) {
	cfg, err := buildHealthCheckConfig(healthOptions{
		Type:     "http",
		Host:     "app.example.com",
		Port:     8080,
		Path:     "/healthz",
		Method:   "HEAD",
		Expect:   []string{"http_2xx", "http_3xx"},
		Interval: 5000,
		Timeout:  1000,
		Fall:     2,
		Rise:     3,
	})
	if err != nil {
		t.Fatalf("buildHealthCheckConfig: %v", err)
	}
	if cfg["is_enabled"] != true || cfg["check_type"] != "http" || cfg["host"] != "app.example.com" || cfg["port"] != 8080 || cfg["path"] != "/healthz" || cfg["method"] != "HEAD" {
		t.Fatalf("bad http config: %+v", cfg)
	}
	expect := cfg["check_http_expect_alive"].([]string)
	if len(expect) != 2 || expect[0] != "http_2xx" || expect[1] != "http_3xx" {
		t.Fatalf("bad expect codes: %+v", expect)
	}
}

func TestBuildHealthCheckConfigRejectsHTTPWithoutExpect(t *testing.T) {
	_, err := buildHealthCheckConfig(healthOptions{Type: "http", Expect: nil, Interval: 30000, Timeout: 1000, Fall: 5, Rise: 2})
	if err == nil || !strings.Contains(err.Error(), "at least one --expect") {
		t.Fatalf("expected missing expect error, got %v", err)
	}
}

func TestBuildHealthCheckConfigTCP(t *testing.T) {
	cfg, err := buildHealthCheckConfig(healthOptions{Type: "tcp", Interval: 30000, Timeout: 1000, Fall: 5, Rise: 2})
	if err != nil {
		t.Fatalf("buildHealthCheckConfig: %v", err)
	}
	if cfg["is_enabled"] != true || cfg["check_type"] != "tcp" {
		t.Fatalf("bad tcp config: %+v", cfg)
	}
	if _, ok := cfg["check_http_expect_alive"]; ok {
		t.Fatalf("tcp config must not include HTTP expect codes: %+v", cfg)
	}
}

func TestApplyHealthConfigPreservesSiteAndSlowAttack(t *testing.T) {
	site := map[string]any{
		"id":   float64(12),
		"name": "app",
		"backend_config": map[string]any{
			"type":        "proxy",
			"servers":     []any{map[string]any{"host": "10.0.0.1", "port": float64(80), "protocol": "http"}},
			"slow_attack": map[string]any{"is_enabled": true},
		},
	}
	cfg := map[string]any{"is_enabled": true, "check_type": "tcp", "interval": 30000, "timeout": 1000, "fall": 5, "rise": 2}
	updated, err := applyHealthConfig(site, cfg)
	if err != nil {
		t.Fatalf("applyHealthConfig: %v", err)
	}
	backend := updated["backend_config"].(map[string]any)
	if backend["slow_attack"].(map[string]any)["is_enabled"] != true {
		t.Fatalf("slow_attack was not preserved: %+v", backend)
	}
	if backend["health_check_config"].(map[string]any)["check_type"] != "tcp" {
		t.Fatalf("health config not applied: %+v", backend)
	}
}

func TestApplyHealthConfigRejectsNonProxyBackend(t *testing.T) {
	_, err := applyHealthConfig(map[string]any{"backend_config": map[string]any{"type": "redirect"}}, map[string]any{"is_enabled": false})
	if err == nil || !strings.Contains(err.Error(), "only proxy backend sites") {
		t.Fatalf("expected proxy backend error, got %v", err)
	}
}

func TestHealthSummaryIncludesConfigAndBackendStatuses(t *testing.T) {
	var site map[string]any
	if err := json.Unmarshal([]byte(`{
		"id": 12,
		"name": "app",
		"health_check_status": "UNHEALTHY",
		"backend_config": {
			"type": "proxy",
			"health_check_config": {"is_enabled": true, "check_type": "http"},
			"servers": [
				{"host": "10.0.0.1", "port": 80, "protocol": "http", "health_check_status": "HEALTHY"},
				{"host": "10.0.0.2", "port": 80, "protocol": "http", "health_check_status": "UNHEALTHY"}
			]
		}
	}`), &site); err != nil {
		t.Fatal(err)
	}
	summary, err := buildHealthSummary("/api/SoftwareReverseProxyWebsiteAPI", site)
	if err != nil {
		t.Fatalf("buildHealthSummary: %v", err)
	}
	if summary.SiteStatus != "UNHEALTHY" || len(summary.Backends) != 2 {
		t.Fatalf("bad summary: %+v", summary)
	}
	if summary.Backends[1].Address != "10.0.0.2:80" || summary.Backends[1].Status != "UNHEALTHY" {
		t.Fatalf("bad backend summary: %+v", summary.Backends[1])
	}
}
