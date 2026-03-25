package commands

import (
	"testing"
)

func TestRootCmd(t *testing.T) {
	// Test that root command exists and has the right properties
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	if rootCmd.Use != "pvflasher" {
		t.Errorf("Use = %q, want %q", rootCmd.Use, "pvflasher")
	}

	if rootCmd.SilenceUsage != true {
		t.Error("SilenceUsage should be true")
	}

	if rootCmd.SilenceErrors != true {
		t.Error("SilenceErrors should be true")
	}
}

func TestRootCmd_Version(t *testing.T) {
	// Version should be set (not empty)
	if rootCmd.Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestRootCmd_Short(t *testing.T) {
	// Short description should be set
	if rootCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}
