package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chaitin/workspace-cli/products/chaitin"
	"github.com/chaitin/workspace-cli/products/xray"
	"github.com/spf13/cobra"
)

type app struct {
	root             *cobra.Command
	aliasSubcommands map[string]struct{}
}

func newApp() *app {
	root := &cobra.Command{
		Use:           "cws",
		Short:         "CLI for Chaitin Tech products",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	a := &app{
		root:             root,
		aliasSubcommands: make(map[string]struct{}),
	}

	a.registerProductCommand(chaitin.NewCommand())

	xrayCmd, err := xray.NewCommand()
	if err != nil {
		log.Fatal(err)
	}
	a.registerProductCommand(xrayCmd)

	// TODO: register more products

	return a
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

	a.aliasSubcommands[cmd.Name()] = struct{}{}
	a.root.AddCommand(cmd)
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
	if err := newApp().execute(); err != nil {
		log.Fatal(err)
	}
}
