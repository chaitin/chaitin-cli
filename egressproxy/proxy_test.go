package egressproxy

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/chaitin/chaitin-cli/config"
	"gopkg.in/yaml.v3"
)

func TestConfigureEnvOverridesConfig(t *testing.T) {
	isolateProxyEnv(t)
	t.Setenv(EnvName, "http://env-proxy.example:3128")
	t.Setenv("HTTPS_PROXY", "http://https-proxy.example:9")
	raw := config.Raw{ConfigKey: yamlString("http://config-proxy.example:3128")}

	if err := Configure(raw); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() { _ = Configure(nil) })

	proxyURL, err := Wrap(&http.Transport{}).Proxy(requestTo("https://product.example"))
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if proxyURL == nil || proxyURL.Host != "env-proxy.example:3128" {
		t.Fatalf("proxy = %v, want env override", proxyURL)
	}
}

func TestConfigureEmptyEnvFallsThroughToConfig(t *testing.T) {
	isolateProxyEnv(t)
	t.Setenv(EnvName, "  ")
	t.Setenv("HTTPS_PROXY", "http://https-proxy.example:9")
	raw := config.Raw{ConfigKey: yamlString("http://config-proxy.example:3128")}

	if err := Configure(raw); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() { _ = Configure(nil) })

	proxyURL, err := Wrap(&http.Transport{}).Proxy(requestTo("https://product.example"))
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if proxyURL == nil || proxyURL.Host != "config-proxy.example:3128" {
		t.Fatalf("proxy = %v, want config proxy", proxyURL)
	}
}

func TestConfigureMissingProxyUsesEnvironmentProxy(t *testing.T) {
	isolateProxyEnv(t)
	t.Setenv("HTTPS_PROXY", "http://https-proxy.example:9")
	useSystemProxy(t, systemProxySettings{httpsProxy: "http://sys-proxy.example:8888"})

	if err := Configure(nil); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() { _ = Configure(nil) })

	proxyURL, err := Wrap(&http.Transport{}).Proxy(requestTo("https://product.example"))
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if proxyURL == nil || proxyURL.Host != "https-proxy.example:9" {
		t.Fatalf("proxy = %v, want HTTPS_PROXY", proxyURL)
	}
}

func TestNoProxySkipsEnvironmentAndSystemProxy(t *testing.T) {
	isolateProxyEnv(t)
	t.Setenv("HTTPS_PROXY", "http://https-proxy.example:9")
	t.Setenv("NO_PROXY", "product.example")
	useSystemProxy(t, systemProxySettings{httpsProxy: "http://sys-proxy.example:8888"})

	if err := Configure(nil); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() { _ = Configure(nil) })

	proxyURL, err := Wrap(&http.Transport{}).Proxy(requestTo("https://product.example"))
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if proxyURL != nil {
		t.Fatalf("proxy = %v, want direct because NO_PROXY matches", proxyURL)
	}
}

func TestAllProxyUsedWhenSchemeProxyUnset(t *testing.T) {
	isolateProxyEnv(t)
	t.Setenv("ALL_PROXY", "http://all-proxy.example:1080")
	useSystemProxy(t, systemProxySettings{httpsProxy: "http://sys-proxy.example:8888"})

	if err := Configure(nil); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() { _ = Configure(nil) })

	proxyURL, err := Wrap(&http.Transport{}).Proxy(requestTo("https://product.example"))
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if proxyURL == nil || proxyURL.Host != "all-proxy.example:1080" {
		t.Fatalf("proxy = %v, want ALL_PROXY", proxyURL)
	}
}

func TestSystemProxyUsedWhenEnvironmentUnset(t *testing.T) {
	isolateProxyEnv(t)
	useSystemProxy(t, systemProxySettings{
		httpProxy:  "http://sys-http.example:8000",
		httpsProxy: "http://sys-https.example:8443",
		bypass:     []string{"localhost", "*.local"},
	})

	if err := Configure(nil); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() { _ = Configure(nil) })

	transport := Wrap(&http.Transport{})
	httpsProxy, err := transport.Proxy(requestTo("https://product.example"))
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if httpsProxy == nil || httpsProxy.Host != "sys-https.example:8443" {
		t.Fatalf("https proxy = %v, want system HTTPS proxy", httpsProxy)
	}
	httpProxy, err := transport.Proxy(requestTo("http://product.example"))
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if httpProxy == nil || httpProxy.Host != "sys-http.example:8000" {
		t.Fatalf("http proxy = %v, want system HTTP proxy", httpProxy)
	}
	direct, err := transport.Proxy(requestTo("https://localhost/api"))
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if direct != nil {
		t.Fatalf("localhost proxy = %v, want direct", direct)
	}
}

func TestNoProxyAndNoSystemProxyIsDirect(t *testing.T) {
	isolateProxyEnv(t)
	useSystemProxy(t, systemProxySettings{})

	if err := Configure(nil); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() { _ = Configure(nil) })

	proxyURL, err := Wrap(&http.Transport{}).Proxy(requestTo("https://product.example"))
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if proxyURL != nil {
		t.Fatalf("proxy = %v, want direct", proxyURL)
	}
}

func TestConfigureReadsConfigWhenEnvUnset(t *testing.T) {
	isolateProxyEnv(t)
	raw := config.Raw{ConfigKey: yamlString("socks5://jump.example:1080")}

	if err := Configure(raw); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() { _ = Configure(nil) })

	transport := Wrap(&http.Transport{})
	proxyURL, err := transport.Proxy(requestTo("https://product.example"))
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
	isolateProxyEnv(t)
	t.Setenv(EnvName, "ftp://jump.example:21")
	err := Configure(nil)
	if err == nil {
		t.Fatal("Configure() error = nil, want invalid scheme")
	}
}

func isolateProxyEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		EnvName,
		"HTTP_PROXY", "http_proxy",
		"HTTPS_PROXY", "https_proxy",
		"ALL_PROXY", "all_proxy",
		"NO_PROXY", "no_proxy",
	} {
		t.Setenv(name, "")
	}
}

func useSystemProxy(t *testing.T, settings systemProxySettings) {
	t.Helper()
	previous := loadSystemProxy
	loadSystemProxy = func() systemProxySettings { return settings }
	t.Cleanup(func() { loadSystemProxy = previous })
}

func requestTo(rawURL string) *http.Request {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return &http.Request{URL: parsed}
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
