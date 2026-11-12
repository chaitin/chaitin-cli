package xray

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/chaitin/chaitin-cli/products/xray/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewCommand() (*cobra.Command, error) {
	cmd, err := cli.MakeCommand()
	if err != nil {
		return nil, err
	}
	if err := configureRequiredFlags(cmd); err != nil {
		return nil, err
	}
	return cmd, nil
}

func ApplyRuntimeConfig(cmd *cobra.Command, cfg config.Raw, dryRun bool) {
	_ = cmd
	cli.SetRuntimeConfig(cfg, dryRun)
}

// configureRequiredFlags restores required-parameter semantics that are only
// emitted as help text by go-swagger's generated Cobra commands. Model flags
// contain a dot-qualified prefix and are deliberately excluded because users
// may provide the same required fields through --body instead.
func configureRequiredFlags(root *cobra.Command) error {
	if err := configureCommandRequiredFlags(root); err != nil {
		return err
	}
	return configureRequiredFlagsForChildren(root)
}

func configureRequiredFlagsForChildren(parent *cobra.Command) error {
	for _, cmd := range parent.Commands() {
		if err := configureCommandRequiredFlags(cmd); err != nil {
			return err
		}
		if err := configureRequiredFlagsForChildren(cmd); err != nil {
			return err
		}
	}
	return nil
}

func configureCommandRequiredFlags(cmd *cobra.Command) error {
	var markErr error
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		if markErr != nil || strings.Contains(flag.Name, ".") || !strings.HasPrefix(flag.Usage, "Required.") {
			return
		}
		markErr = cobra.MarkFlagRequired(cmd.PersistentFlags(), flag.Name)
	})
	if markErr != nil {
		return fmt.Errorf("configure required flags for %s: %w", cmd.CommandPath(), markErr)
	}

	if cmd.Run == nil && cmd.RunE == nil {
		return nil
	}
	previousPreRunE := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if previousPreRunE != nil {
			if err := previousPreRunE(cmd, args); err != nil {
				return err
			}
		}
		return validatePositiveResourceIDs(cmd)
	}
	return nil
}

func validatePositiveResourceIDs(cmd *cobra.Command) error {
	var validationErr error
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if validationErr != nil || !flag.Changed || !isRequiredResourceID(flag) {
			return
		}
		value, err := strconv.ParseInt(flag.Value.String(), 10, 64)
		if err != nil {
			validationErr = fmt.Errorf("invalid --%s: %w", flag.Name, err)
			return
		}
		if value <= 0 {
			validationErr = fmt.Errorf("--%s must be greater than zero", flag.Name)
		}
	})
	return validationErr
}

func isRequiredResourceID(flag *pflag.Flag) bool {
	if flag.Value.Type() != "int" && flag.Value.Type() != "int64" {
		return false
	}
	if _, required := flag.Annotations[cobra.BashCompOneRequiredFlag]; !required {
		return false
	}
	name := strings.ReplaceAll(flag.Name, "-", "_")
	return name == "id" || strings.HasSuffix(name, "_id")
}
