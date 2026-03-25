//go:build darwin

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

type DarwinDeviceWriter struct {
	f *os.File
}

func openDevice(path string) (DeviceWriter, error) {
	path = rawDevicePath(path)

	// Use O_RDWR | O_EXLOCK if possible for exclusive access
	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	// Disable caching for raw device access
	_, _, errno := syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), syscall.F_NOCACHE, 1)
	if errno != 0 {
		// Log warning but continue?
		fmt.Fprintf(os.Stderr, "Warning: failed to set F_NOCACHE on %s: %v\n", path, errno)
	}

	// Try to get exclusive lock
	err = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to lock device (is it in use?): %w", err)
	}

	return &DarwinDeviceWriter{f: f}, nil
}

func (w *DarwinDeviceWriter) Read(p []byte) (int, error) {
	return w.f.Read(p)
}

func (w *DarwinDeviceWriter) Write(p []byte) (int, error) {
	return w.f.Write(p)
}

func (w *DarwinDeviceWriter) Close() error {
	unix.Flock(int(w.f.Fd()), unix.LOCK_UN)
	return w.f.Close()
}

func (w *DarwinDeviceWriter) Seek(offset int64, whence int) (int64, error) {
	return w.f.Seek(offset, whence)
}

func (w *DarwinDeviceWriter) Sync() error {
	return w.f.Sync()
}

func (w *DarwinDeviceWriter) Fd() uintptr {
	return w.f.Fd()
}

// prepareDevice unmounts the disk on macOS before raw writing.
func prepareDevice(path string) error {
	// Use diskutil unmountDisk to unmount all volumes on the device
	cmd := exec.Command("diskutil", "unmountDisk", diskutilPath(path))
	return cmd.Run()
}

func ejectDevice(path string) error {
	// Use diskutil unmountDisk and then eject
	path = diskutilPath(path)
	exec.Command("diskutil", "unmountDisk", path).Run()
	cmd := exec.Command("diskutil", "eject", path)
	return cmd.Run()
}

func rawDevicePath(path string) string {
	if strings.HasPrefix(path, "/dev/disk") {
		return strings.Replace(path, "/dev/disk", "/dev/rdisk", 1)
	}
	return path
}

func diskutilPath(path string) string {
	if strings.HasPrefix(path, "/dev/rdisk") {
		return strings.Replace(path, "/dev/rdisk", "/dev/disk", 1)
	}
	return path
}
