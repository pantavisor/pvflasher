package screens

import (
	"strings"

	"pvflasher/gui/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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
	errorLabel *widget.Label
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
	header, sub := resultHeader(theme.NewErrorThemedResource(theme.ErrorIcon()), "Flash Failed")
	sub.SetText("The target may not be usable until it is flashed successfully.")

	s.errorLabel = widget.NewLabel("")
	s.errorLabel.Wrapping = fyne.TextWrapWord
	s.errorLabel.Selectable = true
	s.errorLabel.TextStyle = fyne.TextStyle{Monospace: true}

	s.tipsBox = container.NewVBox()

	tipsTitle := widget.NewLabelWithStyle("What you can try", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	tryAgainButton := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		if s.callbacks.OnTryAgain != nil {
			s.callbacks.OnTryAgain()
		}
	})
	tryAgainButton.Importance = widget.HighImportance

	viewLogsButton := widget.NewButtonWithIcon("View Log", theme.DocumentIcon(), func() {
		if s.callbacks.OnViewLogs != nil {
			s.callbacks.OnViewLogs()
		}
	})

	s.content = container.NewCenter(util.MinWidth(560, container.NewVBox(
		header,
		util.SectionSpacer(8),
		util.NewSurface(container.NewVBox(s.errorLabel, widget.NewSeparator(), tipsTitle, s.tipsBox)),
		util.SectionSpacer(8),
		container.NewHBox(layout.NewSpacer(), viewLogsButton, tryAgainButton),
	)))
	return s.content
}

// errorTips returns context-sensitive suggestions for an error message.
func errorTips(errMsg string) []string {
	var tips []string
	errLower := strings.ToLower(errMsg)
	if strings.Contains(errLower, "mounted") || strings.Contains(errLower, "busy") {
		tips = append(tips, "Close any file manager windows showing the target, or enable “Unmount without asking”.")
	}
	if strings.Contains(errLower, "permission") || strings.Contains(errLower, "authentication") || strings.Contains(errLower, "not authorized") {
		tips = append(tips, "PvFlasher needs administrator rights to write to the target. Approve the password prompt when asked.")
	}
	if strings.Contains(errLower, "verification") || strings.Contains(errLower, "checksum") {
		tips = append(tips, "The image file may be corrupted. Download it again.")
		tips = append(tips, "The card may be failing. Try a different one.")
	}
	if strings.Contains(errLower, "not found") || strings.Contains(errLower, "no such") {
		tips = append(tips, "Check that the target is still connected, then select it again.")
	}
	if strings.Contains(errLower, "download") || strings.Contains(errLower, "network") {
		tips = append(tips, "Check your internet connection and try again.")
	}
	if len(tips) == 0 {
		tips = append(tips,
			"Reconnect the target and try again.",
			"Open the log for the full error output.")
	}
	return tips
}

// SetError updates the error message and generates tips
func (s *ErrorScreen) SetError(errMsg string) {
	fyne.Do(func() {
		s.errorLabel.SetText(errMsg)
		s.tipsBox.RemoveAll()
		for _, tip := range errorTips(errMsg) {
			l := widget.NewLabel("•  " + tip)
			l.Wrapping = fyne.TextWrapWord
			s.tipsBox.Add(l)
		}
		s.tipsBox.Refresh()
	})
}

// Content returns the screen content
func (s *ErrorScreen) Content() fyne.CanvasObject {
	return s.content
}
