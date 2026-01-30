package util

import (
	"fmt"
	"time"
)

// FormatBytes converts bytes to human-readable format (B, KB, MB, GB, TB)
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatSpeed converts bytes/second to human-readable speed (B/s, KB/s, MB/s, GB/s)
func FormatSpeed(bytesPerSecond float64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytesPerSecond >= GB:
		return fmt.Sprintf("%.2f GB/s", bytesPerSecond/float64(GB))
	case bytesPerSecond >= MB:
		return fmt.Sprintf("%.2f MB/s", bytesPerSecond/float64(MB))
	case bytesPerSecond >= KB:
		return fmt.Sprintf("%.2f KB/s", bytesPerSecond/float64(KB))
	default:
		return fmt.Sprintf("%.0f B/s", bytesPerSecond)
	}
}

// FormatDuration converts time.Duration to human-readable string
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
