package archive

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pvflasher/internal/bmap"
	"pvflasher/internal/image"
)

// IsArchive checks if the path has a tar archive extension
func IsArchive(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".tar" || ext == ".tgz" {
		return true
	}
	if ext == ".gz" {
		// Check if it is .tar.gz
		if strings.HasSuffix(strings.ToLower(path), ".tar.gz") {
			return true
		}
	}
	return false
}

// ArchivePair contains the found image entry and its optional bmap
type ArchivePair struct {
	ImageEntry string
	BmapEntry  string
	Bmap       *bmap.Bmap
}

// GetArchivePair scans the archive for a compatible image and bmap pair
func GetArchivePair(path string) (*ArchivePair, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := image.Decompressor(path, f)
	if err != nil {
		return nil, err
	}

	if c, ok := r.(io.Closer); ok {
		defer c.Close()
	}

	tr := tar.NewReader(r)

	type bmapInfo struct {
		bm       *bmap.Bmap
		filename string
	}

	bmaps := make(map[string]bmapInfo)
	images := make(map[string]string)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		name := header.Name
		baseName := filepath.Base(name)
		lowerBase := strings.ToLower(baseName)

		if strings.HasSuffix(lowerBase, ".bmap") {
			// Read bmap content
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, tr); err != nil {
				continue
			}
			bm, err := bmap.Parse(buf)
			if err == nil {
				// image.wic.bmap -> image.wic
				key := strings.TrimSuffix(baseName, ".bmap")
				bmaps[key] = bmapInfo{bm: bm, filename: name}
			}
		} else {
			// Check for valid image patterns: *.img, *.iso, or *.wic (possibly with compression)
			ext := filepath.Ext(baseName)
			lowerExt := strings.ToLower(ext)

			var key string
			matched := false

			// Supported compression extensions
			isComp := lowerExt == ".gz" || lowerExt == ".bz2" || lowerExt == ".xz" || lowerExt == ".zst" || lowerExt == ".zstd"

			if isComp {
				withoutComp := strings.TrimSuffix(baseName, ext)
				subExt := filepath.Ext(withoutComp)
				lowerSubExt := strings.ToLower(subExt)
				if lowerSubExt == ".img" || lowerSubExt == ".wic" || lowerSubExt == ".iso" {
					matched = true
					key = withoutComp
				}
			} else if lowerExt == ".img" || lowerExt == ".wic" || lowerExt == ".iso" {
				matched = true
				key = baseName
			}

			if matched {
				images[key] = name
			}
		}
	}

	// Find match
	for key, imgEntry := range images {
		if info, ok := bmaps[key]; ok {
			return &ArchivePair{
				ImageEntry: imgEntry,
				BmapEntry:  info.filename,
				Bmap:       info.bm,
			}, nil
		}
	}

	// Fallback: Return first image found if no bmap pair
	for _, imgEntry := range images {
		return &ArchivePair{
			ImageEntry: imgEntry,
			Bmap:       nil,
		}, nil
	}

	return nil, errors.New("no suitable image found in archive")
}

// Extract extracts the image and bmap (if present) from the archive to a temporary directory.
// Returns the paths to the extracted image and bmap, a cleanup function, and any error.
func Extract(archivePath string) (imagePath string, bmapPath string, cleanup func(), err error) {
	pair, err := GetArchivePair(archivePath)
	if err != nil {
		return "", "", nil, err
	}

	tempDir, err := os.MkdirTemp("", "pvflasher-extract-*")
	if err != nil {
		return "", "", nil, err
	}
	cleanup = func() { os.RemoveAll(tempDir) }

	// Re-open archive for extraction
	f, err := os.Open(archivePath)
	if err != nil {
		cleanup()
		return "", "", nil, err
	}
	defer f.Close()

	r, err := image.Decompressor(archivePath, f)
	if err != nil {
		cleanup()
		return "", "", nil, err
	}
	if c, ok := r.(io.Closer); ok {
		defer c.Close()
	}

	tr := tar.NewReader(r)

	foundImage := false
	foundBmap := false

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			cleanup()
			return "", "", nil, err
		}

		if header.Name == pair.ImageEntry {
			destPath := filepath.Join(tempDir, filepath.Base(header.Name))
			outFile, err := os.Create(destPath)
			if err != nil {
				cleanup()
				return "", "", nil, err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				cleanup()
				return "", "", nil, err
			}
			outFile.Close()
			imagePath = destPath
			foundImage = true
		} else if pair.BmapEntry != "" && header.Name == pair.BmapEntry {
			destPath := filepath.Join(tempDir, filepath.Base(header.Name))
			outFile, err := os.Create(destPath)
			if err != nil {
				cleanup()
				return "", "", nil, err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				cleanup()
				return "", "", nil, err
			}
			outFile.Close()
			bmapPath = destPath
			foundBmap = true
		}

		if foundImage && (pair.BmapEntry == "" || foundBmap) {
			break
		}
	}

	if !foundImage {
		cleanup()
		return "", "", nil, errors.New("image entry not found during extraction")
	}

	return imagePath, bmapPath, cleanup, nil
}

type readCloserWrapper struct {
	io.Reader
	closers []io.Closer
}

func (r *readCloserWrapper) Close() error {
	var firstErr error
	for _, c := range r.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// OpenArchiveImage returns a reader for the specific entry in the archive
func OpenArchiveImage(archivePath, entryName string) (io.ReadCloser, int64, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, 0, err
	}

	outerR, err := image.Decompressor(archivePath, f)
	if err != nil {
		f.Close()
		return nil, 0, err
	}

	tr := tar.NewReader(outerR)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			f.Close()
			if c, ok := outerR.(io.Closer); ok {
				c.Close()
			}
			return nil, 0, err
		}

		if header.Name == entryName {
			closers := []io.Closer{f}
			if c, ok := outerR.(io.Closer); ok {
				closers = append([]io.Closer{c}, closers...)
			}

			return &readCloserWrapper{
				Reader: tr,
				closers: closers,
			}, header.Size, nil
		}
	}

	f.Close()
	if c, ok := outerR.(io.Closer); ok {
		c.Close()
	}
	return nil, 0, os.ErrNotExist
}
