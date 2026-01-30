//go:build linux

package platform

import (
	"os"
	"os/exec"
	"syscall"
)

type LinuxDeviceWriter struct {
	f *os.File
}

func openDevice(path string) (DeviceWriter, error) {
	// O_DIRECT requires aligned memory and size, handling that in Writer is complex.
	// For now, let's start with O_EXCL to ensure exclusive access.
	// O_DIRECT can be added later if performance or cache effects are critical.
	// bmaptool uses O_LARGEFILE | O_DIRECT | O_EXCL | O_WRONLY
	f, err := os.OpenFile(path, os.O_RDWR|syscall.O_EXCL, 0666)
	if err != nil {
		return nil, err
	}
	return &LinuxDeviceWriter{f: f}, nil
}

func (w *LinuxDeviceWriter) Read(p []byte) (int, error) {
	return w.f.Read(p)
}

func (w *LinuxDeviceWriter) Write(p []byte) (int, error) {
	return w.f.Write(p)
}

func (w *LinuxDeviceWriter) Close() error {
	return w.f.Close()
}

func (w *LinuxDeviceWriter) Seek(offset int64, whence int) (int64, error) {
	return w.f.Seek(offset, whence)
}

func (w *LinuxDeviceWriter) Sync() error {
	return w.f.Sync()
}

func (w *LinuxDeviceWriter) Fd() uintptr {
	return w.f.Fd()
}

// prepareDevice is a no-op on Linux.
// The O_EXCL flag in openDevice ensures exclusive access.
func prepareDevice(path string) error {
	return nil
}

func ejectDevice(path string) error {
	// Use the 'eject' command which handles unmounting and physical ejection
	cmd := exec.Command("eject", path)
	return cmd.Run()
}
