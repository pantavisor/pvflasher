//go:build windows

package platform

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows IOCTL constants for device ejection
const (
	FSCTL_LOCK_VOLUME               = 0x00090018
	FSCTL_UNLOCK_VOLUME             = 0x0009001C
	FSCTL_DISMOUNT_VOLUME           = 0x00090020
	IOCTL_STORAGE_GET_DEVICE_NUMBER = 0x002D1080
	IOCTL_STORAGE_MEDIA_REMOVAL     = 0x002D4804
	IOCTL_STORAGE_EJECT_MEDIA       = 0x002D4808
)

// PREVENT_MEDIA_REMOVAL structure for IOCTL_STORAGE_MEDIA_REMOVAL
type PREVENT_MEDIA_REMOVAL struct {
	PreventMediaRemoval byte
}

// STORAGE_DEVICE_NUMBER structure for IOCTL_STORAGE_GET_DEVICE_NUMBER
type STORAGE_DEVICE_NUMBER struct {
	DeviceType      uint32
	DeviceNumber    uint32
	PartitionNumber uint32
}

type WindowsDeviceWriter struct {
	f             *os.File
	volumeHandles []windows.Handle // Keep volume handles open to maintain locks
}

func openDevice(path string) (DeviceWriter, error) {
	// On Windows, raw disk access requires special flags and often admin rights.

	// Ensure path is in \\.\PhysicalDriveN format for raw access (case-insensitive check)
	upperPath := strings.ToUpper(path)
	if strings.HasPrefix(upperPath, "PHYSICALDRIVE") {
		path = `\\.\` + path
	}

	// First, lock and dismount all volumes on this device
	volumeHandles, err := lockAndDismountVolumes(path)
	if err != nil {
		// Log warning but continue - might be a new/empty device
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		closeVolumeHandles(volumeHandles)
		return nil, err
	}

	// Open with FILE_FLAG_WRITE_THROUGH for reliability
	// Note: FILE_FLAG_NO_BUFFERING requires sector-aligned I/O which can be problematic
	handle, err := windows.CreateFile(
		pathPtr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_WRITE_THROUGH,
		0,
	)
	if err != nil {
		closeVolumeHandles(volumeHandles)
		return nil, fmt.Errorf("failed to open device %s: %w", path, err)
	}

	f := os.NewFile(uintptr(handle), path)
	return &WindowsDeviceWriter{f: f, volumeHandles: volumeHandles}, nil
}

// lockAndDismountVolumes locks and dismounts all volumes on the device, returning handles that must be kept open
func lockAndDismountVolumes(devicePath string) ([]windows.Handle, error) {
	deviceNum, err := extractDeviceNumber(devicePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse device number: %w", err)
	}

	volumes, err := getVolumesForDevice(deviceNum)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate volumes: %w", err)
	}

	var handles []windows.Handle
	for _, vol := range volumes {
		handle, err := lockAndDismountVolume(vol)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to lock volume %s: %v\n", vol, err)
			continue
		}
		if handle != windows.InvalidHandle {
			handles = append(handles, handle)
		}
	}

	return handles, nil
}

// lockAndDismountVolume locks and dismounts a single volume, returning the handle (caller must keep it open)
func lockAndDismountVolume(volumePath string) (windows.Handle, error) {
	// Ensure path is in \\.\X: format
	if !strings.HasPrefix(volumePath, `\\.\`) {
		volumePath = `\\.\` + volumePath
	}

	volumePtr, err := windows.UTF16PtrFromString(volumePath)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("invalid volume path: %w", err)
	}

	handle, err := windows.CreateFile(
		volumePtr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("failed to open volume: %w", err)
	}

	var bytesReturned uint32

	// Lock volume - this prevents other processes from accessing it
	err = windows.DeviceIoControl(
		handle,
		FSCTL_LOCK_VOLUME,
		nil,
		0,
		nil,
		0,
		&bytesReturned,
		nil,
	)
	if err != nil {
		windows.CloseHandle(handle)
		return windows.InvalidHandle, fmt.Errorf("failed to lock volume: %w", err)
	}

	// Dismount volume - this flushes caches and removes the filesystem
	err = windows.DeviceIoControl(
		handle,
		FSCTL_DISMOUNT_VOLUME,
		nil,
		0,
		nil,
		0,
		&bytesReturned,
		nil,
	)
	if err != nil {
		// Dismount failure is not fatal, continue with locked volume
		fmt.Fprintf(os.Stderr, "Warning: failed to dismount volume %s: %v\n", volumePath, err)
	}

	// Return handle - caller must keep it open to maintain the lock!
	return handle, nil
}

// closeVolumeHandles closes all volume handles (which releases the locks)
func closeVolumeHandles(handles []windows.Handle) {
	for _, h := range handles {
		if h != windows.InvalidHandle {
			// Unlock before closing (best effort)
			var bytesReturned uint32
			windows.DeviceIoControl(h, FSCTL_UNLOCK_VOLUME, nil, 0, nil, 0, &bytesReturned, nil)
			windows.CloseHandle(h)
		}
	}
}

func (w *WindowsDeviceWriter) Read(p []byte) (int, error) {
	return w.f.Read(p)
}

func (w *WindowsDeviceWriter) Write(p []byte) (int, error) {
	return w.f.Write(p)
}

func (w *WindowsDeviceWriter) Close() error {
	err := w.f.Close()
	// Release volume locks after closing the device
	closeVolumeHandles(w.volumeHandles)
	return err
}

func (w *WindowsDeviceWriter) Seek(offset int64, whence int) (int64, error) {
	return w.f.Seek(offset, whence)
}

func (w *WindowsDeviceWriter) Sync() error {
	return w.f.Sync()
}

func (w *WindowsDeviceWriter) Fd() uintptr {
	return w.f.Fd()
}

// extractDeviceNumber parses the device number from a path like \\.\PhysicalDrive1
func extractDeviceNumber(path string) (uint32, error) {
	// Expected format: \\.\PhysicalDrive1 or \\.\PHYSICALDRIVE1 (case-insensitive)
	upperPath := strings.ToUpper(path)
	upperPath = strings.TrimPrefix(upperPath, `\\.\`)
	upperPath = strings.TrimPrefix(upperPath, `PHYSICALDRIVE`)

	if upperPath == "" {
		return 0, fmt.Errorf("invalid device path format: %s", path)
	}

	num, err := strconv.ParseUint(upperPath, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse device number from %s: %w", path, err)
	}

	return uint32(num), nil
}

// getVolumesForDevice returns all volume paths (like "E:", "F:") associated with a physical device
func getVolumesForDevice(deviceNumber uint32) ([]string, error) {
	var volumes []string

	// Buffer for volume name (GUID format: \\?\Volume{GUID}\)
	var volumeNameBuf [windows.MAX_PATH + 1]uint16

	// Find first volume
	hFind, err := windows.FindFirstVolume(&volumeNameBuf[0], uint32(len(volumeNameBuf)))
	if err != nil {
		return nil, fmt.Errorf("FindFirstVolume failed: %w", err)
	}
	defer windows.FindVolumeClose(hFind)

	for {
		volumeName := windows.UTF16ToString(volumeNameBuf[:])

		// Get volume paths for this volume
		volumePaths := getVolumePathsForVolume(volumeName)

		// Check if this volume belongs to our device
		if volumeBelongsToDevice(volumeName, deviceNumber) {
			volumes = append(volumes, volumePaths...)
		}

		// Find next volume
		err = windows.FindNextVolume(hFind, &volumeNameBuf[0], uint32(len(volumeNameBuf)))
		if err != nil {
			if err == syscall.ERROR_NO_MORE_FILES {
				break
			}
			return volumes, fmt.Errorf("FindNextVolume failed: %w", err)
		}
	}

	return volumes, nil
}

// getVolumePathsForVolume returns drive letters (like "E:") for a volume GUID
func getVolumePathsForVolume(volumeName string) []string {
	var paths []string

	// Buffer for path names
	var pathNamesBuf [1024]uint16
	var returnLength uint32

	volumeNamePtr, err := windows.UTF16PtrFromString(volumeName)
	if err != nil {
		return paths
	}

	err = windows.GetVolumePathNamesForVolumeName(
		volumeNamePtr,
		&pathNamesBuf[0],
		uint32(len(pathNamesBuf)),
		&returnLength,
	)
	if err != nil {
		return paths
	}

	// Parse multi-string (null-terminated strings, double-null at end)
	for i := 0; i < len(pathNamesBuf); {
		if pathNamesBuf[i] == 0 {
			break
		}

		// Find end of current string
		end := i
		for end < len(pathNamesBuf) && pathNamesBuf[end] != 0 {
			end++
		}

		path := windows.UTF16ToString(pathNamesBuf[i:end])
		if path != "" {
			// Trim trailing backslash for drive letters (C:\ -> C:)
			path = strings.TrimSuffix(path, `\`)
			paths = append(paths, path)
		}

		i = end + 1
	}

	return paths
}

// volumeBelongsToDevice checks if a volume belongs to the specified physical device
func volumeBelongsToDevice(volumeName string, deviceNumber uint32) bool {
	// Remove trailing backslash and open the volume
	volumePath := strings.TrimSuffix(volumeName, `\`)

	volumePtr, err := windows.UTF16PtrFromString(volumePath)
	if err != nil {
		return false
	}

	handle, err := windows.CreateFile(
		volumePtr,
		0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(handle)

	// Query device number
	var devNum STORAGE_DEVICE_NUMBER
	var bytesReturned uint32

	err = windows.DeviceIoControl(
		handle,
		IOCTL_STORAGE_GET_DEVICE_NUMBER,
		nil,
		0,
		(*byte)(unsafe.Pointer(&devNum)),
		uint32(unsafe.Sizeof(devNum)),
		&bytesReturned,
		nil,
	)
	if err != nil {
		return false
	}

	return devNum.DeviceNumber == deviceNumber
}

// dismountVolume locks and dismounts a volume
func dismountVolume(volumePath string) error {
	// Ensure path is in \\.\X: format
	if !strings.HasPrefix(volumePath, `\\.\`) {
		volumePath = `\\.\` + volumePath
	}

	volumePtr, err := windows.UTF16PtrFromString(volumePath)
	if err != nil {
		return fmt.Errorf("invalid volume path: %w", err)
	}

	handle, err := windows.CreateFile(
		volumePtr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to open volume: %w", err)
	}
	defer windows.CloseHandle(handle)

	var bytesReturned uint32

	// Lock volume
	err = windows.DeviceIoControl(
		handle,
		FSCTL_LOCK_VOLUME,
		nil,
		0,
		nil,
		0,
		&bytesReturned,
		nil,
	)
	if err != nil {
		// Log but continue - volume might be in use
		fmt.Fprintf(os.Stderr, "Warning: failed to lock volume %s: %v\n", volumePath, err)
	}

	// Dismount volume
	err = windows.DeviceIoControl(
		handle,
		FSCTL_DISMOUNT_VOLUME,
		nil,
		0,
		nil,
		0,
		&bytesReturned,
		nil,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to dismount volume %s: %v\n", volumePath, err)
	}

	// Unlock volume
	err = windows.DeviceIoControl(
		handle,
		FSCTL_UNLOCK_VOLUME,
		nil,
		0,
		nil,
		0,
		&bytesReturned,
		nil,
	)
	if err != nil {
		// Unlock failure is non-critical
		fmt.Fprintf(os.Stderr, "Warning: failed to unlock volume %s: %v\n", volumePath, err)
	}

	return nil
}

// ejectPhysicalDevice ejects the physical device
func ejectPhysicalDevice(path string) error {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("invalid device path: %w", err)
	}

	handle, err := windows.CreateFile(
		pathPtr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}
	defer windows.CloseHandle(handle)

	var bytesReturned uint32

	// Allow media removal
	preventRemoval := PREVENT_MEDIA_REMOVAL{PreventMediaRemoval: 0}
	err = windows.DeviceIoControl(
		handle,
		IOCTL_STORAGE_MEDIA_REMOVAL,
		(*byte)(unsafe.Pointer(&preventRemoval)),
		uint32(unsafe.Sizeof(preventRemoval)),
		nil,
		0,
		&bytesReturned,
		nil,
	)
	if err != nil {
		// This can fail for internal drives, which is OK
		if errno, ok := err.(syscall.Errno); ok && errno == windows.ERROR_INVALID_FUNCTION {
			// Not a removable device, silently succeed
			return nil
		}
		fmt.Fprintf(os.Stderr, "Warning: failed to allow media removal: %v\n", err)
	}

	// Eject media
	err = windows.DeviceIoControl(
		handle,
		IOCTL_STORAGE_EJECT_MEDIA,
		nil,
		0,
		nil,
		0,
		&bytesReturned,
		nil,
	)
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			switch errno {
			case windows.ERROR_INVALID_FUNCTION:
				// Not a removable device, this is OK
				return nil
			case windows.ERROR_ACCESS_DENIED:
				return fmt.Errorf("access denied: run as Administrator")
			}
		}
		return fmt.Errorf("failed to eject device: %w", err)
	}

	return nil
}

// prepareDevice is now a no-op on Windows.
// Volume locking and dismounting is handled in openDevice to ensure
// the locks are held throughout the entire write operation.
func prepareDevice(path string) error {
	return nil
}

func ejectDevice(path string) error {
	// Parse device number from path
	deviceNum, err := extractDeviceNumber(path)
	if err != nil {
		return fmt.Errorf("failed to parse device number: %w", err)
	}

	// Get all volumes on this device
	volumes, err := getVolumesForDevice(deviceNum)
	if err != nil {
		// Log warning but continue - might still be able to eject
		fmt.Fprintf(os.Stderr, "Warning: failed to enumerate volumes: %v\n", err)
	}

	// Dismount all volumes
	for _, vol := range volumes {
		if err := dismountVolume(vol); err != nil {
			// Log warning but continue - non-critical
			fmt.Fprintf(os.Stderr, "Warning: failed to dismount volume %s: %v\n", vol, err)
		}
	}

	// Eject the physical device
	return ejectPhysicalDevice(path)
}
