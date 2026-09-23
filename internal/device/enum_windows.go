//go:build windows

package device

import (
	"fmt"
	"github.com/jaypipes/ghw"
	"strings"
)

func newPlatformManager() Manager {
	return &WindowsManager{}
}

type WindowsManager struct{}

func (m *WindowsManager) List() ([]Device, error) {
	block, err := ghw.Block()
	if err != nil {
		return nil, fmt.Errorf("failed to get block info: %w", err)
	}

	var devices []Device
	for _, disk := range block.Disks {
		d := Device{
			Name:      disk.Name, // On Windows this is like PhysicalDrive0
			Size:      int64(disk.SizeBytes),
			Model:     disk.Model,
			Vendor:    disk.Vendor,
			Removable: disk.IsRemovable,
			// WMI captions for USB-attached disks end in "USB Device".
			External: disk.IsRemovable || strings.Contains(strings.ToLower(disk.Model), "usb"),
		}

		// For Windows, ghw should handle basic mount point detection via partitions
		for _, part := range disk.Partitions {
			if part.MountPoint != "" {
				d.MountPoints = append(d.MountPoints, part.MountPoint)
			}
		}

		devices = append(devices, d)
	}
	return devices, nil
}
