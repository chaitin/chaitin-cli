package egressproxy

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/http/httpproxy"
)

// systemProxySettings is the OS HTTP/HTTPS proxy. SOCKS-only system proxies are
// ignored: product targets are http and https.
type systemProxySettings struct {
	httpProxy  string
	httpsProxy string
	bypass     []string
}

var (
	systemProxyOnce   sync.Once
	cachedSystemProxy systemProxySettings
	// loadSystemProxy is replaced by tests. The default reads the OS once.
	loadSystemProxy = func() systemProxySettings {
		systemProxyOnce.Do(func() {
			cachedSystemProxy = readPlatformSystemProxy()
		})
		return cachedSystemProxy
	}
)

func environmentProxy(req *http.Request) (*url.URL, bool, error) {
	if !envProxyConfigured() {
		return nil, false, nil
	}
	if req == nil || req.URL == nil {
		return nil, true, nil
	}
	// Read the variables on each request. net/http.ProxyFromEnvironment caches
	// the first environment it sees, which would ignore later changes and NO_PROXY.
	httpProxy := firstEnv("HTTP_PROXY", "http_proxy")
	httpsProxy := firstEnv("HTTPS_PROXY", "https_proxy")
	allProxy := firstEnv("ALL_PROXY", "all_proxy")
	if httpProxy == "" {
		httpProxy = allProxy
	}
	if httpsProxy == "" {
		httpsProxy = allProxy
	}
	cfg := &httpproxy.Config{
		HTTPProxy:  httpProxy,
		HTTPSProxy: httpsProxy,
		NoProxy:    firstEnv("NO_PROXY", "no_proxy"),
	}
	proxyURL, err := cfg.ProxyFunc()(req.URL)
	return proxyURL, true, err
}

func schemeProxyConfigured() bool {
	return firstEnv("HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy") != ""
}

func envProxyConfigured() bool {
	return schemeProxyConfigured() || firstEnv("ALL_PROXY", "all_proxy") != ""
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func systemProxyURL(req *http.Request) *url.URL {
	if req == nil || req.URL == nil {
		return nil
	}
	if excludedByNoProxy(req.URL) {
		return nil
	}
	return loadSystemProxy().proxyFor(req.URL)
}

func excludedByNoProxy(target *url.URL) bool {
	noProxy := firstEnv("NO_PROXY", "no_proxy")
	if noProxy == "" || target == nil {
		return false
	}
	cfg := &httpproxy.Config{NoProxy: noProxy, HTTPProxy: "http://proxy.invalid", HTTPSProxy: "http://proxy.invalid"}
	proxyURL, err := cfg.ProxyFunc()(target)
	return err == nil && proxyURL == nil
}

func (s systemProxySettings) proxyFor(target *url.URL) *url.URL {
	if target == nil {
		return nil
	}
	host := target.Hostname()
	if host == "" || hostBypassed(host, s.bypass) {
		return nil
	}
	raw := ""
	switch strings.ToLower(target.Scheme) {
	case "http":
		raw = s.httpProxy
	case "https":
		raw = s.httpsProxy
	default:
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil
	}
	return parsed
}

func proxyURLFromHostPort(host, port string) string {
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	if host == "" {
		return ""
	}
	if strings.Contains(host, "://") {
		return host
	}
	if port == "" || port == "0" {
		port = "80"
	}
	return "http://" + net.JoinHostPort(host, port)
}

func hostBypassed(host string, patterns []string) bool {
	host = strings.ToLower(strings.Trim(host, "[]"))
	if host == "" {
		return false
	}
	ip := net.ParseIP(host)
	for _, raw := range patterns {
		pattern := strings.TrimSpace(raw)
		if pattern == "" {
			continue
		}
		if pattern == "*" {
			return true
		}
		if strings.EqualFold(pattern, "<local>") {
			if !strings.Contains(host, ".") {
				return true
			}
			continue
		}
		if strings.Contains(pattern, "/") {
			if ip != nil && cidrContains(pattern, ip) {
				return true
			}
			continue
		}
		if strings.Contains(pattern, "*") {
			if globMatch(strings.ToLower(pattern), host) {
				return true
			}
			continue
		}
		pattern = strings.ToLower(pattern)
		if host == pattern {
			return true
		}
	}
	return false
}

func cidrContains(pattern string, ip net.IP) bool {
	_, network, err := net.ParseCIDR(normalizeCIDR(pattern))
	if err != nil {
		return false
	}
	return network.Contains(ip)
}

func normalizeCIDR(pattern string) string {
	host, bits, ok := strings.Cut(pattern, "/")
	if !ok || strings.Contains(host, ":") {
		return pattern
	}
	parts := strings.Split(host, ".")
	if len(parts) >= 4 {
		return pattern
	}
	for len(parts) < 4 {
		parts = append(parts, "0")
	}
	return strings.Join(parts, ".") + "/" + bits
}

func parseScutilProxy(text string) systemProxySettings {
	var settings systemProxySettings
	httpEnabled := false
	httpsEnabled := false
	httpHost, httpsHost, httpPort, httpsPort := "", "", "", ""
	inExceptions := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if inExceptions {
			if line == "}" || strings.HasPrefix(line, "}") {
				inExceptions = false
				continue
			}
			_, value, ok := strings.Cut(line, ":")
			if ok {
				settings.bypass = append(settings.bypass, strings.TrimSpace(value))
			}
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "ExceptionsList" {
			inExceptions = true
			continue
		}
		switch key {
		case "HTTPEnable":
			httpEnabled = value == "1"
		case "HTTPProxy":
			httpHost = value
		case "HTTPPort":
			httpPort = value
		case "HTTPSEnable":
			httpsEnabled = value == "1"
		case "HTTPSProxy":
			httpsHost = value
		case "HTTPSPort":
			httpsPort = value
		}
	}
	if httpEnabled {
		settings.httpProxy = proxyURLFromHostPort(httpHost, httpPort)
	}
	if httpsEnabled {
		settings.httpsProxy = proxyURLFromHostPort(httpsHost, httpsPort)
	}
	return settings
}

func parseWindowsProxy(text string) systemProxySettings {
	values := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		name := fields[0]
		kind := fields[1]
		if kind != "REG_DWORD" && kind != "REG_SZ" && kind != "REG_EXPAND_SZ" {
			continue
		}
		values[name] = strings.TrimSpace(strings.Join(fields[2:], " "))
	}
	if values["ProxyEnable"] != "0x1" && values["ProxyEnable"] != "0x00000001" && values["ProxyEnable"] != "1" {
		return systemProxySettings{}
	}
	settings := systemProxySettings{bypass: splitList(values["ProxyOverride"])}
	server := values["ProxyServer"]
	if server == "" {
		return systemProxySettings{}
	}
	if strings.Contains(server, "=") {
		for _, part := range splitList(server) {
			scheme, addr, ok := strings.Cut(part, "=")
			if !ok {
				continue
			}
			parsed := ensureProxyURL(addr)
			switch strings.ToLower(strings.TrimSpace(scheme)) {
			case "http":
				settings.httpProxy = parsed
			case "https":
				settings.httpsProxy = parsed
			}
		}
		return settings
	}
	parsed := ensureProxyURL(server)
	settings.httpProxy = parsed
	settings.httpsProxy = parsed
	return settings
}

func parseGsettingsProxy(mode, httpHost, httpPort, httpsHost, httpsPort, ignoreHosts string) systemProxySettings {
	if strings.Trim(strings.TrimSpace(mode), "'") != "manual" {
		return systemProxySettings{}
	}
	return systemProxySettings{
		httpProxy:  proxyURLFromHostPort(unquoteGsettings(httpHost), unquoteGsettings(httpPort)),
		httpsProxy: proxyURLFromHostPort(unquoteGsettings(httpsHost), unquoteGsettings(httpsPort)),
		bypass:     parseGsettingsList(ignoreHosts),
	}
}

func ensureProxyURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, "://") {
		return raw
	}
	host, port, err := net.SplitHostPort(raw)
	if err != nil {
		return proxyURLFromHostPort(raw, "80")
	}
	return proxyURLFromHostPort(host, port)
}

func splitList(raw string) []string {
	raw = strings.ReplaceAll(raw, ",", ";")
	var items []string
	for _, item := range strings.Split(raw, ";") {
		item = strings.TrimSpace(item)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func unquoteGsettings(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "uint32 ")
	return strings.Trim(raw, "'\"")
}

func parseGsettingsList(raw string) []string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	var items []string
	for _, item := range strings.Split(raw, ",") {
		item = unquoteGsettings(item)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func globMatch(pattern, host string) bool {
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == host
	}
	if !strings.HasPrefix(host, parts[0]) {
		return false
	}
	host = host[len(parts[0]):]
	for _, part := range parts[1 : len(parts)-1] {
		index := strings.Index(host, part)
		if index < 0 {
			return false
		}
		host = host[index+len(part):]
	}
	return strings.HasSuffix(host, parts[len(parts)-1])
}
