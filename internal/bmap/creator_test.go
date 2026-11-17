package bmap

import (
	"crypto/sha256"
	"hash"
	"os"
	"path/filepath"
	"testing"

	"pvflasher/internal/image"
)

func TestCreate_Simple(t *testing.T) {
	// Create a temporary image file
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test.img")

	// Write test data (1MB)
	testData := make([]byte, 1024*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	os.WriteFile(imagePath, testData, 0644)

	opts := CreateOptions{
		ImageSize: int64(len(testData)),
		BlockSize: 4096,
	}

	b, err := Create(imagePath, opts)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if b.Version != "2.0" {
		t.Errorf("Version = %q, want %q", b.Version, "2.0")
	}

	if b.ImageSize != int64(len(testData)) {
		t.Errorf("ImageSize = %d, want %d", b.ImageSize, len(testData))
	}

	if b.BlockSize != 4096 {
		t.Errorf("BlockSize = %d, want 4096", b.BlockSize)
	}

	if b.ChecksumType != "sha256" {
		t.Errorf("ChecksumType = %q, want sha256", b.ChecksumType)
	}

	// Should have mapped blocks (non-zero data)
	if b.MappedBlocksCount == 0 {
		t.Error("MappedBlocksCount should be > 0")
	}
}

func TestCreate_ZeroFile(t *testing.T) {
	// A file written with zeros (not sparse): zeros are data, so every block
	// is mapped and a flash writes them over whatever the target held.
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "zero.img")

	testData := make([]byte, 1024*1024)
	os.WriteFile(imagePath, testData, 0644)

	opts := CreateOptions{
		ImageSize: 0, // Should auto-detect
		BlockSize: 0, // Should use default
	}

	b, err := Create(imagePath, opts)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if b.BlockSize != 4096 {
		t.Errorf("BlockSize = %d, want 4096 (default)", b.BlockSize)
	}

	if b.MappedBlocksCount != 256 {
		t.Errorf("MappedBlocksCount = %d, want 256", b.MappedBlocksCount)
	}
}

func TestCreate_SparseFile(t *testing.T) {
	// A file that is one big hole has nothing to write.
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "sparse.img")

	f, err := os.Create(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(1024 * 1024); err != nil {
		t.Fatal(err)
	}
	f.Close()

	b, err := Create(imagePath, CreateOptions{})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if !holesSupported(t, imagePath) {
		t.Skip("filesystem does not report holes")
	}
	if b.MappedBlocksCount != 0 {
		t.Errorf("MappedBlocksCount = %d, want 0", b.MappedBlocksCount)
	}
}

func TestCreate_DataHoleData(t *testing.T) {
	// data, hole, data: only the hole is left out.
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "holey.img")

	const blk = 4096
	const holeBlocks = 64 // larger than any filesystem's allocation unit
	f, err := os.Create(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	block := make([]byte, blk)
	for i := range block {
		block[i] = byte(i%255 + 1)
	}
	f.WriteAt(block, 0)
	f.WriteAt(block, (1+holeBlocks)*blk)
	f.Close()

	b, err := Create(imagePath, CreateOptions{BlockSize: blk})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if !holesSupported(t, imagePath) {
		t.Skip("filesystem does not report holes")
	}
	if b.MappedBlocksCount != 2 {
		t.Errorf("MappedBlocksCount = %d, want 2", b.MappedBlocksCount)
	}
	if len(b.BlockMap) != 2 || b.BlockMap[0].Text != "0" || b.BlockMap[1].Text != "65" {
		t.Errorf("BlockMap = %+v, want ranges 0 and 65", b.BlockMap)
	}
}

// holesSupported reports whether the filesystem holding path reports holes
// (tmpfs, ext4, xfs and btrfs do).
func holesSupported(t *testing.T, path string) bool {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ranges, err := image.GetMappedRanges(f)
	if err != nil {
		t.Fatal(err)
	}
	fi, _ := f.Stat()
	return !(len(ranges) == 1 && ranges[0].Start == 0 && ranges[0].End == fi.Size())
}

func TestCreate_NonExistentFile(t *testing.T) {
	opts := CreateOptions{}
	_, err := Create("/nonexistent/path/to/image.img", opts)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestCreate_MixedData(t *testing.T) {
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "mixed.img")

	// Create file with patterns: data, zeros, data
	data := make([]byte, 12288) // 3 blocks
	// First block: non-zero
	for i := 0; i < 4096; i++ {
		data[i] = byte(i % 256)
	}
	// Second block: zeros (leave as is)
	// Third block: non-zero
	for i := 8192; i < 12288; i++ {
		data[i] = byte(i % 256)
	}
	os.WriteFile(imagePath, data, 0644)

	opts := CreateOptions{
		ImageSize: int64(len(data)),
		BlockSize: 4096,
	}

	b, err := Create(imagePath, opts)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// The zero block in the middle is data too: one range of 3 blocks.
	if b.MappedBlocksCount != 3 {
		t.Errorf("MappedBlocksCount = %d, want 3", b.MappedBlocksCount)
	}
	if len(b.BlockMap) != 1 || b.BlockMap[0].Text != "0-2" {
		t.Errorf("BlockMap = %+v, want one range 0-2", b.BlockMap)
	}
}

func TestCreateRangeFromHasher(t *testing.T) {
	h := GetTestHasher()
	h.Write([]byte("test data"))

	tests := []struct {
		name            string
		start           int64
		end             int64
		wantText        string
		wantChecksumLen int
	}{
		{
			name:            "Single block",
			start:           5,
			end:             5,
			wantText:        "5",
			wantChecksumLen: 64, // SHA256 hex string
		},
		{
			name:            "Range of blocks",
			start:           10,
			end:             20,
			wantText:        "10-20",
			wantChecksumLen: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.Reset()
			h.Write([]byte("test data"))
			r := createRangeFromHasher(tt.start, tt.end, h)

			if r.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", r.Text, tt.wantText)
			}

			if len(r.Checksum) != tt.wantChecksumLen {
				t.Errorf("Checksum length = %d, want %d", len(r.Checksum), tt.wantChecksumLen)
			}
		})
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	bmapPath := filepath.Join(tmpDir, "test.bmap")

	b := &Bmap{
		Version:           "2.0",
		ImageSize:         8192,
		BlockSize:         4096,
		BlocksCount:       2,
		MappedBlocksCount: 1,
		ChecksumType:      "sha256",
		BlockMap: []Range{
			{Text: "0", Checksum: "abc123"},
		},
	}

	err := b.Save(bmapPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(bmapPath); os.IsNotExist(err) {
		t.Fatal("Bmap file was not created")
	}

	// Verify content can be parsed back
	f, err := os.Open(bmapPath)
	if err != nil {
		t.Fatalf("Failed to open saved bmap: %v", err)
	}
	defer f.Close()

	parsed, err := Parse(f)
	if err != nil {
		t.Fatalf("Failed to parse saved bmap: %v", err)
	}

	if parsed.Version != b.Version {
		t.Errorf("Parsed version = %q, want %q", parsed.Version, b.Version)
	}
	if parsed.ImageSize != b.ImageSize {
		t.Errorf("Parsed ImageSize = %d, want %d", parsed.ImageSize, b.ImageSize)
	}
}

func TestSave_InvalidPath(t *testing.T) {
	b := &Bmap{
		Version: "2.0",
	}

	err := b.Save("/nonexistent/path/test.bmap")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestParse_SavedBmap(t *testing.T) {
	// Create a bmap and save it, then parse it back
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test.img")

	// Create test image with some data
	data := make([]byte, 8192)
	for i := range data {
		data[i] = byte(i % 256)
	}
	os.WriteFile(imagePath, data, 0644)

	// Create bmap
	opts := CreateOptions{
		ImageSize: int64(len(data)),
		BlockSize: 4096,
	}
	b, err := Create(imagePath, opts)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Save and re-parse
	bmapPath := filepath.Join(tmpDir, "test.bmap")
	b.Save(bmapPath)

	f, _ := os.Open(bmapPath)
	defer f.Close()

	parsed, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if parsed.BlockSize != b.BlockSize {
		t.Errorf("BlockSize = %d, want %d", parsed.BlockSize, b.BlockSize)
	}
	if len(parsed.BlockMap) != len(b.BlockMap) {
		t.Errorf("BlockMap length = %d, want %d", len(parsed.BlockMap), len(b.BlockMap))
	}
}

// GetTestHasher returns a test SHA256 hasher
func GetTestHasher() *testHasher {
	return &testHasher{h: sha256.New()}
}

type testHasher struct {
	h hash.Hash
}

func (t *testHasher) Write(p []byte) (n int, err error) {
	return t.h.Write(p)
}

func (t *testHasher) Sum(b []byte) []byte {
	return t.h.Sum(b)
}

func (t *testHasher) Reset() {
	t.h.Reset()
}

// Import needed packages at the top
func init() {}
