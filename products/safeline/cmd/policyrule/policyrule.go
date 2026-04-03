package policyrule

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaitin/workspace-cli/products/safeline/cmd"
	"github.com/spf13/cobra"
)

// NewCommand creates the policy-rule command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "policy-rule",
		Short: "Policy rule management commands",
		Long:  "Commands for managing SafeLine policy rules (custom rules).",
	}
	c.AddCommand(newListCmd())
	c.AddCommand(newGetCmd())
	c.AddCommand(newCreateCmd())
	c.AddCommand(newDeleteCmd())
	c.AddCommand(newEnableCmd())
	c.AddCommand(newDisableCmd())
	c.AddCommand(newTargetsCmd())
	return c
}

func newListCmd() *cobra.Command {
	var global bool

	c := &cobra.Command{
		Use:   "list",
		Short: "List policy rules",
		Long: `List policy rules.

Examples:
  # List global rules (default)
  safeline policy-rule list

  # List custom rules
  safeline policy-rule list --global=false`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			scope := "detect:rule:global"
			if !global {
				scope = "detect:rule:custom"
			}
			query := map[string]string{
				"scope":  scope,
				"count":  "20",
				"offset": "0",
			}

			env, err := cl.Do("GET", "/api/FilterV2API", nil, query)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output - parse FilterV2API response structure
			var result struct {
				Total int `json:"total"`
				Items []struct {
					ID        int    `json:"id"`
					Comment   string `json:"comment"`
					Action    string `json:"action"`
					RiskLevel int    `json:"risk_level"`
					IsEnabled bool   `json:"is_enabled"`
					IsGlobal  bool   `json:"is_global"`
				} `json:"items"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}

			headers := []string{"ID", "Comment", "Action", "Risk", "Enabled", "Global"}
			var rows [][]string
			for _, r := range result.Items {
				enabled := "No"
				if r.IsEnabled {
					enabled = "Yes"
				}
				isGlobal := "No"
				if r.IsGlobal {
					isGlobal = "Yes"
				}
				rows = append(rows, []string{
					fmt.Sprintf("%d", r.ID),
					r.Comment,
					r.Action,
					fmt.Sprintf("%d", r.RiskLevel),
					enabled,
					isGlobal,
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}

	c.Flags().BoolVar(&global, "global", true, "List global rules")

	return c
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get policy rule details",
		Long: `Get policy rule details by ID.

Examples:
  safeline policy-rule get 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/PolicyRuleAPI", nil, map[string]string{"id": args[0]})
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Key-value output
			var result struct {
				ID        int                    `json:"id"`
				Comment   string                 `json:"comment"`
				Action    string                 `json:"action"`
				RiskLevel int                    `json:"risk_level"`
				IsEnabled bool                   `json:"is_enabled"`
				IsGlobal  bool                   `json:"is_global"`
				Pattern   map[string]interface{} `json:"pattern"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}

			enabled := "No"
			if result.IsEnabled {
				enabled = "Yes"
			}
			isGlobal := "No"
			if result.IsGlobal {
				isGlobal = "Yes"
			}

			cmd.PrintKeyValue(map[string]string{
				"ID":         fmt.Sprintf("%d", result.ID),
				"Comment":    result.Comment,
				"Action":     result.Action,
				"Risk Level": fmt.Sprintf("%d", result.RiskLevel),
				"Enabled":    enabled,
				"Global":     isGlobal,
			})

			// Print pattern summary
			if result.Pattern != nil {
				patternJSON, _ := json.MarshalIndent(result.Pattern, "", "  ")
				fmt.Printf("\nPattern:\n%s\n", string(patternJSON))
			}
			return nil
		},
	}
}

func newCreateCmd() *cobra.Command {
	var (
		comment     string
		action      string
		riskLevel   int
		enabled     bool
		expireTime  int64
		patternJSON string
		// simple pattern flags
		target string
		cmp    string
		value  string
	)

	c := &cobra.Command{
		Use:   "create",
		Short: "Create a new policy rule",
		Long: `Create a new policy rule.

Two modes are supported:

1. Simple mode (single condition, no logical operators):
   --target <target> --cmp <operator> --value <value>

2. JSON mode (complex conditions):
   --pattern-json '<json>'

Use "safeline policy-rule targets" to list available targets.
Use "safeline policy-rule targets --cmp <target>" to list available operators for a target.

Examples:
  # Simple mode: block requests with /admin in URL
  safeline policy-rule create --comment "Block admin" \
    --target urlpath --cmp infix --value "/admin" \
    --action deny

  # JSON mode: complex conditions
  safeline policy-rule create --comment "Complex rule" \
    --pattern-json '{"$AND":[{"infix":{"urlpath":"admin"}}]}' \
    --action deny`,
		RunE: func(c *cobra.Command, args []string) error {
			if action == "" {
				return fmt.Errorf("--action is required")
			}
			if comment == "" {
				return fmt.Errorf("--comment is required")
			}

			// Build pattern
			var pattern map[string]interface{}
			if patternJSON != "" {
				// JSON mode
				if target != "" || cmp != "" || value != "" {
					return fmt.Errorf("cannot use --target/--cmp/--value with --pattern-json")
				}
				if err := json.Unmarshal([]byte(patternJSON), &pattern); err != nil {
					return fmt.Errorf("invalid pattern JSON: %w", err)
				}
			} else if target != "" {
				// Simple mode
				if cmp == "" {
					return fmt.Errorf("--cmp is required when using --target")
				}
				if value == "" {
					return fmt.Errorf("--value is required when using --target")
				}
				// Build pattern: {"$AND": [{"<cmp>": {"<target>": "<value>"}, "decode_methods": []}]}
				pattern = map[string]interface{}{
					"$AND": []interface{}{
						map[string]interface{}{
							cmp: map[string]interface{}{
								target: value,
							},
							"decode_methods": []string{},
						},
					},
				}
			} else {
				return fmt.Errorf("either --pattern-json or --target/--cmp/--value is required")
			}

			req := map[string]interface{}{
				"comment":     comment,
				"pattern":     pattern,
				"action":      action,
				"risk_level":  riskLevel,
				"is_global":   true,
				"is_enabled":  enabled,
				"rule_type":   1, // GENERAL_RULE
				"expire_time": expireTime,
				"log_option":  "Persistence",
				"attack_type": -1,
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] POST /api/PolicyRuleAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("POST", "/api/PolicyRuleAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Key-value output
			var result struct {
				ID int `json:"id"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}
			cmd.PrintKeyValue(map[string]string{
				"ID":      fmt.Sprintf("%d", result.ID),
				"Comment": comment,
				"Action":  action,
			})
			return nil
		},
	}

	c.Flags().StringVar(&comment, "comment", "", "Rule description (required)")
	c.Flags().StringVar(&action, "action", "", "Action (deny|dry_run|allow) (required)")
	c.Flags().IntVar(&riskLevel, "risk-level", 0, "Risk level (0=none, 1=low, 2=medium, 3=high)")
	c.Flags().BoolVar(&enabled, "enabled", true, "Enable the rule")
	c.Flags().Int64Var(&expireTime, "expire-time", 0, "Expire timestamp (0=never expire)")
	c.Flags().StringVar(&patternJSON, "pattern-json", "", "Pattern JSON (complex mode)")
	c.Flags().StringVar(&target, "target", "", "Target key (simple mode)")
	c.Flags().StringVar(&cmp, "cmp", "", "Comparison operator (simple mode)")
	c.Flags().StringVar(&value, "value", "", "Match value (simple mode)")

	c.MarkFlagRequired("comment")
	c.MarkFlagRequired("action")

	return c
}

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a policy rule",
		Long: `Delete a policy rule by ID.

Examples:
  safeline policy-rule delete 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			req := map[string]interface{}{
				"id__in": []json.Number{json.Number(args[0])},
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] DELETE /api/PolicyRuleAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("DELETE", "/api/PolicyRuleAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("Policy rule %s deleted successfully\n", args[0])
			return nil
		},
	}
}

func newEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <id>",
		Short: "Enable a policy rule",
		Long: `Enable a policy rule by ID.

Examples:
  safeline policy-rule enable 1`,
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
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] PUT /api/EnableDisablePolicyRuleAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("PUT", "/api/EnableDisablePolicyRuleAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("Policy rule %s enabled successfully\n", args[0])
			return nil
		},
	}
}

func newDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <id>",
		Short: "Disable a policy rule",
		Long: `Disable a policy rule by ID.

Examples:
  safeline policy-rule disable 1`,
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
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] PUT /api/EnableDisablePolicyRuleAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("PUT", "/api/EnableDisablePolicyRuleAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("Policy rule %s disabled successfully\n", args[0])
			return nil
		},
	}
}

// Target and CMP types for the targets command
type policyTranslation struct {
	CN          string `json:"cn"`
	EN          string `json:"en"`
	Translation string `json:"translation"`
}

type policyTarget struct {
	Key     string            `json:"key"`
	Name    policyTranslation `json:"name"`
	Comment policyTranslation `json:"comment"`
	Type    interface{}       `json:"type"` // Can be string or object
	Cmp     []string          `json:"cmp"`
}

type policyCmp struct {
	Op   string            `json:"op"`
	Desc policyTranslation `json:"desc"`
}

type policyTargetList struct {
	Targets []policyTarget         `json:"targets"`
	Cmps    map[string][]policyCmp `json:"cmps"`
}

func newTargetsCmd() *cobra.Command {
	var showCmp string

	c := &cobra.Command{
		Use:   "targets",
		Short: "List available targets and operators",
		Long: `List available targets and comparison operators for policy rules.

Examples:
  # List all targets
  safeline policy-rule targets

  # Show operators for a specific target
  safeline policy-rule targets --cmp urlpath`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/ServerControlledConfigAPI", nil, nil)
			if err != nil {
				return err
			}

			// Extract policy_target_list from response
			var result struct {
				PolicyTargetList policyTargetList `json:"policy_target_list"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			if showCmp != "" {
				// Show operators for specific target
				return printTargetCmps(c, result.PolicyTargetList, showCmp)
			}

			// Print all targets (only those supporting str_type)
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			headers := []string{"Key", "Name", "Type", "Description"}
			var rows [][]string
			for _, t := range result.PolicyTargetList.Targets {
				// Filter: only show targets that support str_type only
				typeStr, onlyStrType := getTypeStr(t.Type)
				if !onlyStrType {
					continue
				}
				rows = append(rows, []string{
					t.Key,
					t.Name.EN,
					typeStr,
					t.Comment.EN,
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}

	c.Flags().StringVar(&showCmp, "cmp", "", "Show operators for a specific target")

	return c
}

// getTypeStr returns the type string and whether it only supports str_type
func getTypeStr(t interface{}) (string, bool) {
	switch v := t.(type) {
	case string:
		// Single string type, check if it's kv
		if v == "kv" {
			return v, false
		}
		return v, true
	case map[string]interface{}:
		// Multiple types available, check if only str_type and not kv
		if len(v) == 1 {
			if _, ok := v["str_type"]; ok {
				return "str_type", true
			}
		}
		// Has multiple types or non-str_type, exclude
		var types []string
		for k := range v {
			types = append(types, k)
		}
		return strings.Join(types, ","), false
	}
	return "", false
}

func printTargetCmps(c *cobra.Command, list policyTargetList, targetKey string) error {
	// Find target
	var target *policyTarget
	for i, t := range list.Targets {
		if t.Key == targetKey {
			target = &list.Targets[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("target %q not found", targetKey)
	}

	// Get type string
	typeStr, _ := getTypeStr(target.Type)

	// Print target info
	fmt.Printf("Target: %s (%s)\n", target.Name.EN, target.Key)
	fmt.Printf("Type: %s\n\n", typeStr)
	fmt.Printf("Available operators:\n")

	// Only print cmp groups that this target supports
	headers := []string{"Operator", "Description"}
	var rows [][]string
	for _, cmpKey := range target.Cmp {
		if ops, ok := list.Cmps[cmpKey]; ok {
			for _, op := range ops {
				rows = append(rows, []string{op.Op, op.Desc.EN})
			}
		}
	}
	cmd.PrintTable(headers, rows)
	return nil
}
