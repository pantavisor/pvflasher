package commands

import (
	"context"
	"fmt"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"pvflasher/internal/update"
	"pvflasher/internal/version"
)

var updateCheckOnly bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update pvflasher to the latest release",
	Long: `Check for a newer pvflasher release and install it.

The update is downloaded from GitHub and only installed if its signature
matches the pvflasher release key. Copies installed by a package manager
(deb, rpm, pacman) are not modified; update those with the package manager.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		update.CleanupPrevious()
		rel, err := update.Check(context.Background(), version.Version)
		if err != nil {
			return err
		}
		if rel == nil {
			fmt.Printf("pvflasher %s is the latest version.\n", version.Version)
			return nil
		}
		fmt.Printf("pvflasher v%s is available (you have %s).\n", rel.Version, version.Version)
		if updateCheckOnly {
			return nil
		}
		if !rel.Install.CanSelfUpdate {
			return fmt.Errorf("this copy can't update itself (%s); update it with your package manager or download it from https://github.com/pantavisor/pvflasher/releases/latest", rel.Install.Reason)
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
		fmt.Printf("Updated to v%s: %s\n", rel.Version, rel.Install.Path)
		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "only check whether an update is available")
	rootCmd.AddCommand(updateCmd)
}
