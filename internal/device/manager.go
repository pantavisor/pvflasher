package device

// NewManager creates a new device manager for the current platform
func NewManager() Manager {
	return newPlatformManager()
}