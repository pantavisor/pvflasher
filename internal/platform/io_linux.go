//go:build linux

package platform

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// prepareDevice unmounts all partitions of the device on Linux before raw writing.
func prepareDevice(path string) error {
	// Find the base device name (e.g. "sda" from "/dev/sda")
	devBase := filepath.Base(path)

	// Read /proc/mounts to find mounted partitions
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return fmt.Errorf("failed to read /proc/mounts: %w", err)
	}
	defer f.Close()

	// Collect mount points to unmount. We need to check both:
	// 1. The device column (fields[0]) — normal case: /dev/sda1 mounted on /mnt
	// 2. The mount point column (fields[1]) — reverse case: something mounted ON /dev/sda
	//    This can happen when another filesystem shadows the block device node.
	var mountPoints []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		dev := fields[0]
		mountPoint := fields[1]

		// Check if the device column matches (e.g. /dev/sda, /dev/sda1)
		if strings.HasPrefix(dev, "/dev/") {
			name := filepath.Base(dev)
			if name == devBase || strings.HasPrefix(name, devBase) {
				mountPoints = append(mountPoints, mountPoint)
				continue
			}
		}

		// Check if something is mounted ON the device path or its partitions
		// (e.g. /dev/vdc mounted on /dev/sda)
		if strings.HasPrefix(mountPoint, "/dev/") {
			name := filepath.Base(mountPoint)
			if name == devBase || strings.HasPrefix(name, devBase) {
				mountPoints = append(mountPoints, mountPoint)
			}
		}
	}

	// Unmount each mount point
	for _, mp := range mountPoints {
		cmd := exec.Command("umount", mp)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to unmount %s: %w", mp, err)
		}
	}

	return nil
}

func ejectDevice(path string) error {
	// Use the 'eject' command which handles unmounting and physical ejection
	cmd := exec.Command("eject", path)
	return cmd.Run()
}
