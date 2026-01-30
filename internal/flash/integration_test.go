package flash_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"pvflasher/gui/pantavisor"
	"pvflasher/internal/flash"
)

func TestDownloadAndFlashIntegration(t *testing.T) {
	// 1. Create a dummy image
	imageSize := int64(2 * 1024 * 1024) // 2MB
	imageData := make([]byte, imageSize)
	for i := range imageData {
		imageData[i] = byte(i % 256)
	}

	hasher := sha256.New()
	hasher.Write(imageData)
	imageSHA := hex.EncodeToString(hasher.Sum(nil))

	// 2. Start mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", imageSize))
		w.WriteHeader(http.StatusOK)
		w.Write(imageData)
	}))
	defer server.Close()

	// 3. Download image
	tmpDir, err := os.MkdirTemp("", "pvflasher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	imagePath := filepath.Join(tmpDir, "downloaded.img")
	err = pantavisor.DownloadFileWithSHA(server.URL, imagePath, imageSHA, nil)
	if err != nil {
		t.Fatalf("failed to download image: %v", err)
	}

	// 4. Flash to target file
	targetPath := filepath.Join(tmpDir, "target.img")
	// Create target file first
	targetFile, err := os.Create(targetPath)
	if err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}
	targetFile.Truncate(imageSize + 1024*1024) // 1MB padding
	targetFile.Close()

	opts := flash.Options{
		ImagePath:  imagePath,
		DevicePath: targetPath,
		Force:      true,
		NoVerify:   false,
	}

	flasher := flash.NewFlasher(opts)
	result, err := flasher.Flash(context.Background())
	if err != nil {
		t.Fatalf("flash failed: %v", err)
	}

	if result.BytesWritten != imageSize {
		t.Errorf("expected %d bytes written, got %d", imageSize, result.BytesWritten)
	}

	// 5. Verify content
	flashedData, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read flashed data: %v", err)
	}

	// Compare only the first imageSize bytes
	for i := int64(0); i < imageSize; i++ {
		if flashedData[i] != imageData[i] {
			t.Fatalf("data mismatch at byte %d: expected %v, got %v", i, imageData[i], flashedData[i])
		}
	}

	t.Logf("Integration test passed! Average speed: %.2f MB/s", result.AverageSpeed/(1024*1024))
}
