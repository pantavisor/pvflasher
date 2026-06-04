package pantavisor

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDownloadFromParsedReleases exercises the full path: fetch live releases,
// parse them with the updated schema, then download a real artifact using the
// parsed URL+SHA and confirm checksum validation passes.
//
// Skipped in -short mode since it pulls a real image over the network.
func TestDownloadFromParsedReleases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network download in -short mode")
	}

	releases, err := FetchReleases()
	if err != nil {
		t.Fatalf("FetchReleases: %v", err)
	}

	// Pick docker-x86_64-scarthgap from stable/024 (smallest, ~155MB).
	rw, ok := releases["stable"]["024"]
	if !ok {
		t.Fatal("stable/024 not present in releases")
	}

	var dev *DeviceRelease
	for i := range rw.Devices {
		if rw.Devices[i].Name == "docker-x86_64-scarthgap" {
			dev = &rw.Devices[i]
			break
		}
	}
	if dev == nil {
		t.Fatal("docker-x86_64-scarthgap device not found")
	}
	if dev.FullImage.URL == "" || dev.FullImage.SHA256 == "" {
		t.Fatalf("parsed device missing url/sha: %+v", dev.FullImage)
	}
	t.Logf("downloading %s (sha %s)", dev.FullImage.URL, dev.FullImage.SHA256)

	dest := filepath.Join(t.TempDir(), filepath.Base(dev.FullImage.URL))

	var lastPct float64
	err = DownloadFileWithSHA(dev.FullImage.URL, dest, dev.FullImage.SHA256, func(p DownloadProgress) {
		if p.Percentage-lastPct >= 25 {
			lastPct = p.Percentage
			t.Logf("%s %.0f%%", p.Phase, p.Percentage)
		}
	})
	if err != nil {
		t.Fatalf("DownloadFileWithSHA failed (SHA mismatch or network): %v", err)
	}

	info, statErr := os.Stat(dest)
	if statErr != nil {
		t.Fatalf("downloaded file missing: %v", statErr)
	}
	t.Logf("downloaded and SHA256-verified OK: %d bytes at %s", info.Size(), dest)
}
