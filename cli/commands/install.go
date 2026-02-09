package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"pvflasher/gui/pantavisor"
	"pvflasher/internal/device"
	"pvflasher/internal/flash"
	"pvflasher/internal/platform"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Interactive installer for Pantavisor releases",
	Long:  `Downloads and flashes an official Pantavisor release image to a target device.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Fetch Releases
		fmt.Println("Fetching Pantavisor releases...")
		releases, err := pantavisor.FetchReleases()
		if err != nil {
			return fmt.Errorf("failed to fetch releases: %w", err)
		}

		reader := bufio.NewReader(os.Stdin)

		// 2. Select Channel
		channels := releases.GetChannels()
		if len(channels) == 0 {
			return fmt.Errorf("no release channels found")
		}

		fmt.Println("\nAvailable Channels:")
		for i, ch := range channels {
			fmt.Printf("%d) %s\n", i+1, ch)
		}

		chIdx := promptInt(reader, "Select Channel", 1, len(channels))
		selectedChannel := channels[chIdx-1]

		// 3. Select Version
		versions := releases.GetVersions(selectedChannel)
		if len(versions) == 0 {
			return fmt.Errorf("no versions found for channel %s", selectedChannel)
		}

		fmt.Println("\nAvailable Versions:")
		for i, v := range versions {
			fmt.Printf("%d) %s\n", i+1, v)
		}

		vIdx := promptInt(reader, "Select Version", 1, len(versions))
		selectedVersion := versions[vIdx-1]

		// 4. Select Device/Board
		releaseWrapper := releases[selectedChannel][selectedVersion]
		devices := releaseWrapper.Devices
		// Filter out entries with empty names
		filtered := devices[:0]
		for _, d := range devices {
			if d.Name != "" {
				filtered = append(filtered, d)
			}
		}
		devices = filtered
		if len(devices) == 0 {
			return fmt.Errorf("no devices found for version %s", selectedVersion)
		}

		fmt.Println("\nAvailable Devices:")
		for i, d := range devices {
			fmt.Printf("%d) %s\n", i+1, d.Name)
		}

		dIdx := promptInt(reader, "Select Device", 1, len(devices))
		selectedReleaseDevice := devices[dIdx-1]

		fmt.Printf("\nSelected: %s / %s / %s\n", selectedChannel, selectedVersion, selectedReleaseDevice.Name)

		// 5. Check Cache / Download
		imageURL := selectedReleaseDevice.FullImage.URL
		expectedSHA := selectedReleaseDevice.FullImage.SHA256

		if imageURL == "" {
			return fmt.Errorf("selected release does not have a full image url")
		}

		cachePath, err := pantavisor.GetCachedImagePath(imageURL)
		if err != nil {
			return fmt.Errorf("failed to get cache path: %w", err)
		}

		isValid := pantavisor.ValidateCachedFile(cachePath, expectedSHA)
		if isValid {
			fmt.Printf("Using cached image: %s\n", cachePath)
		} else {
			fmt.Printf("Downloading image to: %s\n", cachePath)
			bar := progressbar.NewOptions64(
				-1,
				progressbar.OptionSetDescription("downloading"),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionSetWidth(30),
				progressbar.OptionThrottle(65*time.Millisecond),
				progressbar.OptionSpinnerType(14),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer:        "=",
					SaucerHead:    ">",
					SaucerPadding: " ",
					BarStart:      "[",
					BarEnd:        "]",
				}),
			)

			var maxSet bool
			err := pantavisor.DownloadFileWithSHA(imageURL, cachePath, expectedSHA, func(p pantavisor.DownloadProgress) {
				if !maxSet && p.Total > 0 {
					bar.ChangeMax64(p.Total)
					maxSet = true
				}
				if p.Phase == "validating" {
					bar.Describe("validating")
				} else {
					downloadedMB := float64(p.Downloaded) / (1024 * 1024)
					speedMBs := p.Speed / (1024 * 1024)
					if p.Total > 0 {
						totalMB := float64(p.Total) / (1024 * 1024)
						bar.Describe(fmt.Sprintf("downloading %.1f MB / %.1f MB | %.1f MB/s", downloadedMB, totalMB, speedMBs))
					} else {
						bar.Describe(fmt.Sprintf("downloading %.1f MB | %.1f MB/s", downloadedMB, speedMBs))
					}
				}
				_ = bar.Set64(p.Downloaded)
			})
			fmt.Println() // Newline after bar
			if err != nil {
				return fmt.Errorf("download failed: %w", err)
			}
			fmt.Println("Download complete and verified.")
		}

		// 6. Select Target Drive
		// Use root privileges if needed (Linux)
		if !platform.IsRoot() {
			fmt.Println("\nScanning for drives (may require root privileges for flashing)...")
		}

		mgr := device.NewManager()
		targetDevs, err := mgr.List()
		if err != nil {
			return fmt.Errorf("failed to list devices: %w", err)
		}

		if len(targetDevs) == 0 {
			return fmt.Errorf("no target devices found. Please insert a USB drive or SD card")
		}

		fmt.Println("\nAvailable Target Drives:")
		for i, d := range targetDevs {
			removable := ""
			if d.Removable {
				removable = "(Removable)"
			}
			sizeGB := float64(d.Size) / (1024 * 1024 * 1024)
			fmt.Printf("%d) %s [%.2f GB] %s %s - %s\n", i+1, d.Name, sizeGB, d.Vendor, d.Model, removable)
		}

		tIdx := promptInt(reader, "Select Target Drive", 1, len(targetDevs))
		targetDevice := targetDevs[tIdx-1]

		fmt.Printf("\nWARNING: All data on %s (%s) will be erased!\n", targetDevice.Name, targetDevice.Model)
		if !promptConfirm(reader, "Are you sure you want to continue?") {
			fmt.Println("Operation cancelled.")
			return nil
		}

		// 7. Flash
		// Use absolute path for cachePath to be safe across context switches
		absCachePath, err := filepath.Abs(cachePath)
		if err != nil {
			return err
		}

		if !platform.IsRoot() {
			fmt.Println("\nPrivileges required for flashing. Requesting elevation...")

			// Construct arguments
			flashArgs := []string{"copy", absCachePath, targetDevice.Name}
			if force {
				flashArgs = append(flashArgs, "--force")
			}
			if noVerify {
				flashArgs = append(flashArgs, "--no-verify")
			}
			if noEject {
				flashArgs = append(flashArgs, "--no-eject")
			}

			elevator := platform.NewElevator()
			cmd, err := elevator.ElevateCommand(flashArgs...)
			if err != nil {
				return fmt.Errorf("failed to prepare elevated command: %w", err)
			}

			// Pass through stdio so the user can interact (if sudo asks for password)
			// and see the output of the copy command.
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("flash command failed: %w", err)
			}
			return nil
		}

		fmt.Printf("\nFlashing %s to %s...\n", cachePath, targetDevice.Name)

		bar := progressbar.DefaultBytes(
			-1,
			"flashing",
		)

		opts := flash.Options{
			ImagePath:  cachePath,
			DevicePath: targetDevice.Name,
			Force:      force,
			NoVerify:   noVerify,
			NoEject:    noEject,
			ProgressCb: func(p flash.Progress) {
				if bar.GetMax() == -1 && p.BytesTotal > 0 {
					bar.ChangeMax64(p.BytesTotal)
				}
				bar.Describe(p.Phase)
				bar.Set64(p.BytesProcessed)
			},
		}

		// Auto-detect bmap
		bmapPath := cachePath + ".bmap"
		if _, err := os.Stat(bmapPath); err == nil {
			opts.BmapPath = bmapPath
			fmt.Println("Auto-detected bmap:", bmapPath)
		}

		f := flash.NewFlasher(opts)
		result, err := f.Flash(context.Background())
		bar.Finish()
		fmt.Println()

		if err != nil {
			return fmt.Errorf("flash failed: %w", err)
		}

		if result != nil {
			fmt.Printf("\n✅ Flash completed successfully!\n")
			fmt.Printf("   Bytes written: %d (%.2f MB)\n", result.BytesWritten, float64(result.BytesWritten)/(1024*1024))
			fmt.Printf("   Duration: %.2fs\n", result.Duration.Seconds())
			fmt.Printf("   Average speed: %.2f MB/s\n", result.AverageSpeed/(1024*1024))
		}

		return nil
	},
}

func promptInt(r *bufio.Reader, label string, min, max int) int {
	for {
		fmt.Printf("%s [%d-%d]: ", label, min, max)
		input, _ := r.ReadString('\n')
		input = strings.TrimSpace(input)
		val, err := strconv.Atoi(input)
		if err == nil && val >= min && val <= max {
			return val
		}
		fmt.Println("Invalid input, please try again.")
	}
}

func promptConfirm(r *bufio.Reader, label string) bool {
	fmt.Printf("%s [y/N]: ", label)
	input, _ := r.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func init() {
	installCmd.Flags().BoolVar(&force, "force", false, "allow writing to mounted devices")
	installCmd.Flags().BoolVar(&noVerify, "no-verify", false, "skip verification after flash")
	installCmd.Flags().BoolVar(&noEject, "no-eject", false, "don't eject device after flash")

	rootCmd.AddCommand(installCmd)
}
