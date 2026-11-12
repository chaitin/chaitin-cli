package main

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// version is overridden at build time with
// `-ldflags "-X main.version=v1.2.3"`. Local builds that do not inject it fall
// back to the metadata the Go toolchain records in the binary.
var version string

const devVersion = "dev"

// pseudoVersionSuffix matches the "<14-digit timestamp>-<12-character revision>"
// tail of a Go pseudo-version such as v0.0.0-20260920052045-2d4b5d8fd518.
var pseudoVersionSuffix = regexp.MustCompile(`-[0-9]{14}-([0-9a-f]{12})$`)

// readBuildInfo is a seam so tests can control the fallback path.
var readBuildInfo = debug.ReadBuildInfo

func resolveVersion() string {
	if injected := strings.TrimSpace(version); injected != "" {
		return injected
	}
	return versionFromBuildInfo(readBuildInfo)
}

func versionFromBuildInfo(read func() (*debug.BuildInfo, bool)) string {
	info, ok := read()
	if !ok || info == nil {
		return devVersion
	}
	// A recorded revision means the binary was built from a VCS checkout.
	if revision := vcsRevision(info); revision != "" {
		return devVersion + "+" + shortRevision(revision)
	}
	moduleVersion := strings.TrimSpace(info.Main.Version)
	// Without VCS settings the version still describes an untagged commit when
	// it is a pseudo-version, which is what `go install <module>@latest` records
	// (for example v0.0.0-20260920052045-2d4b5d8fd518). It reads like a release
	// version but is not one, so report its revision the same way as a local
	// build and only keep Main.Version when it names a real release.
	if revision := pseudoVersionRevision(moduleVersion); revision != "" {
		return devVersion + "+" + shortRevision(revision)
	}
	if moduleVersion != "" && moduleVersion != "(devel)" {
		return moduleVersion
	}
	return devVersion
}

func vcsRevision(info *debug.BuildInfo) string {
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return strings.TrimSpace(setting.Value)
		}
	}
	return ""
}

func pseudoVersionRevision(moduleVersion string) string {
	matches := pseudoVersionSuffix.FindStringSubmatch(moduleVersion)
	if matches == nil {
		return ""
	}
	return matches[1]
}

func shortRevision(revision string) string {
	const maxLength = 12
	if len(revision) > maxLength {
		return revision[:maxLength]
	}
	return revision
}

func newVersionCommand(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the chaitin-cli version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), root.Version)
			return nil
		},
	}
}
