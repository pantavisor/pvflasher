package image

import (
	"compress/bzip2"
	"compress/gzip"
	"io"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

// Decompressor returns a reader that decompresses the source if needed
func Decompressor(path string, r io.Reader) (io.Reader, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".gz":
		return gzip.NewReader(r)
	case ".bz2":
		return bzip2.NewReader(r), nil
	case ".xz":
		return xz.NewReader(r)
	case ".zst", ".zstd":
		d, err := zstd.NewReader(r)
		if err != nil {
			return nil, err
		}
		return d, nil
	default:
		return r, nil
	}
}
