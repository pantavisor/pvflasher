package commands

import (
	"context"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"pvflasher/internal/flash"
)

var verifyCmd = &cobra.Command{
	Use:   "verify [device] [bmap-file]",
	Short: "Verify a device against a bmap file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		devicePath := args[0]
		bmapPath := args[1]

		bar := progressbar.DefaultBytes(
			-1,
			"verifying",
		)

		opts := flash.Options{
			DevicePath: devicePath,
			BmapPath:   bmapPath,
			ProgressCb: func(p flash.Progress) {
                if bar.GetMax() == -1 && p.BytesTotal > 0 {
                    bar.ChangeMax64(p.BytesTotal)
                }
                bar.Set64(p.BytesProcessed)
			},
		}

		v := flash.NewVerifier(opts)
		err := v.Verify(context.Background())
        bar.Finish()
        return err
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
