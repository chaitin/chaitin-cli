package ddr

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestClassifyPathKeepsParameterContextForNestedList(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "top level list",
			path: "/softwaremanager/list",
			want: "list",
		},
		{
			name: "nested list under path parameter",
			path: "/softwaremanager/{software_hash}/list",
			want: "software-hash-list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyOperationName(tt.path, "POST"); got != tt.want {
				t.Fatalf("classifyOperationName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestClassifyPathKeepsResourceContextForNestedDetail(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "simple detail keeps get alias",
			path: "/policy/channel/{policy_id}/detail",
			want: "get",
		},
		{
			name: "result detail keeps resource context",
			path: "/device/filescantask/results/{task_id}/detail",
			want: "results-get",
		},
		{
			name: "remote detail keeps resource context",
			path: "/device/filescantask/remote/{task_id}/detail",
			want: "remote-get",
		},
		{
			name: "remote instance detail keeps resource context",
			path: "/device/filescantask/remote/{task_id}/instance/{instance_id}/detail/",
			want: "remote-instance-get",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyOperationName(tt.path, "POST"); got != tt.want {
				t.Fatalf("classifyOperationName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestGenerateCommandsDeduplicatesSiblingNames(t *testing.T) {
	api := &OpenAPI{
		Paths: map[string]PathItem{
			"/dataasset/{asset_id}/detail": {
				Post: &Operation{Summary: "资产地图详情"},
			},
			"/dataasset/{md5_hash}/detail": {
				Post: &Operation{Summary: "记录详情"},
			},
			"/device/filedeletetask/list": {
				Post: &Operation{Summary: "文件删除列表"},
			},
			"/device/{device_id}/filedeletetask": {
				Post: &Operation{Summary: "文件删除创建"},
			},
			"/ueba": {
				Post: &Operation{Summary: "UEBA添加事件策略"},
			},
			"/ueba/{ueba_id}": {
				Post: &Operation{Summary: "UEBA修改事件策略"},
			},
			"/ueba/log/action/getfieldvalues": {
				Post: &Operation{Summary: "UEBA事件getfieldvalues"},
			},
			"/ueba/log/action/{ueba_category}/getfieldvalues": {
				Post: &Operation{Summary: "UEBA事件getfieldvalues"},
			},
			"/watermarkrule/{wr_id}/detail": {
				Post: &Operation{Summary: "水印详情"},
			},
			"/watermarkrule/{wt_id}/detail": {
				Post: &Operation{Summary: "水印策略详情"},
			},
		},
	}

	commands, err := NewParser().GenerateCommands(api)
	if err != nil {
		t.Fatalf("GenerateCommands() error = %v", err)
	}

	assertNoDuplicateSiblingNames(t, commands...)
	assertCommandNames(t, findCommand(t, commands, "dataasset"), "asset-id-get", "md5-hash-get")
	assertCommandNames(t, findCommand(t, commands, "device"), "device-id-filedeletetask", "filedeletetask")
	assertCommandNames(t, findCommand(t, commands, "ueba"), "create", "log", "ueba-id-create")
	assertCommandNames(t, findCommand(t, commands, "ueba", "log"), "getfieldvalues", "ueba-category-getfieldvalues")
	assertCommandNames(t, findCommand(t, commands, "watermarkrule"), "wr-id-get", "wt-id-get")
}

func TestGeneratedCommandsUsePathPlaceholders(t *testing.T) {
	root := NewCommand()

	// policy/externaldevice 下不应再出现以写死 UUID 命名的命令，极速模式更新应支持 --policy-id
	externaldevice := findCommand(t, root.Commands(), "policy", "externaldevice")
	for _, cmd := range externaldevice.Commands() {
		if strings.Contains(cmd.Name(), "dbc780f7") {
			t.Fatalf("policy externaldevice should not expose a command named after a hardcoded policy ID, got %q", cmd.Name())
		}
	}
	if flag := findCommand(t, root.Commands(), "policy", "externaldevice", "create").Flags().Lookup("policy-id"); flag == nil {
		t.Fatalf("policy externaldevice create should expose a --policy-id flag")
	}

	// system/tedencryptconfig 下不应再出现以写死 UUID 命名的命令，修改配置应支持 --config-id
	tedencryptconfig := findCommand(t, root.Commands(), "system", "tedencryptconfig")
	for _, cmd := range tedencryptconfig.Commands() {
		if strings.Contains(cmd.Name(), "37852520") {
			t.Fatalf("system tedencryptconfig should not expose a command named after a hardcoded config ID, got %q", cmd.Name())
		}
	}
	if flag := findCommand(t, root.Commands(), "system", "tedencryptconfig", "create").Flags().Lookup("config-id"); flag == nil {
		t.Fatalf("system tedencryptconfig create should expose a --config-id flag")
	}
}

func findCommand(t *testing.T, commands []*cobra.Command, names ...string) *cobra.Command {
	t.Helper()
	current := commands
	var cmd *cobra.Command
	for _, name := range names {
		cmd = nil
		for _, candidate := range current {
			if candidate.Name() == name {
				cmd = candidate
				break
			}
		}
		if cmd == nil {
			t.Fatalf("command %q not found under %q; have %q", name, strings.Join(names, " "), commandNames(current))
		}
		current = cmd.Commands()
	}
	return cmd
}

func assertCommandNames(t *testing.T, cmd *cobra.Command, want ...string) {
	t.Helper()
	have := commandNames(cmd.Commands())
	for _, name := range want {
		if !containsString(have, name) {
			t.Fatalf("%s command names = %q, want %q", cmd.CommandPath(), have, name)
		}
	}
}

func assertNoDuplicateSiblingNames(t *testing.T, commands ...*cobra.Command) {
	t.Helper()
	seen := make(map[string]struct{})
	for _, cmd := range commands {
		if _, ok := seen[cmd.Name()]; ok {
			t.Fatalf("duplicate command name %q", cmd.Name())
		}
		seen[cmd.Name()] = struct{}{}
		assertNoDuplicateSiblingNames(t, cmd.Commands()...)
	}
}

func commandNames(commands []*cobra.Command) []string {
	names := make([]string, 0, len(commands))
	for _, cmd := range commands {
		names = append(names, cmd.Name())
	}
	sort.Strings(names)
	return names
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
