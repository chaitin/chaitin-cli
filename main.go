package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/chaitin/chaitin-cli/config"
	"github.com/chaitin/chaitin-cli/products/agentcompose"
	"github.com/chaitin/chaitin-cli/products/apisec"
	"github.com/chaitin/chaitin-cli/products/chaitin"
	"github.com/chaitin/chaitin-cli/products/cloudatlas"
	"github.com/chaitin/chaitin-cli/products/cloudwalker"
	"github.com/chaitin/chaitin-cli/products/codeforce"
	"github.com/chaitin/chaitin-cli/products/codeinsight"
	"github.com/chaitin/chaitin-cli/products/cosmos"
	"github.com/chaitin/chaitin-cli/products/ddr"
	"github.com/chaitin/chaitin-cli/products/dsensor"
	"github.com/chaitin/chaitin-cli/products/insight"
	"github.com/chaitin/chaitin-cli/products/monkeyscan"
	"github.com/chaitin/chaitin-cli/products/safeline"
	safelinece "github.com/chaitin/chaitin-cli/products/safeline-ce"
	"github.com/chaitin/chaitin-cli/products/safeline3"
	"github.com/chaitin/chaitin-cli/products/tanswer"
	"github.com/chaitin/chaitin-cli/products/veinmind"
	"github.com/chaitin/chaitin-cli/products/xray"
	"github.com/spf13/cobra"
)

type app struct {
	root             *cobra.Command
	aliasSubcommands map[string]struct{}
	productNames     map[string]struct{}
	config           config.Raw
	configPath       string
	configLoaded     bool
	dryRun           bool
}

const (
	defaultConfigDir  = ".chaitin-cli"
	defaultConfigFile = "config.yaml"
)

func newApp() (*app, error) {
	root := &cobra.Command{
		Use:           "chaitin-cli",
		Short:         "CLI for Chaitin Tech products",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.Version = resolveVersion()
	root.SetVersionTemplate("{{.Version}}\n")

	a := &app{
		root:             root,
		aliasSubcommands: make(map[string]struct{}),
		productNames:     make(map[string]struct{}),
		configPath:       defaultConfigPath(),
	}

	root.AddCommand(newVersionCommand(root))

	root.PersistentFlags().StringVarP(&a.configPath, "config", "c", a.configPath, "Config file path; when unset, CLI tries recognized ./config.yaml before ~/.chaitin-cli/config.yaml")
	root.PersistentFlags().BoolVar(&a.dryRun, "dry-run", false, "Do not send requests; commands that support dry-run print a request summary")

	a.registerProductCommand(chaitin.NewCommand())
	a.registerProductCommand(agentcompose.NewCommand())
	a.registerProductCommand(cloudatlas.NewCommand())
	a.registerProductCommand(apisec.NewCommand())
	a.registerProductCommand(safelinece.NewCommand())
	a.registerProductCommand(safeline3.NewCommand())
	a.registerProductCommand(cloudwalker.NewCommand())
	a.registerProductCommand(codeforce.NewCommand())
	a.registerProductCommand(codeinsight.NewCommand())
	a.registerProductCommand(ddr.NewCommand())
	a.registerProductCommand(dsensor.NewCommand())
	a.registerProductCommand(monkeyscan.NewCommand())
	a.registerProductCommand(tanswer.NewCommand())
	a.registerProductCommand(veinmind.NewCommand())
	a.registerProductCommand(cosmos.NewCommand())
	a.registerProductCommand(insight.NewCommand())

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
	a.productNames[cmd.Name()] = struct{}{}
	a.root.AddCommand(cmd)
}

func (a *app) wrapProductCommand(cmd *cobra.Command) {
	oldPreRun := cmd.PersistentPreRun
	oldPreRunE := cmd.PersistentPreRunE

	cmd.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		if err := a.ensureRuntimeConfigLoaded(); err != nil {
			if cmd.Name() == "agent-compose" {
				return agentcompose.RuntimeConfigError(command, err)
			}
			return err
		}

		switch cmd.Name() {
		case "agent-compose":
			agentcompose.ApplyRuntimeConfig(command, a.config, a.writeConfigPath(), a.dryRun)
		case "cloudAtlas":
			cloudatlas.ApplyRuntimeConfig(command, a.config, a.dryRun)
		case "apisec":
			apisec.ApplyRuntimeConfig(command, a.config, a.dryRun)
		case "safeline-ce":
			safelinece.ApplyRuntimeConfig(command, a.config, a.dryRun)
		case "safeline-3":
			safeline3.ApplyRuntimeConfig(command, a.config, a.dryRun)
		case "cloudwalker":
			cloudwalker.ApplyRuntimeConfig(command, a.config)
		case "codeforce":
			codeforce.ApplyRuntimeConfig(command, a.config, a.dryRun)
		case "codeinsight":
			codeinsight.ApplyRuntimeConfig(command, a.config, a.dryRun)
		case "ddr":
			ddr.ApplyRuntimeConfig(command, a.config, a.writeConfigPath(), a.dryRun)
		case "dsensor":
			dsensor.ApplyRuntimeConfig(command, a.config, a.dryRun)
		case "monkeyscan":
			monkeyscan.ApplyRuntimeConfig(command, a.config, a.writeConfigPath(), a.dryRun)
		case "tanswer":
			tanswer.ApplyRuntimeConfig(command, a.config)
		case "veinmind":
			veinmind.ApplyRuntimeConfig(command, a.config, a.dryRun)
		case "cosmos":
			cosmos.ApplyRuntimeConfig(command, a.config, a.dryRun)
		case "insight":
			insight.ApplyRuntimeConfig(command, a.config, a.dryRun)
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

	cfg, err := a.loadRuntimeConfig()
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

func (a *app) loadRuntimeConfig() (config.Raw, error) {
	if a.configFlagChanged() {
		return loadConfigFile(a.configPath)
	}

	localCfg, localRecognized, err := a.loadLocalConfigIfRecognized()
	if err == nil && localRecognized {
		return localCfg, nil
	}

	globalCfg, err := loadConfigFile(defaultConfigPath())
	if err != nil {
		return nil, err
	}
	return globalCfg, nil
}

func (a *app) writeConfigPath() string {
	if a.configFlagChanged() {
		return a.configPath
	}
	if _, ok, err := a.loadLocalConfigIfRecognized(); err == nil && ok {
		return localConfigPath()
	}
	return defaultConfigPath()
}

func (a *app) loadLocalConfigIfRecognized() (config.Raw, bool, error) {
	localCfg, err := loadConfigFile(localConfigPath())
	if err != nil {
		return nil, false, err
	}
	return localCfg, looksLikeCLIConfig(localCfg, a.productNames), nil
}

func (a *app) configFlagChanged() bool {
	if a.root == nil {
		return a.configPath != "" && a.configPath != defaultConfigPath()
	}
	flag := a.root.PersistentFlags().Lookup("config")
	return flag != nil && flag.Changed
}

func looksLikeCLIConfig(cfg config.Raw, productNames map[string]struct{}) bool {
	for name := range cfg {
		if _, ok := productNames[name]; ok {
			return true
		}
	}
	return false
}

func localConfigPath() string {
	return filepath.Join(".", defaultConfigFile)
}

func defaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return localConfigPath()
	}
	return filepath.Join(homeDir, defaultConfigDir, defaultConfigFile)
}

func runMain() int {
	app, err := newApp()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	app.root.SetContext(ctx)
	if err := app.execute(); err != nil {
		if renderer, ok := err.(interface{ RenderError(io.Writer) error }); ok {
			_ = renderer.RenderError(app.root.ErrOrStderr())
		} else {
			fmt.Fprintln(app.root.ErrOrStderr(), "Error:", err)
		}
		if coder, ok := err.(interface{ ExitCode() int }); ok {
			return coder.ExitCode()
		}
		return 1
	}
	return 0
}

func main() { os.Exit(runMain()) }
