package screens

import (
	"strings"

	"pvflasher/gui/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ErrorScreenCallbacks defines callbacks for error screen events
type ErrorScreenCallbacks struct {
	OnTryAgain func()
	OnViewLogs func()
}

// ErrorScreen represents the error display screen
type ErrorScreen struct {
	callbacks ErrorScreenCallbacks

	// Widgets
	errorLabel *util.ColoredLabel
	tipsBox    *fyne.Container

	// Content
	content fyne.CanvasObject
}

// NewErrorScreen creates a new error screen
func NewErrorScreen(callbacks ErrorScreenCallbacks) *ErrorScreen {
	return &ErrorScreen{
		callbacks: callbacks,
	}
}

// Build constructs and returns the screen UI
func (s *ErrorScreen) Build() fyne.CanvasObject {
	titleBar := util.CreateTitleBar("❌ Flash Operation Failed")

	s.errorLabel = util.NewColoredLabel("Error details will appear here", util.ColorError)
	s.tipsBox = container.NewVBox()

	errorCard := widget.NewCard("", "", container.NewVBox(
		util.SubHeadingLabel("Error Details"),
		util.SectionSpacer(4),
		s.errorLabel,
	))

	tipsCard := widget.NewCard("", "", container.NewVBox(
		util.SubHeadingLabel("💡 Troubleshooting Tips"),
		util.SectionSpacer(4),
		s.tipsBox,
	))

	tryAgainButton := util.PrimaryButton("🔄 Try Again", func() {
		if s.callbacks.OnTryAgain != nil {
			s.callbacks.OnTryAgain()
		}
	})

	viewLogsButton := util.PrimaryButton("📋 View Logs", func() {
		if s.callbacks.OnViewLogs != nil {
			s.callbacks.OnViewLogs()
		}
	})

	// Create background
	background := canvas.NewRectangle(util.CurrentBackgroundColor())

	// Main content
	contentBox := container.NewVBox(
		titleBar,
		errorCard,
		tipsCard,
		util.SectionSpacer(12),
		container.NewHBox(tryAgainButton, viewLogsButton),
	)

	// Stack background with content
	s.content = container.NewStack(
		background,
		container.NewPadded(contentBox),
	)

	return s.content
}

// SetError updates the error message and generates tips
func (s *ErrorScreen) SetError(errMsg string) {
	fyne.Do(func() {
		s.errorLabel.SetText(errMsg)
		s.tipsBox.RemoveAll()

		tips := []string{}

		// Context-sensitive tips
		errLower := strings.ToLower(errMsg)
		if strings.Contains(errLower, "mounted") {
			tips = append(tips, "• Device is mounted: Try using the 'Force' option or unmount the device first")
		}
		if strings.Contains(errLower, "permission") {
			tips = append(tips, "• Permission denied: Try running with admin/root privileges")
		}
		if strings.Contains(errLower, "verification") {
			tips = append(tips, "• Verification failed: Check that the image file is not corrupted")
			tips = append(tips, "• Try flashing with 'Force' option to skip verification")
		}
		if strings.Contains(errLower, "device") || strings.Contains(errLower, "not found") {
			tips = append(tips, "• Device not found: Check that the device is properly connected")
			tips = append(tips, "• Try refreshing the device list")
		}

		if len(tips) == 0 {
			tips = append(tips, "• Check that the image file is valid and not corrupted")
			tips = append(tips, "• Ensure the target device is properly connected")
			tips = append(tips, "• Try again or check the logs for more details")
		}

		for _, tip := range tips {
			s.tipsBox.Add(util.NewColoredLabel(tip, util.CurrentSecondaryTextColor()))
		}
		s.tipsBox.Refresh()
	})
}

// Content returns the screen content
func (s *ErrorScreen) Content() fyne.CanvasObject {
	return s.content
}
