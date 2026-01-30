package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"pvflasher/internal/version"
)

var rootCmd = &cobra.Command{
	Use:           "pvflasher",
	Short:         "A cross-platform image pvflasher",
	Version:       version.Version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	// Attach to console for CLI output (Windows GUI apps need this)
	AttachConsoleIfNeeded()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
