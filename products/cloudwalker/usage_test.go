package cloudwalker

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chaitin/chaitin-cli/products/cloudwalker/cli/host_asset"
	"github.com/spf13/cobra"
)

func TestCommandHelpExplainsUnsetBools(t *testing.T) {
	cmd := findCommand(NewCommand(), []string{"host_asset", "get_host_asset_list"})
	if cmd == nil {
		t.Fatal("get_host_asset_list not found")
	}

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.Help(); err != nil {
		t.Fatal(err)
	}
	help := buf.String()
	if !strings.Contains(help, "未在命令行显式设置的 bool 参数不会发送给服务端") {
		t.Fatalf("help missing bool note: %s", help)
	}
	if !strings.Contains(help, "--flag=false") {
		t.Fatalf("help missing explicit false guidance: %s", help)
	}
}

func TestSetGroupHelpUsesIDSelect(t *testing.T) {
	cmd := findCommand(NewCommand(), []string{"host_asset", "set_group"})
	if cmd == nil {
		t.Fatal("set_group not found")
	}

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.Help(); err != nil {
		t.Fatal(err)
	}
	help := buf.String()
	for _, want := range []string{
		"目标业务组使用 --gid",
		"[{\"id\": 196}]",
		"--host-id 和 --gids 不能代替",
		"未在命令行显式设置的 bool 参数不会发送给服务端",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("set_group help missing %q", want)
		}
	}
}

func TestSetGroupHostIDBecomesSelect(t *testing.T) {
	cmd := setGroupCmd(t)
	t.Cleanup(func() { resetSetGroupFlags(cmd) })
	resetSetGroupFlags(cmd)

	if err := cmd.Flags().Set("host-id", "3"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("host-id", "4"); err != nil {
		t.Fatal(err)
	}
	if err := applySetGroupHostIDSelect(cmd); err != nil {
		t.Fatal(err)
	}
	if host_asset.SetGroupSelectJSON != "[{\"id\":3},{\"id\":4}]" {
		t.Fatalf("select = %s", host_asset.SetGroupSelectJSON)
	}
}

func TestSetGroupHostIDDoesNotOverrideExplicitSelect(t *testing.T) {
	cmd := setGroupCmd(t)
	t.Cleanup(func() { resetSetGroupFlags(cmd) })
	resetSetGroupFlags(cmd)

	explicit := "[{\"id\":1}]"
	if err := cmd.Flags().Set("select", explicit); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("host-id", "3"); err != nil {
		t.Fatal(err)
	}
	if err := applySetGroupHostIDSelect(cmd); err != nil {
		t.Fatal(err)
	}
	if host_asset.SetGroupSelectJSON != explicit {
		t.Fatalf("explicit select overwritten: %s", host_asset.SetGroupSelectJSON)
	}
}

func TestSetGroupHostIDDoesNotOverrideSelectAll(t *testing.T) {
	cmd := setGroupCmd(t)
	t.Cleanup(func() { resetSetGroupFlags(cmd) })
	resetSetGroupFlags(cmd)

	if err := cmd.Flags().Set("select-all", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("host-id", "3"); err != nil {
		t.Fatal(err)
	}
	if err := applySetGroupHostIDSelect(cmd); err != nil {
		t.Fatal(err)
	}
	if host_asset.SetGroupSelectJSON != "" {
		t.Fatalf("select-all should not synthesize select, got %s", host_asset.SetGroupSelectJSON)
	}
}

func setGroupCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := findCommand(NewCommand(), []string{"host_asset", "set_group"})
	if cmd == nil {
		t.Fatal("set_group not found")
	}
	return cmd
}

func resetSetGroupFlags(cmd *cobra.Command) {
	host_asset.SetGroupSelectJSON = ""
	for _, name := range []string{"select", "select-all", "host-id"} {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			continue
		}
		flag.Changed = false
		if err := flag.Value.Set(flag.DefValue); err != nil {
			_ = flag.Value.Set("")
		}
	}
}
