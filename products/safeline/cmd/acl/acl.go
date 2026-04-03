package acl

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaitin/workspace-cli/products/safeline/cmd"
	"github.com/spf13/cobra"
)

// NewCommand creates the acl command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "acl",
		Short: "ACL management commands",
		Long:  "Commands for managing SafeLine ACL (rate limiting) rules.",
	}
	c.AddCommand(newTemplateCmd())
	c.AddCommand(newRuleCmd())
	return c
}

func newTemplateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "template",
		Short: "ACL template management",
		Long:  "Commands for managing ACL templates (rate limiting rules).",
	}
	c.AddCommand(newTemplateListCmd())
	c.AddCommand(newTemplateGetCmd())
	c.AddCommand(newTemplateCreateCmd())
	c.AddCommand(newTemplateEnableCmd())
	c.AddCommand(newTemplateDisableCmd())
	c.AddCommand(newTemplateDeleteCmd())
	return c
}

func newTemplateListCmd() *cobra.Command {
	var name string

	c := &cobra.Command{
		Use:   "list",
		Short: "List all ACL templates",
		Long: `List all ACL templates (rate limiting rules).

Examples:
  safeline acl template list
  safeline acl template list --name "limit"`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			query := map[string]string{
				"scope": "detect:rule_template:template",
			}
			if name != "" {
				query["name__like"] = name
			}

			env, err := cl.Do("GET", "/api/FilterV2API", nil, query)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output - FilterV2API returns array directly
			var results []struct {
				ID           int    `json:"id"`
				Name         string `json:"name"`
				TemplateType string `json:"template_type"`
				IsEnabled    bool   `json:"is_enabled"`
				DryRun       bool   `json:"dry_run"`
			}
			if err := json.Unmarshal(env.Data, &results); err != nil {
				return err
			}

			headers := []string{"ID", "Name", "Type", "Enabled", "Watch"}
			var rows [][]string
			for _, r := range results {
				enabled := "No"
				if r.IsEnabled {
					enabled = "Yes"
				}
				watch := "No"
				if r.DryRun {
					watch = "Yes"
				}
				rows = append(rows, []string{
					fmt.Sprintf("%d", r.ID),
					r.Name,
					r.TemplateType,
					enabled,
					watch,
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}

	c.Flags().StringVar(&name, "name", "", "Filter by name (fuzzy match)")

	return c
}

func newTemplateGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get ACL template details",
		Long: `Get ACL template details by ID.

Examples:
  safeline acl template get 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			query := map[string]string{
				"scope":     "detect:rule_template:template",
				"id__exact": args[0],
			}

			env, err := cl.Do("GET", "/api/FilterV2API", nil, query)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Parse response - could be array or single object
			var results []struct {
				ID           int                    `json:"id"`
				Name         string                 `json:"name"`
				TemplateType string                 `json:"template_type"`
				IsEnabled    bool                   `json:"is_enabled"`
				DryRun       bool                   `json:"dry_run"`
				MatchMethod  map[string]interface{} `json:"match_method"`
				Action       map[string]interface{} `json:"action"`
			}
			if err := json.Unmarshal(env.Data, &results); err != nil {
				return err
			}

			if len(results) == 0 {
				return fmt.Errorf("template not found: %s", args[0])
			}

			result := results[0]
			enabled := "No"
			if result.IsEnabled {
				enabled = "Yes"
			}
			dryRun := "No"
			if result.DryRun {
				dryRun = "Yes"
			}

			actionStr := fmt.Sprintf("%v", result.Action["action"])
			targetType := fmt.Sprintf("%v", result.MatchMethod["target_type"])

			cmd.PrintKeyValue(map[string]string{
				"ID":          fmt.Sprintf("%d", result.ID),
				"Name":        result.Name,
				"Type":        result.TemplateType,
				"Target Type": targetType,
				"Action":      actionStr,
				"Enabled":     enabled,
				"Watch":       dryRun,
			})
			return nil
		},
	}
}

func newTemplateCreateCmd() *cobra.Command {
	var (
		name            string
		templateType    string
		targetType      string
		scope           string
		period          int
		limit           int
		action          string
		limitRateLimit  int
		limitRatePeriod int
		targets         string
		ipGroups        string
		expirePeriod    int
		watch           bool
		enabled         bool
	)

	c := &cobra.Command{
		Use:   "create",
		Short: "Create a new ACL template",
		Long: `Create a new ACL template (rate limiting rule).

Template types:
  - manual: Manual rule (requires --targets or --ip-groups)
  - auto: Automatic rule (requires --period and --limit)

Target types:
  - cidr: IP/CIDR
  - session: Session
  - fingerprint: Fingerprint

Scope types:
  - all: All requests
  - url: Specific URL
  - prefix: URL prefix

Actions:
  - forbid: Block access
  - limit_rate: Rate limit (requires --limit-rate-limit and --limit-rate-period)

Examples:
  # Manual rule blocking specific IPs
  safeline acl template create --name "Block IPs" --template-type manual \
    --target-type cidr --action forbid --targets "192.168.1.100,10.0.0.50"

  # Manual rule using IP group
  safeline acl template create --name "Block Group" --template-type manual \
    --target-type cidr --action forbid --ip-groups 1,2

  # Auto rule limiting requests
  safeline acl template create --name "Rate Limit" --template-type auto \
    --period 60 --limit 100 --action forbid`,
		RunE: func(c *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			// Map CLI values to API enum values
			targetTypeMap := map[string]string{
				"cidr":        "CIDR",
				"session":     "Session",
				"fingerprint": "Fingerprint",
			}
			scopeMap := map[string]string{
				"all":    "All",
				"url":    "URL",
				"prefix": "URL Prefix",
			}

			apiTargetType := targetTypeMap[targetType]
			if apiTargetType == "" {
				apiTargetType = "CIDR"
			}
			apiScope := scopeMap[scope]
			if apiScope == "" {
				apiScope = "All"
			}

			// Build match_method
			matchMethod := map[string]interface{}{
				"scope":       apiScope,
				"target_type": apiTargetType,
			}

			if templateType == "auto" {
				if period <= 0 || limit <= 0 {
					return fmt.Errorf("--period and --limit are required for auto template type")
				}
				matchMethod["period"] = period
				matchMethod["limit"] = limit
			}

			// Build action
			actionMap := map[string]interface{}{
				"action": action,
			}
			if action == "limit_rate" {
				if limitRateLimit <= 0 || limitRatePeriod <= 0 {
					return fmt.Errorf("--limit-rate-limit and --limit-rate-period are required for limit_rate action")
				}
				actionMap["limit_rate_limit"] = limitRateLimit
				actionMap["limit_rate_period"] = limitRatePeriod
			}

			// Build request
			req := map[string]interface{}{
				"name":          name,
				"template_type": templateType,
				"match_method":  matchMethod,
				"action":        actionMap,
				"is_enabled":    enabled,
				"dry_run":       watch,
				"forbidden_page_config": map[string]interface{}{
					"action":      "response",
					"status_code": 403,
					"path":        "",
				},
			}

			if targetType == "cidr" {
				if targets != "" {
					req["targets"] = strings.Split(targets, ",")
				}
				if ipGroups != "" {
					req["target_ip_groups"] = strings.Split(ipGroups, ",")
				}
			}

			if expirePeriod > 0 {
				req["expire_period"] = expirePeriod
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run (CLI dry-run, different from template's watch mode)
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] POST /api/ACLRuleTemplateAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("POST", "/api/ACLRuleTemplateAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Key-value output
			var result struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}
			cmd.PrintKeyValue(map[string]string{
				"ID":   fmt.Sprintf("%d", result.ID),
				"Name": result.Name,
			})
			return nil
		},
	}

	c.Flags().StringVar(&name, "name", "", "Template name (required)")
	c.Flags().StringVar(&templateType, "template-type", "manual", "Template type (manual|auto)")
	c.Flags().StringVar(&targetType, "target-type", "cidr", "Target type (cidr|session|fingerprint)")
	c.Flags().StringVar(&scope, "scope", "all", "Scope (all|url|prefix)")
	c.Flags().IntVar(&period, "period", 0, "Time period in seconds (required for auto type)")
	c.Flags().IntVar(&limit, "limit", 0, "Request limit count (required for auto type)")
	c.Flags().StringVar(&action, "action", "forbid", "Action (forbid|limit_rate)")
	c.Flags().IntVar(&limitRateLimit, "limit-rate-limit", 0, "Rate limit count (required for limit_rate action)")
	c.Flags().IntVar(&limitRatePeriod, "limit-rate-period", 0, "Rate limit period in seconds (required for limit_rate action)")
	c.Flags().StringVar(&targets, "targets", "", "Comma-separated list of IPs/CIDRs")
	c.Flags().StringVar(&ipGroups, "ip-groups", "", "Comma-separated list of IP group IDs")
	c.Flags().IntVar(&expirePeriod, "expire-period", 0, "Expiration period in seconds")
	c.Flags().BoolVar(&watch, "watch", false, "Enable watch (observation) mode for the template")
	c.Flags().BoolVar(&enabled, "enabled", true, "Enable the template")

	c.MarkFlagRequired("name")

	return c
}

func newTemplateEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <id>",
		Short: "Enable an ACL template",
		Long: `Enable an ACL template by ID.

Examples:
  safeline acl template enable 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			req := map[string]interface{}{
				"id":     json.Number(args[0]),
				"action": "enable",
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] PUT /api/EnableDisableACLRuleTemplateAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("PUT", "/api/EnableDisableACLRuleTemplateAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("ACL template %s enabled successfully\n", args[0])
			return nil
		},
	}
}

func newTemplateDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <id>",
		Short: "Disable an ACL template",
		Long: `Disable an ACL template by ID.

Examples:
  safeline acl template disable 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			req := map[string]interface{}{
				"id":     json.Number(args[0]),
				"action": "disable",
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] PUT /api/EnableDisableACLRuleTemplateAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("PUT", "/api/EnableDisableACLRuleTemplateAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("ACL template %s disabled successfully\n", args[0])
			return nil
		},
	}
}

func newTemplateDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an ACL template",
		Long: `Delete an ACL template by ID.

Examples:
  safeline acl template delete 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			req := map[string]interface{}{
				"id__in": []string{args[0]},
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] DELETE /api/ACLRuleTemplateAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("DELETE", "/api/ACLRuleTemplateAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("ACL template %s deleted successfully\n", args[0])
			return nil
		},
	}
}

// Rule commands
func newRuleCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "rule",
		Short: "ACL rule management",
		Long:  "Commands for managing ACL rules (blocked users).",
	}
	c.AddCommand(newRuleListCmd())
	c.AddCommand(newRuleDeleteCmd())
	c.AddCommand(newRuleClearCmd())
	return c
}

func newRuleListCmd() *cobra.Command {
	var templateID int

	c := &cobra.Command{
		Use:   "list",
		Short: "List ACL rules",
		Long: `List ACL rules (blocked users) for a template.

Examples:
  safeline acl rule list --template-id 1`,
		RunE: func(c *cobra.Command, args []string) error {
			if templateID <= 0 {
				return fmt.Errorf("--template-id is required")
			}

			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/ACLRuleAPI", nil, map[string]string{"acl_rule_template_id": fmt.Sprintf("%d", templateID)})
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output
			var results []struct {
				ID         int    `json:"id"`
				Target     string `json:"target"`
				CreateTime string `json:"create_time"`
				ExpireTime string `json:"expire_time"`
			}
			if err := json.Unmarshal(env.Data, &results); err != nil {
				return err
			}

			headers := []string{"ID", "Target", "Create Time", "Expire Time"}
			var rows [][]string
			for _, r := range results {
				expireTime := r.ExpireTime
				if expireTime == "" || expireTime == "null" {
					expireTime = "Never"
				}
				rows = append(rows, []string{
					fmt.Sprintf("%d", r.ID),
					r.Target,
					r.CreateTime,
					expireTime,
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}

	c.Flags().IntVar(&templateID, "template-id", 0, "Template ID (required)")
	c.MarkFlagRequired("template-id")

	return c
}

func newRuleDeleteCmd() *cobra.Command {
	var addToWhitelist bool

	c := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an ACL rule",
		Long: `Delete an ACL rule (unblock a user) by ID.

Examples:
  safeline acl rule delete 1
  safeline acl rule delete 1 --add-to-whitelist`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			req := map[string]interface{}{
				"id":                json.Number(args[0]),
				"add_to_white_list": addToWhitelist,
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] DELETE /api/ACLRuleAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("DELETE", "/api/ACLRuleAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("ACL rule %s deleted successfully\n", args[0])
			return nil
		},
	}

	c.Flags().BoolVar(&addToWhitelist, "add-to-whitelist", false, "Add the user to whitelist")

	return c
}

func newRuleClearCmd() *cobra.Command {
	var templateID int
	var addToWhitelist bool

	c := &cobra.Command{
		Use:   "clear",
		Short: "Clear all ACL rules for a template",
		Long: `Clear all ACL rules (unblock all users) for a template.

Examples:
  safeline acl rule clear --template-id 1
  safeline acl rule clear --template-id 1 --add-to-whitelist`,
		RunE: func(c *cobra.Command, args []string) error {
			if templateID <= 0 {
				return fmt.Errorf("--template-id is required")
			}

			req := map[string]interface{}{
				"acl_rule_template_id": templateID,
				"add_to_white_list":    addToWhitelist,
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] DELETE /api/ClearACLRuleAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("DELETE", "/api/ClearACLRuleAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("All ACL rules for template %d cleared successfully\n", templateID)
			return nil
		},
	}

	c.Flags().IntVar(&templateID, "template-id", 0, "Template ID (required)")
	c.Flags().BoolVar(&addToWhitelist, "add-to-whitelist", false, "Add users to whitelist")
	c.MarkFlagRequired("template-id")

	return c
}
