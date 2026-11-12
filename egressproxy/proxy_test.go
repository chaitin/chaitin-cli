package egressproxy

import (
	"net/http"
	"os"
	"testing"

	"github.com/chaitin/chaitin-cli/config"
	"gopkg.in/yaml.v3"
)

func TestConfigureEnvOverridesConfig(t *testing.T) {
	t.Setenv(EnvName, "http://env-proxy.example:3128")
	raw := config.Raw{ConfigKey: yamlString("http://config-proxy.example:3128")}

	if err := Configure(raw); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() { _ = Configure(nil) })

	transport := Wrap(&http.Transport{})
	proxyURL, err := transport.Proxy(&http.Request{})
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if proxyURL == nil || proxyURL.Host != "env-proxy.example:3128" {
		t.Fatalf("proxy = %v, want env override", proxyURL)
	}
}

func TestConfigureEmptyEnvDisablesConfigProxy(t *testing.T) {
	t.Setenv(EnvName, "  ")
	raw := config.Raw{ConfigKey: yamlString("http://config-proxy.example:3128")}

	if err := Configure(raw); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() { _ = Configure(nil) })

	transport := Wrap(&http.Transport{})
	proxyURL, err := transport.Proxy(&http.Request{})
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if proxyURL != nil {
		t.Fatalf("proxy = %v, want direct connection", proxyURL)
	}
}

func TestConfigureReadsConfigWhenEnvUnset(t *testing.T) {
	original, existed := os.LookupEnv(EnvName)
	if err := os.Unsetenv(EnvName); err != nil {
		t.Fatalf("unset %s: %v", EnvName, err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(EnvName, original)
		} else {
			_ = os.Unsetenv(EnvName)
		}
	})
	raw := config.Raw{ConfigKey: yamlString("socks5://jump.example:1080")}

	if err := Configure(raw); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() { _ = Configure(nil) })

	transport := Wrap(&http.Transport{})
	proxyURL, err := transport.Proxy(&http.Request{})
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if proxyURL != nil {
		t.Fatalf("socks proxy URL = %v, want nil so DialContext owns the hop", proxyURL)
	}
	if transport.DialContext == nil {
		t.Fatal("DialContext was not installed")
	}
}

func TestConfigureRejectsUnsupportedScheme(t *testing.T) {
	t.Setenv(EnvName, "ftp://jump.example:21")
	err := Configure(nil)
	if err == nil {
		t.Fatal("Configure() error = nil, want invalid scheme")
	}
}

func yamlString(value string) yaml.Node {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(value), &node); err != nil {
		panic(err)
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) == 1 {
		return *node.Content[0]
	}
	return node
}
