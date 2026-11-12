package tanswer

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRootCommandRegistersAgentReadableCommands(t *testing.T) {
	cmd := NewCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	expected := []string{"auth", "manifest", "system", "alarm", "file-alarm", "asset", "metadata", "policy", "response", "api"}
	for _, name := range expected {
		found := false
		for _, child := range cmd.Commands() {
			if child.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected child command %q", name)
		}
	}
}

func TestTAnswerHelpGuidesUsersWithoutProductDocuments(t *testing.T) {
	var out bytes.Buffer
	cmd := NewCommand()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	for _, want := range []string{
		"TANSWER_URL", "TANSWER_API_KEY", "auth check", "manifest",
		"--help", "preview", "confirm", "semantic commands", "明确确认",
	} {
		if !strings.Contains(strings.ToLower(out.String()), strings.ToLower(want)) {
			t.Fatalf("help missing %q:\n%s", want, out.String())
		}
	}
}

func TestTAnswerRootHelpDirectsUsersToProductHelp(t *testing.T) {
	cmd := NewCommand()
	if !strings.Contains(cmd.Short, "chaitin-cli tanswer --help") {
		t.Fatalf("root command short help does not direct users to product help: %q", cmd.Short)
	}
	if strings.Contains(cmd.Long, "\n\t") {
		t.Fatalf("root command help contains unintended indentation: %q", cmd.Long)
	}
}

func TestHumanCommandReferenceMatchesRuntimeManifest(t *testing.T) {
	raw, err := os.ReadFile("COMMAND_REFERENCE.md")
	if err != nil {
		t.Fatalf("read command reference: %v", err)
	}
	reference := string(raw)
	for _, entry := range BuildCommandManifest().Commands {
		if !commandReferenceContains(reference, entry.Name) {
			t.Fatalf("command reference missing runtime command %q", entry.Name)
		}
	}
	if !strings.Contains(reference, "前五类为只读查询；后三类均带有 `--preview`") {
		t.Fatalf("command reference must accurately describe the read-only and preview examples")
	}
}

func commandReferenceContains(reference, command string) bool {
	for _, line := range strings.Split(reference, "\n") {
		line = strings.TrimSpace(line)
		if line == command || strings.HasPrefix(line, command+" ") {
			return true
		}
	}
	return false
}

func TestEnglishRepositoryGuideIncludesTAnswerOnboarding(t *testing.T) {
	raw, err := os.ReadFile("../../README.en.md")
	if err != nil {
		t.Fatalf("read English README: %v", err)
	}
	english := string(raw)
	for _, want := range []string{
		"### T-Answer Quick Start",
		"chaitin-cli tanswer auth check",
		"chaitin-cli tanswer manifest",
		"explicit confirmation",
	} {
		if !strings.Contains(english, want) {
			t.Fatalf("English README missing %q", want)
		}
	}
}

func TestCompanySkillMatchesRuntimeConfigDiscovery(t *testing.T) {
	raw, err := os.ReadFile("../../skills/chaitin-cli/SKILL.md")
	if err != nil {
		t.Fatalf("read company skill: %v", err)
	}
	skill := string(raw)
	for _, want := range []string{
		"recognized `./config.yaml`",
		"`~/.chaitin-cli/config.yaml`",
		"`flags > environment/.env > recognized ./config.yaml > ~/.chaitin-cli/config.yaml`",
	} {
		if !strings.Contains(skill, want) {
			t.Fatalf("company skill configuration guidance missing %q", want)
		}
	}
	if strings.Contains(skill, "Config file path (default: `./config.yaml`)") {
		t.Fatal("company skill must not describe ./config.yaml as the unconditional --config default")
	}
}

func TestTAnswerRuntimeAndRepositoryGuidanceListAllConfigurationSources(t *testing.T) {
	var out bytes.Buffer
	cmd := NewCommand()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("render tanswer help: %v", err)
	}

	configSources := []string{
		"--url", "TANSWER_URL", "tanswer.url",
		"--api-key", "TANSWER_API_KEY", "tanswer.api_key",
		"--timeout", "TANSWER_TIMEOUT", "tanswer.timeout",
		"--insecure", "TANSWER_INSECURE", "tanswer.insecure",
	}
	for _, source := range configSources {
		if !containsGuidanceToken(out.String(), source) {
			t.Errorf("tanswer help configuration guidance missing %q:\n%s", source, out.String())
		}
	}

	for _, path := range []string{"../../README.md", "../../README.en.md", "../../skills/chaitin-cli/SKILL.md", "README.md", "COMMAND_REFERENCE.md"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read %s: %v", path, err)
			continue
		}
		for _, source := range configSources {
			if !containsGuidanceToken(string(raw), source) {
				t.Errorf("%s configuration guidance missing %q", path, source)
			}
		}
	}
}

func TestGuidanceTokenMatchingRequiresBoundaries(t *testing.T) {
	if containsGuidanceToken("TANSWER_TIMEOUT_REGRESSED", "TANSWER_TIMEOUT") {
		t.Fatal("a configuration token must not match a longer identifier")
	}
	if !containsGuidanceToken("set TANSWER_TIMEOUT=30s", "TANSWER_TIMEOUT") {
		t.Fatal("expected standalone configuration token to match")
	}
}

func containsGuidanceToken(text string, token string) bool {
	for offset := 0; ; {
		index := strings.Index(text[offset:], token)
		if index < 0 {
			return false
		}
		index += offset
		end := index + len(token)
		beforeIsToken := index > 0 && isGuidanceTokenCharacter(text[index-1])
		afterIsToken := end < len(text) && isGuidanceTokenCharacter(text[end])
		if !beforeIsToken && !afterIsToken {
			return true
		}
		offset = end
	}
}

func isGuidanceTokenCharacter(value byte) bool {
	return value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9' ||
		value == '_' || value == '.' || value == '-'
}

func TestManifestCommandsAreDiscoverableFromHelp(t *testing.T) {
	for _, entry := range BuildCommandManifest().Commands {
		var out bytes.Buffer
		root := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
		args := strings.Fields(strings.TrimPrefix(entry.FullCommand, "chaitin-cli "))
		command, _, err := root.Find(args)
		if err != nil || command == nil {
			t.Fatalf("manifest command %s is not registered: %v", entry.FullCommand, err)
		}
		if strings.TrimSpace(command.Long) == "" {
			t.Fatalf("manifest command %s has no detailed help", entry.FullCommand)
		}
		args = append(args, "--help")
		root.SetArgs(args)

		if err := root.Execute(); err != nil {
			t.Fatalf("%s help returned error: %v", entry.FullCommand, err)
		}
		help := out.String()
		if !strings.Contains(help, command.Long) {
			t.Fatalf("%s did not render detailed help:\n%s", entry.FullCommand, help)
		}
		if entry.RequiresConfirmation && (!strings.Contains(strings.ToLower(help), "preview") || !strings.Contains(strings.ToLower(help), "confirm")) {
			t.Fatalf("protected command %s help must mention preview and confirm:\n%s", entry.FullCommand, help)
		}
	}
}
