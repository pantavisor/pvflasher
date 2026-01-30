package platform

import "os/exec"

// Elevator defines the interface for requesting administrative privileges
type Elevator interface {
	// IsAdmin returns true if the current process has administrative privileges
	IsAdmin() bool
	// ElevateCommand returns an exec.Cmd configured to run the current executable
	// with the given arguments as an administrator.
	ElevateCommand(args ...string) (*exec.Cmd, error)
}

// NewElevator returns an Elevator implementation for the current platform
func NewElevator() Elevator {
	return newPlatformElevator()
}
