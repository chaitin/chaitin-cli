package safelinece

import (
	"bytes"
	"testing"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	got := stdout.String()
	want := "SafeLine CE CLI"
	if !bytes.Contains([]byte(got), []byte(want)) {
		t.Fatalf("unexpected output: got %q, want to contain %q", got, want)
	}
}
