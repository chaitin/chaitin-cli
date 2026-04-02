package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chaitin/workspace-cli/config"
	"github.com/chaitin/workspace-cli/products/chaitin"
	"github.com/chaitin/workspace-cli/products/cloudwalker"
	"github.com/chaitin/workspace-cli/products/tanswer"
	"github.com/chaitin/workspace-cli/products/xray"
	"github.com/spf13/cobra"
)

type app struct {
	root             *cobra.Command
	aliasSubcommands map[string]struct{}
	config           config.Raw
	dryRun           bool
}

func newApp() (*app, error) {
	cfg, err := loadConfigFile(configPathFromCWD())
	if err != nil {
		return nil, err
	}

	root := &cobra.Command{
		Use:           "cws",
		Short:         "CLI for Chaitin Tech products",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	a := &app{
		root:             root,
		aliasSubcommands: make(map[string]struct{}),
		config:           cfg,
	}

	root.PersistentFlags().BoolVar(&a.dryRun, "dry-run", false, "Do not send requests for commands that support dry-run")

	a.registerProductCommand(chaitin.NewCommand())
	a.registerProductCommand(cloudwalker.NewCommand())
	a.registerProductCommand(tanswer.NewCommand())

	xrayCmd, err := xray.NewCommand()
	if err != nil {
		return nil, err
	}
	a.registerProductCommand(xrayCmd)

	// TODO: register more products

	return a, nil
}

func (a *app) execute() error {
	a.rewriteArgsForAlias()
	return a.root.Execute()
}

func (a *app) rewriteArgsForAlias() {
	argv0 := normalizeBinaryName(os.Args[0])
	if argv0 == "" || argv0 == a.root.Name() {
		return
	}

	if _, ok := a.aliasSubcommands[argv0]; !ok {
		return
	}

	args := make([]string, 0, len(os.Args))
	args = append(args, os.Args[0], argv0)
	args = append(args, os.Args[1:]...)
	a.root.SetArgs(args[1:])
}

func (a *app) registerProductCommand(cmd *cobra.Command) {
	if cmd == nil || cmd.Name() == "" {
		return
	}

	a.wrapProductCommand(cmd)
	a.aliasSubcommands[cmd.Name()] = struct{}{}
	a.root.AddCommand(cmd)
}

func (a *app) wrapProductCommand(cmd *cobra.Command) {
	oldPreRun := cmd.PersistentPreRun
	oldPreRunE := cmd.PersistentPreRunE

	cmd.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		switch cmd.Name() {
		case "cloudwalker":
			cloudwalker.ApplyRuntimeConfig(command, a.config)
		case "tanswer":
			tanswer.ApplyRuntimeConfig(command, a.config)
		case "xray":
			xray.ApplyRuntimeConfig(command, a.config, a.dryRun)
			// TODO: register more products
		}

		if oldPreRun != nil {
			oldPreRun(command, args)
		}
		if oldPreRunE != nil {
			return oldPreRunE(command, args)
		}
		return nil
	}
}

func normalizeBinaryName(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.TrimSpace(base)
	if base == "." || base == string(filepath.Separator) {
		return ""
	}
	return base
}

func main() {
	app, err := newApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.execute(); err != nil {
		log.Fatal(err)
	}
}
