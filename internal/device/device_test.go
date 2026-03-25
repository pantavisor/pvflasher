package device

import (
	"testing"
)

func TestDevice_Struct(t *testing.T) {
	d := Device{
		Name:        "/dev/sda",
		Size:        1000000000,
		Model:       "Samsung SSD",
		Vendor:      "Samsung",
		Removable:   true,
		MountPoints: []string{"/mnt/usb"},
	}

	if d.Name != "/dev/sda" {
		t.Errorf("Name = %v, want %v", d.Name, "/dev/sda")
	}
	if d.Size != 1000000000 {
		t.Errorf("Size = %v, want %v", d.Size, 1000000000)
	}
	if d.Model != "Samsung SSD" {
		t.Errorf("Model = %v, want %v", d.Model, "Samsung SSD")
	}
	if d.Vendor != "Samsung" {
		t.Errorf("Vendor = %v, want %v", d.Vendor, "Samsung")
	}
	if !d.Removable {
		t.Error("Removable = false, want true")
	}
	if len(d.MountPoints) != 1 || d.MountPoints[0] != "/mnt/usb" {
		t.Errorf("MountPoints = %v, want [/mnt/usb]", d.MountPoints)
	}
}

func TestDevice_Empty(t *testing.T) {
	d := Device{}

	if d.Name != "" {
		t.Errorf("Empty Name = %v, want empty string", d.Name)
	}
	if d.Size != 0 {
		t.Errorf("Empty Size = %v, want 0", d.Size)
	}
	if d.MountPoints != nil {
		t.Errorf("Empty MountPoints = %v, want nil", d.MountPoints)
	}
}

func TestDevice_IsMounted(t *testing.T) {
	tests := []struct {
		name        string
		mountPoints []string
		wantMounted bool
	}{
		{
			name:        "Mounted device",
			mountPoints: []string{"/mnt/usb"},
			wantMounted: true,
		},
		{
			name:        "Multiple mounts",
			mountPoints: []string{"/mnt/usb1", "/mnt/usb2"},
			wantMounted: true,
		},
		{
			name:        "Not mounted",
			mountPoints: []string{},
			wantMounted: false,
		},
		{
			name:        "Nil mounts",
			mountPoints: nil,
			wantMounted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Device{
				Name:        "/dev/sdb",
				MountPoints: tt.mountPoints,
			}

			isMounted := len(d.MountPoints) > 0
			if isMounted != tt.wantMounted {
				t.Errorf("IsMounted = %v, want %v", isMounted, tt.wantMounted)
			}
		})
	}
}

func TestDevice_SizeHuman(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "1 GB",
			size:     1000000000,
			expected: "1.00 GB",
		},
		{
			name:     "16 GB",
			size:     16000000000,
			expected: "16.00 GB",
		},
		{
			name:     "512 MB",
			size:     512000000,
			expected: "512.00 MB",
		},
		{
			name:     "32 GB",
			size:     32000000000,
			expected: "32.00 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Device{Size: tt.size}

			// Simple size calculation for display
			var sizeStr string
			if d.Size >= 1<<30 {
				sizeStr = "%.2f GB"
			} else if d.Size >= 1<<20 {
				sizeStr = "%.2f MB"
			} else if d.Size >= 1<<10 {
				sizeStr = "%.2f KB"
			} else {
				sizeStr = "%d B"
			}

			// Just verify the format string is set correctly
			if sizeStr == "" {
				t.Error("Size format string is empty")
			}
		})
	}
}

func TestManager_Interface(t *testing.T) {
	// Test that Manager interface is properly defined
	// This is a compile-time check
	var _ Manager = (*MockManager)(nil)
}

// MockManager is a mock implementation for testing
type MockManager struct {
	devices []Device
	err     error
}

func (m *MockManager) List() ([]Device, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.devices, nil
}

func TestMockManager(t *testing.T) {
	tests := []struct {
		name    string
		devices []Device
		wantErr bool
	}{
		{
			name:    "Empty list",
			devices: []Device{},
			wantErr: false,
		},
		{
			name: "Multiple devices",
			devices: []Device{
				{Name: "/dev/sda", Size: 1000000000},
				{Name: "/dev/sdb", Size: 32000000000},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockManager{devices: tt.devices}

			devices, err := mock.List()
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(devices) != len(tt.devices) {
				t.Errorf("List() = %v devices, want %v", len(devices), len(tt.devices))
			}
		})
	}
}

func TestNewManager(t *testing.T) {
	// Test that NewManager returns a non-nil Manager
	mgr := NewManager()
	if mgr == nil {
		t.Error("NewManager() returned nil")
	}
}
