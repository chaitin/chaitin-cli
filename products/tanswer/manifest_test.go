package tanswer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestManifestCommandOutputsAIReadableCommandMetadata(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "manifest"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("manifest output is not envelope json: %v\n%s", err, out.String())
	}
	if env.Task != "查看 CLI 命令清单" {
		t.Fatalf("Task = %q", env.Task)
	}

	raw, err := json.Marshal(env.Data)
	if err != nil {
		t.Fatalf("marshal Data: %v", err)
	}
	var manifest CommandManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("manifest data shape mismatch: %v\n%s", err, string(raw))
	}

	if manifest.Product != "tanswer" {
		t.Fatalf("Product = %q", manifest.Product)
	}
	requireManifestCommand(t, manifest, "tanswer manifest", "foundation", "read")
	requireManifestCommand(t, manifest, "tanswer alarm overview", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm timeline", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm list", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm high-priority", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm detail", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm by-attacker", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm by-victim", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm by-threat", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm important-assets", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm attacker-rank", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm victim-rank", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm phase-distribution", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer alarm related", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer file-alarm overview", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer file-alarm malicious", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer file-alarm webshell", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer file-alarm sandbox", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer file-alarm detail", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer system status", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer asset list", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer asset detail", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer asset group-tree", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer asset download-template", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer asset export", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer asset create", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer asset update", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer asset delete", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer asset batch-maintain", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer asset batch-tag", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer asset group-create", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer asset group-rename", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer asset group-delete", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer asset tree-move", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer asset import", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer metadata protocol", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer metadata search", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer metadata detail", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer metadata near-alarm", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer metadata config", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer metadata config-update", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy detection-whitelist", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer policy detection-whitelist-create", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy detection-whitelist-update", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy detection-whitelist-enable", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy detection-whitelist-disable", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy detection-whitelist-delete", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy detection-whitelist-from-alarm", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy detection-whitelist-export", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer policy detection-whitelist-import", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy custom-intelligence", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer policy custom-intelligence-create", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy custom-intelligence-update", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy custom-intelligence-enable", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy custom-intelligence-disable", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy custom-intelligence-delete", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer policy custom-intelligence-export", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer policy custom-intelligence-import", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response block-policies", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer response block-policy-create", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response block-policy-update", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response block-policy-enable", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response block-policy-disable", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response block-policy-delete", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response block-records", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer response whitelist", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer response whitelist-create", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response whitelist-update", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response whitelist-enable", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response whitelist-disable", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response whitelist-delete", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response block-policy-from-alarm", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response whitelist-from-alarm", "semantic_shortcut", "write_high")
	requireManifestCommand(t, manifest, "tanswer response devices", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer response device-records", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer response auto-policies", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer response auto-list", "semantic_shortcut", "read")
	requireManifestCommand(t, manifest, "tanswer api", "openapi_fallback", "read_write")
	api := findManifestCommand(t, manifest, "tanswer api")
	if !api.RequiresConfirmation {
		t.Fatal("raw API fallback must require confirmation for potentially mutating requests")
	}
	rawAPI, err := json.Marshal(api)
	if err != nil {
		t.Fatalf("marshal raw API manifest command: %v", err)
	}
	var rawAPIContract struct {
		ConfirmationCondition string   `json:"confirmation_condition"`
		PreviewOutputFields   []string `json:"preview_output_fields"`
	}
	if err := json.Unmarshal(rawAPI, &rawAPIContract); err != nil {
		t.Fatalf("unmarshal raw API manifest contract: %v", err)
	}
	if rawAPIContract.ConfirmationCondition != "required for non-GET/HEAD requests" {
		t.Fatalf("raw API confirmation condition = %q", rawAPIContract.ConfirmationCondition)
	}
	for _, field := range []string{"requires_confirmation", "confirmation_token", "target", "change_summary", "impact", "risk_warnings"} {
		if !containsString(rawAPIContract.PreviewOutputFields, field) {
			t.Fatalf("raw API preview output fields = %#v, missing %q", rawAPIContract.PreviewOutputFields, field)
		}
	}
	for _, example := range api.Examples {
		var previewFlag, confirmFlag *ManifestFlag
		for index := range api.Flags {
			flag := &api.Flags[index]
			switch flag.Name {
			case "--preview":
				previewFlag = flag
			case "--confirm":
				confirmFlag = flag
			}
		}
		if previewFlag == nil || !strings.Contains(previewFlag.Description, "non-GET/HEAD") {
			t.Fatalf("raw API preview flag contract = %#v", previewFlag)
		}
		if confirmFlag == nil || !strings.Contains(confirmFlag.Description, "CONFIRM_TANSWER_RAW_API_WRITE") {
			t.Fatalf("raw API confirm flag contract = %#v", confirmFlag)
		}

		if strings.Contains(example.Command, "OpsService.GetBaseInfo") {
			t.Fatalf("raw API example must not bypass the system status semantic command: %q", example.Command)
		}
	}
	requireManifestCommand(t, manifest, "tanswer auth check", "foundation", "read")
}

func TestSemanticProtectedWritesDeclareCompletePreviewContract(t *testing.T) {
	manifest := BuildCommandManifest()
	requiredFields := []string{
		"requires_confirmation",
		"confirmed",
		"operation_type",
		"risk_level",
		"target",
		"change_summary",
		"impact",
		"risk_warnings",
		"confirmation_token",
		"confirmation_note",
	}

	protectedCount := 0
	for _, command := range manifest.Commands {
		if command.Layer != "semantic_shortcut" || !command.RequiresConfirmation {
			continue
		}
		protectedCount++
		if command.ConfirmationCondition != "required after preview and explicit user confirmation" {
			t.Errorf("%s confirmation condition = %q", command.Name, command.ConfirmationCondition)
		}
		for _, field := range requiredFields {
			if !containsString(command.PreviewOutputFields, field) {
				t.Errorf("%s preview output fields = %#v, missing %q", command.Name, command.PreviewOutputFields, field)
			}
		}
	}
	if protectedCount == 0 {
		t.Fatal("manifest has no protected semantic write commands")
	}
}

func TestAlarmByThreatManifestDeclaresAtLeastOneThreatSelector(t *testing.T) {
	command := findManifestCommand(t, BuildCommandManifest(), "tanswer alarm by-threat")
	if len(command.InputConstraints) != 1 {
		t.Fatalf("input constraints = %#v", command.InputConstraints)
	}
	constraint := command.InputConstraints[0]
	if constraint.Rule != "at_least_one" {
		t.Fatalf("constraint rule = %q", constraint.Rule)
	}
	if len(constraint.Flags) != 3 || !containsString(constraint.Flags, "--name") || !containsString(constraint.Flags, "--tag") || !containsString(constraint.Flags, "--phase") {
		t.Fatalf("constraint flags = %#v", constraint.Flags)
	}
}

func TestPolicyDetectionWhitelistWriteManifestRequiresExpiry(t *testing.T) {
	for _, name := range []string{
		"tanswer policy detection-whitelist-create",
		"tanswer policy detection-whitelist-update",
		"tanswer policy detection-whitelist-from-alarm",
	} {
		command := findManifestCommand(t, BuildCommandManifest(), name)
		found := false
		for _, constraint := range command.InputConstraints {
			if constraint.Rule == "at_least_one" && containsString(constraint.Flags, "--expire") && containsString(constraint.Flags, "--valid-time") {
				found = true
			}
		}
		if !found {
			t.Errorf("%s input constraints = %#v, want expiry constraint", name, command.InputConstraints)
		}
	}
}

func findManifestCommand(t *testing.T, manifest CommandManifest, name string) ManifestCommand {
	t.Helper()
	for _, cmd := range manifest.Commands {
		if cmd.Name == name {
			return cmd
		}
	}
	t.Fatalf("manifest missing command %q", name)
	return ManifestCommand{}
}

func TestManifestHelpExplainsPurpose(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "manifest", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查看 CLI 命令清单", "AI Agent", "risk", "output"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestManifestIncludesRuntimeBootstrapGuidance(t *testing.T) {
	manifest := BuildCommandManifest()
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}

	var payload struct {
		Bootstrap struct {
			Authentication struct {
				URL struct {
					Flag       string `json:"flag"`
					Env        string `json:"env"`
					ConfigPath string `json:"config_path"`
				} `json:"url"`
				APIKey struct {
					Flag   string `json:"flag"`
					Env    string `json:"env"`
					Secret bool   `json:"secret"`
				} `json:"api_key"`
				VerifyCommand string   `json:"verify_command"`
				Precedence    []string `json:"precedence"`
			} `json:"authentication"`
			AgentProtocol []string `json:"agent_protocol"`
		} `json:"bootstrap"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal bootstrap: %v", err)
	}

	auth := payload.Bootstrap.Authentication
	if auth.URL.Flag != "--url" || auth.URL.Env != "TANSWER_URL" || auth.URL.ConfigPath != "tanswer.url" {
		t.Fatalf("URL bootstrap = %#v", auth.URL)
	}
	if auth.APIKey.Flag != "--api-key" || auth.APIKey.Env != "TANSWER_API_KEY" || !auth.APIKey.Secret {
		t.Fatalf("API key bootstrap = %#v", auth.APIKey)
	}
	if auth.VerifyCommand != "chaitin-cli tanswer auth check" {
		t.Fatalf("verify command = %q", auth.VerifyCommand)
	}
	if len(auth.Precedence) != 4 || auth.Precedence[0] != "command flag" {
		t.Fatalf("configuration precedence = %#v", auth.Precedence)
	}
	if !containsString(payload.Bootstrap.AgentProtocol, "discover commands and unknown flags with --help") {
		t.Fatalf("agent protocol = %#v", payload.Bootstrap.AgentProtocol)
	}
	if !containsString(payload.Bootstrap.AgentProtocol, "wait for explicit user confirmation after preview before executing a protected write") {
		t.Fatalf("agent protocol must require explicit user confirmation: %#v", payload.Bootstrap.AgentProtocol)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func requireManifestCommand(t *testing.T, manifest CommandManifest, name string, layer string, risk string) {
	t.Helper()
	for _, cmd := range manifest.Commands {
		if cmd.Name == name {
			if cmd.Layer != layer {
				t.Fatalf("%s layer = %q", name, cmd.Layer)
			}
			if cmd.RiskLevel != risk {
				t.Fatalf("%s risk = %q", name, cmd.RiskLevel)
			}
			if risk == "write_high" && !cmd.RequiresConfirmation {
				t.Fatalf("%s requires_confirmation = false", name)
			}
			if cmd.OutputType == "" {
				t.Fatalf("%s output type is empty", name)
			}
			if len(cmd.Examples) == 0 {
				t.Fatalf("%s examples are empty", name)
			}
			return
		}
	}
	t.Fatalf("manifest missing command %q", name)
}
