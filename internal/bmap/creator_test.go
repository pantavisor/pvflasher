package bmap

import (
	"crypto/sha256"
	"hash"
	"os"
	"path/filepath"
	"testing"
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
	// Create a file with all zeros (sparse)
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "zero.img")

	// Write 1MB of zeros
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

	// All-zero file should have no mapped blocks
	if b.MappedBlocksCount != 0 {
		t.Errorf("MappedBlocksCount = %d, want 0", b.MappedBlocksCount)
	}
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

	// Should have 2 mapped blocks (first and third)
	if b.MappedBlocksCount != 2 {
		t.Errorf("MappedBlocksCount = %d, want 2", b.MappedBlocksCount)
	}
}

func TestIsAllZero(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "All zeros",
			data:     make([]byte, 100),
			expected: true,
		},
		{
			name:     "First byte non-zero",
			data:     append([]byte{1}, make([]byte, 99)...),
			expected: false,
		},
		{
			name:     "Last byte non-zero",
			data:     append(make([]byte, 99), 1),
			expected: false,
		},
		{
			name:     "Middle byte non-zero",
			data:     append(append(make([]byte, 50), byte(1)), make([]byte, 49)...),
			expected: false,
		},
		{
			name:     "Empty slice",
			data:     []byte{},
			expected: true,
		},
		{
			name:     "Single non-zero",
			data:     []byte{1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAllZero(tt.data)
			if result != tt.expected {
				t.Errorf("isAllZero() = %v, want %v", result, tt.expected)
			}
		})
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
