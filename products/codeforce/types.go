package codeforce

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	productName           = "codeforce"
	defaultTimeoutSeconds = 120
	defaultOutputFilePerm = 0o600
	defaultOutputDirPerm  = 0o755
	defaultListPageSize   = 100
	headerMaskedSecret    = "***"
	codeforceUserAgent    = "chaitin-cli-codeforce/1.0"
)

const (
	accountTypeAdmin   = "admin"
	accountTypeUser    = "user"
	accountTypeOpenAPI = "openapi"
)

type Config struct {
	URL            string `yaml:"url"`
	AccessToken    string `yaml:"access_token"`
	APIKey         string `yaml:"api_key"`
	AccountType    string `yaml:"account_type"`
	Insecure       bool   `yaml:"insecure"`
	TimeoutSeconds int    `yaml:"timeout"`
}

type dryRunRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Path    string            `json:"path"`
	Query   map[string]string `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    any               `json:"body,omitempty"`
	Files   map[string]string `json:"files,omitempty"`
}

type refSpec struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type reportParseResult struct {
	Type                              string           `json:"type"`
	MaxSelectedVulnerabilitiesPerTask int              `json:"max_selected_vulnerabilities_per_task"`
	MaxMinioZipFileSizeMB             int              `json:"max_minio_zip_file_size_mb"`
	Vulnerabilities                   []map[string]any `json:"vulnerabilities,omitempty"`
	ScaVulnerabilities                []map[string]any `json:"sca_vulnerabilities,omitempty"`
}

var (
	runtimeCfg Config
	dryRun     bool
)

func (c Config) normalized() Config {
	c.URL = strings.TrimSpace(c.URL)
	c.AccessToken = strings.TrimSpace(firstNonEmpty(c.AccessToken, c.APIKey))
	c.APIKey = strings.TrimSpace(c.APIKey)
	c.AccountType = strings.ToLower(strings.TrimSpace(c.AccountType))
	if c.AccountType == "" {
		c.AccountType = accountTypeAdmin
	}
	if c.URL != "" && !strings.HasPrefix(c.URL, "http://") && !strings.HasPrefix(c.URL, "https://") {
		c.URL = "https://" + c.URL
	}
	c.URL = strings.TrimRight(c.URL, "/")
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = defaultTimeoutSeconds
	}
	return c
}

func (c Config) validate() error {
	if strings.TrimSpace(c.URL) == "" {
		return fmt.Errorf("codeforce url is not configured; set --url, CODEFORCE_URL, or config file codeforce.url")
	}
	if strings.TrimSpace(c.AccessToken) == "" {
		return fmt.Errorf("codeforce access token is not configured; set --access-token/--api-key, CODEFORCE_ACCESS_TOKEN/CODEFORCE_API_KEY, or config file codeforce.access_token")
	}
	switch c.AccountType {
	case accountTypeAdmin, accountTypeUser, accountTypeOpenAPI:
		return nil
	default:
		return fmt.Errorf("unsupported codeforce account type %q; use admin, user, or openapi", c.AccountType)
	}
}

func ApplyRuntimeConfig(cmd *cobra.Command, cfg config.Raw, isDryRun bool) {
	productCfg, err := config.DecodeProduct[Config](cfg, productName)
	if err != nil {
		return
	}
	runtimeCfg = productCfg
	dryRun = isDryRun
}

func outputJSON(w io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal output: %w", err)
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func outputOK(cmd *cobra.Command, fields map[string]any) error {
	payload := map[string]any{"ok": true}
	for k, v := range fields {
		payload[k] = v
	}
	return outputJSON(cmd.OutOrStdout(), payload)
}

func outputDryRun(cmd *cobra.Command, requests []dryRunRequest) error {
	return outputJSON(cmd.OutOrStdout(), map[string]any{
		"ok":       true,
		"dry_run":  true,
		"requests": redactSecrets(requests),
	})
}

func getConfigFromCommand(cmd *cobra.Command) (Config, error) {
	cfg := Config{
		URL:            getProductStringFlag(cmd, "url"),
		AccessToken:    firstNonEmpty(getProductStringFlag(cmd, "access-token"), getProductStringFlag(cmd, "api-key")),
		AccountType:    getProductStringFlag(cmd, "account-type"),
		Insecure:       getProductBoolFlag(cmd, "insecure"),
		TimeoutSeconds: getProductIntFlag(cmd, "timeout"),
	}.normalized()
	if err := cfg.validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func applyRuntimeConfig(cmd *cobra.Command) {
	cfg := runtimeCfg.normalized()
	if !productFlagChanged(cmd, "url") && cfg.URL != "" {
		_ = setProductFlag(cmd, "url", cfg.URL)
	}
	if !productFlagChanged(cmd, "access-token") && cfg.AccessToken != "" {
		_ = setProductFlag(cmd, "access-token", cfg.AccessToken)
	}
	if !productFlagChanged(cmd, "api-key") && cfg.AccessToken != "" {
		_ = setProductFlag(cmd, "api-key", cfg.AccessToken)
	}
	if !productFlagChanged(cmd, "account-type") && cfg.AccountType != "" {
		_ = setProductFlag(cmd, "account-type", cfg.AccountType)
	}
	if !productFlagChanged(cmd, "timeout") && cfg.TimeoutSeconds > 0 {
		_ = setProductFlag(cmd, "timeout", strconv.Itoa(cfg.TimeoutSeconds))
	}
}

func lookupFlag(cmd *cobra.Command, name string) *pflag.Flag {
	if flag := cmd.Flags().Lookup(name); flag != nil {
		return flag
	}
	if flag := cmd.InheritedFlags().Lookup(name); flag != nil {
		return flag
	}
	return cmd.PersistentFlags().Lookup(name)
}

func setFlag(cmd *cobra.Command, name, value string) error {
	if cmd.Flags().Lookup(name) != nil {
		return cmd.Flags().Set(name, value)
	}
	if cmd.InheritedFlags().Lookup(name) != nil {
		return cmd.InheritedFlags().Set(name, value)
	}
	return cmd.PersistentFlags().Set(name, value)
}

func getStringFlag(cmd *cobra.Command, name string) string {
	if value, err := cmd.Flags().GetString(name); err == nil {
		return value
	}
	if value, err := cmd.InheritedFlags().GetString(name); err == nil {
		return value
	}
	return ""
}

func getBoolFlag(cmd *cobra.Command, name string) bool {
	if value, err := cmd.Flags().GetBool(name); err == nil {
		return value
	}
	if value, err := cmd.InheritedFlags().GetBool(name); err == nil {
		return value
	}
	return false
}

func getIntFlag(cmd *cobra.Command, name string) int {
	if value, err := cmd.Flags().GetInt(name); err == nil {
		return value
	}
	if value, err := cmd.InheritedFlags().GetInt(name); err == nil {
		return value
	}
	return 0
}

func productFlagChanged(cmd *cobra.Command, name string) bool {
	if flag := cmd.InheritedFlags().Lookup(name); flag != nil {
		return flag.Changed
	}
	if flag := cmd.PersistentFlags().Lookup(name); flag != nil {
		return flag.Changed
	}
	if flag := cmd.Flags().Lookup(name); flag != nil {
		return flag.Changed
	}
	return false
}

func setProductFlag(cmd *cobra.Command, name, value string) error {
	if flag := cmd.InheritedFlags().Lookup(name); flag != nil {
		return cmd.InheritedFlags().Set(name, value)
	}
	if flag := cmd.PersistentFlags().Lookup(name); flag != nil {
		return cmd.PersistentFlags().Set(name, value)
	}
	return setFlag(cmd, name, value)
}

func getProductStringFlag(cmd *cobra.Command, name string) string {
	if value, err := cmd.InheritedFlags().GetString(name); err == nil {
		return value
	}
	if value, err := cmd.PersistentFlags().GetString(name); err == nil {
		return value
	}
	if value, err := cmd.Flags().GetString(name); err == nil {
		return value
	}
	return ""
}

func getProductBoolFlag(cmd *cobra.Command, name string) bool {
	if value, err := cmd.InheritedFlags().GetBool(name); err == nil {
		return value
	}
	if value, err := cmd.PersistentFlags().GetBool(name); err == nil {
		return value
	}
	if value, err := cmd.Flags().GetBool(name); err == nil {
		return value
	}
	return false
}

func getProductIntFlag(cmd *cobra.Command, name string) int {
	if value, err := cmd.InheritedFlags().GetInt(name); err == nil {
		return value
	}
	if value, err := cmd.PersistentFlags().GetInt(name); err == nil {
		return value
	}
	if value, err := cmd.Flags().GetInt(name); err == nil {
		return value
	}
	return 0
}

func dryRunURL(cfg Config, path string, query map[string]string) string {
	base := strings.TrimRight(cfg.URL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := base + path
	if len(query) == 0 {
		return u
	}
	values := url.Values{}
	for key, value := range query {
		if value != "" {
			values.Set(key, value)
		}
	}
	if encoded := values.Encode(); encoded != "" {
		return u + "?" + encoded
	}
	return u
}

func makeDryRunRequest(cfg Config, method, path string, query map[string]string, body any, files map[string]string) dryRunRequest {
	return dryRunRequest{
		Method: method,
		URL:    dryRunURL(cfg, path, query),
		Path:   path,
		Query:  query,
		Headers: map[string]string{
			"Authorization": "Bearer " + headerMaskedSecret,
			"X-API-Key":     headerMaskedSecret,
			"Token":         headerMaskedSecret,
			"Cookie":        headerMaskedSecret,
		},
		Body:  body,
		Files: files,
	}
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func mustGetString(cmd *cobra.Command, name string) string {
	return strings.TrimSpace(getStringFlag(cmd, name))
}

func mustGetBool(cmd *cobra.Command, name string) bool {
	return getBoolFlag(cmd, name)
}

func mustGetInt(cmd *cobra.Command, name string) int {
	return getIntFlag(cmd, name)
}

func parseStringMapJSON(raw, field string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf("parse %s: %w", field, err)
	}
	return value, nil
}

func parseJSONStringArray(raw, field string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, fmt.Errorf("parse %s: %w", field, err)
	}
	return values, nil
}

func mustJSONEncode(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ensureDirForFile(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, defaultOutputDirPerm)
}

func writeOutputFile(path string, data []byte) error {
	if err := ensureDirForFile(path); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(path, data, defaultOutputFilePerm); err != nil {
		return fmt.Errorf("write output file %s: %w", path, err)
	}
	return nil
}

func parseRefSpec(raw, field string, required bool) (*refSpec, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required {
			return nil, fmt.Errorf("%s is required", field)
		}
		return nil, nil
	}
	refType, name, ok := strings.Cut(raw, ":")
	if !ok {
		return nil, fmt.Errorf("%s must use <branch|tag>:<name>", field)
	}
	refType = strings.ToLower(strings.TrimSpace(refType))
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%s name is required", field)
	}
	switch refType {
	case "branch", "tag":
		return &refSpec{Type: refType, Name: name}, nil
	default:
		return nil, fmt.Errorf("%s type must be branch or tag", field)
	}
}

func parseBoolString(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func rawQuery(query url.Values) map[string]string {
	if len(query) == 0 {
		return nil
	}
	out := make(map[string]string, len(query))
	for key, values := range query {
		if len(values) == 0 {
			continue
		}
		out[key] = values[0]
	}
	return out
}

func redactSecrets(value any) any {
	switch v := value.(type) {
	case []dryRunRequest:
		out := make([]dryRunRequest, len(v))
		for i := range v {
			out[i] = dryRunRequest{
				Method:  v[i].Method,
				URL:     v[i].URL,
				Path:    v[i].Path,
				Query:   v[i].Query,
				Headers: redactStringMap(v[i].Headers),
				Body:    redactSecrets(v[i].Body),
				Files:   v[i].Files,
			}
		}
		return out
	case []map[string]any:
		out := make([]map[string]any, 0, len(v))
		for _, item := range v {
			out = append(out, redactSecrets(item).(map[string]any))
		}
		return out
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, redactSecrets(item))
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			if isSecretKey(key) {
				if s, ok := item.(string); ok && s == "" {
					out[key] = ""
				} else {
					out[key] = headerMaskedSecret
				}
				continue
			}
			out[key] = redactSecrets(item)
		}
		return out
	case map[string]string:
		return redactStringMap(v)
	default:
		return value
	}
}

func redactStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		if isSecretKey(key) {
			if value == "" {
				out[key] = ""
			} else {
				out[key] = headerMaskedSecret
			}
			continue
		}
		out[key] = value
	}
	return out
}

func isSecretKey(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "token") ||
		strings.Contains(key, "password") ||
		strings.Contains(key, "authorization") ||
		strings.Contains(key, "cookie") ||
		strings.Contains(key, "api_key")
}
