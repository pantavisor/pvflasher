//go:build !windows

package commands

// AttachConsoleIfNeeded is a no-op on non-Windows platforms
func AttachConsoleIfNeeded() {
	// Console is already available on Unix-like systems
}
