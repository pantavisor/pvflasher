//go:build !windows

package device

import "testing"

func TestUnsafeReason(t *testing.T) {
	tests := []struct {
		name string
		dev  Device
		safe bool
	}{
		{"sd card", Device{External: true, Size: 32e9}, true},
		{"auto-mounted usb", Device{External: true, Size: 32e9, MountPoints: []string{"/run/media/user/BOOT"}}, true},
		{"macos volume", Device{External: true, Size: 32e9, MountPoints: []string{"/Volumes/BOOT"}}, true},
		{"internal nvme", Device{External: false, Size: 1e12}, false},
		{"usb disk with /home/projects", Device{External: true, Size: 1e12, MountPoints: []string{"/home/projects"}}, false},
		{"usb root", Device{External: true, Size: 1e12, MountPoints: []string{"/"}}, false},
		{"manual /mnt mount", Device{External: true, Size: 1e12, MountPoints: []string{"/mnt/data"}}, false},
		{"empty card reader", Device{External: true, Size: 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := tt.dev.UnsafeReason()
			if (reason == "") != tt.safe {
				t.Errorf("UnsafeReason() = %q, want safe=%v", reason, tt.safe)
			}
		})
	}
}
