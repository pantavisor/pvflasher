package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsArchive(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"image.tar", true},
		{"image.tgz", true},
		{"image.tar.gz", true},
		{"image.TAR.GZ", true},
		{"image.Tar.Gz", true},
		{"image.img", false},
		{"image.iso", false},
		{"image.zip", false},
		{"image.txt", false},
		{"image.gz", false},        // Just .gz, not .tar.gz
		{"archive.tar.bz2", false}, // Not supported
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := IsArchive(tt.path)
			if result != tt.expected {
				t.Errorf("IsArchive(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func createTestTarGz(t *testing.T, files map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("Failed to write header: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write content: %v", err)
		}
	}

	return archivePath
}

func TestGetArchivePair_ImageOnly(t *testing.T) {
	files := map[string]string{
		"image.img": "this is image content",
	}
	archivePath := createTestTarGz(t, files)

	pair, err := GetArchivePair(archivePath)
	if err != nil {
		t.Fatalf("GetArchivePair failed: %v", err)
	}

	if pair.ImageEntry != "image.img" {
		t.Errorf("ImageEntry = %q, want %q", pair.ImageEntry, "image.img")
	}
	if pair.BmapEntry != "" {
		t.Errorf("BmapEntry = %q, want empty", pair.BmapEntry)
	}
	if pair.Bmap != nil {
		t.Error("Bmap should be nil")
	}
}

func TestGetArchivePair_WithBmap(t *testing.T) {
	// Create a valid bmap XML template
	templateBmap := `<?xml version="1.0" ?>
<bmap version="2.0">
	<ImageSize> 4096 </ImageSize>
	<BlockSize> 4096 </BlockSize>
	<BlocksCount> 1 </BlocksCount>
	<MappedBlocksCount> 1 </MappedBlocksCount>
	<ChecksumType> sha256 </ChecksumType>
	<BmapFileChecksum> PLACEHOLDER </BmapFileChecksum>
	<BlockMap>
		<Range chksum="abc"> 0 </Range>
	</BlockMap>
</bmap>`

	// Calculate valid checksum for the template
	zeroed := strings.Replace(templateBmap, "PLACEHOLDER", strings.Repeat("0", 64), 1)
	h := sha256.New()
	h.Write([]byte(zeroed))
	validChecksum := fmt.Sprintf("%x", h.Sum(nil))
	bmapContent := strings.Replace(templateBmap, "PLACEHOLDER", validChecksum, 1)

	files := map[string]string{
		"image.wic":      "this is image content",
		"image.wic.bmap": bmapContent,
	}
	archivePath := createTestTarGz(t, files)

	pair, err := GetArchivePair(archivePath)
	if err != nil {
		t.Fatalf("GetArchivePair failed: %v", err)
	}

	if pair.ImageEntry != "image.wic" {
		t.Errorf("ImageEntry = %q, want %q", pair.ImageEntry, "image.wic")
	}
	if pair.BmapEntry != "image.wic.bmap" {
		t.Errorf("BmapEntry = %q, want %q", pair.BmapEntry, "image.wic.bmap")
	}
	if pair.Bmap == nil {
		t.Error("Bmap should not be nil")
	}
}

func TestGetArchivePair_CompressedImage(t *testing.T) {
	files := map[string]string{
		"image.img.gz": "this is compressed image content",
	}
	archivePath := createTestTarGz(t, files)

	pair, err := GetArchivePair(archivePath)
	if err != nil {
		t.Fatalf("GetArchivePair failed: %v", err)
	}

	if pair.ImageEntry != "image.img.gz" {
		t.Errorf("ImageEntry = %q, want %q", pair.ImageEntry, "image.img.gz")
	}
}

func TestGetArchivePair_MultipleImages(t *testing.T) {
	// When multiple images exist, it should return the first one with a bmap
	files := map[string]string{
		"image1.img": "image1 content",
		"image2.img": "image2 content",
	}
	archivePath := createTestTarGz(t, files)

	pair, err := GetArchivePair(archivePath)
	if err != nil {
		t.Fatalf("GetArchivePair failed: %v", err)
	}

	// Should return the first image found
	if pair.ImageEntry == "" {
		t.Error("ImageEntry should not be empty")
	}
}

func TestGetArchivePair_NoImage(t *testing.T) {
	files := map[string]string{
		"readme.txt": "this is not an image",
	}
	archivePath := createTestTarGz(t, files)

	_, err := GetArchivePair(archivePath)
	if err == nil {
		t.Error("Expected error for archive without image")
	}
}

func TestGetArchivePair_NonExistentFile(t *testing.T) {
	_, err := GetArchivePair("/nonexistent/path/to/archive.tar.gz")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestGetArchivePair_ISOImage(t *testing.T) {
	files := map[string]string{
		"image.iso": "this is iso content",
	}
	archivePath := createTestTarGz(t, files)

	pair, err := GetArchivePair(archivePath)
	if err != nil {
		t.Fatalf("GetArchivePair failed: %v", err)
	}

	if pair.ImageEntry != "image.iso" {
		t.Errorf("ImageEntry = %q, want %q", pair.ImageEntry, "image.iso")
	}
}

func TestExtract(t *testing.T) {
	files := map[string]string{
		"image.img": "this is image content for extraction",
	}
	archivePath := createTestTarGz(t, files)

	imagePath, bmapPath, cleanup, err := Extract(archivePath)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	defer cleanup()

	// Verify image was extracted
	content, err := os.ReadFile(imagePath)
	if err != nil {
		t.Fatalf("Failed to read extracted image: %v", err)
	}
	if string(content) != files["image.img"] {
		t.Errorf("Extracted content mismatch: got %q, want %q", string(content), files["image.img"])
	}

	// bmapPath should be empty for this test
	if bmapPath != "" {
		t.Errorf("BmapPath = %q, want empty", bmapPath)
	}

	// Verify the file exists in temp directory
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Error("Extracted file does not exist")
	}
}

func TestExtract_WithBmap(t *testing.T) {
	// Create a valid bmap XML template
	templateBmap := `<?xml version="1.0" ?>
<bmap version="2.0">
	<ImageSize> 4096 </ImageSize>
	<BlockSize> 4096 </BlockSize>
	<BlocksCount> 1 </BlocksCount>
	<MappedBlocksCount> 1 </MappedBlocksCount>
	<ChecksumType> sha256 </ChecksumType>
	<BmapFileChecksum> PLACEHOLDER </BmapFileChecksum>
	<BlockMap>
		<Range chksum="abc"> 0 </Range>
	</BlockMap>
</bmap>`

	// Calculate valid checksum for the template
	zeroed := strings.Replace(templateBmap, "PLACEHOLDER", strings.Repeat("0", 64), 1)
	h := sha256.New()
	h.Write([]byte(zeroed))
	validChecksum := fmt.Sprintf("%x", h.Sum(nil))
	bmapContent := strings.Replace(templateBmap, "PLACEHOLDER", validChecksum, 1)

	files := map[string]string{
		"image.wic":      "this is wic content",
		"image.wic.bmap": bmapContent,
	}
	archivePath := createTestTarGz(t, files)

	imagePath, bmapPath, cleanup, err := Extract(archivePath)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	defer cleanup()

	// Verify both files were extracted
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Error("Image file does not exist")
	}
	if _, err := os.Stat(bmapPath); os.IsNotExist(err) {
		t.Error("Bmap file does not exist")
	}

	// Verify bmap content
	content, err := os.ReadFile(bmapPath)
	if err != nil {
		t.Fatalf("Failed to read extracted bmap: %v", err)
	}
	if string(content) != bmapContent {
		t.Errorf("Bmap content mismatch")
	}
}

func TestExtract_InvalidArchive(t *testing.T) {
	_, _, _, err := Extract("/nonexistent/archive.tar.gz")
	if err == nil {
		t.Error("Expected error for invalid archive")
	}
}

func TestReadCloserWrapper(t *testing.T) {
	// Create mock closers
	var closed []string
	closer1 := &mockCloser{
		name:   "closer1",
		closed: &closed,
	}
	closer2 := &mockCloser{
		name:   "closer2",
		closed: &closed,
	}

	wrapper := &readCloserWrapper{
		Reader:  bytes.NewReader([]byte("test")),
		closers: []io.Closer{closer1, closer2},
	}

	// Close should close both closers in order (FIFO)
	err := wrapper.Close()
	if err != nil {
		t.Errorf("Close error: %v", err)
	}

	if len(closed) != 2 {
		t.Errorf("Expected 2 closers to be closed, got %d", len(closed))
	}

	// Should be closed in order (FIFO)
	if closed[0] != "closer1" {
		t.Errorf("First closed should be closer1, got %s", closed[0])
	}
	if closed[1] != "closer2" {
		t.Errorf("Second closed should be closer2, got %s", closed[1])
	}
}

type mockCloser struct {
	name   string
	closed *[]string
	err    error
}

func (m *mockCloser) Close() error {
	*m.closed = append(*m.closed, m.name)
	return m.err
}

func TestOpenArchiveImage(t *testing.T) {
	files := map[string]string{
		"image.img": "this is image content",
	}
	archivePath := createTestTarGz(t, files)

	reader, size, err := OpenArchiveImage(archivePath, "image.img")
	if err != nil {
		t.Fatalf("OpenArchiveImage failed: %v", err)
	}
	defer reader.Close()

	// Check size
	expectedSize := int64(len(files["image.img"]))
	if size != expectedSize {
		t.Errorf("Size = %d, want %d", size, expectedSize)
	}

	// Read content
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if string(content) != files["image.img"] {
		t.Errorf("Content mismatch: got %q, want %q", string(content), files["image.img"])
	}
}

func TestOpenArchiveImage_NotFound(t *testing.T) {
	files := map[string]string{
		"image.img": "this is image content",
	}
	archivePath := createTestTarGz(t, files)

	_, _, err := OpenArchiveImage(archivePath, "nonexistent.img")
	if err == nil {
		t.Error("Expected error for non-existent entry")
	}
}

func TestOpenArchiveImage_InvalidArchive(t *testing.T) {
	_, _, err := OpenArchiveImage("/nonexistent/archive.tar.gz", "image.img")
	if err == nil {
		t.Error("Expected error for invalid archive")
	}
}
