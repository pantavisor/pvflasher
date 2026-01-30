//go:build !windows

package image

import (
	"errors"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// GetMappedRanges returns the ranges of the file that contain data (non-holes)
func GetMappedRanges(f *os.File) ([]ByteRange, error) {
	var ranges []ByteRange
	offset := int64(0)

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := fi.Size()

	for offset < size {
		// Find next data
		dataStart, err := unix.Seek(int(f.Fd()), offset, unix.SEEK_DATA)
		if err != nil {
			// If ENXIO, no more data
			if errors.Is(err, syscall.ENXIO) {
				break
			}
			// Fallback for filesystems not supporting SEEK_DATA
			if err == syscall.EINVAL || err == syscall.ENOTSUP {
				return []ByteRange{{Start: 0, End: size}}, nil
			}
			return nil, err
		}

		// Find next hole
		holeStart, err := unix.Seek(int(f.Fd()), dataStart, unix.SEEK_HOLE)
		if err != nil {
			// Should not happen usually if SEEK_DATA succeeded
			holeStart = size
		}

		if holeStart > size {
			holeStart = size
		}

		ranges = append(ranges, ByteRange{Start: dataStart, End: holeStart})
		offset = holeStart
	}

	return ranges, nil
}
