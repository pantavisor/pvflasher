package cards

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"pvflasher/gui/pantavisor"
	"pvflasher/gui/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ImageExtensions lists the image formats pvflasher can write.
var ImageExtensions = []string{".img", ".iso", ".wic", ".gz", ".bz2", ".xz", ".zst", ".zip", ".tar", ".tgz"}

// IsImageFile reports whether path has a supported image extension.
func IsImageFile(path string) bool {
	lower := strings.ToLower(path)
	for _, ext := range ImageExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

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

	tile       *util.SelectionTile
	bmapStatus *widget.Label

	// Release data, loaded in the background
	mu          sync.Mutex
	releaseData pantavisor.Releases
	releaseErr  error
	loaded      bool
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
	c.tile = util.NewSelectionTile(theme.FileIcon(), "No image selected", "Choose a file or a Pantavisor release")

	c.bmapStatus = widget.NewLabel("")
	c.bmapStatus.Alignment = fyne.TextAlignCenter
	c.bmapStatus.Importance = widget.LowImportance
	c.bmapStatus.SizeName = theme.SizeNameCaptionText
	c.bmapStatus.Truncation = fyne.TextTruncateEllipsis

	fileBtn := widget.NewButtonWithIcon("Choose File…", theme.FolderOpenIcon(), c.showFileDialog)
	pvBtn := widget.NewButtonWithIcon("Pantavisor Release…", theme.DownloadIcon(), c.showPantavisorDialog)

	go c.loadReleases()

	return util.NewSurface(container.NewBorder(
		util.StepHeader("1", "Image"),
		container.NewVBox(fileBtn, pvBtn),
		nil, nil,
		container.NewVBox(layout.NewSpacer(), c.tile.Object, c.bmapStatus, layout.NewSpacer()),
	))
}

func (c *ImageCard) showFileDialog() {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		go func() {
			path, ok, _ := pickImageNative()
			fyne.Do(func() {
				switch {
				case !ok:
					c.showFyneFileDialog() // native picker unavailable
				case path != "":
					c.SetLocalImage(path)
				}
			})
		}()
		return
	}
	c.showFyneFileDialog()
}

// showFyneFileDialog is Fyne's file dialog, which uses the desktop portal on
// Linux and is the fallback elsewhere.
func (c *ImageCard) showFyneFileDialog() {
	d := dialog.NewFileOpen(func(uri fyne.URIReadCloser, err error) {
		if err != nil || uri == nil {
			return
		}
		path := uri.URI().Path()
		uri.Close()
		c.SetLocalImage(path)
	}, c.window)
	d.SetFilter(storage.NewExtensionFileFilter(ImageExtensions))
	if home, err := os.UserHomeDir(); err == nil {
		if lister, err := storage.ListerForURI(storage.NewFileURI(home)); err == nil {
			d.SetLocation(lister)
		}
	}
	d.Resize(fyne.NewSize(760, 520))
	d.Show()
}

// SetLocalImage selects a local file, e.g. from the file dialog or drag-and-drop.
func (c *ImageCard) SetLocalImage(path string) {
	detail := "Local file"
	if info, err := os.Stat(path); err == nil {
		detail = util.FormatBytes(info.Size()) + " · " + filepath.Dir(path)
	}
	c.tile.Set(theme.FileImageIcon(), filepath.Base(path), detail)
	c.bmapStatus.SetText("")
	if c.callbacks.CheckBmapStatus != nil {
		c.bmapStatus.SetText(c.callbacks.CheckBmapStatus(path))
	}
	if c.callbacks.OnLocalImageSelected != nil {
		c.callbacks.OnLocalImageSelected(path)
	}
}

func (c *ImageCard) setRelease(channel, version string, rel *pantavisor.DeviceRelease) {
	c.tile.Set(theme.DownloadIcon(), rel.Title(), fmt.Sprintf("Pantavisor %s · %s", version, pantavisor.ChannelLabel(channel)))
	c.bmapStatus.SetText("Downloaded when flashing starts")
	if c.callbacks.OnPantavisorSelected != nil {
		c.callbacks.OnPantavisorSelected(rel)
	}
}

func (c *ImageCard) loadReleases() {
	releases, err := pantavisor.FetchReleases()
	c.mu.Lock()
	c.releaseData = releases
	c.releaseErr = err
	c.loaded = true
	c.mu.Unlock()
}

// showPantavisorDialog lets the user pick a channel, version and board.
func (c *ImageCard) showPantavisorDialog() {
	c.mu.Lock()
	releases, loadErr, loaded := c.releaseData, c.releaseErr, c.loaded
	c.mu.Unlock()

	if !loaded {
		dialog.ShowInformation("Pantavisor Releases", "The release list is still loading. Try again in a moment.", c.window)
		go c.loadReleases()
		return
	}
	if loadErr != nil {
		dialog.ShowError(fmt.Errorf("could not load Pantavisor releases: %w", loadErr), c.window)
		go c.loadReleases()
		return
	}

	var selected *pantavisor.DeviceRelease
	var confirm *widget.Button

	// Channels and boards are shown with their pantavisor.io/downloads names.
	channelKeys := releases.GetChannels()
	channelByLabel := map[string]string{}
	var channelLabels []string
	for _, ch := range channelKeys {
		l := pantavisor.ChannelLabel(ch)
		channelByLabel[l] = ch
		channelLabels = append(channelLabels, l)
	}
	boardByTitle := map[string]*pantavisor.DeviceRelease{}

	info := widget.NewLabel("")
	info.Wrapping = fyne.TextWrapWord
	info.Importance = widget.LowImportance
	machine := widget.NewLabel("")
	machine.TextStyle = fyne.TextStyle{Monospace: true}
	machine.Truncation = fyne.TextTruncateEllipsis
	details := container.NewVBox(machine, info)
	details.Hide()

	version := widget.NewSelect(nil, nil)
	version.PlaceHolder = "Select a version"
	version.Disable()
	board := widget.NewSelect(nil, nil)
	board.PlaceHolder = "Select a device"
	board.Disable()
	channel := widget.NewSelect(channelLabels, nil)
	channel.PlaceHolder = "Select a channel"

	clearBoard := func() {
		selected = nil
		confirm.Disable()
		details.Hide()
	}
	channel.OnChanged = func(label string) {
		version.SetOptions(releases.GetVersions(channelByLabel[label]))
		version.ClearSelected()
		version.Enable()
		board.SetOptions(nil)
		board.ClearSelected()
		board.Disable()
		clearBoard()
	}
	version.OnChanged = func(v string) {
		if v == "" {
			return
		}
		devices := releases[channelByLabel[channel.Selected]][v].Devices
		boardByTitle = map[string]*pantavisor.DeviceRelease{}
		titles := []string{}
		for i := range devices {
			t := devices[i].Title()
			if _, dup := boardByTitle[t]; dup {
				t += " (" + devices[i].Name + ")"
			}
			boardByTitle[t] = &devices[i]
			titles = append(titles, t)
		}
		sort.Strings(titles)
		board.SetOptions(titles)
		board.ClearSelected()
		board.Enable()
		clearBoard()
	}
	board.OnChanged = func(title string) {
		clearBoard()
		rel, ok := boardByTitle[title]
		if !ok {
			return
		}
		selected = rel
		machine.SetText(filepath.Base(rel.FullImage.URL))
		info.SetText(rel.Description)
		info.Hidden = rel.Description == ""
		details.Show()
		confirm.Enable()
	}

	intro := widget.NewLabel("Official images from pantavisor.io/downloads. The image is downloaded and checksum-verified when flashing starts.")
	intro.Wrapping = fyne.TextWrapWord

	form := widget.NewForm(
		widget.NewFormItem("Channel", channel),
		widget.NewFormItem("Version", version),
		widget.NewFormItem("Device", board),
	)

	d := dialog.NewCustomWithoutButtons("Pantavisor Release", container.NewVBox(intro, form, details), c.window)
	confirm = widget.NewButton("Use This Image", func() {
		if selected != nil {
			c.setRelease(channelByLabel[channel.Selected], version.Selected, selected)
		}
		d.Hide()
	})
	confirm.Importance = widget.HighImportance
	confirm.Disable()
	d.SetButtons([]fyne.CanvasObject{widget.NewButton("Cancel", d.Hide), confirm})
	d.Resize(fyne.NewSize(560, 0))
	d.Show()

	// Preselect the recommended channel and its newest version.
	if len(channelLabels) > 0 {
		channel.SetSelected(channelLabels[0])
		if len(version.Options) > 0 {
			version.SetSelected(version.Options[0])
		}
	}
}

// Reset clears the card state
func (c *ImageCard) Reset() {
	c.tile.Set(theme.FileIcon(), "No image selected", "Choose a file or a Pantavisor release")
	c.bmapStatus.SetText("")
}
