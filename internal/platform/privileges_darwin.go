//go:build darwin

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type DarwinElevator struct{}

func newPlatformElevator() Elevator {
	return &DarwinElevator{}
}

func (e *DarwinElevator) IsAdmin() bool {
	return os.Geteuid() == 0
}

func (e *DarwinElevator) ElevateCommand(args ...string) (*exec.Cmd, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	// Use osascript to run with administrator privileges.
	// This prompts the user with a standard macOS system dialog.
	script := fmt.Sprintf("do shell script \"%s %s\" with administrator privileges", exe, strings.Join(args, " "))
	return exec.Command("osascript", "-e", script), nil
}

// Legacy helpers for compatibility
func IsRoot() bool {
	return (&DarwinElevator{}).IsAdmin()
}

func RelaunchWithSudo() error {
	elevator := &DarwinElevator{}
	cmd, err := elevator.ElevateCommand(os.Args[1:]...)
	if err != nil {
		return err
	}
	return cmd.Run()
}
