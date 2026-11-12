package xray

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestGeneratedOperationRequiredFlagsAreMarked(t *testing.T) {
	cmd := newTestCommand(t)
	requiredCount := 0
	visitCommands(cmd, func(command *cobra.Command) {
		command.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
			if strings.Contains(flag.Name, ".") || !strings.HasPrefix(flag.Usage, "Required.") {
				return
			}
			requiredCount++
			if _, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]; !ok {
				t.Errorf("%s --%s is not marked required", command.CommandPath(), flag.Name)
			}
		})
	})
	if requiredCount == 0 {
		t.Fatal("expected generated required operation flags")
	}
}

func TestDeletePlanRequiredID(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "missing", args: []string{"plan", "DeletePlanID"}, wantErr: `required flag(s) "id" not set`},
		{name: "zero", args: []string{"plan", "DeletePlanID", "--id=0"}, wantErr: "--id must be greater than zero"},
		{name: "valid", args: []string{"plan", "DeletePlanID", "--id=1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestCommand(t)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestRequiredBodyModelFlagsAreNotMarked(t *testing.T) {
	cmd := newTestCommand(t)
	found := false
	visitCommands(cmd, func(command *cobra.Command) {
		command.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
			if !strings.Contains(flag.Name, ".") || !strings.HasPrefix(flag.Usage, "Required.") {
				return
			}
			found = true
			if _, marked := flag.Annotations[cobra.BashCompOneRequiredFlag]; marked {
				t.Errorf("body model flag --%s must not be marked required", flag.Name)
			}
		})
	})
	if !found {
		t.Fatal("expected at least one required body model flag")
	}
}

func newTestCommand(t *testing.T) *cobra.Command {
	t.Helper()
	t.Setenv("XRAY_URL", "https://example.invalid/api/v2")
	cmd, err := NewCommand()
	if err != nil {
		t.Fatal(err)
	}
	ApplyRuntimeConfig(cmd, nil, true)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return cmd
}

func visitCommands(command *cobra.Command, visit func(*cobra.Command)) {
	visit(command)
	for _, child := range command.Commands() {
		visitCommands(child, visit)
	}
}
