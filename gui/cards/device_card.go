package cards

import (
	"fmt"
	"strings"

	"pvflasher/gui/util"
	"pvflasher/internal/device"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// DeviceCardCallbacks defines callbacks for device card events
type DeviceCardCallbacks struct {
	OnDeviceSelected func(devicePath string)
	OnDeviceCleared  func()
}

// DeviceCard represents the device selection card
type DeviceCard struct {
	window    fyne.Window
	callbacks DeviceCardCallbacks

	// Widgets
	SelectedDeviceLabel *util.ColoredLabel
	DeviceListSelect    *widget.Select
}

// NewDeviceCard creates a new device selection card
func NewDeviceCard(window fyne.Window, callbacks DeviceCardCallbacks) *DeviceCard {
	return &DeviceCard{
		window:    window,
		callbacks: callbacks,
	}
}

// Build constructs and returns the card UI
func (c *DeviceCard) Build() fyne.CanvasObject {
	stepLabel := util.StepLabel("STEP 2")
	titleLabel := util.SubHeadingLabel("Select Target Device")

	c.SelectedDeviceLabel = util.NewThemedLabelBold("No device selected")

	c.DeviceListSelect = widget.NewSelect([]string{}, func(s string) {
		parts := strings.Split(s, " ")
		if len(parts) > 0 {
			devicePath := parts[0]
			c.SelectedDeviceLabel.SetText(devicePath)

			if c.callbacks.OnDeviceSelected != nil {
				c.callbacks.OnDeviceSelected(devicePath)
			}
		}
	})

	refreshButton := util.PrimaryButton("🔄 Refresh", func() {
		c.RefreshDeviceList()
	})

	c.RefreshDeviceList()

	header := container.NewVBox(
		stepLabel,
		util.SectionSpacer(4),
		titleLabel,
		util.SectionSpacer(4),
	)

	warningBox := container.NewVBox(
		util.ErrorBox("🚨 DESTRUCTIVE ACTION", "This operation will erase the selected device. Make sure you have selected the correct target!"),
	)

	contentBox := container.NewVBox(
		warningBox,
		util.SectionSpacer(0),
		util.InstructionLabel("Selected Device:"),
		util.SectionSpacer(4),
		c.SelectedDeviceLabel,
		util.SectionSpacer(4),
		util.InstructionLabel("Available Devices:"),
		util.SectionSpacer(4),
		c.DeviceListSelect,
	)

	// Use border to place button at bottom with full width
	cardContent := container.NewBorder(
		header,        // top
		refreshButton, // bottom (button with full width)
		nil,           // left
		nil,           // right
		contentBox,    // center
	)

	return util.StyledCardWithBorder(cardContent)
}

// RefreshDeviceList refreshes the device list
func (c *DeviceCard) RefreshDeviceList() {
	mgr := device.NewManager()
	devices, err := mgr.List()
	if err != nil {
		c.DeviceListSelect.Options = []string{"Error: " + err.Error()}
		return
	}

	options := []string{}
	for _, d := range devices {
		// Skip system drives entirely - they should never be flashing targets
		if c.isSystemDrive(d.MountPoints) {
			continue
		}

		warning := ""
		if len(d.MountPoints) > 0 {
			warning = " ⚠️ MOUNTED"
		}

		sizeStr := fmt.Sprintf("%.0f GB", float64(d.Size)/1e9)
		options = append(options, fmt.Sprintf("%s (%s - %s)%s", d.Name, d.Vendor, sizeStr, warning))
	}
	c.DeviceListSelect.Options = options
	c.DeviceListSelect.PlaceHolder = "(Select one)"
}

// isSystemDrive checks if any mount point indicates a system/boot drive
func (c *DeviceCard) isSystemDrive(mountPoints []string) bool {
	systemMounts := []string{"/", "/boot", "/boot/efi", "/home", "/usr", "/var", "/etc"}
	for _, mp := range mountPoints {
		for _, sm := range systemMounts {
			if mp == sm {
				return true
			}
		}
		// Also check for Windows system drives
		if len(mp) >= 2 && mp[1] == ':' {
			drive := strings.ToUpper(string(mp[0]))
			if drive == "C" {
				return true
			}
		}
	}
	return false
}

// Reset clears the card state
func (c *DeviceCard) Reset() {
	c.SelectedDeviceLabel.SetText("No device selected")
	c.DeviceListSelect.ClearSelected()
}
