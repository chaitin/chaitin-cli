package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/chaitin/chaitin-cli/egressproxy"
	"github.com/chaitin/chaitin-cli/products/safeline/pkg/auth"
	"github.com/chaitin/chaitin-cli/products/safeline/pkg/client"
	"github.com/spf13/cobra"
)

// Global flags - these are set by the parent safeline command via SetFlags
var (
	URL                   string
	APIKey                string
	Output                string
	Insecure              bool
	DryRun                bool
	ServerVersion         string // detected server version, empty if unknown
	VersionOverride       string
	OperationModeOverride string
	ConfigVersion         string
	ConfigOperationMode   string
)

// SetFlags sets the global flags from the safeline package.
// This must be called by the safeline package before commands run.
func SetFlags(url, apiKey, output string, insecure bool, dryRun bool, versionOverride, operationModeOverride, configVersion, configOperationMode string) {
	URL = url
	APIKey = apiKey
	Output = output
	Insecure = insecure
	DryRun = dryRun
	VersionOverride = versionOverride
	OperationModeOverride = operationModeOverride
	ConfigVersion = configVersion
	ConfigOperationMode = configOperationMode
}

// SetServerVersion sets the detected server version for subcommand use.
func SetServerVersion(v string) {
	ServerVersion = v
}

// GetServerVersion returns the detected server version (empty if unknown).
func GetServerVersion() string {
	return ServerVersion
}

// NewClient creates an authenticated Skyview client.
func NewClient() *client.Client {
	var transport http.RoundTripper

	// Create base transport
	if Insecure {
		transport = egressproxy.Wrap(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		})
	} else {
		transport = http.DefaultTransport
	}

	// Wrap with auth transport
	authTransport := &auth.Transport{
		Base:  transport,
		Token: APIKey,
	}

	httpClient := &http.Client{
		Transport: authTransport,
	}

	return client.New(URL, httpClient)
}

// PrintResult prints the result in the specified format.
func PrintResult(c *cobra.Command, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if Output == "json" {
		var pretty interface{}
		if err := json.Unmarshal(jsonData, &pretty); err != nil {
			fmt.Fprintln(c.OutOrStdout(), string(jsonData))
			return nil
		}
		enc := json.NewEncoder(c.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(pretty)
	}

	// Table output - just print JSON for now, can be enhanced later
	fmt.Fprintln(c.OutOrStdout(), string(jsonData))
	return nil
}

// PrintEnvelope prints the data from an API response envelope.
func PrintEnvelope(c *cobra.Command, env *client.Envelope) error {
	if env.IsWarning() {
		fmt.Fprintf(c.ErrOrStderr(), "WARNING: %s\n", env.GetWarningText())
	}

	if Output == "json" {
		var pretty interface{}
		if err := json.Unmarshal(env.Data, &pretty); err != nil {
			fmt.Fprintln(c.OutOrStdout(), string(env.Data))
			return nil
		}
		enc := json.NewEncoder(c.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(pretty)
	}

	fmt.Fprintln(c.OutOrStdout(), string(env.Data))
	return nil
}

// ReadInput reads JSON input from file or stdin.
func ReadInput(file string) (io.Reader, error) {
	if file == "" || file == "-" {
		return os.Stdin, nil
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", file, err)
	}
	return f, nil
}

// ParseStringSlice parses a comma-separated string into a slice.
func ParseStringSlice(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// ParseIntSlice parses a comma-separated string of integers into a slice.
func ParseIntSlice(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(p, "%d", &n); err != nil {
			return nil, fmt.Errorf("invalid integer: %s", p)
		}
		result = append(result, n)
	}
	return result, nil
}

// GetOutput returns the current output format.
func GetOutput() string {
	return Output
}

// IsDryRun returns whether dry-run mode is enabled.
func IsDryRun() bool {
	return DryRun
}

// DryRunPrint prints a dry-run message and returns true if dry-run is enabled.
// Returns false if dry-run is not enabled, allowing the caller to proceed with the actual operation.
func DryRunPrint(method, path string, body io.Reader) bool {
	if !DryRun {
		return false
	}

	fmt.Fprintf(os.Stderr, "[DRY-RUN] %s %s\n", method, path)
	if body != nil {
		bodyBytes, err := io.ReadAll(body)
		if err == nil && len(bodyBytes) > 0 {
			fmt.Fprintf(os.Stderr, "Body: %s\n", string(bodyBytes))
		}
	}
	return true
}

// IsWriteMethod returns true if the HTTP method is a write operation.
func IsWriteMethod(method string) bool {
	switch method {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

// PrintTable prints data in table format with given headers and rows.
func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print header
	for _, h := range headers {
		fmt.Fprintf(w, "%s\t", h)
	}
	fmt.Fprintln(w)

	// Print rows
	for _, row := range rows {
		for _, cell := range row {
			fmt.Fprintf(w, "%s\t", cell)
		}
		fmt.Fprintln(w)
	}

	w.Flush()
}

// PrintKeyValue prints data in key-value format.
func PrintKeyValue(data map[string]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for k, v := range data {
		fmt.Fprintf(w, "%s:\t%s\n", k, v)
	}
	w.Flush()
}
