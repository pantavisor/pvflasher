package flash

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"pvflasher/internal/archive"
	"pvflasher/internal/bmap"
	"pvflasher/internal/device"
	"pvflasher/internal/image"
	"pvflasher/internal/platform"
)

type Flasher struct {
	opts Options
}

// normalizeDevicePath normalizes a device path for comparison.
// On Windows, removes \\.\  prefix and converts to uppercase.
// On Unix, returns the path as-is.
func normalizeDevicePath(path string) string {
	// Remove Windows device path prefix
	path = strings.TrimPrefix(path, `\\.\`)
	// Normalize case for Windows
	return strings.ToUpper(path)
}

func NewFlasher(opts Options) *Flasher {
	return &Flasher{opts: opts}
}

func (f *Flasher) Flash(ctx context.Context) (*FlashResult, error) {
	// 0. Safety Check: Is it mounted?
	if !f.opts.Force {
		mgr := device.NewManager()
		devs, err := mgr.List()
		if err == nil {
			for _, d := range devs {
				if normalizeDevicePath(d.Name) == normalizeDevicePath(f.opts.DevicePath) && len(d.MountPoints) > 0 {
					return nil, fmt.Errorf("device %s is mounted at %v; use force to override", d.Name, d.MountPoints)
				}
			}
		}
	}

	// 1. Prepare and Open Device
	// Dismount volumes before raw device access (critical on Windows)
	if err := platform.PrepareDevice(f.opts.DevicePath); err != nil {
		return nil, fmt.Errorf("failed to prepare device: %w", err)
	}

	// Open device for writing
	dev, err := platform.OpenDevice(f.opts.DevicePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %w", err)
	}
	defer dev.Close()

	// 2. Open Image & 3. Load Bmap
	var imgReader io.Reader
	var bm *bmap.Bmap
	var sourceSize int64
	var counter *image.CountingReader

	if archive.IsArchive(f.opts.ImagePath) {
		f.reportPhase("extracting")
		extractedImage, extractedBmap, cleanExtract, err := archive.Extract(f.opts.ImagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract archive: %w", err)
		}
		defer cleanExtract()

		f.opts.ImagePath = extractedImage
		// Use extracted bmap if not explicitly provided
		if f.opts.BmapPath == "" && extractedBmap != "" {
			f.opts.BmapPath = extractedBmap
		}
	}

	imgFile, err := os.Open(f.opts.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer imgFile.Close()

	fi, err := imgFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat image: %w", err)
	}
	sourceSize = fi.Size()
	counter = &image.CountingReader{Reader: imgFile}

	imgReader, err = image.Decompressor(f.opts.ImagePath, counter)
	if err != nil {
		return nil, fmt.Errorf("failed to create decompressor: %w", err)
	}

	// If BmapPath was explicitly provided, it overrides archive bmap
	if f.opts.BmapPath != "" {
		bmapFile, err := os.Open(f.opts.BmapPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open bmap: %w", err)
		}
		defer bmapFile.Close()
		bm, err = bmap.Parse(bmapFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bmap: %w", err)
		}
	}

	// Wrap in ForwardSeeker
	seeker := image.NewForwardSeeker(imgReader)

	// 4. Flash Loop
	startTime := time.Now()
	var totalBytes int64
	var writtenBytes int64

	bufSize := 1024 * 1024 // 1MB buffer
	buf := make([]byte, bufSize)

	if bm != nil {
		// Bmap-optimized copy
		totalBytes = bm.MappedBlocksCount * int64(bm.BlockSize)
		// If last block is mapped but partial, adjust totalBytes for progress reporting
		lastBlockIdx := bm.BlocksCount - 1
		for _, rng := range bm.BlockMap {
			pr, _ := rng.Parse()
			if pr.End == lastBlockIdx {
				padding := (bm.BlocksCount * int64(bm.BlockSize)) - bm.ImageSize
				totalBytes -= padding
				break
			}
		}

		for _, rng := range bm.BlockMap {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			parsedRange, err := rng.Parse()
			if err != nil {
				return nil, err
			}

			startByte := parsedRange.Start * int64(bm.BlockSize)
			endByte := (parsedRange.End + 1) * int64(bm.BlockSize)

			// Cap endByte to actual ImageSize to handle partial last block
			if endByte > bm.ImageSize {
				endByte = bm.ImageSize
			}

			countByte := endByte - startByte

			// Seek to start in image
			_, err = seeker.Seek(startByte, io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("failed to seek to block %d: %w", parsedRange.Start, err)
			}

			// Seek device to start
			_, err = dev.Seek(startByte, io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("failed to seek device to %d: %w", startByte, err)
			}

			// Copy loop for this range
			remaining := countByte
			for remaining > 0 {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}

				toRead := int64(len(buf))
				if remaining < toRead {
					toRead = remaining
				}

				n, err := io.ReadFull(seeker, buf[:toRead])
				if err != nil {
					return nil, fmt.Errorf("read error at block %d: %w", parsedRange.Start, err)
				}

				// Write all bytes, handling partial writes
				written := 0
				for written < n {
					w, werr := dev.Write(buf[written:n])
					if werr != nil {
						return nil, fmt.Errorf("write error at block %d: %w", parsedRange.Start, werr)
					}
					if w == 0 {
						return nil, fmt.Errorf("write returned 0 bytes at block %d", parsedRange.Start)
					}
					written += w
				}

				remaining -= int64(n)
				writtenBytes += int64(n)
				f.reportProgress(writtenBytes, totalBytes, counter.Count, sourceSize, startTime)
			}
		}
	} else {
		// Raw copy (Full image)
		totalBytes = sourceSize // Fallback

		for {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			n, err := seeker.Read(buf)
			if n > 0 {
				// Write all bytes, handling partial writes
				written := 0
				for written < n {
					w, werr := dev.Write(buf[written:n])
					if werr != nil {
						return nil, fmt.Errorf("write error at offset %d: %w", writtenBytes+int64(written), werr)
					}
					if w == 0 {
						return nil, fmt.Errorf("write returned 0 bytes at offset %d", writtenBytes+int64(written))
					}
					written += w
				}
				writtenBytes += int64(n)
				f.reportProgress(writtenBytes, totalBytes, counter.Count, sourceSize, startTime)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("read error: %w", err)
			}
		}
	}

	// 5. Sync
	f.reportPhase("syncing")
	if err := dev.Sync(); err != nil {
		return nil, fmt.Errorf("failed to sync device: %w", err)
	}

	// 6. Verification
	verificationDone := false
	if !f.opts.NoVerify {
		f.reportPhase("verifying")
		// Close current dev handle to allow exclusive access for verifier
		dev.Close()

		v := NewVerifier(f.opts)
		if bm != nil {
			v.SetBmap(bm)
		}

		if err := v.Verify(ctx); err != nil {
			return nil, fmt.Errorf("verification failed: %w", err)
		}
		verificationDone = true
	} else {
		dev.Close()
	}

	// 7. Eject
	deviceEjected := false
	if !f.opts.NoEject {
		f.reportPhase("ejecting")
		if err := platform.EjectDevice(f.opts.DevicePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to eject device: %v\n", err)
		} else {
			deviceEjected = true
		}
	}

	// Calculate final statistics
	duration := time.Since(startTime)
	avgSpeed := float64(writtenBytes) / duration.Seconds()

	// Calculate blocks written (for bmap mode)
	var blocksWritten int64
	if bm != nil {
		blocksWritten = bm.MappedBlocksCount
	} else {
		// For raw copy, calculate approximate blocks
		blockSize := int64(4096) // Standard block size
		blocksWritten = (writtenBytes + blockSize - 1) / blockSize
	}

	result := &FlashResult{
		BytesWritten:     writtenBytes,
		BlocksWritten:    blocksWritten,
		Duration:         duration,
		AverageSpeed:     avgSpeed,
		UsedBmap:         bm != nil,
		VerificationDone: verificationDone,
		DeviceEjected:    deviceEjected,
	}

	return result, nil
}

func (f *Flasher) reportPhase(phase string) {
	if f.opts.ProgressCb != nil {
		f.opts.ProgressCb(Progress{
			Phase:      phase,
			Percentage: 100,
		})
	}
}

func (f *Flasher) reportProgress(written, total, sourceRead, sourceTotal int64, start time.Time) {
	if f.opts.ProgressCb != nil {
		elapsed := time.Since(start).Seconds()
		var speed float64
		if elapsed > 0 {
			speed = float64(written) / elapsed
		}

		var percentage float64
		if sourceTotal > 0 {
			percentage = float64(sourceRead) / float64(sourceTotal) * 100
		} else if total > 0 {
			percentage = float64(written) / float64(total) * 100
		}

		f.opts.ProgressCb(Progress{
			Phase:          "writing",
			BytesProcessed: written,
			BytesTotal:     total,
			Percentage:     percentage,
			Speed:          speed,
		})
	}
}
