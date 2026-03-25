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
	commandLine, err := darwinCommandLine(args...)
	if err != nil {
		return nil, err
	}

	// Use osascript to run with administrator privileges.
	// This prompts the user with a standard macOS system dialog.
	script := fmt.Sprintf("do shell script \"%s\" with administrator privileges", escapeAppleScript(commandLine))
	return exec.Command("osascript", "-e", script), nil
}

// Legacy helpers for compatibility
func IsRoot() bool {
	return (&DarwinElevator{}).IsAdmin()
}

func RelaunchWithSudo() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command("sudo", append([]string{exe}, os.Args[1:]...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}

func darwinCommandLine(args ...string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	parts := []string{shellQuote(exe)}
	for _, arg := range args {
		parts = append(parts, shellQuote(arg))
	}
	return strings.Join(parts, " "), nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func escapeAppleScript(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}
