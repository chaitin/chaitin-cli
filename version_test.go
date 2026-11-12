package main

import (
	"runtime/debug"
	"strings"
	"testing"
)

func stubVersion(t *testing.T, injected string, info *debug.BuildInfo, ok bool) {
	t.Helper()
	origVersion := version
	origRead := readBuildInfo
	t.Cleanup(func() {
		version = origVersion
		readBuildInfo = origRead
	})
	version = injected
	readBuildInfo = func() (*debug.BuildInfo, bool) { return info, ok }
}

func runRootCommand(t *testing.T, args ...string) string {
	t.Helper()
	setTestHome(t, t.TempDir())
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp() error = %v", err)
	}
	var stdout, stderr strings.Builder
	app.root.SetOut(&stdout)
	app.root.SetErr(&stderr)
	app.root.SetArgs(args)
	if err := app.execute(); err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty so version output stays script-friendly on stdout", stderr.String())
	}
	return stdout.String()
}

func TestVersionCommandPrintsInjectedVersion(t *testing.T) {
	stubVersion(t, "v1.2.3", nil, false)

	got := runRootCommand(t, "version")
	if got != "v1.2.3\n" {
		t.Fatalf("version output = %q, want %q", got, "v1.2.3\n")
	}
}

func TestVersionCommandFallsBackToBuildInfoModuleVersion(t *testing.T) {
	stubVersion(t, "", &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}}, true)

	got := runRootCommand(t, "version")
	if got != "v9.9.9\n" {
		t.Fatalf("version output = %q, want %q", got, "v9.9.9\n")
	}
}

func TestVersionCommandFallsBackToDevWithoutBuildInfo(t *testing.T) {
	stubVersion(t, "", nil, false)

	got := runRootCommand(t, "version")
	if got != devVersion+"\n" {
		t.Fatalf("version output = %q, want %q", got, devVersion+"\n")
	}
}

func TestRootVersionFlagMatchesVersionCommand(t *testing.T) {
	stubVersion(t, "v1.2.3", nil, false)

	flagOutput := runRootCommand(t, "--version")
	commandOutput := runRootCommand(t, "version")
	if flagOutput != commandOutput {
		t.Fatalf("--version = %q, version = %q", flagOutput, commandOutput)
	}
}

func TestVersionFromBuildInfo(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{
			name: "module version without vcs revision",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v1.0.0"}},
			ok:   true,
			want: "v1.0.0",
		},
		{
			name: "stamped pseudo-version falls back to vcs revision",
			info: &debug.BuildInfo{
				Main:     debug.Module{Version: "v0.0.0-20260922074332-dc0b69e60f23"},
				Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "dc0b69e60f23628db60df018446d27c6053038d8"}},
			},
			ok:   true,
			want: "dev+dc0b69e60f23",
		},
		{
			name: "pseudo-version without vcs settings",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v0.0.0-20260920052045-2d4b5d8fd518"}},
			ok:   true,
			want: "dev+2d4b5d8fd518",
		},
		{
			name: "devel falls back to vcs revision",
			info: &debug.BuildInfo{
				Main:     debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "abcdef1234567890"}},
			},
			ok:   true,
			want: "dev+abcdef123456",
		},
		{
			name: "devel without settings",
			info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:   true,
			want: "dev",
		},
		{
			name: "empty build info",
			info: &debug.BuildInfo{},
			ok:   true,
			want: "dev",
		},
		{
			name: "build info unavailable",
			ok:   false,
			want: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := tt.info
			ok := tt.ok
			got := versionFromBuildInfo(func() (*debug.BuildInfo, bool) { return info, ok })
			if got != tt.want {
				t.Fatalf("versionFromBuildInfo() = %q, want %q", got, tt.want)
			}
		})
	}
}
