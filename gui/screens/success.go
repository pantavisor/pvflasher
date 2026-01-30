package screens

import (
	"fmt"
	"image/color"
	"strings"

	"pvflasher/gui/util"
	"pvflasher/internal/flash"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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

	// Content
	content fyne.CanvasObject
}

// NewSuccessScreen creates a new success screen
func NewSuccessScreen(callbacks SuccessScreenCallbacks) *SuccessScreen {
	return &SuccessScreen{
		callbacks: callbacks,
	}
}

// Build constructs and returns the screen UI
func (s *SuccessScreen) Build() fyne.CanvasObject {
	// Header Section - Green background with Checkmark
	headerText := canvas.NewText("Flash Completed Successfully!", color.White)
	headerText.TextSize = 24
	headerText.TextStyle = fyne.TextStyle{Bold: true}
	headerText.Alignment = fyne.TextAlignCenter

	checkMark := canvas.NewText("✓", color.White)
	checkMark.TextSize = 48
	checkMark.TextStyle = fyne.TextStyle{Bold: true}
	checkMark.Alignment = fyne.TextAlignCenter

	headerContent := container.NewVBox(
		util.SectionSpacer(24),
		container.NewCenter(checkMark),
		util.SectionSpacer(12),
		container.NewCenter(headerText),
		util.SectionSpacer(24),
	)

	headerBg := canvas.NewRectangle(util.ColorSuccess)
	header := container.NewStack(headerBg, headerContent)

	// Safety Message
	safetyLabel := widget.NewLabel("It is now safe to remove the device.")
	safetyLabel.Alignment = fyne.TextAlignCenter

	// Flash Statistics Header
	statsHeader := util.SubHeadingLabel("Flash Statistics:")

	// Stats Grid
	s.statsGrid = container.NewGridWithColumns(2)

	// Buttons
	anotherButton := util.PrimaryButton("Flash Another Device", func() {
		if s.callbacks.OnFlashAnother != nil {
			s.callbacks.OnFlashAnother()
		}
	})

	viewLogsButton := widget.NewButton("View Logs", func() {
		if s.callbacks.OnViewLogs != nil {
			s.callbacks.OnViewLogs()
		}
	})
	viewLogsButton.Importance = widget.MediumImportance

	// Use a grid for buttons to ensure equal width
	buttonRow := container.NewGridWithColumns(2,
		anotherButton,
		viewLogsButton,
	)

	// Main content
	contentBox := container.NewVBox(
		header,
		container.NewPadded(container.NewVBox(
			util.SectionSpacer(20),
			container.NewCenter(safetyLabel),
			util.SectionSpacer(20),
			statsHeader,
			util.SectionSpacer(12),
			s.statsGrid,
			util.SectionSpacer(30),
			buttonRow,
		)),
	)

	// Background
	background := canvas.NewRectangle(util.CurrentBackgroundColor())

	// Stack background with content (scrollable)
	s.content = container.NewStack(
		background,
		container.NewScroll(contentBox),
	)

	return s.content
}

// UpdateStats updates the statistics display
func (s *SuccessScreen) UpdateStats(result *flash.FlashResult) {
	fyne.Do(func() {
		s.statsGrid.RemoveAll()
		if result == nil {
			s.statsGrid.Add(util.NewColoredLabel("No result data available", util.CurrentSecondaryTextColor()))
			return
		}

		// Row 1
		s.statsGrid.Add(s.createStatCard("Data Written", util.FormatBytes(result.BytesWritten)))
		s.statsGrid.Add(s.createStatCard("Blocks Written", fmt.Sprintf("%d", result.BlocksWritten)))

		// Row 2
		s.statsGrid.Add(s.createStatCard("Duration", util.FormatDuration(result.Duration)))
		s.statsGrid.Add(s.createStatCard("Average Speed", util.FormatSpeed(result.AverageSpeed)))

		// Row 3
		method := "Raw Copy"
		if result.UsedBmap {
			method = "Bmap (Optimized)"
		}
		s.statsGrid.Add(s.createStatCard("Method", method))

		verification := "⊘ Skipped"
		if result.VerificationDone {
			verification = "✓ Passed"
		}
		s.statsGrid.Add(s.createStatCard("Verification", verification))

		s.statsGrid.Refresh()
	})
}

func (s *SuccessScreen) createStatCard(label, value string) fyne.CanvasObject {
	lbl := canvas.NewText(strings.ToUpper(label)+":", util.CurrentSecondaryTextColor())
	lbl.TextSize = 10
	lbl.TextStyle = fyne.TextStyle{Bold: true}

	val := canvas.NewText(value, util.CurrentTextColor())
	val.TextSize = 18
	val.TextStyle = fyne.TextStyle{Bold: true}

	content := container.NewVBox(
		lbl,
		util.SectionSpacer(4),
		val,
	)

	return util.StyledCardWithBorderAndPadding(content, 12)
}

// Content returns the screen content
func (s *SuccessScreen) Content() fyne.CanvasObject {
	return s.content
}
