package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/chaitin/chaitin-cli/products/chaitin"
	"github.com/chaitin/chaitin-cli/products/cloudwalker"
	"github.com/chaitin/chaitin-cli/products/ddr"
	"github.com/chaitin/chaitin-cli/products/safeline"
	safelinece "github.com/chaitin/chaitin-cli/products/safeline-ce"
	"github.com/chaitin/chaitin-cli/products/tanswer"
	"github.com/chaitin/chaitin-cli/products/xray"
	"github.com/spf13/cobra"
)

type app struct {
	root             *cobra.Command
	aliasSubcommands map[string]struct{}
	config           config.Raw
	configPath       string
	configLoaded     bool
	dryRun           bool
}

const defaultConfigPath = "config.yaml"

func newApp() (*app, error) {
	root := &cobra.Command{
		Use:           "chaitin-cli",
		Short:         "CLI for Chaitin Tech products",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	a := &app{
		root:             root,
		aliasSubcommands: make(map[string]struct{}),
		configPath:       defaultConfigPathFromCWD(),
	}

	root.PersistentFlags().StringVarP(&a.configPath, "config", "c", a.configPath, "Config file path")
	root.PersistentFlags().BoolVar(&a.dryRun, "dry-run", false, "Do not send requests for commands that support dry-run")

	a.registerProductCommand(chaitin.NewCommand())
	a.registerProductCommand(safelinece.NewCommand())
	a.registerProductCommand(cloudwalker.NewCommand())
	a.registerProductCommand(ddr.NewCommand())
	a.registerProductCommand(tanswer.NewCommand())

	xrayCmd, err := xray.NewCommand()
	if err != nil {
		return nil, err
	}
	a.registerProductCommand(xrayCmd)

	safelineCmd := safeline.NewCommand()
	safeline.RegisterModules(safelineCmd)
	a.registerProductCommand(safelineCmd)

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
		if err := a.ensureRuntimeConfigLoaded(); err != nil {
			return err
		}

		switch cmd.Name() {
		case "safeline-ce":
			safelinece.ApplyRuntimeConfig(command, a.config, a.dryRun)
		case "cloudwalker":
			cloudwalker.ApplyRuntimeConfig(command, a.config)
		case "ddr":
			ddr.ApplyRuntimeConfig(command, a.config, a.configPath, a.dryRun)
		case "tanswer":
			tanswer.ApplyRuntimeConfig(command, a.config)
		case "xray":
			xray.ApplyRuntimeConfig(command, a.config, a.dryRun)
		case "safeline":
			safeline.ApplyRuntimeConfig(command, a.config)
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

func (a *app) ensureRuntimeConfigLoaded() error {
	if a.configLoaded {
		return nil
	}

	if err := config.LoadEnvFile(filepath.Join(".", ".env")); err != nil {
		return err
	}

	cfg, err := loadConfigFile(a.configPath)
	if err != nil {
		return err
	}

	a.config = cfg
	a.configLoaded = true
	return nil
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

func loadConfigFile(path string) (config.Raw, error) {
	return config.Load(path)
}

func defaultConfigPathFromCWD() string {
	return filepath.Join(".", defaultConfigPath)
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
