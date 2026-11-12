package cloudwalker

import (
	"encoding/json"
	"strings"

	"github.com/chaitin/chaitin-cli/products/cloudwalker/cli/host_asset"
	"github.com/spf13/cobra"
)

const boolFlagNote = `

未在命令行显式设置的 bool 参数不会发送给服务端。需要发送 false 时必须写成 --flag=false。--select-all 只作用于已经写明的过滤条件，不会再附带未设置的 bool。`

const setGroupNote = `
修改主机业务组时，目标业务组使用 --gid。要修改的主机使用 --select，键名是 id，例如 [{"id": 196}]，或使用 --select-all。--host-id 和 --gids 不能代替 --select。`

func init() {
	annotateCommands(rootCmd)
	patchSetGroup(rootCmd)
}

func annotateCommands(cmd *cobra.Command) {
	if cmd.Long == "" {
		cmd.Long = strings.TrimSpace(boolFlagNote)
	} else if !strings.Contains(cmd.Long, "未在命令行显式设置的 bool") {
		cmd.Long = strings.TrimRight(cmd.Long, "\n") + boolFlagNote
	}
	for _, child := range cmd.Commands() {
		annotateCommands(child)
	}
}

func patchSetGroup(root *cobra.Command) {
	cmd := findCommand(root, []string{"host_asset", "set_group"})
	if cmd == nil {
		return
	}
	if !strings.Contains(cmd.Long, "目标业务组使用 --gid") {
		cmd.Long = strings.TrimRight(cmd.Long, "\n") + setGroupNote
	}
	if flag := cmd.Flags().Lookup("select"); flag != nil {
		flag.Usage = "要修改的主机 (JSON，键名是 id，例如 [{\"id\": 196}])。--host-id 和 --gids 不能代替本参数；只设置了 --host-id 且未设置 --select 和 --select-all 时，会自动转换成 [{\"id\": ...}]。目标业务组请使用 --gid。"
	}
	if flag := cmd.Flags().Lookup("gid"); flag != nil {
		flag.Usage = "目标业务组 ID。请用本参数指定要移入的业务组，不要用 --gids 代替。"
	}

	original := cmd.Run
	cmd.Run = func(c *cobra.Command, args []string) {
		if err := applySetGroupHostIDSelect(c); err != nil {
			c.PrintErrln("Error:", err)
			return
		}
		if original != nil {
			original(c, args)
		}
	}
}

func findCommand(cmd *cobra.Command, path []string) *cobra.Command {
	if len(path) == 0 {
		return cmd
	}
	for _, child := range cmd.Commands() {
		if child.Name() == path[0] {
			return findCommand(child, path[1:])
		}
	}
	return nil
}

// applySetGroupHostIDSelect turns an explicit --host-id into select entries
// when the user did not choose hosts another way. The demo server accepts
// {"id": ...} and ignores a bare host_id field.
func applySetGroupHostIDSelect(cmd *cobra.Command) error {
	selectFlag := cmd.Flags().Lookup("select")
	selectAllFlag := cmd.Flags().Lookup("select-all")
	hostIDFlag := cmd.Flags().Lookup("host-id")
	if selectFlag == nil || selectAllFlag == nil || hostIDFlag == nil {
		return nil
	}
	if selectFlag.Changed || selectAllFlag.Changed || !hostIDFlag.Changed {
		return nil
	}

	ids, err := cmd.Flags().GetFloat64Slice("host-id")
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	items := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		items = append(items, map[string]any{"id": id})
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return err
	}
	host_asset.SetGroupSelectJSON = string(raw)
	return nil
}
