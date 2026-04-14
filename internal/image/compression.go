package image

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

// IsCompressed returns true if the file path has a known compression extension
func IsCompressed(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".gz", ".bz2", ".xz", ".zst", ".zstd":
		return true
	default:
		return false
	}
}

// Magic byte signatures for compression formats.
var (
	magicGzip = []byte{0x1f, 0x8b}
	magicXZ   = []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}
	magicBZ2  = []byte{0x42, 0x5a}                   // "BZ"
	magicZstd = []byte{0x28, 0xb5, 0x2f, 0xfd}
)

// detectMagic peeks at the first bytes of r and returns the detected
// compression format ("gz", "xz", "bz2", "zstd") or "" if unrecognised.
// The returned reader replays the peeked bytes transparently.
func detectMagic(r io.Reader) (format string, buffered *bufio.Reader) {
	br := bufio.NewReaderSize(r, 8)
	peek, _ := br.Peek(6) // 6 bytes covers the longest magic (xz)

	switch {
	case len(peek) >= 2 && peek[0] == magicGzip[0] && peek[1] == magicGzip[1]:
		return "gz", br
	case len(peek) >= 6 && matchBytes(peek, magicXZ):
		return "xz", br
	case len(peek) >= 4 && matchBytes(peek, magicZstd):
		return "zstd", br
	case len(peek) >= 2 && peek[0] == magicBZ2[0] && peek[1] == magicBZ2[1]:
		return "bz2", br
	default:
		return "", br
	}
}

func matchBytes(data, magic []byte) bool {
	if len(data) < len(magic) {
		return false
	}
	for i, b := range magic {
		if data[i] != b {
			return false
		}
	}
	return true
}

// Decompressor returns a reader that decompresses the source if needed.
//
// Detection strategy:
//  1. The file extension is checked first as a hint.
//  2. The first few bytes (magic signature) are inspected to confirm.
//  3. If the extension says compressed but the magic doesn't match, we
//     trust the magic and return an appropriate decompressor (or the raw
//     reader if the content is uncompressed). A warning is logged.
func Decompressor(path string, r io.Reader) (io.Reader, error) {
	ext := strings.ToLower(filepath.Ext(path))

	// For extensions that don't suggest compression, return as-is.
	switch ext {
	case ".gz", ".bz2", ".xz", ".zst", ".zstd":
		// Fall through to magic-byte detection below.
	default:
		return r, nil
	}

	magic, br := detectMagic(r)

	// Strip leading dot for comparison: ".gz" → "gz", ".zstd" → "zstd".
	extName := strings.TrimPrefix(ext, ".")

	// Extension and magic agree — use the matching decompressor.
	if magic == extName || (extName == "zst" && magic == "zstd") || (extName == "zstd" && magic == "zstd") {
		return newDecompressor(magic, br)
	}

	// Magic detected a DIFFERENT compression than the extension claims.
	if magic != "" {
		fmt.Printf("WARNING: %s has %s extension but content is %s-compressed (detected by magic bytes); using %s decompressor\n",
			filepath.Base(path), ext, magic, magic)
		return newDecompressor(magic, br)
	}

	// Extension says compressed but content has no recognisable magic.
	// Likely an uncompressed file with a wrong extension.
	// Treat as raw/uncompressed so the caller can handle it.
	fmt.Printf("WARNING: %s has %s extension but content does not match any known compression format; treating as uncompressed\n",
		filepath.Base(path), ext)
	return br, nil
}

// newDecompressor creates a decompressor reader for the given format.
func newDecompressor(format string, r io.Reader) (io.Reader, error) {
	switch format {
	case "gz":
		return gzip.NewReader(r)
	case "bz2":
		return bzip2.NewReader(r), nil
	case "xz":
		return xz.NewReader(r)
	case "zstd":
		d, err := zstd.NewReader(r)
		if err != nil {
			return nil, err
		}
		return d, nil
	default:
		return nil, fmt.Errorf("unsupported compression format: %s", format)
	}
}
