package device

// Device represents a storage device
type Device struct {
	Name        string   `json:"name"`        // e.g. /dev/sda, PhysicalDrive1
	Size        int64    `json:"size"`        // Size in bytes
	Model       string   `json:"model"`       // Device model
	Vendor      string   `json:"vendor"`      // Device vendor
	Removable   bool     `json:"removable"`   // Is removable
	MountPoints []string `json:"mountPoints"` // List of mount points
}

// Manager defines the interface for device enumeration
type Manager interface {
	List() ([]Device, error)
}
