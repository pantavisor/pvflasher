package gui

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"pvflasher/assets"
	"pvflasher/gui/cards"
	"pvflasher/gui/pantavisor"
	"pvflasher/gui/screens"
	"pvflasher/gui/util"
	"pvflasher/internal/update"
	"pvflasher/internal/version"
	"pvflasher/pkg/flash"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// App represents the Fyne GUI application
type App struct {
	// Fyne components
	fyneApp fyne.App
	window  fyne.Window

	// State
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	cmd      *exec.Cmd
	lastLogs []string

	// User selections
	selectedImage  string
	selectedDevice string
	bmapPath       string
	forceChecked   bool
	verifyChecked  bool
	ejectChecked   bool

	// Screen state
	mainContent     fyne.CanvasObject
	progressContent fyne.CanvasObject
	successContent  fyne.CanvasObject
	errorContent    fyne.CanvasObject

	// Progress state for thread-safe updates
	progressChan    chan flash.Progress
	progressMu      sync.RWMutex
	lastProgress    flash.Progress
	progressChanged bool

	// Result storage
	lastResult *flash.FlashResult
	lastError  string

	// Card components
	imageCard   *cards.ImageCard
	deviceCard  *cards.DeviceCard
	optionsCard *cards.OptionsCard

	// Screen components
	progressScreen *screens.ProgressScreen
	successScreen  *screens.SuccessScreen
	errorScreen    *screens.ErrorScreen

	// Pantavisor state
	selectedRel *pantavisor.DeviceRelease

	// Window chrome
	body           *fyne.Container
	logo           *canvas.Image
	headerTitle    *canvas.Text
	headerSubtitle *canvas.Text
	settingsBtn    *widget.Button

	// Theme preference: "system", "light" or "dark"
	themePref    string
	themeApplied bool

	// Self-update
	pendingUpdate *update.Release
	updateLink    *widget.Hyperlink

	// Flash lifecycle
	flashing         bool
	cancelled        bool
	unmountConfirmed bool
}

// NewApp creates a new Fyne app instance
func NewApp() *App {
	return &App{
		lastLogs:      []string{},
		verifyChecked: true,
		ejectChecked:  true,
	}
}

// Run starts the Fyne application
func (a *App) Run() {
	a.fyneApp = app.NewWithID(AppID)
	a.fyneApp.SetIcon(fyne.NewStaticResource("icon.png", assets.AppIconPNG))

	config, err := util.LoadConfig()
	if err != nil {
		config = util.DefaultConfig()
	}
	a.themePref = config.Theme
	a.applyTheme()
	// Follow the desktop's light/dark switch while the preference is "system".
	a.fyneApp.Settings().AddListener(func(fyne.Settings) { a.applyTheme() })

	a.window = a.fyneApp.NewWindow("PvFlasher")
	a.window.SetMaster()

	a.buildMainView()
	a.buildProgressScreen()
	a.buildSuccessScreen()
	a.buildErrorScreen()

	a.body = container.NewStack()
	a.window.SetContent(container.NewBorder(a.buildHeader(), nil, nil, nil, a.body))
	a.showMainView()

	a.window.SetOnDropped(a.onDropped)
	a.window.SetCloseIntercept(a.onCloseRequested)

	a.window.Resize(fyne.NewSize(920, 600))
	a.window.CenterOnScreen()
	a.maybeCheckForUpdates()
	a.window.ShowAndRun()
}

// AppID is the reverse-DNS application identifier. It becomes the Wayland
// app_id / X11 WM_CLASS, which desktops use to match the window to its
// .desktop launcher, icon and taskbar entry.
const AppID = "com.pantacor.pvflasher"

// applyTheme resolves the light/dark preference and applies it if it changed.
func (a *App) applyTheme() {
	mode := util.ThemeModeLight
	switch a.themePref {
	case "dark":
		mode = util.ThemeModeDark
	case "light":
	default:
		if a.fyneApp.Settings().ThemeVariant() == theme.VariantDark {
			mode = util.ThemeModeDark
		}
	}
	t := util.GetTheme()
	if a.themeApplied && t.Mode() == mode {
		return
	}
	a.themeApplied = true
	t.SetMode(mode)
	a.fyneApp.Settings().SetTheme(t)
	if a.logo != nil {
		a.logo.Resource = pantacorLogoResource()
		a.logo.Refresh()
	}
	if a.headerTitle != nil {
		a.headerTitle.Color = util.CurrentTextColor()
		a.headerTitle.Refresh()
		a.headerSubtitle.Color = util.CurrentSecondaryTextColor()
		a.headerSubtitle.Refresh()
	}
}

// setScreen swaps the window body below the header.
func (a *App) setScreen(content fyne.CanvasObject) {
	a.body.Objects = []fyne.CanvasObject{content}
	a.body.Refresh()
}

// buildHeader builds the app bar shown above every screen.
func (a *App) buildHeader() fyne.CanvasObject {
	_, icon := util.BigIcon(fyne.NewStaticResource("icon.png", assets.AppIconPNG), 32)

	// canvas.Text avoids the label padding so the two lines sit tight together;
	// applyTheme keeps their colours in sync with the theme.
	a.headerTitle = canvas.NewText("PvFlasher", util.CurrentTextColor())
	a.headerTitle.TextSize = 17
	a.headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	a.headerSubtitle = canvas.NewText("Write OS images to SD cards and USB drives", util.CurrentSecondaryTextColor())
	a.headerSubtitle.TextSize = 12
	text := container.New(layout.NewCustomPaddedVBoxLayout(2), a.headerTitle, a.headerSubtitle)

	a.settingsBtn = widget.NewButtonWithIcon("", theme.SettingsIcon(), a.showSettingsDialog)
	a.settingsBtn.Importance = widget.LowImportance

	bar := container.NewBorder(nil, nil,
		container.NewHBox(container.NewCenter(icon), util.HorizontalSpacer(4), container.NewCenter(text)),
		container.NewCenter(a.settingsBtn),
	)
	return container.NewVBox(util.Inset(8, bar), widget.NewSeparator())
}

// buildMainView builds the main selection view
func (a *App) buildMainView() {
	a.imageCard = cards.NewImageCard(a.window, cards.ImageCardCallbacks{
		OnLocalImageSelected: func(path string) {
			a.mu.Lock()
			a.selectedImage = path
			a.selectedRel = nil // Clear any Pantavisor selection
			a.mu.Unlock()
			a.updateFlashButtonState()
		},
		OnPantavisorSelected: func(rel *pantavisor.DeviceRelease) {
			a.mu.Lock()
			a.selectedRel = rel
			a.selectedImage = "" // Clear local image
			a.bmapPath = ""
			a.mu.Unlock()
			a.updateFlashButtonState()
		},
		CheckBmapStatus: a.checkBmapStatus,
	})

	a.deviceCard = cards.NewDeviceCard(a.window, cards.DeviceCardCallbacks{
		OnDeviceSelected: func(devicePath string) {
			a.SetSelectedDevice(devicePath)
			a.updateFlashButtonState()
		},
		OnDeviceCleared: func() {
			a.SetSelectedDevice("")
			a.updateFlashButtonState()
		},
	})

	a.optionsCard = cards.NewOptionsCard(cards.OptionsCardCallbacks{
		OnForceChanged:  func(b bool) { a.SetForceChecked(b) },
		OnVerifyChanged: func(b bool) { a.SetVerifyChecked(b) },
		OnEjectChanged:  func(b bool) { a.SetEjectChecked(b) },
		OnStartFlash:    func() { a.startFlash() },
	})

	steps := container.NewGridWithColumns(3,
		a.imageCard.Build(),
		a.deviceCard.Build(),
		a.optionsCard.Build(),
	)

	a.mainContent = container.NewBorder(nil, a.buildFooter(), nil, nil, util.Inset(16, steps))
}

// buildFooter shows branding, links and the version.
func (a *App) buildFooter() fyne.CanvasObject {
	a.logo = canvas.NewImageFromResource(pantacorLogoResource())
	a.logo.FillMode = canvas.ImageFillContain
	a.logo.SetMinSize(fyne.NewSize(96, 20))

	// The version opens Settings, which shows the latest version too.
	ver := widget.NewHyperlink(versionText(), nil)
	ver.OnTapped = a.showSettingsDialog
	ver.SizeName = theme.SizeNameCaptionText

	// Shown once an update is found; opens the update dialog.
	a.updateLink = widget.NewHyperlink("", nil)
	a.updateLink.OnTapped = a.showUpdateDialog
	a.updateLink.TextStyle = fyne.TextStyle{Bold: true}
	a.updateLink.Hide()

	links := container.NewHBox(
		a.updateLink,
		widget.NewHyperlink("pantacor.com", mustParseURL("https://pantacor.com/")),
		widget.NewHyperlink("pantavisor.io", mustParseURL("https://pantavisor.io/")),
		ver,
	)
	return container.NewVBox(
		widget.NewSeparator(),
		container.NewBorder(nil, nil, util.Inset(8, a.logo), links),
	)
}

// versionText formats the build version for display.
func versionText() string {
	v := strings.TrimPrefix(version.Version, "v")
	if v == "" || v[0] < '0' || v[0] > '9' {
		return "development build"
	}
	return "v" + v
}

func mustParseURL(raw string) *url.URL {
	parsed, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return parsed
}

func pantacorLogoResource() fyne.Resource {
	if util.GetTheme().IsDark() {
		return fyne.NewStaticResource("pantacor-logo-dark.svg", assets.PantacorLogoDarkSVG)
	}
	return fyne.NewStaticResource("pantacor-logo.svg", assets.PantacorLogoSVG)
}

// onDropped accepts an image file dragged onto the window.
func (a *App) onDropped(_ fyne.Position, uris []fyne.URI) {
	if a.isFlashing() || len(uris) == 0 {
		return
	}
	path := uris[0].Path()
	if !cards.IsImageFile(path) {
		dialog.ShowInformation("Unsupported File",
			filepath.Base(path)+" is not a supported image.\n\nSupported: "+strings.Join(cards.ImageExtensions, " "),
			a.window)
		return
	}
	a.imageCard.SetLocalImage(path)
}

// onCloseRequested asks before quitting while a flash is running.
func (a *App) onCloseRequested() {
	if !a.isFlashing() {
		a.window.Close()
		return
	}
	d := dialog.NewConfirm("Quit While Flashing?",
		"Flashing is still in progress. Quitting now leaves the target unusable until it is flashed again.",
		func(ok bool) {
			if ok {
				a.CancelFlash()
				a.window.Close()
			}
		}, a.window)
	d.SetConfirmText("Quit")
	d.SetDismissText("Keep Flashing")
	d.SetConfirmImportance(widget.DangerImportance)
	d.Show()
}

func (a *App) isFlashing() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.flashing
}

// buildProgressScreen builds the progress display screen
func (a *App) buildProgressScreen() {
	a.progressScreen = screens.NewProgressScreen(screens.ProgressScreenCallbacks{
		OnCancel: a.confirmCancel,
	})
	a.progressContent = a.progressScreen.Build()

	// Initialize progress channel
	a.progressChan = make(chan flash.Progress, 10)
}

// buildSuccessScreen builds the success display screen
func (a *App) buildSuccessScreen() {
	a.successScreen = screens.NewSuccessScreen(screens.SuccessScreenCallbacks{
		OnFlashAnother: a.flashAnother,
		OnViewLogs:     a.ShowLogsDialog,
	})
	a.successContent = a.successScreen.Build()
}

// buildErrorScreen builds the error display screen
func (a *App) buildErrorScreen() {
	a.errorScreen = screens.NewErrorScreen(screens.ErrorScreenCallbacks{
		OnTryAgain: a.showMainView,
		OnViewLogs: a.ShowLogsDialog,
	})
	a.errorContent = a.errorScreen.Build()
}

// updateFlashButtonState enables/disables flash button based on selections
func (a *App) updateFlashButtonState() {
	a.mu.Lock()
	hasImage := a.selectedImage != "" || a.selectedRel != nil
	hasDevice := a.selectedDevice != ""
	a.mu.Unlock()

	if a.optionsCard != nil {
		a.optionsCard.SetFlashEnabled(hasImage && hasDevice)
	}
}

// showMainView returns to the selection view, keeping the current choices.
func (a *App) showMainView() {
	a.settingsBtn.Enable()
	a.setScreen(a.mainContent)
	a.deviceCard.StartPolling()
}

// flashAnother keeps the image but clears the (possibly ejected) target.
func (a *App) flashAnother() {
	a.SetSelectedDevice("")
	a.deviceCard.Reset()
	a.updateFlashButtonState()
	a.showMainView()
}

// showProgressScreen shows the progress view
func (a *App) showProgressScreen() {
	// Reinitialize progress channel
	a.mu.Lock()
	if a.progressChan != nil {
		close(a.progressChan)
	}
	a.progressChan = make(chan flash.Progress, 10)
	a.mu.Unlock()

	a.deviceCard.StopPolling()
	a.settingsBtn.Disable()
	a.setScreen(a.progressContent)
	go a.progressListener()
}

// showSuccessScreen shows the success view
func (a *App) showSuccessScreen() {
	fyne.Do(func() {
		a.settingsBtn.Enable()
		a.setScreen(a.successContent)
	})
}

// showErrorScreen shows the error view
func (a *App) showErrorScreen() {
	fyne.Do(func() {
		a.settingsBtn.Enable()
		a.setScreen(a.errorContent)
	})
}

// checkBmapStatus checks for bmap file and returns status string
func (a *App) checkBmapStatus(imagePath string) string {
	bmapPath := a.CheckBmap(imagePath)
	if bmapPath != "" {
		a.SetBmapPath(bmapPath)
		return "Block map found — only used blocks are written"
	}
	return "No block map — the full image is written"
}

// CheckBmap returns the path of the auto-discovered bmap file
func (a *App) CheckBmap(imagePath string) string {
	candidates := []string{
		imagePath + ".bmap",
	}

	ext := filepath.Ext(imagePath)
	switch strings.ToLower(ext) {
	case ".gz", ".bz2", ".xz", ".zst", ".zstd", ".zip":
		base := strings.TrimSuffix(imagePath, ext)
		candidates = append(candidates, base+".bmap")
	case ".tar":
		candidates = append(candidates, imagePath+".bmap")
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

// GetLogs returns recent log lines
func (a *App) GetLogs() []string {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.lastLogs) == 0 {
		return []string{"No logs available"}
	}
	return a.lastLogs
}

// ShowLogsDialog displays logs in a dialog
func (a *App) ShowLogsDialog() {
	logsText := strings.Join(a.GetLogs(), "\n")
	text := widget.NewLabel(logsText)
	text.TextStyle = fyne.TextStyle{Monospace: true}
	text.Selectable = true
	text.Wrapping = fyne.TextWrapBreak
	scroll := container.NewVScroll(text)
	scroll.SetMinSize(fyne.NewSize(640, 360))

	d := dialog.NewCustom("Flash Log", "Close", scroll, a.window)
	copyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		a.fyneApp.Clipboard().SetContent(logsText)
	})
	d.SetButtons([]fyne.CanvasObject{copyBtn, widget.NewButton("Close", d.Hide)})
	d.Show()
}

// showSettingsDialog shows appearance and help settings
func (a *App) showSettingsDialog() {
	labels := map[string]string{"system": "Follow system", "light": "Light", "dark": "Dark"}
	values := map[string]string{"Follow system": "system", "Light": "light", "Dark": "dark"}

	appearance := widget.NewRadioGroup([]string{"Follow system", "Light", "Dark"}, nil)
	pref := a.themePref
	if _, ok := labels[pref]; !ok {
		pref = "system"
	}
	appearance.SetSelected(labels[pref])
	appearance.Required = true
	appearance.OnChanged = func(sel string) {
		a.themePref = values[sel]
		a.applyTheme()
		config, err := util.LoadConfig()
		if err != nil {
			config = util.DefaultConfig()
		}
		config.Theme = a.themePref
		_ = util.SaveConfig(config)
	}

	help := widget.NewHyperlink("Troubleshooting guide",
		mustParseURL("https://github.com/pantavisor/pvflasher/blob/main/docs/TROUBLESHOOTING.md"))

	config, _ := util.LoadConfig()
	autoUpdate := widget.NewCheck("Check for updates automatically", func(on bool) {
		c, _ := util.LoadConfig()
		c.DisableUpdateCheck = !on
		_ = util.SaveConfig(c)
	})
	autoUpdate.SetChecked(!config.DisableUpdateCheck)
	if !update.IsReleaseVersion(version.Version) {
		autoUpdate.Disable()
	}

	form := widget.NewForm(
		widget.NewFormItem("Appearance", appearance),
		widget.NewFormItem("Help", help),
	)
	var d dialog.Dialog
	updatesTitle := widget.NewLabelWithStyle("Updates", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	updates := a.updatesPanel(func() { d.Hide() })
	d = dialog.NewCustom("Settings", "Close", container.NewVBox(
		form,
		widget.NewSeparator(),
		updatesTitle,
		updates,
		autoUpdate,
	), a.window)
	d.Resize(fyne.NewSize(420, 0))
	d.Show()
}

// Setters for app state

func (a *App) SetSelectedImage(path string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.selectedImage = path
}

func (a *App) SetSelectedDevice(path string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.selectedDevice = path
}

func (a *App) SetBmapPath(path string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bmapPath = path
}

func (a *App) SetForceChecked(checked bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.forceChecked = checked
}

func (a *App) SetVerifyChecked(checked bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.verifyChecked = checked
}

func (a *App) SetEjectChecked(checked bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ejectChecked = checked
}
