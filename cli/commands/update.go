package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"pvflasher/internal/update"
	"pvflasher/internal/version"
)

var (
	updateCheckOnly bool
	updateYes       bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Show the latest version and update pvflasher",
	Long: `Show the running and latest pvflasher versions and, when a newer release
exists, offer to install it.

The update is downloaded from GitHub and only installed if its signature
matches the pvflasher release key. Copies installed by a package manager
(deb, rpm, pacman) are not modified; update those with the package manager.`,
	Example: `  pvflasher update           # show versions, ask before installing
  pvflasher update --check   # only show versions
  pvflasher update --yes     # install without asking (scripts)`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		update.CleanupPrevious()
		st, err := checkStatus()
		if err != nil {
			return err
		}
		printStatus(st)
		rel := st.Update
		if rel == nil || updateCheckOnly {
			return nil
		}
		if !rel.Install.CanSelfUpdate {
			fmt.Printf("\nThis copy can't update itself (%s).\nUpdate it with your package manager or download it from https://github.com/pantavisor/pvflasher/releases/latest\n", rel.Install.Reason)
			return nil
		}
		if !updateYes {
			if !isTerminal(os.Stdin) {
				fmt.Println("\nRun 'pvflasher update --yes' to install it.")
				return nil
			}
			fmt.Printf("\nUpdate to v%s now? [y/N] ", rel.Version)
			answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			if a := strings.ToLower(strings.TrimSpace(answer)); a != "y" && a != "yes" {
				fmt.Println("Not updated.")
				return nil
			}
		}

		bar := progressbar.DefaultBytes(-1, "downloading")
		err = update.Apply(context.Background(), rel, func(done, total int64) {
			if bar.GetMax() == -1 && total > 0 {
				bar.ChangeMax64(total)
			}
			bar.Set64(done)
		})
		bar.Finish()
		if err != nil {
			return err
		}
		fmt.Printf("\nUpdated to v%s (%s).\n", rel.Version, rel.Install.Path)
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the running and latest pvflasher versions",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := checkStatus()
		if err != nil {
			// Offline is fine: still show what is running.
			fmt.Printf("Current version: %s\nLatest version:  unknown (%v)\n", version.Version, err)
			return nil
		}
		printStatus(st)
		if st.Update != nil {
			fmt.Println("\nRun 'pvflasher update' to install it.")
		}
		return nil
	},
}

func checkStatus() (*update.Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return update.Check(ctx, version.Version)
}

func printStatus(st *update.Status) {
	state := "up to date"
	switch {
	case st.DevBuild:
		state = "development build, updates are disabled"
	case st.Update != nil:
		state = "update available"
	}
	fmt.Printf("Current version: %s\n", st.Current)
	fmt.Printf("Latest version:  v%s (%s)\n", st.Latest, state)
	if st.Update != nil && st.Update.Notes != "" {
		fmt.Printf("\nWhat's new in v%s:\n%s\n", st.Update.Version, indent(st.Update.Notes))
	}
}

func indent(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = "  " + l
		}
	}
	return strings.Join(lines, "\n")
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "only show the current and latest versions")
	updateCmd.Flags().BoolVarP(&updateYes, "yes", "y", false, "install an available update without asking")
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(versionCmd)
}
