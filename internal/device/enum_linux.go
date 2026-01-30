//go:build linux

package device

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jaypipes/ghw"
)

func newPlatformManager() Manager {
	return &LinuxManager{}
}

type LinuxManager struct{}

func (m *LinuxManager) List() ([]Device, error) {
	block, err := ghw.Block()
	if err != nil {
		return nil, fmt.Errorf("failed to get block info: %w", err)
	}

	mounts, err := getMounts()
	if err != nil {
		// Log error but continue
		fmt.Fprintf(os.Stderr, "Warning: failed to get mount points: %v\n", err)
	}

	var devices []Device
	for _, disk := range block.Disks {
		devName := "/dev/" + disk.Name
		d := Device{
			Name:        devName,
			Size:        int64(disk.SizeBytes),
			Model:       disk.Model,
			Vendor:      disk.Vendor,
			Removable:   disk.IsRemovable,
			MountPoints: mounts[devName],
		}

		// Also check partitions for mounts
		for _, part := range disk.Partitions {
			partName := "/dev/" + part.Name
			if partMounts, ok := mounts[partName]; ok {
				d.MountPoints = append(d.MountPoints, partMounts...)
			}
		}

		devices = append(devices, d)
	}
	return devices, nil
}

func getMounts() (map[string][]string, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	mounts := make(map[string][]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 {
			dev := fields[0]
			mountPoint := fields[1]
			if strings.HasPrefix(dev, "/dev/") {
				mounts[dev] = append(mounts[dev], mountPoint)
			}
		}
	}
	return mounts, nil
}
