package main

import (
	"bytes"
	"testing"

	"github.com/Goooler/gh-changelog/pkg/cmd"
)

func TestRootCommandVersion(t *testing.T) {
	var stdout bytes.Buffer
	rootCmd := cmd.NewRootCmd("v1.2.3")
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"--version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	const want = "gh-changelog version v1.2.3\n"
	if got := stdout.String(); got != want {
		t.Fatalf("version output = %q, want %q", got, want)
	}
}
