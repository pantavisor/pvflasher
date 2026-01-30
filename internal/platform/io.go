package platform

import (
	"io"
)

// DeviceWriter is an interface for writing to a block device
type DeviceWriter interface {
	io.Writer
	io.Reader
	io.Seeker
	io.Closer
	Sync() error
	Fd() uintptr // Useful for ioctl if needed
}

// PrepareDevice prepares the device for raw writing by dismounting volumes.
// On Windows, this dismounts all volumes on the physical device.
// On Linux/macOS, this is a no-op as the device open handles exclusivity.
func PrepareDevice(path string) error {
	return prepareDevice(path)
}

// OpenDevice opens a block device for writing with platform-specific optimizations
func OpenDevice(path string) (DeviceWriter, error) {
	return openDevice(path)
}

// EjectDevice attempts to unmount and eject the device
func EjectDevice(path string) error {
	return ejectDevice(path)
}
