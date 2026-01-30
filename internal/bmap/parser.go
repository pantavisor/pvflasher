package bmap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parse reads a bmap XML from the reader and returns a Bmap struct
func Parse(r io.Reader) (*Bmap, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read bmap content: %w", err)
	}

	var b Bmap
	if err := xml.Unmarshal(content, &b); err != nil {
		return nil, fmt.Errorf("failed to decode bmap xml: %w", err)
	}
	b.Trim()

	// Verify integrity
	if err := verifyIntegrity(&b, content); err != nil {
		return nil, fmt.Errorf("bmap integrity check failed: %w", err)
	}

	return &b, nil
}

func verifyIntegrity(b *Bmap, content []byte) error {
	var expectedChecksum string
	var algo string

	if b.BmapFileChecksum != "" {
		expectedChecksum = b.BmapFileChecksum
		algo = b.ChecksumType
	} else if b.BmapFileSHA1 != "" {
		expectedChecksum = b.BmapFileSHA1
		algo = "sha1"
	}

	if expectedChecksum == "" {
		return nil // No checksum to verify
	}

	// Find the checksum in the content
	checksumBytes := []byte(expectedChecksum)
	pos := bytes.Index(content, checksumBytes)
	if pos == -1 {
		return fmt.Errorf("checksum string not found in bmap file")
	}

	// Replace with zeros
	tempContent := make([]byte, len(content))
	copy(tempContent, content)
	for i := 0; i < len(checksumBytes); i++ {
		tempContent[pos+i] = '0'
	}

	hasher, err := GetHasher(algo)
	if err != nil {
		return err
	}
	hasher.Write(tempContent)
	actualChecksum := fmt.Sprintf("%x", hasher.Sum(nil))

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// Trim removes leading and trailing whitespace from string fields
func (b *Bmap) Trim() {
	b.Version = strings.TrimSpace(b.Version)
	b.ChecksumType = strings.TrimSpace(b.ChecksumType)
	b.BmapFileChecksum = strings.TrimSpace(b.BmapFileChecksum)
	b.BmapFileSHA1 = strings.TrimSpace(b.BmapFileSHA1)
	for i := range b.BlockMap {
		b.BlockMap[i].Checksum = strings.TrimSpace(b.BlockMap[i].Checksum)
		b.BlockMap[i].Text = strings.TrimSpace(b.BlockMap[i].Text)
	}
}

// BlockRange represents a parsed range of blocks
type BlockRange struct {
	Start    int64
	End      int64 // Inclusive
	Count    int64
	Checksum string
}

// Parse converts the XML Range text into a BlockRange
func (r Range) Parse() (BlockRange, error) {
	text := strings.TrimSpace(r.Text)
	parts := strings.Split(text, "-")

	var start, end int64
	var err error

	if len(parts) == 1 {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return BlockRange{}, fmt.Errorf("invalid range start '%s': %w", parts[0], err)
		}
		end = start
	} else if len(parts) == 2 {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return BlockRange{}, fmt.Errorf("invalid range start '%s': %w", parts[0], err)
		}
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return BlockRange{}, fmt.Errorf("invalid range end '%s': %w", parts[1], err)
		}
	} else {
		return BlockRange{}, fmt.Errorf("invalid range format '%s'", text)
	}

	return BlockRange{
		Start:    start,
		End:      end,
		Count:    end - start + 1,
		Checksum: r.Checksum,
	}, nil
}
