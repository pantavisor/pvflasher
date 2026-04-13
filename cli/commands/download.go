package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"pvflasher/gui/pantavisor"
)

var outputPath string

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a Pantavisor release image",
	Long:  `Downloads an official Pantavisor release image without flashing it.`,
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

		// 5. Download
		imageURL := selectedReleaseDevice.FullImage.URL
		expectedSHA := selectedReleaseDevice.FullImage.SHA256

		if imageURL == "" {
			return fmt.Errorf("selected release does not have a full image url")
		}

		// Determine destination path
		destPath := outputPath
		if destPath == "" {
			// Default: use cache
			destPath, err = pantavisor.GetCachedImagePath(imageURL)
			if err != nil {
				return fmt.Errorf("failed to get cache path: %w", err)
			}
		} else {
			// If output is a directory, append the filename from the URL
			info, statErr := os.Stat(destPath)
			if statErr == nil && info.IsDir() {
				urlBase := filepath.Base(imageURL)
				destPath = filepath.Join(destPath, urlBase)
			}
		}

		isValid := pantavisor.ValidateCachedFile(destPath, expectedSHA)
		if isValid {
			fmt.Printf("Image already exists and is valid: %s\n", destPath)
			return nil
		}

		fmt.Printf("Downloading image to: %s\n", destPath)
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
		err = pantavisor.DownloadFileWithSHA(imageURL, destPath, expectedSHA, func(p pantavisor.DownloadProgress) {
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
		fmt.Println()
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		fmt.Printf("Download complete and verified: %s\n", destPath)
		return nil
	},
}

func init() {
	downloadCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output path (file or directory; defaults to cache)")
	rootCmd.AddCommand(downloadCmd)
}
