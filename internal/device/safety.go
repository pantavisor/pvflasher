package device

import (
	"os"
	"runtime"
	"strings"
)

// autoMountRoots are where desktops auto-mount inserted media. A disk mounted
// anywhere else (/, /home, /home/projects, /mnt/data, ...) is in use by the
// system and must never be offered as a flash target.
var autoMountRoots = []string{"/run/media/", "/media/", "/Volumes/"}

// UnsafeReason explains why d should not be offered as a flash target, or
// returns "" when it is a plausible target (external media, only mounted
// where the desktop auto-mounts removable drives).
func (d Device) UnsafeReason() string {
	if !d.External {
		return "internal disk"
	}
	if d.Size <= 0 {
		return "no media inserted"
	}
	for _, mp := range d.MountPoints {
		if !isAutoMount(mp) {
			return "in use by the system (mounted at " + mp + ")"
		}
	}
	return ""
}

func isAutoMount(mountPoint string) bool {
	if runtime.GOOS == "windows" {
		// Any drive letter except the system drive (usually C:).
		sys := strings.ToUpper(strings.TrimSuffix(os.Getenv("SystemDrive"), `\`))
		if sys == "" {
			sys = "C:"
		}
		return !strings.HasPrefix(strings.ToUpper(mountPoint), sys)
	}
	for _, root := range autoMountRoots {
		if strings.HasPrefix(mountPoint, root) {
			return true
		}
	}
	return false
}
