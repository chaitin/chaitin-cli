package host_asset

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestApplyHostIDSelectSynthesizesSelect(t *testing.T) {
	cmd := &cobra.Command{Use: "set_group"}
	var hostIDs []float64
	var selected []map[string]interface{}
	var selectAll bool
	cmd.Flags().Float64SliceVar(&hostIDs, "host-id", nil, "host id")
	cmd.Flags().BoolVar(&selectAll, "select-all", false, "select all")
	cmd.Flags().String("select", "", "select")
	if err := cmd.Flags().Set("host-id", "3"); err != nil {
		t.Fatal(err)
	}

	applyHostIDSelect(cmd, hostIDs, &selected, &selectAll)

	if len(selected) != 1 || selected[0]["id"] != 3.0 {
		t.Fatalf("select = %#v, want [{id: 3}]", selected)
	}
}

func TestApplyHostIDSelectKeepsExplicitSelect(t *testing.T) {
	cmd := &cobra.Command{Use: "set_group"}
	var hostIDs []float64
	selected := []map[string]interface{}{{"id": 9.0}}
	var selectAll bool
	cmd.Flags().Float64SliceVar(&hostIDs, "host-id", nil, "host id")
	cmd.Flags().BoolVar(&selectAll, "select-all", false, "select all")
	cmd.Flags().String("select", "", "select")
	if err := cmd.Flags().Set("host-id", "3"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("select", `[{"id":9}]`); err != nil {
		t.Fatal(err)
	}

	applyHostIDSelect(cmd, hostIDs, &selected, &selectAll)

	if len(selected) != 1 || selected[0]["id"] != 9.0 {
		t.Fatalf("explicit select was overwritten: %#v", selected)
	}
}

func TestApplyHostIDSelectKeepsSelectAll(t *testing.T) {
	cmd := &cobra.Command{Use: "set_group"}
	var hostIDs []float64
	var selected []map[string]interface{}
	selectAll := false
	cmd.Flags().Float64SliceVar(&hostIDs, "host-id", nil, "host id")
	cmd.Flags().BoolVar(&selectAll, "select-all", false, "select all")
	cmd.Flags().String("select", "", "select")
	if err := cmd.Flags().Set("host-id", "3"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("select-all", "true"); err != nil {
		t.Fatal(err)
	}
	selectAll = true

	applyHostIDSelect(cmd, hostIDs, &selected, &selectAll)

	if selected != nil {
		t.Fatalf("select-all should not synthesize select, got %#v", selected)
	}
}
