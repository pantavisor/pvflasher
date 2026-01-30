package flash

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"pvflasher/internal/archive"
	"pvflasher/internal/bmap"
	"pvflasher/internal/image"
	"pvflasher/internal/platform"
)

type Verifier struct {
	opts       Options
	bm         *bmap.Bmap
	imageEntry string
}

func NewVerifier(opts Options) *Verifier {
	return &Verifier{opts: opts}
}

func (v *Verifier) SetBmap(bm *bmap.Bmap) {
	v.bm = bm
}

func (v *Verifier) SetImageEntry(entry string) {
	v.imageEntry = entry
}

// Verify checks the device content against the bmap checksums
func (v *Verifier) Verify(ctx context.Context) error {
	if v.bm != nil || v.opts.BmapPath != "" {
		return v.verifyWithBmap(ctx)
	}
	return v.verifyRaw(ctx)
}

func (v *Verifier) verifyWithBmap(ctx context.Context) error {
	// 1. Open Device
	dev, err := platform.OpenDevice(v.opts.DevicePath)
	if err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}
	defer dev.Close()

	// 2. Load Bmap
	var bm *bmap.Bmap
	if v.bm != nil {
		bm = v.bm
	} else {
		bmapFile, err := os.Open(v.opts.BmapPath)
		if err != nil {
			return err
		}
		defer bmapFile.Close()

		bm, err = bmap.Parse(bmapFile)
		if err != nil {
			return err
		}
	}

	// 3. Verification Loop
	startTime := time.Now()
	totalBytes := bm.MappedBlocksCount * int64(bm.BlockSize)
	var verifiedBytes int64

	bufSize := 1024 * 1024 // 1MB
	buf := make([]byte, bufSize)

	for _, rng := range bm.BlockMap {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		parsedRange, err := rng.Parse()
		if err != nil {
			return err
		}

		startByte := parsedRange.Start * int64(bm.BlockSize)
		endByte := (parsedRange.End + 1) * int64(bm.BlockSize)

		// Cap endByte to actual ImageSize to handle partial last block
		if endByte > bm.ImageSize {
			endByte = bm.ImageSize
		}

		countByte := endByte - startByte

		// Seek device to start
		_, err = dev.Seek(startByte, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to seek device to %d: %w", startByte, err)
		}

		hasher, err := bmap.GetHasher(bm.ChecksumType)
		if err != nil {
			return err
		}

		// Read loop
		remaining := countByte
		for remaining > 0 {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			toRead := int64(len(buf))
			if remaining < toRead {
				toRead = remaining
			}

			n, err := io.ReadFull(dev, buf[:toRead])
			if err != nil {
				return fmt.Errorf("read error verifying block %d: %w", parsedRange.Start, err)
			}

			hasher.Write(buf[:n])
			remaining -= int64(n)
			verifiedBytes += int64(n)

			v.reportProgress(verifiedBytes, totalBytes, startTime)
		}

		// Compare checksums
		calculatedSum := fmt.Sprintf("%x", hasher.Sum(nil))
		if calculatedSum != parsedRange.Checksum {
			return fmt.Errorf("checksum mismatch at range %s: expected %s, got %s", rng.Text, parsedRange.Checksum, calculatedSum)
		}
	}

	return nil
}

func (v *Verifier) verifyRaw(ctx context.Context) error {
	// 1. Open Device
	dev, err := platform.OpenDevice(v.opts.DevicePath)
	if err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}
	defer dev.Close()

	// 2. Open Image
	var imgReader io.Reader
	var cleanup func()
	var totalBytes int64

	if archive.IsArchive(v.opts.ImagePath) {
		entry := v.imageEntry
		if entry == "" {
			pair, err := archive.GetArchivePair(v.opts.ImagePath)
			if err != nil {
				return fmt.Errorf("failed to scan archive: %w", err)
			}
			entry = pair.ImageEntry
		}

		rc, _, err := archive.OpenArchiveImage(v.opts.ImagePath, entry)
		if err != nil {
			return fmt.Errorf("failed to open archive entry: %w", err)
		}
		cleanup = func() { rc.Close() }

		// Note: We don't have exact total size for progress in raw verification from archive
		// unless we knew uncompressed size. We can skip totalBytes or use compressed size as estimate.
		totalBytes = 0

		imgReader, err = image.Decompressor(entry, rc)
		if err != nil {
			cleanup()
			return err
		}
	} else {
		imgFile, err := os.Open(v.opts.ImagePath)
		if err != nil {
			return fmt.Errorf("failed to open image: %w", err)
		}
		cleanup = func() { imgFile.Close() }

		fi, _ := imgFile.Stat()
		totalBytes = fi.Size()

		imgReader, err = image.Decompressor(v.opts.ImagePath, imgFile)
		if err != nil {
			cleanup()
			return err
		}
	}
	defer cleanup()

	// 3. Compare Loop
	startTime := time.Now()
	var verifiedBytes int64
	bufSize := 1024 * 1024
	bufImg := make([]byte, bufSize)
	bufDev := make([]byte, bufSize)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		nImg, errImg := io.ReadFull(imgReader, bufImg)
		if nImg > 0 {
			nDev, errDev := io.ReadFull(dev, bufDev[:nImg])
			if errDev != nil {
				return fmt.Errorf("failed to read from device during verification: %w", errDev)
			}
			if nDev != nImg {
				return fmt.Errorf("short read from device during verification: expected %d, got %d", nImg, nDev)
			}

			for i := 0; i < nImg; i++ {
				if bufImg[i] != bufDev[i] {
					return fmt.Errorf("verification failed: mismatch at byte %d", verifiedBytes+int64(i))
				}
			}

			verifiedBytes += int64(nImg)
			v.reportProgress(verifiedBytes, totalBytes, startTime)
		}

		if errImg == io.EOF || errImg == io.ErrUnexpectedEOF {
			break
		}
		if errImg != nil {
			return fmt.Errorf("failed to read from image during verification: %w", errImg)
		}
	}

	return nil
}

func (v *Verifier) reportProgress(verified, total int64, start time.Time) {
	if v.opts.ProgressCb != nil {
		elapsed := time.Since(start).Seconds()
		var speed float64
		if elapsed > 0 {
			speed = float64(verified) / elapsed
		}
		var percentage float64
		if total > 0 {
			percentage = float64(verified) / float64(total) * 100
		}
		v.opts.ProgressCb(Progress{
			Phase:          "verifying",
			BytesProcessed: verified,
			BytesTotal:     total,
			Percentage:     percentage,
			Speed:          speed,
		})
	}
}
