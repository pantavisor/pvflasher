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

// RootCmd returns the root cobra command with all subcommands registered.
// This allows external projects to embed pvflasher commands into their
// own CLI trees.
func RootCmd() *cobra.Command {
	return rootCmd
}

// RegisterCommands adds all pvflasher subcommands (copy, list, verify,
// create, install) to a parent cobra command. This is the recommended
// way for external tools to integrate pvflasher capabilities.
//
// Example:
//
//	import pvcli "pvflasher/cli/commands"
//	flashCmd := &cobra.Command{Use: "flash", Short: "Flash tools"}
//	pvcli.RegisterCommands(flashCmd)
//	rootCmd.AddCommand(flashCmd)
func RegisterCommands(parent *cobra.Command) {
	parent.AddCommand(copyCmd)
	parent.AddCommand(listCmd)
	parent.AddCommand(verifyCmd)
	parent.AddCommand(createCmd)
	parent.AddCommand(installCmd)
	parent.AddCommand(downloadCmd)
}

func Execute() {
	// Attach to console for CLI output (Windows GUI apps need this)
	AttachConsoleIfNeeded()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
