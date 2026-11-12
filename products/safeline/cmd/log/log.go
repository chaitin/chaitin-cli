package log

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaitin/chaitin-cli/products/safeline/cmd"
	"github.com/chaitin/chaitin-cli/products/safeline/version"
	"github.com/spf13/cobra"
)

// NewCommand creates the log command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "log",
		Short: "Log query commands",
		Long:  "Commands for querying SafeLine logs (detection, access, rate-limit).",
	}
	c.AddCommand(newDetectCmd())
	c.AddCommand(newAccessCmd())
	c.AddCommand(newRateLimitCmd())
	return c
}

// Detection log commands
func newDetectCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "detect",
		Short: "Detection log commands",
		Long:  "Commands for querying detection (attack) logs.",
	}
	c.AddCommand(newDetectListCmd())
	c.AddCommand(newDetectGetCmd())
	return c
}

func newDetectListCmd() *cobra.Command {
	var count int
	var currentPage int
	var targetPage int
	var tailSort string
	var headSort string
	var srcIP string
	var srcCIDR string
	var host string
	var urlPath string
	var method string
	var statusCode string
	var eventID string
	var filters []string

	c := &cobra.Command{
		Use:   "list",
		Short: "List detection logs",
		Long: `List detection (attack) logs using ES search_after pagination.

Parameters:
  --count        Number of logs to return (default: 20)
  --current-page Current page number, starts from 0 (default: 0)
  --target-page  Target page number, starts from 1 (default: 1)
  --tail-sort    Sort cursor for forward pagination (JSON array from last item's sort field)
  --head-sort    Sort cursor for backward pagination (JSON array from first item's sort field)
  --src-ip       Filter by exact source IP
  --src-cidr     Filter by source CIDR
  --host         Filter by host substring
  --url-path     Filter by URL path substring
  --method       Filter by HTTP method
  --status-code  Filter by exact status code
  --event-id     Filter by exact event ID
  --filter       Raw FilterV2 key=value filter (repeatable)

Pagination:
  - First query: current_page=0, target_page=1, no sort cursor needed
  - Forward (next page): current_page < target_page, use --tail-sort with last item's sort
  - Backward (prev page): current_page > target_page, use --head-sort with first item's sort
  - Max span: 10000 records between current_page and target_page

Examples:
  safeline log detect list
  safeline log detect list --count 50
  safeline log detect list --src-ip 1.2.3.4
  safeline log detect list --src-cidr 1.2.3.0/24 --host example.com
  safeline log detect list --filter attack_type__in=5 --filter risk_level__in=2
  safeline log detect list --current-page 1 --target-page 2 --tail-sort '[1743628800, "abc123"]'`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			query := map[string]string{
				"scope":        "log:detect_log:optim_limit",
				"exclude_body": "true",
				"current_page": fmt.Sprintf("%d", currentPage),
				"target_page":  fmt.Sprintf("%d", targetPage),
				"count":        fmt.Sprintf("%d", count),
			}
			addOptionalQuery(query, "src_ip__exact", srcIP)
			addOptionalQuery(query, "src_ip__net_contained_or_equal", srcCIDR)
			addOptionalQuery(query, "host__contains", host)
			addOptionalQuery(query, "url_path__contains", urlPath)
			addOptionalQuery(query, "method__in", method)
			addOptionalQuery(query, "status_code__exact", statusCode)
			addOptionalQuery(query, "event_id__exact", eventID)
			if err := addRawFilters(query, filters); err != nil {
				return err
			}

			// Filter params based on server version
			svrVer := cmd.GetServerVersion()
			query = version.FilterParams("log/detect/list", query, svrVer)

			// Build fallback query (strip version-sensitive params)
			fallbackQuery := version.StripFallbackParams("log/detect/list", query)

			var multiQuery map[string][]string
			if tailSort != "" {
				values, err := parseSortValues(tailSort)
				if err != nil {
					return fmt.Errorf("invalid tail-sort: %w", err)
				}
				multiQuery = map[string][]string{"tail_sort": values}
			}
			if headSort != "" {
				values, err := parseSortValues(headSort)
				if err != nil {
					return fmt.Errorf("invalid head-sort: %w", err)
				}
				if multiQuery == nil {
					multiQuery = map[string][]string{}
				}
				multiQuery["head_sort"] = values
			}

			env, err := cl.DoMultiWithFallback("GET", "/api/FilterV2API", nil, query, multiQuery, fallbackQuery, nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output - parse paginated response
			var paged struct {
				Items []struct {
					EventID   string        `json:"event_id"`
					Timestamp string        `json:"timestamp"`
					SrcIP     string        `json:"src_ip"`
					Host      string        `json:"host"`
					URL       string        `json:"url_path"`
					Attack    string        `json:"attack_type"`
					Sort      []interface{} `json:"sort"`
				} `json:"items"`
				Total int `json:"total"`
			}
			if err := json.Unmarshal(env.Data, &paged); err != nil {
				return err
			}

			headers := []string{"Event ID", "Timestamp", "Src IP", "Host", "URL", "Attack", "Sort"}
			var rows [][]string
			for _, r := range paged.Items {
				sortStr := "[]"
				if len(r.Sort) > 0 {
					if b, err := json.Marshal(r.Sort); err == nil {
						sortStr = string(b)
					}
				}
				rows = append(rows, []string{
					r.EventID,
					r.Timestamp,
					r.SrcIP,
					truncate(r.Host, 20),
					truncate(r.URL, 30),
					truncate(r.Attack, 15),
					sortStr,
				})
			}
			cmd.PrintTable(headers, rows)
			fmt.Printf("\nTotal: %d\n", paged.Total)
			return nil
		},
	}

	c.Flags().IntVar(&count, "count", 20, "Number of logs to return")
	c.Flags().IntVar(&currentPage, "current-page", 0, "Current page number (starts from 0)")
	c.Flags().IntVar(&targetPage, "target-page", 1, "Target page number (starts from 1)")
	c.Flags().StringVar(&tailSort, "tail-sort", "", "Sort cursor for forward pagination (JSON array)")
	c.Flags().StringVar(&headSort, "head-sort", "", "Sort cursor for backward pagination (JSON array)")
	c.Flags().StringVar(&srcIP, "src-ip", "", "Filter by exact source IP")
	c.Flags().StringVar(&srcCIDR, "src-cidr", "", "Filter by source CIDR")
	c.Flags().StringVar(&host, "host", "", "Filter by host substring")
	c.Flags().StringVar(&urlPath, "url-path", "", "Filter by URL path substring")
	c.Flags().StringVar(&method, "method", "", "Filter by HTTP method")
	c.Flags().StringVar(&statusCode, "status-code", "", "Filter by exact status code")
	c.Flags().StringVar(&eventID, "event-id", "", "Filter by exact event ID")
	c.Flags().StringArrayVar(&filters, "filter", nil, "Raw FilterV2 query filter in key=value form; can be repeated")

	return c
}

func addOptionalQuery(query map[string]string, key, value string) {
	if value = strings.TrimSpace(value); value != "" {
		query[key] = value
	}
}

func addRawFilters(query map[string]string, filters []string) error {
	for _, filter := range filters {
		key, value, ok := strings.Cut(filter, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if !ok || key == "" || value == "" {
			return fmt.Errorf("invalid --filter %q: use key=value", filter)
		}
		query[key] = value
	}
	return nil
}

func newDetectGetCmd() *cobra.Command {
	var eventID string
	var timestamp string

	c := &cobra.Command{
		Use:   "get",
		Short: "Get detection log details",
		Long: `Get detection log details by event ID.

Parameters:
  --event-id   Event ID (required, get from 'list' command)
  --timestamp  Timestamp in Unix format (optional, narrows the query)

Examples:
  safeline log detect get --event-id "6edb4c7eb69042cd996045e3ee5526d9"
  safeline log detect get --event-id "6edb4c7eb69042cd996045e3ee5526d9" --timestamp "1774857841"`,
		RunE: func(c *cobra.Command, args []string) error {
			if eventID == "" {
				return fmt.Errorf("--event-id is required")
			}

			cl := cmd.NewClient()

			query := map[string]string{
				"scope":           "log:detect_log:detail",
				"event_id__exact": eventID,
			}
			if timestamp != "" {
				query["timestamp__exact"] = timestamp
			}

			env, err := cl.Do("GET", "/api/FilterV2API", nil, query)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Parse response - FilterV2API returns array
			var results []struct {
				EventID   string `json:"event_id"`
				Timestamp string `json:"timestamp"`
				SrcIP     string `json:"src_ip"`
				Host      string `json:"host"`
				URL       string `json:"url"`
				Method    string `json:"method"`
				Attack    string `json:"attack_type"`
				RiskLevel string `json:"risk_level"`
				Action    struct {
					Translation string `json:"translation"`
				} `json:"action"`
			}
			if err := json.Unmarshal(env.Data, &results); err != nil {
				return err
			}

			if len(results) == 0 {
				return fmt.Errorf("log not found")
			}

			result := results[0]
			cmd.PrintKeyValue(map[string]string{
				"Event ID":   result.EventID,
				"Timestamp":  result.Timestamp,
				"Src IP":     result.SrcIP,
				"Host":       result.Host,
				"Method":     result.Method,
				"URL":        result.URL,
				"Attack":     result.Attack,
				"Risk Level": result.RiskLevel,
				"Action":     result.Action.Translation,
			})
			return nil
		},
	}

	c.Flags().StringVar(&eventID, "event-id", "", "Event ID (required)")
	c.Flags().StringVar(&timestamp, "timestamp", "", "Timestamp in Unix format (optional)")

	c.MarkFlagRequired("event-id")

	return c
}

// Access log commands
func newAccessCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "access",
		Short: "Access log commands",
		Long:  "Commands for querying access (request) logs.",
	}
	c.AddCommand(newAccessListCmd())
	c.AddCommand(newAccessGetCmd())
	return c
}

func newAccessListCmd() *cobra.Command {
	var count int
	var currentPage int
	var targetPage int
	var tailSort string
	var headSort string

	c := &cobra.Command{
		Use:   "list",
		Short: "List access logs",
		Long: `List access (request) logs using ES search_after pagination.

Parameters:
  --count        Number of logs to return (default: 20)
  --current-page Current page number, starts from 0 (default: 0)
  --target-page  Target page number, starts from 1 (default: 1)
  --tail-sort    Sort cursor for forward pagination (JSON array from last item's sort field)
  --head-sort    Sort cursor for backward pagination (JSON array from first item's sort field)

Pagination:
  - First query: current_page=0, target_page=1, no sort cursor needed
  - Forward (next page): current_page < target_page, use --tail-sort with last item's sort
  - Backward (prev page): current_page > target_page, use --head-sort with first item's sort
  - Max span: 10000 records between current_page and target_page

Examples:
  safeline log access list
  safeline log access list --count 50
  safeline log access list --current-page 1 --target-page 2 --tail-sort '[1743628800, "abc123"]'`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			query := map[string]string{
				"scope":        "log:fullaccess:optim_limit",
				"exclude_body": "true",
				"current_page": fmt.Sprintf("%d", currentPage),
				"target_page":  fmt.Sprintf("%d", targetPage),
				"count":        fmt.Sprintf("%d", count),
			}

			// Filter params based on server version
			svrVer := cmd.GetServerVersion()
			query = version.FilterParams("log/access/list", query, svrVer)

			// Build fallback query (strip version-sensitive params)
			fallbackQuery := version.StripFallbackParams("log/access/list", query)

			var multiQuery map[string][]string
			if tailSort != "" {
				values, err := parseSortValues(tailSort)
				if err != nil {
					return fmt.Errorf("invalid tail-sort: %w", err)
				}
				multiQuery = map[string][]string{"tail_sort": values}
			}
			if headSort != "" {
				values, err := parseSortValues(headSort)
				if err != nil {
					return fmt.Errorf("invalid head-sort: %w", err)
				}
				if multiQuery == nil {
					multiQuery = map[string][]string{}
				}
				multiQuery["head_sort"] = values
			}

			env, err := cl.DoMultiWithFallback("GET", "/api/FilterV2API", nil, query, multiQuery, fallbackQuery, nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output - parse paginated response
			var paged struct {
				Items []struct {
					EventID      string        `json:"event_id"`
					ReqStartTime string        `json:"req_start_time"`
					SrcIP        string        `json:"src_ip"`
					Host         string        `json:"host"`
					URL          string        `json:"url_path"`
					Method       string        `json:"method"`
					StatusCode   int           `json:"status_code"`
					Sort         []interface{} `json:"sort"`
				} `json:"items"`
				Total int `json:"total"`
			}
			if err := json.Unmarshal(env.Data, &paged); err != nil {
				return err
			}

			headers := []string{"Event ID", "Time", "Src IP", "Host", "Method", "Status", "URL", "Sort"}
			var rows [][]string
			for _, r := range paged.Items {
				sortStr := "[]"
				if len(r.Sort) > 0 {
					if b, err := json.Marshal(r.Sort); err == nil {
						sortStr = string(b)
					}
				}
				rows = append(rows, []string{
					r.EventID,
					r.ReqStartTime,
					r.SrcIP,
					truncate(r.Host, 15),
					r.Method,
					fmt.Sprintf("%d", r.StatusCode),
					truncate(r.URL, 25),
					sortStr,
				})
			}
			cmd.PrintTable(headers, rows)
			fmt.Printf("\nTotal: %d\n", paged.Total)
			return nil
		},
	}

	c.Flags().IntVar(&count, "count", 20, "Number of logs to return")
	c.Flags().IntVar(&currentPage, "current-page", 0, "Current page number (starts from 0)")
	c.Flags().IntVar(&targetPage, "target-page", 1, "Target page number (starts from 1)")
	c.Flags().StringVar(&tailSort, "tail-sort", "", "Sort cursor for forward pagination (JSON array)")
	c.Flags().StringVar(&headSort, "head-sort", "", "Sort cursor for backward pagination (JSON array)")

	return c
}

func newAccessGetCmd() *cobra.Command {
	var eventID string
	var reqStartTime string

	c := &cobra.Command{
		Use:   "get",
		Short: "Get access log details",
		Long: `Get access log details by event ID and request start time.

Parameters:
  --event-id       Event ID (required, get from 'list' command)
  --req-start-time Request start time in Unix format (required)

Examples:
  safeline log access get --event-id "1e1ef8e9b21d42cd996045e3ee5526d9" --req-start-time "1775117700"`,
		RunE: func(c *cobra.Command, args []string) error {
			if eventID == "" {
				return fmt.Errorf("--event-id is required")
			}
			if reqStartTime == "" {
				return fmt.Errorf("--req-start-time is required")
			}

			cl := cmd.NewClient()

			query := map[string]string{
				"scope":                 "log:fullaccess:detail",
				"event_id__exact":       eventID,
				"req_start_time__exact": reqStartTime,
			}

			env, err := cl.Do("GET", "/api/FilterV2API", nil, query)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Parse response - FilterV2API returns array
			var results []struct {
				EventID      string `json:"event_id"`
				ReqStartTime string `json:"req_start_time"`
				SrcIP        string `json:"src_ip"`
				Host         string `json:"host"`
				URL          string `json:"url"`
				Method       string `json:"method"`
				StatusCode   int    `json:"status_code"`
				ResponseTime int    `json:"response_time"`
			}
			if err := json.Unmarshal(env.Data, &results); err != nil {
				return err
			}

			if len(results) == 0 {
				return fmt.Errorf("log not found")
			}

			result := results[0]
			cmd.PrintKeyValue(map[string]string{
				"Event ID":       result.EventID,
				"Req Start Time": result.ReqStartTime,
				"Src IP":         result.SrcIP,
				"Host":           result.Host,
				"Method":         result.Method,
				"URL":            result.URL,
				"Status Code":    fmt.Sprintf("%d", result.StatusCode),
				"Response Time":  fmt.Sprintf("%dms", result.ResponseTime),
			})
			return nil
		},
	}

	c.Flags().StringVar(&eventID, "event-id", "", "Event ID (required)")
	c.Flags().StringVar(&reqStartTime, "req-start-time", "", "Request start time in Unix format (required)")

	c.MarkFlagRequired("event-id")
	c.MarkFlagRequired("req-start-time")

	return c
}

// Rate-limit log commands
func newRateLimitCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "rate-limit",
		Short:   "Rate-limit log commands",
		Long:    "Commands for querying rate-limit (ACL) logs.",
		Aliases: []string{"rl"},
	}
	c.AddCommand(newRateLimitListCmd())
	return c
}

func newRateLimitListCmd() *cobra.Command {
	var count int
	var offset int

	c := &cobra.Command{
		Use:   "list",
		Short: "List rate-limit logs",
		Long: `List rate-limit (ACL) logs.

Parameters:
  --count   Number of logs to return (default: 20)
  --offset  Offset for pagination (default: 0)

Examples:
  safeline log rate-limit list
  safeline log rate-limit list --count 50
  safeline log rate-limit list --count 20 --offset 20`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			query := map[string]string{}
			if count > 0 {
				query["count"] = fmt.Sprintf("%d", count)
			}
			if offset > 0 {
				query["offset"] = fmt.Sprintf("%d", offset)
			}

			env, err := cl.Do("GET", "/api/ACLRuleExecutionLogAPI", nil, query)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output - parse paginated response
			var paged struct {
				Items []struct {
					ID        int    `json:"id"`
					Timestamp string `json:"timestamp"`
					Target    string `json:"target"`
					Count     int    `json:"count"`
					Template  struct {
						Name string `json:"name"`
					} `json:"acl_rule_template"`
				} `json:"items"`
				Total int `json:"total"`
			}
			if err := json.Unmarshal(env.Data, &paged); err != nil {
				return err
			}

			headers := []string{"ID", "Timestamp", "Target", "Count", "Rule"}
			var rows [][]string
			for _, r := range paged.Items {
				rows = append(rows, []string{
					fmt.Sprintf("%d", r.ID),
					r.Timestamp,
					truncate(r.Target, 20),
					fmt.Sprintf("%d", r.Count),
					truncate(r.Template.Name, 20),
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}

	c.Flags().IntVar(&count, "count", 20, "Number of logs to return")
	c.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")

	return c
}

// truncate truncates a string to the given length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// parseSortValues parses a JSON array into string values for query parameters.
// e.g., '[1775117699000, "abc123"]' -> ["1775117699000", "abc123"]
func parseSortValues(s string) ([]string, error) {
	var values []interface{}
	if err := json.Unmarshal([]byte(s), &values); err != nil {
		return nil, err
	}
	result := make([]string, len(values))
	for i, v := range values {
		switch val := v.(type) {
		case float64:
			result[i] = fmt.Sprintf("%.0f", val)
		case string:
			result[i] = val
		default:
			result[i] = fmt.Sprintf("%v", val)
		}
	}
	return result, nil
}
