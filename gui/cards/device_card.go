package cards

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"pvflasher/gui/util"
	"pvflasher/internal/device"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// devicePollInterval is how often the device list is re-scanned so that
// inserted or removed cards show up without a manual refresh.
const devicePollInterval = 3 * time.Second

// DeviceCardCallbacks defines callbacks for device card events
type DeviceCardCallbacks struct {
	OnDeviceSelected func(devicePath string)
	OnDeviceCleared  func()
}

// DeviceCard represents the device selection card
type DeviceCard struct {
	window    fyne.Window
	callbacks DeviceCardCallbacks

	tile          *util.SelectionTile
	deviceSelect  *widget.Select
	mountedNotice fyne.CanvasObject
	mountedLabel  *widget.Label
	hiddenLink    *widget.Hyperlink

	mu       sync.Mutex
	devices  map[string]device.Device // option label -> device
	hidden   []device.Device          // drives never offered as targets
	lastScan string
	polling  bool
	stopPoll chan struct{}
}

// NewDeviceCard creates a new device selection card
func NewDeviceCard(window fyne.Window, callbacks DeviceCardCallbacks) *DeviceCard {
	return &DeviceCard{
		window:    window,
		callbacks: callbacks,
		devices:   map[string]device.Device{},
	}
}

// Build constructs and returns the card UI
func (c *DeviceCard) Build() fyne.CanvasObject {
	c.tile = util.NewSelectionTile(theme.StorageIcon(), "No target selected", "Insert an SD card or USB drive")

	c.deviceSelect = widget.NewSelect(nil, c.onSelected)
	c.deviceSelect.PlaceHolder = "Select target…"

	refresh := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() { go c.scan() })
	refresh.Importance = widget.LowImportance

	c.mountedLabel, c.mountedNotice = util.Notice(theme.NewWarningThemedResource(theme.WarningIcon()), "")
	c.mountedNotice.Hide()

	_, eraseNotice := util.Notice(theme.NewWarningThemedResource(theme.WarningIcon()),
		"Everything on the target will be erased.")

	c.hiddenLink = widget.NewHyperlink("", nil)
	c.hiddenLink.SizeName = theme.SizeNameCaptionText
	c.hiddenLink.Alignment = fyne.TextAlignCenter
	c.hiddenLink.OnTapped = c.showHiddenDrives
	c.hiddenLink.Hide()

	go c.scan()

	return util.NewSurface(container.NewBorder(
		util.StepHeader("2", "Target"),
		container.NewVBox(eraseNotice, container.NewBorder(nil, nil, nil, refresh, c.deviceSelect), c.hiddenLink),
		nil, nil,
		container.NewVBox(layout.NewSpacer(), c.tile.Object, c.mountedNotice, layout.NewSpacer()),
	))
}

// DeviceTitle returns a human-readable name for a device.
func DeviceTitle(d device.Device) string {
	var parts []string
	for _, p := range []string{d.Vendor, d.Model} {
		p = strings.TrimSpace(strings.ReplaceAll(p, "_", " "))
		if p != "" && !strings.EqualFold(p, "unknown") && !slices.Contains(parts, p) {
			parts = append(parts, p)
		}
	}
	name := strings.Join(parts, " ")
	if name == "" {
		name = "Removable drive"
	}
	return name
}

func deviceOption(d device.Device) string {
	label := fmt.Sprintf("%s — %s (%s)", d.Name, DeviceTitle(d), util.FormatBytes(d.Size))
	if len(d.MountPoints) > 0 {
		label += " · mounted"
	}
	return label
}

// scan re-enumerates devices and updates the list if anything changed.
func (c *DeviceCard) scan() {
	devices, err := device.NewManager().List()

	var keys []string
	var hidden []device.Device
	found := map[string]device.Device{}
	if err == nil {
		for _, d := range devices {
			// Internal disks, drives in use by the system and empty card
			// readers are never offered as flashing targets.
			if d.UnsafeReason() != "" {
				hidden = append(hidden, d)
				continue
			}
			label := deviceOption(d)
			keys = append(keys, label)
			found[label] = d
		}
	}
	var signature string
	for _, d := range hidden {
		signature += "hidden:" + deviceOption(d) + "|" + d.UnsafeReason() + "\n"
	}
	for _, k := range keys {
		signature += k + "|" + strings.Join(found[k].MountPoints, ",") + "\n"
	}

	c.mu.Lock()
	if signature == c.lastScan && err == nil {
		c.mu.Unlock()
		return
	}
	c.lastScan = signature
	c.devices = found
	c.hidden = hidden
	c.mu.Unlock()

	fyne.DoAndWait(func() {
		if err != nil {
			c.deviceSelect.PlaceHolder = "Could not list devices"
			c.deviceSelect.SetOptions(nil)
			return
		}
		if len(keys) == 0 {
			c.deviceSelect.PlaceHolder = "No removable drives found"
		} else {
			c.deviceSelect.PlaceHolder = "Select target…"
		}
		switch len(hidden) {
		case 0:
		case 1:
			c.hiddenLink.SetText("Why is 1 drive hidden?")
		default:
			c.hiddenLink.SetText(fmt.Sprintf("Why are %d drives hidden?", len(hidden)))
		}
		c.hiddenLink.Hidden = len(hidden) == 0
		c.hiddenLink.Refresh()
		selected := c.deviceSelect.Selected
		c.deviceSelect.SetOptions(keys)
		if _, ok := found[selected]; selected != "" && !ok {
			// The selected drive was removed.
			c.Reset()
			if c.callbacks.OnDeviceCleared != nil {
				c.callbacks.OnDeviceCleared()
			}
		} else if ok {
			c.showDevice(found[selected])
		}
	})
}

// showHiddenDrives lists the drives filtered out of the target list and why.
func (c *DeviceCard) showHiddenDrives() {
	c.mu.Lock()
	hidden := slices.Clone(c.hidden)
	c.mu.Unlock()

	intro := widget.NewLabel("To protect your system, these drives are never offered as flashing targets.")
	intro.Wrapping = fyne.TextWrapWord

	list := container.NewVBox()
	for _, d := range hidden {
		name := widget.NewLabelWithStyle(fmt.Sprintf("%s — %s (%s)", d.Name, DeviceTitle(d), util.FormatBytes(d.Size)),
			fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		name.Truncation = fyne.TextTruncateEllipsis
		reason := widget.NewLabel(capitalize(d.UnsafeReason()))
		reason.Importance = widget.LowImportance
		reason.SizeName = theme.SizeNameCaptionText
		reason.Wrapping = fyne.TextWrapWord
		list.Add(container.NewBorder(nil, nil,
			container.NewVBox(container.NewGridWrap(fyne.NewSize(24, 24), widget.NewIcon(theme.StorageIcon()))),
			nil, container.NewVBox(name, reason)))
	}

	d := dialog.NewCustom("Hidden Drives", "Close", container.NewVBox(intro, widget.NewSeparator(), list), c.window)
	d.Resize(fyne.NewSize(520, 0))
	d.Show()
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// StartPolling re-scans devices periodically until StopPolling is called.
func (c *DeviceCard) StartPolling() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.polling {
		return
	}
	c.polling = true
	c.stopPoll = make(chan struct{})
	stop := c.stopPoll
	go func() {
		t := time.NewTicker(devicePollInterval)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				c.scan()
			}
		}
	}()
}

// StopPolling stops the background device scan.
func (c *DeviceCard) StopPolling() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.polling {
		close(c.stopPoll)
		c.polling = false
	}
}

func (c *DeviceCard) onSelected(label string) {
	c.mu.Lock()
	d, ok := c.devices[label]
	c.mu.Unlock()
	if !ok {
		return
	}
	c.showDevice(d)
	if c.callbacks.OnDeviceSelected != nil {
		c.callbacks.OnDeviceSelected(d.Name)
	}
}

func (c *DeviceCard) showDevice(d device.Device) {
	c.tile.Set(theme.StorageIcon(), DeviceTitle(d), d.Name+" · "+util.FormatBytes(d.Size))
	if len(d.MountPoints) > 0 {
		c.mountedLabel.SetText("Mounted — will be unmounted before writing")
		c.mountedNotice.Show()
	} else {
		c.mountedNotice.Hide()
	}
}

// SelectedDevice returns the currently selected device, if any.
func (c *DeviceCard) SelectedDevice() (device.Device, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.devices[c.deviceSelect.Selected]
	return d, ok
}

// Reset clears the card state
func (c *DeviceCard) Reset() {
	c.tile.Set(theme.StorageIcon(), "No target selected", "Insert an SD card or USB drive")
	c.mountedNotice.Hide()
	c.deviceSelect.ClearSelected()
}
