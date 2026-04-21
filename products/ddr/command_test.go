package ddr

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

	if !bytes.Contains(stdout.Bytes(), []byte("DDR CLI")) {
		t.Fatalf("unexpected output: %q", stdout.String())
	}
}
