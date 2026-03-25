package image

import (
	"testing"
)

func TestByteRange(t *testing.T) {
	tests := []struct {
		name  string
		start int64
		end   int64
	}{
		{
			name:  "Simple range",
			start: 0,
			end:   100,
		},
		{
			name:  "Offset range",
			start: 1024,
			end:   2048,
		},
		{
			name:  "Large range",
			start: 0,
			end:   1 << 30, // 1GB
		},
		{
			name:  "Single byte range",
			start: 50,
			end:   51,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := ByteRange{
				Start: tt.start,
				End:   tt.end,
			}

			if br.Start != tt.start {
				t.Errorf("Start = %v, want %v", br.Start, tt.start)
			}
			if br.End != tt.end {
				t.Errorf("End = %v, want %v", br.End, tt.end)
			}
		})
	}
}

func TestByteRange_Length(t *testing.T) {
	tests := []struct {
		name       string
		start      int64
		end        int64
		wantLength int64
	}{
		{
			name:       "Normal range",
			start:      0,
			end:        100,
			wantLength: 100,
		},
		{
			name:       "Offset range",
			start:      500,
			end:        1500,
			wantLength: 1000,
		},
		{
			name:       "Single byte",
			start:      100,
			end:        101,
			wantLength: 1,
		},
		{
			name:       "Empty range",
			start:      100,
			end:        100,
			wantLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := ByteRange{
				Start: tt.start,
				End:   tt.end,
			}

			length := br.End - br.Start
			if length != tt.wantLength {
				t.Errorf("Length = %v, want %v", length, tt.wantLength)
			}
		})
	}
}

func TestByteRange_Slice(t *testing.T) {
	data := []byte("Hello, World! This is a test.")

	tests := []struct {
		name  string
		start int64
		end   int64
		want  string
	}{
		{
			name:  "First word",
			start: 0,
			end:   5,
			want:  "Hello",
		},
		{
			name:  "Second word",
			start: 7,
			end:   12,
			want:  "World",
		},
		{
			name:  "Entire string",
			start: 0,
			end:   int64(len(data)),
			want:  string(data),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.end > int64(len(data)) {
				t.Skip("Range exceeds data length")
			}

			result := string(data[tt.start:tt.end])
			if result != tt.want {
				t.Errorf("Got %q, want %q", result, tt.want)
			}
		})
	}
}
