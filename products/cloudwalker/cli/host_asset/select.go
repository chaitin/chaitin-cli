package host_asset

import "github.com/spf13/cobra"

// applyHostIDSelect turns --host-id into the select filter that strict
// HostAssetService writes require.
//
// SetGroup embeds the list filter, so --host-id is a real filter and is still
// sent. The write path additionally requires select or select_all; without
// either of them the daemon returns "no select filter set". When the user
// passed --host-id and did not pass --select or --select-all, synthesize
// select entries as [{"id": <host-id>}]. An explicit selection wins.
func applyHostIDSelect(cmd *cobra.Command, hostIDs []float64, selected *[]map[string]interface{}, selectAll *bool) {
	if cmd == nil || selected == nil || selectAll == nil {
		return
	}
	if !cmd.Flags().Changed("host-id") || len(hostIDs) == 0 {
		return
	}
	if cmd.Flags().Changed("select") || (cmd.Flags().Changed("select-all") && *selectAll) {
		return
	}

	selects := make([]map[string]interface{}, 0, len(hostIDs))
	for _, id := range hostIDs {
		selects = append(selects, map[string]interface{}{"id": id})
	}
	*selected = selects
}
