package pantavisor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

const ReleasesURL = "https://pantavisor-ci.s3.amazonaws.com/meta-pantavisor/releases.json"

// GetCacheDir returns the path to the image cache directory
func GetCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	cacheDir := filepath.Join(homeDir, ".pvflasher", "images")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	return cacheDir, nil
}

// GetCachedImagePath returns the path where an image should be cached
func GetCachedImagePath(url string) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, filepath.Base(url)), nil
}

// ValidateCachedFile checks if a cached file exists and has the correct SHA256
func ValidateCachedFile(filePath string, expectedSHA256 string) bool {
	if expectedSHA256 == "" {
		// No SHA to validate, just check if file exists
		_, err := os.Stat(filePath)
		return err == nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false
	}

	actualSHA := hex.EncodeToString(hasher.Sum(nil))
	return actualSHA == expectedSHA256
}

type Artifact struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// Docs describes the documentation bundle for a release version.
type Docs struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	URL  string `json:"url"`
}

type DeviceRelease struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	FullImage   Artifact `json:"full_image"`
	PVRExports  Artifact `json:"pvrexports"`
	BSP         Artifact `json:"bsp"`
	SDK         Artifact `json:"sdk,omitempty"`
}

// ReleaseWrapper handles the JSON structure for a release version. A version
// can be either a bare list of devices or an object containing a list of
// devices alongside metadata such as docs and a timestamp.
//
// The devices list itself may contain a trailing marker entry that carries
// only a timestamp (e.g. {"timestamp": "..."}) instead of device data; such
// entries are extracted into Timestamp rather than treated as devices.
type ReleaseWrapper struct {
	Docs      *Docs
	Devices   []DeviceRelease
	Timestamp string
}

func (rw *ReleaseWrapper) UnmarshalJSON(data []byte) error {
	// Bare list form: [device, device, ...].
	if trimmed := bytes.TrimSpace(data); len(trimmed) > 0 && trimmed[0] == '[' {
		return rw.parseDevices(data)
	}

	// Object form: {docs, devices, timestamp/release-date}. A version may carry
	// only docs (no devices), so a missing/empty devices list is not an error.
	var obj struct {
		Docs        *Docs           `json:"docs"`
		Devices     json.RawMessage `json:"devices"`
		Timestamp   string          `json:"timestamp"`
		ReleaseDate string          `json:"release-date"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("failed to parse release: %w", err)
	}

	rw.Docs = obj.Docs
	rw.Timestamp = obj.Timestamp
	if rw.Timestamp == "" {
		rw.Timestamp = obj.ReleaseDate
	}

	if len(obj.Devices) == 0 || string(obj.Devices) == "null" {
		return nil
	}
	return rw.parseDevices(obj.Devices)
}

// parseDevices parses a JSON array of device entries. Entries that carry only a
// timestamp marker (no device name) are folded into rw.Timestamp; entries
// without a name are otherwise skipped.
func (rw *ReleaseWrapper) parseDevices(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse release devices: %w", err)
	}

	for _, item := range raw {
		var marker struct {
			Name      string `json:"name"`
			Timestamp string `json:"timestamp"`
		}
		if err := json.Unmarshal(item, &marker); err == nil && marker.Name == "" {
			if marker.Timestamp != "" && rw.Timestamp == "" {
				rw.Timestamp = marker.Timestamp
			}
			// Nameless entry (timestamp marker or otherwise) — not a device.
			continue
		}

		var dev DeviceRelease
		if err := json.Unmarshal(item, &dev); err != nil {
			return fmt.Errorf("failed to parse device: %w", err)
		}
		rw.Devices = append(rw.Devices, dev)
	}

	return nil
}

// Releases maps Channel -> Version -> Release Info
type Releases map[string]map[string]ReleaseWrapper

func FetchReleases() (Releases, error) {
	resp, err := http.Get(ReleasesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch releases: status %d", resp.StatusCode)
	}

	var releases Releases
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// Title returns the board's official display name, falling back to its
// machine name for releases that predate display names.
func (d DeviceRelease) Title() string {
	if d.DisplayName != "" {
		return d.DisplayName
	}
	return d.Name
}

// channelOrder and channelLabels match the naming on pantavisor.io/downloads.
var (
	channelOrder  = []string{"stable", "release-candidate"}
	channelLabels = map[string]string{
		"stable":            "Stable (Recommended)",
		"release-candidate": "Release Candidate",
	}
)

// ChannelLabel returns the user-facing name of a channel.
func ChannelLabel(channel string) string {
	if l, ok := channelLabels[channel]; ok {
		return l
	}
	return channel
}

// GetChannels returns the channels, known ones first in website order.
func (r Releases) GetChannels() []string {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	rank := func(c string) int {
		if i := slices.Index(channelOrder, c); i >= 0 {
			return i
		}
		return len(channelOrder)
	}
	sort.Slice(keys, func(i, j int) bool {
		if ri, rj := rank(keys[i]), rank(keys[j]); ri != rj {
			return ri < rj
		}
		return keys[i] < keys[j]
	})
	return keys
}

// GetVersions returns the versions of a channel that have devices, newest first.
func (r Releases) GetVersions(channel string) []string {
	versionsMap, ok := r[channel]
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(versionsMap))
	for k, v := range versionsMap {
		if len(v.Devices) > 0 {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return compareVersions(keys[i], keys[j]) > 0 // Descending
	})
	return keys
}

var versionChunk = regexp.MustCompile(`\d+|\D+`)

// compareVersions orders versions like "028-rc9" < "028-rc10" < "028-rc10.1"
// by comparing digit runs numerically and everything else lexically.
func compareVersions(a, b string) int {
	ca, cb := versionChunk.FindAllString(a, -1), versionChunk.FindAllString(b, -1)
	for i := 0; i < len(ca) && i < len(cb); i++ {
		na, errA := strconv.Atoi(ca[i])
		nb, errB := strconv.Atoi(cb[i])
		if errA == nil && errB == nil {
			if na != nb {
				return na - nb
			}
			continue
		}
		if c := strings.Compare(ca[i], cb[i]); c != 0 {
			return c
		}
	}
	return len(ca) - len(cb)
}

type DownloadProgress struct {
	Total      int64
	Downloaded int64
	Percentage float64
	Phase      string // "downloading" or "validating"
	Speed      float64
}

// DownloadFile downloads a file from URL to destPath with progress reporting
func DownloadFile(url string, destPath string, progressCb func(DownloadProgress)) error {
	return DownloadFileWithSHA(url, destPath, "", progressCb)
}

// DownloadFileWithSHA downloads a file and validates its SHA256 checksum
func DownloadFileWithSHA(url string, destPath string, expectedSHA256 string, progressCb func(DownloadProgress)) error {
	return DownloadFileWithSHAContext(context.Background(), url, destPath, expectedSHA256, progressCb)
}

// DownloadFileWithSHAContext is DownloadFileWithSHA with cancellation.
func DownloadFileWithSHAContext(ctx context.Context, url string, destPath string, expectedSHA256 string, progressCb func(DownloadProgress)) error {
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := downloadWithValidation(ctx, url, destPath, expectedSHA256, progressCb, attempt)
		if err == nil {
			return nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// If it's a validation error, don't retry
		if _, ok := err.(*SHA256MismatchError); ok {
			return err
		}

		// Wait before retrying (exponential backoff)
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
	}

	return fmt.Errorf("download failed after %d attempts: %w", maxRetries, lastErr)
}

// SHA256MismatchError indicates the downloaded file's checksum doesn't match
type SHA256MismatchError struct {
	Expected string
	Actual   string
}

func (e *SHA256MismatchError) Error() string {
	return fmt.Sprintf("SHA256 mismatch: expected %s, got %s", e.Expected, e.Actual)
}

func downloadWithValidation(ctx context.Context, url string, destPath string, expectedSHA256 string, progressCb func(DownloadProgress), attempt int) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Minute, // Long timeout for large files
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %s", resp.Status)
	}

	// Create temp file for download
	tmpPath := destPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	size := resp.ContentLength
	hasher := sha256.New()

	// Create a multi-writer to write to both file and hasher
	multiWriter := io.MultiWriter(out, hasher)

	// Create a proxy reader to track progress
	reader := &ProgressReader{
		Reader:    resp.Body,
		Total:     size,
		Phase:     "downloading",
		Cb:        progressCb,
		StartTime: time.Now(),
	}

	_, err = io.Copy(multiWriter, reader)
	out.Close() // Close before rename

	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("download interrupted: %w", err)
	}

	// Validate SHA256 if provided
	if expectedSHA256 != "" {
		if progressCb != nil {
			progressCb(DownloadProgress{
				Total:      size,
				Downloaded: size,
				Percentage: 100,
				Phase:      "validating",
			})
		}

		actualSHA := hex.EncodeToString(hasher.Sum(nil))
		if actualSHA != expectedSHA256 {
			os.Remove(tmpPath)
			return &SHA256MismatchError{Expected: expectedSHA256, Actual: actualSHA}
		}
	}

	// Rename temp file to final destination
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to finalize download: %w", err)
	}

	return nil
}

type ProgressReader struct {
	io.Reader
	Total      int64
	Downloaded int64
	Phase      string
	Cb         func(DownloadProgress)
	StartTime  time.Time
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Downloaded += int64(n)

	if pr.Cb != nil {
		percent := 0.0
		if pr.Total > 0 {
			percent = float64(pr.Downloaded) / float64(pr.Total) * 100
		}

		speed := 0.0
		elapsed := time.Since(pr.StartTime).Seconds()
		if elapsed > 0 {
			speed = float64(pr.Downloaded) / elapsed
		}

		pr.Cb(DownloadProgress{
			Total:      pr.Total,
			Downloaded: pr.Downloaded,
			Percentage: percent,
			Phase:      pr.Phase,
			Speed:      speed,
		})
	}

	return n, err
}
