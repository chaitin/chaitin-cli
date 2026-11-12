package egressproxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/chaitin/chaitin-cli/config"
	"golang.org/x/net/proxy"
)

const (
	// EnvName is the environment variable that overrides the config file.
	EnvName = "CHAITIN_CLI_PROXY"
	// ConfigKey is the top-level config.yaml key for the CLI egress proxy.
	ConfigKey = "CHAITIN_CLI_PROXY"
)

type setting struct {
	kind    string
	httpURL *url.URL
	dialer  proxy.ContextDialer
}

var (
	mu      sync.RWMutex
	current setting
	wrapped sync.Map
)

func init() {
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		Wrap(transport)
	}
}

// Configure installs the egress proxy. A set CHAITIN_CLI_PROXY environment
// variable wins over the config file, including when it is empty.
func Configure(raw config.Raw) error {
	value, fromEnv, err := resolve(raw)
	if err != nil {
		return err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		mu.Lock()
		current = setting{}
		mu.Unlock()
		return nil
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return invalidProxy(value, fromEnv, fmt.Errorf("expected an absolute http, https, socks5, or socks5h URL"))
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		mu.Lock()
		current = setting{kind: "http", httpURL: parsed}
		mu.Unlock()
		return nil
	case "socks5", "socks5h":
		socksURL := *parsed
		socksURL.Scheme = "socks5"
		dialer, err := proxy.FromURL(&socksURL, proxy.Direct)
		if err != nil {
			return invalidProxy(value, fromEnv, err)
		}
		contextDialer, ok := dialer.(proxy.ContextDialer)
		if !ok {
			contextDialer = contextDialerFunc(func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			})
		}
		mu.Lock()
		current = setting{kind: "socks", dialer: contextDialer}
		mu.Unlock()
		return nil
	default:
		return invalidProxy(value, fromEnv, fmt.Errorf("unsupported scheme %q", parsed.Scheme))
	}
}

// Wrap makes transport honor CHAITIN_CLI_PROXY. When that proxy is unset, the
// transport keeps its previous proxy and dialer.
func Wrap(transport *http.Transport) *http.Transport {
	if transport == nil {
		return nil
	}
	if _, loaded := wrapped.LoadOrStore(transport, struct{}{}); loaded {
		return transport
	}

	prevProxy := transport.Proxy
	prevDial := transport.DialContext
	transport.Proxy = func(req *http.Request) (*url.URL, error) {
		kind, httpURL := snapshot()
		switch kind {
		case "http":
			return httpURL, nil
		case "socks":
			return nil, nil
		}
		if prevProxy != nil {
			return prevProxy(req)
		}
		return nil, nil
	}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if dialer := socksDialer(); dialer != nil {
			return dialer.DialContext(ctx, network, addr)
		}
		if prevDial != nil {
			return prevDial(ctx, network, addr)
		}
		var dialer net.Dialer
		return dialer.DialContext(ctx, network, addr)
	}
	return transport
}

func resolve(raw config.Raw) (string, bool, error) {
	if value, ok := os.LookupEnv(EnvName); ok {
		return value, true, nil
	}
	if raw == nil {
		return "", false, nil
	}
	node, ok := raw[ConfigKey]
	if !ok || node.Kind == 0 {
		return "", false, nil
	}
	var value string
	if err := node.Decode(&value); err != nil {
		return "", false, fmt.Errorf("invalid config %s: %w", ConfigKey, err)
	}
	return value, false, nil
}

func snapshot() (string, *url.URL) {
	mu.RLock()
	defer mu.RUnlock()
	return current.kind, current.httpURL
}

func socksDialer() proxy.ContextDialer {
	mu.RLock()
	defer mu.RUnlock()
	if current.kind != "socks" {
		return nil
	}
	return current.dialer
}

func invalidProxy(value string, fromEnv bool, err error) error {
	source := "config " + ConfigKey
	if fromEnv {
		source = EnvName
	}
	return fmt.Errorf("invalid %s %q: %w", source, value, err)
}

type contextDialerFunc func(ctx context.Context, network, addr string) (net.Conn, error)

func (f contextDialerFunc) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return f(ctx, network, addr)
}
