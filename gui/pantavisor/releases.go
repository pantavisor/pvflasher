package pantavisor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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

type DeviceRelease struct {
	Name       string   `json:"name"`
	FullImage  Artifact `json:"full_image"`
	PVRExports Artifact `json:"pvrexports"`
	BSP        Artifact `json:"bsp"`
	SDK        Artifact `json:"sdk,omitempty"`
}

// ReleaseWrapper handles the inconsistent JSON structure where a release
// can be either a list of devices or an object containing a list of devices.
type ReleaseWrapper struct {
	Devices   []DeviceRelease
	Timestamp string
}

func (rw *ReleaseWrapper) UnmarshalJSON(data []byte) error {
	// Try unmarshalling as a list first
	var list []DeviceRelease
	if err := json.Unmarshal(data, &list); err == nil {
		rw.Devices = list
		return nil
	}

	// Try unmarshalling as an object
	var obj struct {
		Devices   []DeviceRelease `json:"devices"`
		Timestamp string          `json:"timestamp"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		rw.Devices = obj.Devices
		rw.Timestamp = obj.Timestamp
		return nil
	}

	return fmt.Errorf("failed to parse release data")
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

// GetChannels returns sorted list of channels
func (r Releases) GetChannels() []string {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// GetVersions returns sorted list of versions for a channel (descending)
func (r Releases) GetVersions(channel string) []string {
	versionsMap, ok := r[channel]
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(versionsMap))
	for k := range versionsMap {
		keys = append(keys, k)
	}
	// Sort descending (assuming versions are comparable strings, or just alphanumeric)
	// For better sorting, we might need semantic version parsing, but string sort is a start.
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j] // Descending
	})
	return keys
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
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := downloadWithValidation(url, destPath, expectedSHA256, progressCb, attempt)
		if err == nil {
			return nil
		}
		lastErr = err

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

func downloadWithValidation(url string, destPath string, expectedSHA256 string, progressCb func(DownloadProgress), attempt int) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Minute, // Long timeout for large files
	}

	resp, err := client.Get(url)
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
