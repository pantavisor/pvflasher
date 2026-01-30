package image

import "os"

// OpenFile is a helper to open a regular file
func OpenFile(path string) (*os.File, error) {
    return os.Open(path)
}
