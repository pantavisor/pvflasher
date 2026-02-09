package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"pvflasher/internal/flash"
	"pvflasher/internal/platform"
)

var bmapFile string
var force bool
var noVerify bool
var noEject bool
var jsonOutput bool

var copyCmd = &cobra.Command{
	Use:   "copy [image] [device]",
	Short: "Write an image to a device using bmap if available",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		imagePath := args[0]
		devicePath := args[1]

		// Only require root for block devices on Linux
		if !jsonOutput && !platform.IsRoot() {
			fi, err := os.Stat(devicePath)
			if err == nil {
				// If it's not a regular file, it might be a block device
				if !fi.Mode().IsRegular() {
					fmt.Println("This command requires root privileges for non-regular files. Attempting to relaunch with sudo...")
					return platform.RelaunchWithSudo()
				}
			} else if !os.IsNotExist(err) {
				// Some other error, better be safe
				return platform.RelaunchWithSudo()
			}
		}

		var bar *progressbar.ProgressBar
		if !jsonOutput {
			bar = progressbar.DefaultBytes(
				-1,
				"flashing",
			)
		}

		// Auto-discover bmap if not set
		if bmapFile == "" {
			candidates := []string{
				imagePath + ".bmap",
			}

			// If image has an extension like .gz, .bz2, etc, try removing it
			ext := filepath.Ext(imagePath)
			switch strings.ToLower(ext) {
			case ".gz", ".bz2", ".xz", ".zst", ".zstd", ".zip":
				base := strings.TrimSuffix(imagePath, ext)
				candidates = append(candidates, base+".bmap")
			}

			for _, c := range candidates {
				if _, err := os.Stat(c); err == nil {
					bmapFile = c
					if !jsonOutput {
						fmt.Println("Auto-detected bmap:", bmapFile)
					}
					break
				}
			}
		}

		opts := flash.Options{
			ImagePath:  imagePath,
			DevicePath: devicePath,
			BmapPath:   bmapFile,
			Force:      force,
			NoVerify:   noVerify,
			NoEject:    noEject,
			ProgressCb: func(p flash.Progress) {
				if jsonOutput {
					data, _ := json.Marshal(p)
					fmt.Println(string(data))
				} else {
					if bar != nil {
						if bar.GetMax() == -1 && p.BytesTotal > 0 {
							bar.ChangeMax64(p.BytesTotal)
						}
						bar.Describe(p.Phase)
						bar.Set64(p.BytesProcessed)
					}
				}
			},
		}

		f := flash.NewFlasher(opts)
		result, err := f.Flash(context.Background())
		if bar != nil {
			bar.Finish()
		}
		if err == nil && result != nil {
			if jsonOutput {
				data, _ := json.Marshal(result)
				fmt.Println(string(data))
			} else {
				fmt.Printf("\n✅ Flash completed successfully!\n")
				fmt.Printf("   Bytes written: %d (%.2f MB)\n", result.BytesWritten, float64(result.BytesWritten)/(1024*1024))
				fmt.Printf("   Duration: %.2fs\n", result.Duration.Seconds())
				fmt.Printf("   Average speed: %.2f MB/s\n", result.AverageSpeed/(1024*1024))
			}
		}
		return err
	},
}

func init() {
	copyCmd.Flags().StringVar(&bmapFile, "bmap", "", "path to .bmap file")
	copyCmd.Flags().BoolVar(&force, "force", false, "allow writing to mounted devices")
	copyCmd.Flags().BoolVar(&noVerify, "no-verify", false, "skip verification after flash")
	copyCmd.Flags().BoolVar(&noEject, "no-eject", false, "don't eject device after flash")
	copyCmd.Flags().BoolVar(&jsonOutput, "json", false, "output progress in JSON format")
	rootCmd.AddCommand(copyCmd)
}
