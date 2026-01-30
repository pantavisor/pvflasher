//go:build windows

package commands

import (
	"os"
	"sync"
	"syscall"
)

var (
	kernel32      = syscall.NewLazyDLL("kernel32.dll")
	attachConsole = kernel32.NewProc("AttachConsole")
	allocConsole  = kernel32.NewProc("AllocConsole")

	consoleOnce sync.Once
)

const ATTACH_PARENT_PROCESS = ^uintptr(0) // -1

// AttachConsoleIfNeeded attaches to the parent console for CLI output.
// This is needed because Windows GUI apps (-H=windowsgui) don't have a console.
// Only call this when running in CLI mode, not GUI mode.
func AttachConsoleIfNeeded() {
	consoleOnce.Do(func() {
		// Try to attach to parent console (e.g., cmd.exe, PowerShell)
		r, _, _ := attachConsole.Call(ATTACH_PARENT_PROCESS)
		if r == 0 {
			// No parent console - don't allocate a new one for GUI apps
			// Only allocate if we really need CLI output
			return
		}

		// Reopen stdout/stderr to the console
		// This is necessary after attaching to a console
		hOut, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
		hErr, _ := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)

		if hOut != syscall.InvalidHandle && hOut != 0 {
			os.Stdout = os.NewFile(uintptr(hOut), "stdout")
		}
		if hErr != syscall.InvalidHandle && hErr != 0 {
			os.Stderr = os.NewFile(uintptr(hErr), "stderr")
		}
	})
}
