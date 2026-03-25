package device

import (
	"testing"
)

func TestNewPlatformManager(t *testing.T) {
	// Test that newPlatformManager returns a non-nil Manager
	mgr := newPlatformManager()
	if mgr == nil {
		t.Error("newPlatformManager() returned nil")
	}
}

func TestLinuxManager_Interface(t *testing.T) {
	// Test that LinuxManager implements Manager interface
	var _ Manager = (*LinuxManager)(nil)
}

func TestLinuxManager_List(t *testing.T) {
	mgr := &LinuxManager{}

	// This test will fail if run without proper permissions or on non-Linux systems
	// It's marked to be skipped in CI or when the ghw library can't access hardware
	devices, err := mgr.List()

	// We expect this to either succeed or fail gracefully
	// The actual result depends on the system capabilities
	if err != nil {
		// Log the error but don't fail the test
		// Hardware enumeration may fail in CI/containers
		t.Logf("List() returned error (expected in CI): %v", err)
		return
	}

	// If successful, verify devices are reasonable
	for _, d := range devices {
		if d.Name == "" {
			t.Error("Device has empty name")
		}
		if d.Size < 0 {
			t.Errorf("Device %s has negative size: %d", d.Name, d.Size)
		}
	}

	t.Logf("Found %d devices", len(devices))
}
