package flash

import (
	"errors"
	"io"
)

// devicePipe overlaps image decompression with device writes. The producer
// (the caller's goroutine in flasher.go) fills borrowed buffers from the
// decompressed image stream and submits them tagged with the absolute device
// offset; a single consumer goroutine seeks the device only when offsets are
// discontinuous, then writes. This lets the CPU-bound decompress run ahead of
// the slower USB/SD/eMMC writes instead of alternating with them.
//
// Buffers are recycled through a fixed free-list, so there is no per-chunk
// allocation and the producer fills a free-list buffer directly (no extra copy).
// Read-ahead is bounded by numBufs*bufSize.
type devicePipe struct {
	dev    io.WriteSeeker
	free   chan []byte
	filled chan pipeChunk
	done   chan struct{} // closed by the consumer if a write fails
	fin    chan struct{} // closed when the consumer goroutine exits

	onProgress func(written, sourceRead int64)

	// written/err are owned by the consumer goroutine and are only read by the
	// producer after finish() observes <-fin (a happens-before edge).
	written int64
	err     error
}

type pipeChunk struct {
	off        int64
	buf        []byte // data to write (sub-slice of a free-list buffer)
	sourceRead int64  // compressed bytes read so far, for progress reporting
}

var errZeroWrite = errors.New("device write returned 0 bytes")

func newDevicePipe(dev io.WriteSeeker, numBufs, bufSize int, onProgress func(written, sourceRead int64)) *devicePipe {
	p := &devicePipe{
		dev:        dev,
		free:       make(chan []byte, numBufs),
		filled:     make(chan pipeChunk, numBufs),
		done:       make(chan struct{}),
		fin:        make(chan struct{}),
		onProgress: onProgress,
	}
	for i := 0; i < numBufs; i++ {
		p.free <- make([]byte, bufSize)
	}
	go p.consume()
	return p
}

func (p *devicePipe) consume() {
	defer close(p.fin)
	cur := int64(-1) // unknown device position
	for c := range p.filled {
		if c.off != cur {
			if _, err := p.dev.Seek(c.off, io.SeekStart); err != nil {
				p.err = err
				close(p.done)
				return
			}
			cur = c.off
		}
		for w := 0; w < len(c.buf); {
			n, err := p.dev.Write(c.buf[w:])
			if err != nil {
				p.err = err
				close(p.done)
				return
			}
			if n == 0 {
				p.err = errZeroWrite
				close(p.done)
				return
			}
			w += n
		}
		cur += int64(len(c.buf))
		p.written += int64(len(c.buf))
		if p.onProgress != nil {
			p.onProgress(p.written, c.sourceRead)
		}
		p.free <- c.buf[:cap(c.buf)]
	}
}

// get returns a buffer to fill, or ok=false if the consumer has aborted.
func (p *devicePipe) get() (buf []byte, ok bool) {
	select {
	case b := <-p.free:
		return b[:cap(b)], true
	case <-p.done:
		return nil, false
	}
}

// recycle returns an unused buffer (e.g. a zero-length read) to the free-list.
func (p *devicePipe) recycle(buf []byte) {
	select {
	case p.free <- buf[:cap(buf)]:
	case <-p.done:
	}
}

// submit hands a filled buffer (data == buf[:n]) to the consumer for writing at
// device offset off. Returns false if the consumer has aborted.
func (p *devicePipe) submit(off int64, data []byte, sourceRead int64) bool {
	select {
	case p.filled <- pipeChunk{off: off, buf: data, sourceRead: sourceRead}:
		return true
	case <-p.done:
		return false
	}
}

// finish signals end-of-input and waits for the consumer to drain. It returns
// the total bytes written and the first write error, if any.
func (p *devicePipe) finish() (int64, error) {
	close(p.filled)
	<-p.fin
	return p.written, p.err
}
