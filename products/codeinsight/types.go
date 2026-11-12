package codeinsight

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	productName            = "codeinsight"
	defaultTimeoutSeconds  = 120
	defaultReportWait      = 300
	defaultReportPoll      = 5
	defaultReportFormat    = 4
	defaultReportSpecs     = 1
	defaultReportFilePerm  = 0o600
	defaultReportDirPerm   = 0o755
	headerMaskedSecret     = "***"
	repoConfigModeSingle   = 1
	repoConfigModeMulti    = 2
	codeInsightUserAgent   = "chaitin-cli-codeinsight/1.0"
	defaultRepoConfigLimit = 200
)

type Config struct {
	URL            string `yaml:"url"`
	AccessToken    string `yaml:"access_token"`
	Token          string `yaml:"token"`
	Insecure       bool   `yaml:"insecure"`
	TimeoutSeconds int    `yaml:"timeout"`
}

type projectRecord struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Raw  any    `json:"raw,omitempty"`
}

type repoConfigRecord struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	RepoType   int    `json:"repo_type"`
	ServerHost string `json:"server_host,omitempty"`
	Raw        any    `json:"raw,omitempty"`
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

var (
	runtimeCfg Config
	dryRun     bool
)

var languageIDs = map[string]int64{
	"java":       3,
	"c":          4,
	"c++":        4,
	"cpp":        4,
	"c/c++":      4,
	"python":     5,
	"php":        6,
	"javascript": 7,
	"js":         7,
	"typescript": 7,
	"ts":         7,
	"c#":         8,
	"csharp":     8,
	"go":         9,
	"golang":     9,
	"swift":      10,
	"ruby":       11,
	"rust":       12,
	"kotlin":     17,
}

func (c Config) normalized() Config {
	c.URL = strings.TrimSpace(c.URL)
	c.AccessToken = strings.TrimSpace(firstNonEmpty(c.AccessToken, c.Token))
	c.Token = strings.TrimSpace(c.Token)
	if c.URL != "" && !strings.HasPrefix(c.URL, "http://") && !strings.HasPrefix(c.URL, "https://") {
		c.URL = "http://" + c.URL
	}
	c.URL = strings.TrimRight(c.URL, "/")
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = defaultTimeoutSeconds
	}
	return c
}

func (c Config) validate() error {
	if strings.TrimSpace(c.URL) == "" {
		return fmt.Errorf("codeinsight url is not configured; set --url, CODEINSIGHT_URL, or config file codeinsight.url")
	}
	if strings.TrimSpace(c.AccessToken) == "" {
		return fmt.Errorf("codeinsight access token is not configured; set --access-token, CODEINSIGHT_TOKEN/CODEINSIGHT_ACCESS_TOKEN, or config file codeinsight.access_token")
	}
	return nil
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

func resolveLanguageID(language string, languageID int64) (int64, error) {
	if languageID > 0 {
		return languageID, nil
	}
	language = strings.ToLower(strings.TrimSpace(language))
	if language == "" {
		return 0, fmt.Errorf("请提供 --language 或 --language-id")
	}
	id, ok := languageIDs[language]
	if !ok {
		return 0, fmt.Errorf("不支持的语言 %q，请使用 --language-id 显式指定", language)
	}
	return id, nil
}

func repoTypeID(repoType string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(repoType)) {
	case "", "git":
		return 1, nil
	case "svn":
		return 2, nil
	case "ftp":
		return 3, nil
	case "tfs":
		return 4, nil
	default:
		return 0, fmt.Errorf("不支持的仓库类型 %q", repoType)
	}
}

func repoTypeName(repoType int) string {
	switch repoType {
	case 1:
		return "git"
	case 2:
		return "svn"
	case 3:
		return "ftp"
	case 4:
		return "tfs"
	default:
		return "unknown"
	}
}

func authTypeID(authType string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(authType)) {
	case "", "access_token":
		return 2, nil
	case "username_password":
		return 1, nil
	case "anonymous":
		return 3, nil
	default:
		return 0, fmt.Errorf("不支持的认证类型 %q", authType)
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

func firstNonNilMapValue(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := m[key]; ok && value != nil {
			return value
		}
	}
	return nil
}

func stringValue(m map[string]any, keys ...string) string {
	value := firstNonNilMapValue(m, keys...)
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case nil:
		return ""
	default:
		return fmt.Sprint(v)
	}
}

func int64Value(m map[string]any, keys ...string) int64 {
	return toInt64(firstNonNilMapValue(m, keys...))
}

func toInt64(value any) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		i, _ := v.Int64()
		return i
	case string:
		i, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return i
	default:
		return 0
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

func flagChanged(cmd *cobra.Command, name string) bool {
	flag := lookupFlag(cmd, name)
	return flag != nil && flag.Changed
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

func getConfigFromCommand(cmd *cobra.Command) (Config, error) {
	cfg := Config{
		URL:            getProductStringFlag(cmd, "url"),
		AccessToken:    firstNonEmpty(getProductStringFlag(cmd, "token"), getProductStringFlag(cmd, "access-token")),
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
	if !productFlagChanged(cmd, "token") && cfg.AccessToken != "" {
		_ = setProductFlag(cmd, "token", cfg.AccessToken)
		if !productFlagChanged(cmd, "access-token") {
			_ = setProductFlag(cmd, "access-token", cfg.AccessToken)
		}
	}
	if !productFlagChanged(cmd, "timeout") && cfg.TimeoutSeconds > 0 {
		_ = setProductFlag(cmd, "timeout", strconv.Itoa(cfg.TimeoutSeconds))
	}
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
			"Token":         headerMaskedSecret,
		},
		Body:  body,
		Files: files,
	}
}

func repoURLFromName(serverHost, repo string) string {
	serverHost = strings.TrimRight(strings.TrimSpace(serverHost), "/")
	repo = strings.Trim(strings.TrimSpace(repo), "/")
	if repo == "" {
		return serverHost
	}
	if strings.HasPrefix(repo, "http://") || strings.HasPrefix(repo, "https://") {
		return repo
	}
	if strings.HasSuffix(repo, ".git") {
		return serverHost + "/" + repo
	}
	return serverHost + "/" + repo + ".git"
}

func basename(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Base(path)
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
		strings.Contains(key, "cookie")
}
