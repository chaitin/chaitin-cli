package cloudwalker

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/chaitin/chaitin-cli/products/cloudwalker/cli/host_asset"
	"github.com/chaitin/chaitin-cli/products/cloudwalker/client"
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
	if !strings.Contains(help, "未在命令行显式设置的 bool 开关不会发送给服务端") {
		t.Fatalf("help missing bool note: %s", help)
	}
	if !strings.Contains(help, "--flag=false") {
		t.Fatalf("help missing explicit false guidance: %s", help)
	}
}

func TestSetGroupHelpUsesHostIDSelect(t *testing.T) {
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
		"未在命令行显式设置的 bool 开关不会发送给服务端",
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
	ids, err := cmd.Flags().GetFloat64Slice("host-id")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("top-level host-id still set: %v", ids)
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

func TestSetGroupHostIDDoesNotOverrideSelectAllTrue(t *testing.T) {
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
		t.Fatalf("select-all=true should not synthesize select, got %s", host_asset.SetGroupSelectJSON)
	}
}

func TestSetGroupSelectAllFalseStillConvertsHostID(t *testing.T) {
	cmd := setGroupCmd(t)
	t.Cleanup(func() { resetSetGroupFlags(cmd) })
	resetSetGroupFlags(cmd)

	if err := cmd.Flags().Set("select-all", "false"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("host-id", "3"); err != nil {
		t.Fatal(err)
	}
	if err := applySetGroupHostIDSelect(cmd); err != nil {
		t.Fatal(err)
	}
	if host_asset.SetGroupSelectJSON != "[{\"id\":3}]" {
		t.Fatalf("select = %s", host_asset.SetGroupSelectJSON)
	}
}

func TestSetGroupExecuteOmitsTopLevelHostID(t *testing.T) {
	root := NewCommand()
	cmd := findCommand(root, []string{"host_asset", "set_group"})
	if cmd == nil {
		t.Fatal("set_group not found")
	}
	t.Cleanup(func() {
		resetSetGroupFlags(cmd)
		client.SetDryRun(false)
	})
	resetSetGroupFlags(cmd)
	client.SetDryRun(true)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SilenceUsage = true

	stdout := captureStdout(t, func() {
		root.SetArgs([]string{"host_asset", "set_group", "--host-id", "3", "--gid", "5"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	params := dryRunParams(t, stdout)
	selectItems, ok := params["select"].([]any)
	if !ok || len(selectItems) != 1 {
		t.Fatalf("select = %#v", params["select"])
	}
	item, ok := selectItems[0].(map[string]any)
	if !ok || item["id"] != float64(3) {
		t.Fatalf("select item = %#v", selectItems[0])
	}
	if params["host_id"] != nil {
		t.Fatalf("top-level host_id = %#v, want null", params["host_id"])
	}
	if params["gid"] != float64(5) {
		t.Fatalf("gid = %#v", params["gid"])
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
	if flag := cmd.Flags().Lookup("host-id"); flag != nil {
		_ = clearFloat64Slice(flag)
	}
	for _, name := range []string{"select", "select-all", "gid"} {
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

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	out, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func dryRunParams(t *testing.T, stdout string) map[string]any {
	t.Helper()
	const marker = "[DRY-RUN] Request Body:"
	idx := strings.Index(stdout, marker)
	if idx < 0 {
		t.Fatalf("dry-run body not found: %s", stdout)
	}
	raw := strings.TrimSpace(stdout[idx+len(marker):])
	var body struct {
		Params map[string]any `json:"params"`
	}
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatalf("unmarshal dry-run body: %v\n%s", err, raw)
	}
	return body.Params
}
