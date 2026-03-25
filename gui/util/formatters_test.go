package util

import (
	"strings"
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
		{2.5 * 1024 * 1024, "2.50 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{2.5 * 1024 * 1024 * 1024, "2.50 GB"},
		{1024 * 1024 * 1024 * 1024, "1.00 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatSpeed(t *testing.T) {
	tests := []struct {
		bytesPerSec float64
		expected    string
	}{
		{0, "0 B/s"},
		{100, "100 B/s"},
		{1024, "1.00 KB/s"},
		{1536, "1.50 KB/s"},
		{1024 * 1024, "1.00 MB/s"},
		{2.5 * 1024 * 1024, "2.50 MB/s"},
		{100 * 1024 * 1024, "100.00 MB/s"},
		{1024 * 1024 * 1024, "1.00 GB/s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatSpeed(tt.bytesPerSec)
			if !strings.HasPrefix(result, tt.expected[:len(tt.expected)-3]) || !strings.HasSuffix(result, tt.expected[len(tt.expected)-3:]) {
				// Allow for floating point differences
				if result != tt.expected {
					t.Errorf("FormatSpeed(%f) = %q, want %q", tt.bytesPerSec, result, tt.expected)
				}
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0 * time.Second, "0s"},
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m 30s"},
		{3600 * time.Second, "1h 0m 0s"},
		{3661 * time.Second, "1h 1m 1s"},
		{7322 * time.Second, "2h 2m 2s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes_Boundary(t *testing.T) {
	// Test boundary values
	tests := []struct {
		bytes    int64
		contains string
	}{
		{1023, "1023 B"},
		{1024, "1.00 KB"},
		{1025, "1.00 KB"},
		{1024*1024 - 1, "1024.00 KB"},
		{1024 * 1024, "1.00 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.contains, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("FormatBytes(%d) = %q, should contain %q", tt.bytes, result, tt.contains)
			}
		})
	}
}
