package image

import (
	"errors"
	"io"
)

// CountingReader wraps an io.Reader and counts the number of bytes read
type CountingReader struct {
	io.Reader
	Count int64
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.Count += int64(n)
	return
}

// ForwardSeeker allows seeking forward only on a reader
type ForwardSeeker struct {
	r      io.Reader
	offset int64
}

// NewForwardSeeker creates a new ForwardSeeker
func NewForwardSeeker(r io.Reader) *ForwardSeeker {
	return &ForwardSeeker{r: r}
}

func (s *ForwardSeeker) Read(p []byte) (n int, err error) {
	n, err = s.r.Read(p)
	s.offset += int64(n)
	return
}

// Seek implements io.Seeker but only supports forward seeking
func (s *ForwardSeeker) Seek(offset int64, whence int) (int64, error) {
	var target int64
	switch whence {
	case io.SeekStart:
		target = offset
	case io.SeekCurrent:
		target = s.offset + offset
	default:
		return s.offset, errors.New("only SeekCurrent and SeekStart are supported")
	}

	if target < s.offset {
		return s.offset, errors.New("cannot seek backwards in a stream")
	}

	if target == s.offset {
		return s.offset, nil
	}

	delta := target - s.offset
	// discard delta bytes
	copied, err := io.CopyN(io.Discard, s.r, delta)
	s.offset += copied
	if err != nil {
		return s.offset, err
	}

	return s.offset, nil
}
