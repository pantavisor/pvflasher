package screens

import (
	"fmt"

	"pvflasher/gui/util"
	"pvflasher/pkg/flash"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SuccessScreenCallbacks defines callbacks for success screen events
type SuccessScreenCallbacks struct {
	OnFlashAnother func()
	OnViewLogs     func()
}

// SuccessScreen represents the success display screen
type SuccessScreen struct {
	callbacks SuccessScreenCallbacks

	// Widgets
	statsGrid *fyne.Container
	subtitle  *widget.Label

	// Content
	content fyne.CanvasObject
}

// NewSuccessScreen creates a new success screen
func NewSuccessScreen(callbacks SuccessScreenCallbacks) *SuccessScreen {
	return &SuccessScreen{
		callbacks: callbacks,
	}
}

func layoutSpacer() fyne.CanvasObject { return layout.NewSpacer() }

// resultHeader builds the icon + title + subtitle block shared by the result screens.
func resultHeader(icon fyne.Resource, title string) (fyne.CanvasObject, *widget.Label) {
	_, iconBox := util.BigIcon(icon, 56)
	heading := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	heading.SizeName = theme.SizeNameHeadingText
	sub := widget.NewLabel("")
	sub.Alignment = fyne.TextAlignCenter
	sub.Wrapping = fyne.TextWrapWord
	sub.Importance = widget.LowImportance
	return container.NewVBox(iconBox, heading, sub), sub
}

// Build constructs and returns the screen UI
func (s *SuccessScreen) Build() fyne.CanvasObject {
	header, sub := resultHeader(theme.NewSuccessThemedResource(theme.ConfirmIcon()), "Flash Complete")
	s.subtitle = sub

	s.statsGrid = container.NewGridWithColumns(3)

	anotherButton := widget.NewButtonWithIcon("Flash Another", theme.ViewRefreshIcon(), func() {
		if s.callbacks.OnFlashAnother != nil {
			s.callbacks.OnFlashAnother()
		}
	})
	anotherButton.Importance = widget.HighImportance

	viewLogsButton := widget.NewButtonWithIcon("View Log", theme.DocumentIcon(), func() {
		if s.callbacks.OnViewLogs != nil {
			s.callbacks.OnViewLogs()
		}
	})

	s.content = container.NewCenter(util.MinWidth(560, container.NewVBox(
		header,
		util.SectionSpacer(8),
		util.NewSurface(s.statsGrid),
		util.SectionSpacer(8),
		container.NewHBox(layout.NewSpacer(), viewLogsButton, anotherButton),
	)))
	return s.content
}

// UpdateStats updates the statistics display
func (s *SuccessScreen) UpdateStats(result *flash.FlashResult) {
	fyne.Do(func() {
		s.statsGrid.RemoveAll()
		if result == nil {
			s.subtitle.SetText("The image was written to the target.")
			s.statsGrid.Add(widget.NewLabel("No statistics available"))
			return
		}

		if result.DeviceEjected {
			s.subtitle.SetText("The target was ejected and can be removed safely.")
		} else {
			s.subtitle.SetText("Eject the target from your system before removing it.")
		}

		method := "Full copy"
		if result.UsedBmap {
			method = "Bmap (sparse)"
		}
		verification := "Skipped"
		if result.VerificationDone {
			verification = "Passed"
		}

		s.statsGrid.Add(statItem("Written", util.FormatBytes(result.BytesWritten)))
		s.statsGrid.Add(statItem("Duration", util.FormatDuration(result.Duration)))
		s.statsGrid.Add(statItem("Average speed", util.FormatSpeed(result.AverageSpeed)))
		s.statsGrid.Add(statItem("Blocks", fmt.Sprintf("%d", result.BlocksWritten)))
		s.statsGrid.Add(statItem("Method", method))
		s.statsGrid.Add(statItem("Verification", verification))
		s.statsGrid.Refresh()
	})
}

func statItem(label, value string) fyne.CanvasObject {
	l := widget.NewLabel(label)
	l.Importance = widget.LowImportance
	l.SizeName = theme.SizeNameCaptionText
	v := widget.NewLabelWithStyle(value, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewVBox(l, v)
}

// Content returns the screen content
func (s *SuccessScreen) Content() fyne.CanvasObject {
	return s.content
}
