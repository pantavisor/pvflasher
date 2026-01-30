package gui

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"pvflasher/gui/cards"
	"pvflasher/gui/pantavisor"
	"pvflasher/gui/screens"
	"pvflasher/gui/util"
	"pvflasher/internal/flash"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	a.fyneApp = app.New()

	// Initialize theme based on system preference
	if a.fyneApp.Settings().ThemeVariant() == theme.VariantDark {
		util.GetTheme().SetMode(util.ThemeModeDark)
	} else {
		util.GetTheme().SetMode(util.ThemeModeLight)
	}

	a.fyneApp.Settings().SetTheme(util.GetTheme())

	a.window = a.fyneApp.NewWindow("PvFlasher")
	a.window.Resize(fyne.NewSize(950, 650))
	a.window.SetFixedSize(true)
	a.window.CenterOnScreen()

	// Build all views
	a.buildMainView()
	a.buildProgressScreen()
	a.buildSuccessScreen()
	a.buildErrorScreen()

	// Set the main view as the content
	a.window.SetContent(a.mainContent)
	a.window.ShowAndRun()
}

// buildMainView builds the main selection view
func (a *App) buildMainView() {
	// Create background rectangle
	background := canvas.NewRectangle(util.CurrentBackgroundColor())

	// Theme toggle button
	themeToggle := util.ThemeToggleButton(a.fyneApp, func() {
		a.rebuildAllViews()
	})

	// Title bar with theme toggle
	titleBar := util.CreateTitleBarWithAction("⚡ PvFlasher - USB Image Flasher", themeToggle)

	// Step 1: Image Selection Card
	a.imageCard = cards.NewImageCard(a.window, cards.ImageCardCallbacks{
		OnLocalImageSelected: func(path string) {
			a.SetSelectedImage(path)
			a.selectedRel = nil // Clear any Pantavisor selection
			a.updateFlashButtonState()
		},
		OnPantavisorSelected: func(rel *pantavisor.DeviceRelease) {
			a.mu.Lock()
			a.selectedRel = rel
			a.selectedImage = "" // Clear local image
			a.mu.Unlock()
			a.updateFlashButtonState()
		},
		CheckBmapStatus: a.checkBmapStatus,
	})
	imageCardUI := a.imageCard.Build()

	// Step 2: Device Selection Card
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
	deviceCardUI := a.deviceCard.Build()

	// Step 3: Flash Options Card
	a.optionsCard = cards.NewOptionsCard(cards.OptionsCardCallbacks{
		OnForceChanged:  func(b bool) { a.SetForceChecked(b) },
		OnVerifyChanged: func(b bool) { a.SetVerifyChecked(b) },
		OnEjectChanged:  func(b bool) { a.SetEjectChecked(b) },
		OnStartFlash:    func() { a.startFlash() },
	})
	optionsCardUI := a.optionsCard.Build()

	// Wrap each card with fixed width constraint
	cardWidth := float32(280)
	wrapCard := func(card fyne.CanvasObject) fyne.CanvasObject {
		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(cardWidth, 0))
		return container.NewStack(spacer, card)
	}

	// Use GridWithColumns to ensure all cards have equal height
	cardsContainer := container.NewGridWithColumns(3,
		wrapCard(imageCardUI),
		wrapCard(deviceCardUI),
		wrapCard(optionsCardUI),
	)

	// Center the cards container
	centeredCards := container.NewCenter(cardsContainer)

	// Main content with generous spacing
	contentBox := container.NewVBox(
		util.SectionSpacer(24),
		centeredCards,
		util.SectionSpacer(24),
	)

	// Make content scrollable
	scrollableContent := container.NewScroll(contentBox)

	// Combine title bar (fixed) with scrollable content
	mainLayout := container.NewBorder(
		titleBar,
		nil,
		nil,
		nil,
		scrollableContent,
	)

	// Combine background and content using Stack
	a.mainContent = container.NewStack(
		background,
		mainLayout,
	)
}

// buildProgressScreen builds the progress display screen
func (a *App) buildProgressScreen() {
	a.progressScreen = screens.NewProgressScreen(screens.ProgressScreenCallbacks{
		OnCancel: a.CancelFlash,
	})
	a.progressContent = a.progressScreen.Build()

	// Initialize progress channel
	a.progressChan = make(chan flash.Progress, 10)
}

// buildSuccessScreen builds the success display screen
func (a *App) buildSuccessScreen() {
	a.successScreen = screens.NewSuccessScreen(screens.SuccessScreenCallbacks{
		OnFlashAnother: a.resetToMainView,
		OnViewLogs:     a.ShowLogsDialog,
	})
	a.successContent = a.successScreen.Build()
}

// buildErrorScreen builds the error display screen
func (a *App) buildErrorScreen() {
	a.errorScreen = screens.NewErrorScreen(screens.ErrorScreenCallbacks{
		OnTryAgain: a.resetToMainView,
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

// resetToMainView returns to the main view
func (a *App) resetToMainView() {
	a.mu.Lock()
	a.selectedImage = ""
	a.selectedDevice = ""
	a.bmapPath = ""
	a.selectedRel = nil
	a.mu.Unlock()

	// Reset cards
	if a.imageCard != nil {
		a.imageCard.Reset()
	}
	if a.deviceCard != nil {
		a.deviceCard.Reset()
	}
	if a.optionsCard != nil {
		a.optionsCard.SetFlashEnabled(false)
	}

	a.window.SetContent(a.mainContent)
}

// rebuildAllViews rebuilds all views to apply theme changes
func (a *App) rebuildAllViews() {
	// Store current state
	selectedImage := a.selectedImage
	selectedDevice := a.selectedDevice
	bmapPath := a.bmapPath

	// Rebuild all views
	a.buildMainView()
	a.buildProgressScreen()
	a.buildSuccessScreen()
	a.buildErrorScreen()

	// Restore state
	a.selectedImage = selectedImage
	a.selectedDevice = selectedDevice
	a.bmapPath = bmapPath

	// Update UI labels if needed
	if selectedImage != "" && a.imageCard != nil {
		a.imageCard.SelectedImageLabel.SetText(fmt.Sprintf("📁 %s", filepath.Base(selectedImage)))
	}
	if selectedDevice != "" && a.deviceCard != nil {
		a.deviceCard.SelectedDeviceLabel.SetText(selectedDevice)
	}
	a.updateFlashButtonState()

	// Refresh the window content
	a.window.SetContent(a.mainContent)
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

	a.window.SetContent(a.progressContent)
	go a.progressListener()
}

// showSuccessScreen shows the success view
func (a *App) showSuccessScreen() {
	fyne.Do(func() {
		a.window.SetContent(a.successContent)
	})
}

// showErrorScreen shows the error view
func (a *App) showErrorScreen() {
	fyne.Do(func() {
		a.window.SetContent(a.errorContent)
	})
}

// checkBmapStatus checks for bmap file and returns status string
func (a *App) checkBmapStatus(imagePath string) string {
	bmapPath := a.CheckBmap(imagePath)
	if bmapPath != "" {
		a.SetBmapPath(bmapPath)
		return fmt.Sprintf("✅ Found: %s", filepath.Base(bmapPath))
	}
	return "💡 No bmap file found (will use full image)"
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
	logs := a.GetLogs()
	logsText := strings.Join(logs, "\n")
	textWidget := widget.NewLabel(logsText)
	textWidget.Wrapping = fyne.TextWrapBreak
	scrollable := container.NewVScroll(textWidget)
	scrollable.SetMinSize(fyne.NewSize(600, 400))
	dialog.ShowCustom("Flash Operation Logs", "Close", scrollable, a.window)
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
