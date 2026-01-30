//go:build windows

package image

import (
	"os"
)

// GetMappedRanges returns the ranges of the file that contain data.
// On Windows, we fallback to treating the whole file as data for now.
func GetMappedRanges(f *os.File) ([]ByteRange, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return []ByteRange{{Start: 0, End: fi.Size()}}, nil
}
