package cards

import (
	"fmt"
	"path/filepath"
	"sort"
	"sync"

	"pvflasher/gui/pantavisor"
	"pvflasher/gui/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// ImageCardCallbacks defines callbacks for image card events
type ImageCardCallbacks struct {
	OnLocalImageSelected func(path string)
	OnPantavisorSelected func(rel *pantavisor.DeviceRelease)
	CheckBmapStatus      func(path string) string
}

// ImageCard represents the image selection card
type ImageCard struct {
	window    fyne.Window
	callbacks ImageCardCallbacks

	// Widgets
	SelectedImageLabel *util.ColoredLabel
	bmapStatusLabel    *util.ColoredLabel
	ChannelSelect      *widget.Select
	VersionSelect      *widget.Select
	DeviceSelect       *widget.Select
	DownloadBtn        *widget.Button
	DownloadStatus     *util.ColoredLabel

	// State
	mu          sync.Mutex
	releaseData pantavisor.Releases
	SelectedRel *pantavisor.DeviceRelease
}

// NewImageCard creates a new image selection card
func NewImageCard(window fyne.Window, callbacks ImageCardCallbacks) *ImageCard {
	return &ImageCard{
		window:    window,
		callbacks: callbacks,
	}
}

// Build constructs and returns the card UI
func (c *ImageCard) Build() fyne.CanvasObject {
	stepLabel := util.StepLabel("STEP 1")
	titleLabel := util.SubHeadingLabel("Select Image")

	// --- Tab 1: Local File ---
	c.SelectedImageLabel = util.NewThemedLabelBold("No image selected")
	c.bmapStatusLabel = util.NewColoredLabel("", util.CurrentSecondaryTextColor())

	selectButton := util.PrimaryButton("📂 Browse Local File", func() {
		fileDialog := dialog.NewFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err == nil && uri != nil {
				path := uri.URI().Path()
				c.SelectedImageLabel.SetText(fmt.Sprintf("📁 %s", filepath.Base(path)))
				if c.callbacks.CheckBmapStatus != nil {
					c.bmapStatusLabel.SetText(c.callbacks.CheckBmapStatus(path))
				}
				if c.callbacks.OnLocalImageSelected != nil {
					c.callbacks.OnLocalImageSelected(path)
				}
				uri.Close()
			}
		}, c.window)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".img", ".iso", ".wic", ".gz", ".bz2", ".xz", ".zst", ".zip", ".tar", ".tgz"}))
		fileDialog.Resize(fyne.NewSize(1200, 700))
		fileDialog.Show()
	})

	localFileContent := container.NewVBox(
		util.SectionSpacer(12),
		util.InstructionLabel("Choose a disk image from your computer:"),
		util.SectionSpacer(8),
		c.SelectedImageLabel,
		c.bmapStatusLabel,
	)

	// --- Tab 2: Pantavisor Download ---
	c.DownloadBtn = util.SuccessButton("✓ Select Image", c.selectPantavisorImage)
	c.DownloadBtn.Disable()
	pvContent := c.createPantavisorTab()

	// Create Tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("📁 Local", localFileContent),
		container.NewTabItem("☁️ Pantavisor", pvContent),
	)

	// Footer for buttons to ensure alignment across cards
	footer := container.NewStack(selectButton)

	tabs.OnSelected = func(t *container.TabItem) {
		if t.Text == "📁 Local File" {
			footer.Objects = []fyne.CanvasObject{selectButton}
		} else {
			footer.Objects = []fyne.CanvasObject{c.DownloadBtn}
		}
		footer.Refresh()
	}

	// Header section with step label and title
	header := container.NewVBox(
		stepLabel,
		util.SectionSpacer(4),
		titleLabel,
		util.SectionSpacer(4),
	)

	// Use border layout - header at top, tabs in center, footer at bottom
	cardContent := container.NewBorder(
		header, // top
		footer, // bottom
		nil,    // left
		nil,    // right
		tabs,   // center
	)

	return util.StyledCardWithBorder(cardContent)
}

func (c *ImageCard) createPantavisorTab() fyne.CanvasObject {
	// Initialize widgets
	c.ChannelSelect = widget.NewSelect([]string{}, c.onChannelSelected)
	c.ChannelSelect.PlaceHolder = "Select Channel"

	c.VersionSelect = widget.NewSelect([]string{}, c.onVersionSelected)
	c.VersionSelect.PlaceHolder = "Select Version"
	c.VersionSelect.Disable()

	c.DeviceSelect = widget.NewSelect([]string{}, c.onDeviceSelected)
	c.DeviceSelect.PlaceHolder = "Select Device"
	c.DeviceSelect.Disable()

	c.DownloadStatus = util.NewColoredLabel("", util.CurrentSecondaryTextColor())
	c.DownloadStatus.Hide()

	// Load data in background
	go c.loadReleases()

	return container.NewVBox(
		util.SectionSpacer(4),
		util.InstructionLabel("Channel:"),
		c.ChannelSelect,
		util.SectionSpacer(4),
		util.InstructionLabel("Version:"),
		c.VersionSelect,
		util.SectionSpacer(4),
		util.InstructionLabel("Target Device:"),
		c.DeviceSelect,
		util.SectionSpacer(8),
		c.DownloadStatus,
	)
}

func (c *ImageCard) selectPantavisorImage() {
	c.mu.Lock()
	rel := c.SelectedRel
	c.mu.Unlock()

	if rel == nil {
		return
	}

	c.SelectedImageLabel.SetText(fmt.Sprintf("☁️ %s (%s)", rel.Name, filepath.Base(rel.FullImage.URL)))
	c.DownloadStatus.SetText("Image will be downloaded when flashing")
	c.DownloadStatus.SetColor(util.CurrentSecondaryTextColor())
	c.DownloadStatus.Show()

	if c.callbacks.OnPantavisorSelected != nil {
		c.callbacks.OnPantavisorSelected(rel)
	}
}

func (c *ImageCard) loadReleases() {
	fyne.Do(func() {
		c.DownloadStatus.SetText("Fetching releases...")
		c.DownloadStatus.Show()
	})

	releases, err := pantavisor.FetchReleases()
	if err != nil {
		fyne.Do(func() {
			c.DownloadStatus.SetText("Error: " + err.Error())
			c.DownloadStatus.SetColor(util.ColorError)
		})
		return
	}

	c.mu.Lock()
	c.releaseData = releases
	c.mu.Unlock()

	channels := releases.GetChannels()
	fyne.Do(func() {
		c.ChannelSelect.Options = channels
		c.ChannelSelect.Refresh()
		c.DownloadStatus.Hide()
	})
}

func (c *ImageCard) onChannelSelected(channel string) {
	c.mu.Lock()
	versions := c.releaseData.GetVersions(channel)
	c.mu.Unlock()

	if len(versions) > 0 {
		c.VersionSelect.Options = versions
		c.VersionSelect.Enable()
	} else {
		c.VersionSelect.Options = []string{}
		c.VersionSelect.Disable()
	}
	c.VersionSelect.ClearSelected()
	c.VersionSelect.Refresh()

	c.DeviceSelect.Options = []string{}
	c.DeviceSelect.Disable()
	c.DeviceSelect.ClearSelected()
	c.DeviceSelect.Refresh()

	c.DownloadBtn.Disable()
}

func (c *ImageCard) onVersionSelected(version string) {
	channel := c.ChannelSelect.Selected
	if channel == "" || version == "" {
		return
	}

	c.mu.Lock()
	wrapper, ok := c.releaseData[channel][version]
	c.mu.Unlock()

	if !ok {
		return
	}

	devices := make([]string, 0)
	for _, d := range wrapper.Devices {
		devices = append(devices, d.Name)
	}
	sort.Strings(devices)

	c.DeviceSelect.Options = devices
	c.DeviceSelect.Enable()
	c.DeviceSelect.Refresh()

	c.DownloadBtn.Disable()
}

func (c *ImageCard) onDeviceSelected(deviceName string) {
	channel := c.ChannelSelect.Selected
	version := c.VersionSelect.Selected

	if channel == "" || version == "" || deviceName == "" {
		return
	}

	c.mu.Lock()
	wrapper := c.releaseData[channel][version]
	var selectedDevice *pantavisor.DeviceRelease
	for i := range wrapper.Devices {
		if wrapper.Devices[i].Name == deviceName {
			selectedDevice = &wrapper.Devices[i]
			break
		}
	}
	c.SelectedRel = selectedDevice
	c.mu.Unlock()

	if selectedDevice != nil {
		c.DownloadBtn.Enable()
	}
}

// Reset clears the card state
func (c *ImageCard) Reset() {
	c.SelectedImageLabel.SetText("No image selected")
	c.bmapStatusLabel.SetText("")

	c.mu.Lock()
	c.SelectedRel = nil
	c.mu.Unlock()

	if c.ChannelSelect != nil {
		c.ChannelSelect.ClearSelected()
	}
	if c.VersionSelect != nil {
		c.VersionSelect.ClearSelected()
		c.VersionSelect.Disable()
	}
	if c.DeviceSelect != nil {
		c.DeviceSelect.ClearSelected()
		c.DeviceSelect.Disable()
	}
	if c.DownloadBtn != nil {
		c.DownloadBtn.Disable()
	}
	if c.DownloadStatus != nil {
		c.DownloadStatus.Hide()
	}
}
