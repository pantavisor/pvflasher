package bmap

import (
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"pvflasher/internal/image"
)

// CreateOptions configures the bmap generation
type CreateOptions struct {
	ImageSize int64
	BlockSize int
}

// Create generates a Bmap struct from an image file
func Create(imagePath string, opts CreateOptions) (*Bmap, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if opts.ImageSize == 0 {
		fi, err := f.Stat()
		if err != nil {
			return nil, err
		}
		opts.ImageSize = fi.Size()
	}

	if opts.BlockSize == 0 {
		opts.BlockSize = 4096 // Default 4KB
	}

	// Get mapped ranges (sparse detection)
	ranges, err := image.GetMappedRanges(f)
	if err != nil {
		return nil, fmt.Errorf("failed to detect mapped ranges: %w", err)
	}

	blocksCount := opts.ImageSize / int64(opts.BlockSize)
	if opts.ImageSize%int64(opts.BlockSize) != 0 {
		blocksCount++
	}

	bm := &Bmap{
		Version:      "2.0",
		ImageSize:    opts.ImageSize,
		BlockSize:    opts.BlockSize,
		BlocksCount:  blocksCount,
		ChecksumType: "sha256",
		BlockMap:     make([]Range, 0),
	}

	var mappedBlocksCount int64
	buf := make([]byte, opts.BlockSize)

	for _, byteRange := range ranges {
		firstBlock := byteRange.Start / int64(opts.BlockSize)
		lastBlock := (byteRange.End - 1) / int64(opts.BlockSize)

		if lastBlock >= blocksCount {
			lastBlock = blocksCount - 1
		}

		var currentRangeStart int64 = -1
		hasher := sha256.New()

		_, err = f.Seek(firstBlock*int64(opts.BlockSize), io.SeekStart)
		if err != nil {
			return nil, err
		}

		for blk := firstBlock; blk <= lastBlock; blk++ {
			n, err := io.ReadFull(f, buf)
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				return nil, err
			}
			if n == 0 {
				break
			}

			// Check if block is all zeros (to skip even if mapped by FS)
			if isAllZero(buf[:n]) {
				if currentRangeStart != -1 {
					bm.BlockMap = append(bm.BlockMap, createRangeFromHasher(currentRangeStart, blk-1, hasher))
					currentRangeStart = -1
					hasher.Reset()
				}
				continue
			}

			mappedBlocksCount++
			if currentRangeStart == -1 {
				currentRangeStart = blk
			}
			hasher.Write(buf[:n])
		}

		if currentRangeStart != -1 {
			bm.BlockMap = append(bm.BlockMap, createRangeFromHasher(currentRangeStart, lastBlock, hasher))
		}
	}

	bm.MappedBlocksCount = mappedBlocksCount
	return bm, nil
}

func isAllZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

func createRangeFromHasher(start, end int64, h io.Writer) Range {
	sum := fmt.Sprintf("%x", h.(interface{ Sum([]byte) []byte }).Sum(nil))
	r := Range{
		Checksum: sum,
	}
	if start == end {
		r.Text = fmt.Sprintf("%d", start)
	} else {
		r.Text = fmt.Sprintf("%d-%d", start, end)
	}
	return r
}

// Save writes the Bmap to a file
func (b *Bmap) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header
	f.WriteString("<?xml version=\"1.0\" ?>\n")
	
	// Comment with human readable info like bmaptool
	f.WriteString(fmt.Sprintf("<!-- Bmap for image %d bytes, mapped %d blocks -->\n", b.ImageSize, b.MappedBlocksCount))

	enc := xml.NewEncoder(f)
	enc.Indent("", "    ")
	if err := enc.Encode(b); err != nil {
		return err
	}
    f.WriteString("\n")
	return nil
}
