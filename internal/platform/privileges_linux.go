//go:build linux

package platform

import (
	"os"
	"os/exec"
)

type LinuxElevator struct{}

func newPlatformElevator() Elevator {
	return &LinuxElevator{}
}

func (e *LinuxElevator) IsAdmin() bool {
	return os.Geteuid() == 0
}

// HasPolkitAgent checks if a graphical polkit authentication agent is available
// by attempting a no-op pkexec call with the internal agent disabled.
func HasPolkitAgent() bool {
	if _, err := exec.LookPath("pkexec"); err != nil {
		return false
	}
	// Try running a harmless command with --disable-internal-agent.
	// If no graphical agent is registered, pkexec exits with 127.
	cmd := exec.Command("pkexec", "--disable-internal-agent", "true")
	err := cmd.Run()
	return err == nil
}

func (e *LinuxElevator) ElevateCommand(args ...string) (*exec.Cmd, error) {
	var exe string
	var err error

	// If running from AppImage, use the AppImage path.
	// This is crucial because root might not have permissions to access the FUSE mount point
	// where the binary is located, but root can always run the AppImage itself.
	if appimagePath := os.Getenv("APPIMAGE"); appimagePath != "" {
		exe = appimagePath
	} else {
		exe, err = os.Executable()
		if err != nil {
			return nil, err
		}
	}

	fullArgs := append([]string{exe}, args...)

	// Prefer pkexec with --disable-internal-agent to avoid the textual agent
	// failing in GUI apps that have no controlling terminal.
	if _, err := exec.LookPath("pkexec"); err == nil {
		pkexecArgs := append([]string{"--disable-internal-agent"}, fullArgs...)
		return exec.Command("pkexec", pkexecArgs...), nil
	}

	return exec.Command("sudo", fullArgs...), nil
}

// IsRoot checks if the current process has root privileges (legacy helper)
func IsRoot() bool {
	return (&LinuxElevator{}).IsAdmin()
}

// RelaunchWithSudo attempts to relaunch the current application with sudo or pkexec (legacy helper)
func RelaunchWithSudo() error {
	elevator := &LinuxElevator{}
	cmd, err := elevator.ElevateCommand(os.Args[1:]...)
	if err != nil {
		return err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	err = cmd.Run()
	if err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
