//go:build darwin

package device

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"howett.net/plist"
)

func newPlatformManager() Manager {
	return &DarwinManager{}
}

type DarwinManager struct{}

type diskUtilList struct {
	AllDisks []string `plist:"AllDisks"`
}

type diskUtilInfo struct {
	DeviceIdentifier string `plist:"DeviceIdentifier"`
	Size             int64  `plist:"Size"`
	Removable        bool   `plist:"Removable"`
	Model            string `plist:"Model"`
	Vendor           string `plist:"Vendor"`
	MountPoint       string `plist:"MountPoint"`
	Partitions       []struct {
		DeviceIdentifier string `plist:"DeviceIdentifier"`
		MountPoint       string `plist:"MountPoint"`
	} `plist:"Partitions"`
}

func (m *DarwinManager) List() ([]Device, error) {
	cmd := exec.Command("diskutil", "list", "-plist")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run diskutil list: %w", err)
	}

	var list diskUtilList
	if _, err := plist.Unmarshal(stdout.Bytes(), &list); err != nil {
		return nil, fmt.Errorf("failed to parse diskutil list plist: %w", err)
	}

	var devices []Device
	for _, devID := range list.AllDisks {
		// Only look at whole disks, e.g., disk0, disk1, not partitions like disk0s1
		if !isWholeDisk(devID) {
			continue
		}

		info, err := getDiskInfo(devID)
		if err != nil {
			continue
		}

		d := Device{
			Name:      "/dev/" + info.DeviceIdentifier,
			Size:      info.Size,
			Model:     info.Model,
			Vendor:    info.Vendor,
			Removable: info.Removable,
		}

		if info.MountPoint != "" {
			d.MountPoints = append(d.MountPoints, info.MountPoint)
		}

		for _, p := range info.Partitions {
			if p.MountPoint != "" {
				d.MountPoints = append(d.MountPoints, p.MountPoint)
			}
		}

		devices = append(devices, d)
	}
	return devices, nil
}

func isWholeDisk(devID string) bool {
	// diskX is a whole disk, diskXsY is a partition
	return !strings.Contains(devID, "s")
}

func getDiskInfo(devID string) (*diskUtilInfo, error) {
	cmd := exec.Command("diskutil", "info", "-plist", devID)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var info diskUtilInfo
	if _, err := plist.Unmarshal(stdout.Bytes(), &info); err != nil {
		return nil, err
	}
	return &info, nil
}
