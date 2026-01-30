//go:build windows

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows"
)

type WindowsElevator struct{}

func newPlatformElevator() Elevator {
	return &WindowsElevator{}
}

func (e *WindowsElevator) IsAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

func (e *WindowsElevator) ElevateCommand(args ...string) (*exec.Cmd, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	// Use PowerShell to start the process with elevation (RunAs)
	psArgs := fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList '%s' -Verb RunAs", exe, strings.Join(args, " "))
	return exec.Command("powershell", "-Command", psArgs), nil
}

// Legacy helpers for compatibility
func IsRoot() bool {
	return (&WindowsElevator{}).IsAdmin()
}

func RelaunchWithSudo() error {
	elevator := &WindowsElevator{}
	cmd, err := elevator.ElevateCommand(os.Args[1:]...)
	if err != nil {
		return err
	}
	return cmd.Run()
}
