package agentcompose

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/spf13/cobra"
)

const productName = "agent-compose"

type productConfig struct {
	URL            string `yaml:"url"`
	APIToken       string `yaml:"api_token"`
	APIKey         string `yaml:"api_key,omitempty"`
	DefaultProject string `yaml:"default_project"`
	Timeout        string `yaml:"timeout"`
	Insecure       bool   `yaml:"insecure"`
}

func (c productConfig) token() string {
	if strings.TrimSpace(c.APIToken) != "" {
		return c.APIToken
	}
	return c.APIKey
}

type runtimeOptions struct {
	URL         string
	Project     string
	Timeout     time.Duration
	Insecure    bool
	JSON        bool
	Token       string
	ConfigPath  string
	DryRun      bool
	TokenSource string

	timeoutText string
}

func ApplyRuntimeConfig(cmd *cobra.Command, raw config.Raw, configPath string, dryRun bool) {
	state := stateFromCommand(cmd)
	if state == nil {
		return
	}
	cfg, err := config.DecodeProduct[productConfig](raw, productName)
	if err != nil {
		state.configErr = err
		return
	}
	if !cmd.Flags().Changed("url") && !cmd.InheritedFlags().Changed("url") {
		state.options.URL = cfg.URL
	}
	if !cmd.Flags().Changed("project") && !cmd.InheritedFlags().Changed("project") {
		state.options.Project = cfg.DefaultProject
	}
	if !cmd.Flags().Changed("timeout") && !cmd.InheritedFlags().Changed("timeout") && cfg.Timeout != "" {
		state.options.timeoutText = cfg.Timeout
	}
	if !cmd.Flags().Changed("insecure") && !cmd.InheritedFlags().Changed("insecure") {
		state.options.Insecure = cfg.Insecure
	}
	state.options.Token = cfg.token()
	state.options.TokenSource = tokenSource(cfg)
	state.options.ConfigPath = configPath
	state.options.DryRun = dryRun
}

type commandState struct {
	options   runtimeOptions
	configErr error
}

func stateFromCommand(cmd *cobra.Command) *commandState {
	for current := cmd; current != nil; current = current.Parent() {
		ctx := current.Context()
		if ctx == nil {
			continue
		}
		if state, ok := ctx.Value(stateKey{}).(*commandState); ok {
			return state
		}
	}
	return nil
}

type stateKey struct{}

func RuntimeConfigError(cmd *cobra.Command, err error) error {
	state := stateFromCommand(cmd)
	jsonOutput := state != nil && state.options.JSON
	return usageError(err.Error(), jsonOutput)
}

func (s *commandState) prepare(cmd *cobra.Command) error {
	if s.configErr != nil {
		return usageError(s.configErr.Error(), s.options.JSON)
	}
	if s.options.DryRun {
		if operation := dryRunWriteOperation(cmd); operation != "" {
			return unsupportedDryRun(operation, s.options.JSON)
		}
	}
	if s.options.timeoutText == "" {
		s.options.timeoutText = "30s"
	}
	timeout, err := time.ParseDuration(s.options.timeoutText)
	if err != nil || timeout <= 0 {
		return usageError("--timeout must be a positive duration", s.options.JSON)
	}
	s.options.Timeout = timeout

	if commandNeedsURL(cmd) {
		normalized, err := normalizeBaseURL(s.options.URL)
		if err != nil {
			return usageError(err.Error(), s.options.JSON)
		}
		s.options.URL = normalized
	}
	if commandNeedsToken(cmd) && strings.TrimSpace(s.options.Token) == "" {
		return newError("authentication_failed", "Token is not configured; run agent-compose auth login", exitAuth, s.options.JSON)
	}
	if s.options.Insecure && commandNeedsURL(cmd) {
		fmt.Fprintln(cmd.ErrOrStderr(), "Warning: TLS certificate verification is disabled")
	}
	return nil
}

func dryRunWriteOperation(cmd *cobra.Command) string {
	parent := ""
	if cmd.Parent() != nil {
		parent = cmd.Parent().Name()
	}
	switch {
	case parent == "auth" && (cmd.Name() == "login" || cmd.Name() == "logout"):
		return "auth " + cmd.Name()
	case parent == productName && cmd.Name() == "run":
		return "run"
	case parent == "run" && (cmd.Name() == "start" || cmd.Name() == "stop"):
		return "run " + cmd.Name()
	case parent == "scheduler" && (cmd.Name() == "invoke" || cmd.Name() == "trigger" || cmd.Name() == "stop"):
		return "scheduler " + cmd.Name()
	case parent == "sandbox" && (cmd.Name() == "stop" || cmd.Name() == "resume" || cmd.Name() == "rm"):
		return "sandbox " + cmd.Name()
	case parent == productName && cmd.Name() == "exec":
		return "exec"
	default:
		return ""
	}
}

func commandNeedsToken(cmd *cobra.Command) bool {
	if cmd.Parent() == nil || cmd.Parent().Name() != "auth" {
		return true
	}
	return cmd.Name() == "status"
}

func commandNeedsURL(cmd *cobra.Command) bool {
	return !(cmd.Name() == "logout" && cmd.Parent() != nil && cmd.Parent().Name() == "auth")
}

func normalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("Agent Compose URL is required; use --url or AGENT_COMPOSE_URL")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("Agent Compose URL must include a valid scheme and host")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("Agent Compose URL must not contain a query or fragment")
	}
	if parsed.User != nil {
		return "", fmt.Errorf("Agent Compose URL must not contain user information")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("Agent Compose URL scheme must be http or https")
	}
	parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	if parsed.Path == "/" {
		parsed.Path = ""
	}
	return strings.TrimSuffix(parsed.String(), "/"), nil
}

func (s *commandState) httpClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if s.options.Insecure {
		transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true} //nolint:gosec -- explicit user flag
	}
	return &http.Client{Transport: &statusTransport{base: &bearerTransport{token: s.options.Token, base: transport}}}
}

type httpStatusError struct{ status int }

func (e *httpStatusError) Error() string { return http.StatusText(e.status) }

type statusTransport struct{ base http.RoundTripper }

func (t *statusTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	response, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == http.StatusBadGateway || response.StatusCode == http.StatusServiceUnavailable {
		_ = response.Body.Close()
		return nil, &httpStatusError{status: response.StatusCode}
	}
	return response, nil
}

type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	if t.token != "" {
		clone.Header.Set("Authorization", "Bearer "+t.token)
	}
	return t.base.RoundTrip(clone)
}

func envTokenActive() bool {
	return envAPIToken() != "" || envAPIKey() != ""
}

func tokenSource(cfg productConfig) string {
	if strings.TrimSpace(cfg.token()) == "" {
		return "none"
	}
	if strings.TrimSpace(cfg.APIToken) != "" {
		if envAPIToken() != "" {
			return "environment"
		}
		return "config"
	}
	if envAPIKey() != "" {
		return "environment"
	}
	return "config"
}

func envAPIToken() string {
	return strings.TrimSpace(os.Getenv("AGENT_COMPOSE_API_TOKEN"))
}

func envAPIKey() string {
	return strings.TrimSpace(os.Getenv("AGENT_COMPOSE_API_KEY"))
}
