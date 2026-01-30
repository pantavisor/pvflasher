package screens

import (
	"pvflasher/gui/util"
	"pvflasher/internal/flash"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ProgressScreenCallbacks defines callbacks for progress screen events
type ProgressScreenCallbacks struct {
	OnCancel func()
}

// ProgressScreen represents the progress display screen
type ProgressScreen struct {
	callbacks ProgressScreenCallbacks

	// Widgets
	ProgressBar *widget.ProgressBar
	PhaseLabel  *util.ColoredLabel
	SpeedLabel  *util.ColoredLabel
	BytesLabel  *util.ColoredLabel

	// Content
	content fyne.CanvasObject
}

// NewProgressScreen creates a new progress screen
func NewProgressScreen(callbacks ProgressScreenCallbacks) *ProgressScreen {
	return &ProgressScreen{
		callbacks: callbacks,
	}
}

// Build constructs and returns the screen UI
func (s *ProgressScreen) Build() fyne.CanvasObject {
	titleBar := util.CreateTitleBar("⚡ Flashing in Progress...")

	s.ProgressBar = util.StyledProgressBar()
	s.PhaseLabel = util.NewThemedLabel("Starting...")
	s.SpeedLabel = util.NewThemedLabel("Speed: 0 MB/s")
	s.BytesLabel = util.NewThemedLabel("0 B / 0 B")

	cancelButton := util.WarningButton("⏹️ Cancel Operation", func() {
		if s.callbacks.OnCancel != nil {
			s.callbacks.OnCancel()
		}
	})

	progressCard := widget.NewCard("", "", container.NewVBox(
		util.SubHeadingLabel("Transfer Progress"),
		util.SectionSpacer(4),
		s.ProgressBar,
	))

	statusCard := widget.NewCard("", "", container.NewVBox(
		util.SubHeadingLabel("Transfer Status"),
		util.SectionSpacer(4),
		s.PhaseLabel,
		util.SectionSpacer(4),
		s.SpeedLabel,
		util.SectionSpacer(4),
		s.BytesLabel,
	))

	// Create background
	background := canvas.NewRectangle(util.CurrentBackgroundColor())

	// Main content
	contentBox := container.NewVBox(
		titleBar,
		progressCard,
		statusCard,
		util.SectionSpacer(12),
		container.NewCenter(cancelButton),
	)

	// Stack background with content
	s.content = container.NewStack(
		background,
		container.NewPadded(contentBox),
	)

	return s.content
}

// UpdateProgress updates the progress display
func (s *ProgressScreen) UpdateProgress(p flash.Progress) {
	fyne.Do(func() {
		// Update progress bar
		if p.BytesTotal > 0 {
			s.ProgressBar.SetValue(float64(p.BytesProcessed) / float64(p.BytesTotal))
		} else {
			s.ProgressBar.SetValue(p.Percentage / 100.0)
		}

		// Update phase label
		s.PhaseLabel.SetText(p.Phase)

		// Update speed label
		s.SpeedLabel.SetText("Speed: " + util.FormatSpeed(p.Speed))

		// Update bytes label
		s.BytesLabel.SetText(util.FormatBytes(p.BytesProcessed) + " / " + util.FormatBytes(p.BytesTotal))
	})
}

// SetPhase updates just the phase label
func (s *ProgressScreen) SetPhase(phase string) {
	fyne.Do(func() {
		s.PhaseLabel.SetText(phase)
	})
}

// SetProgress updates just the progress bar value
func (s *ProgressScreen) SetProgress(value float64) {
	fyne.Do(func() {
		s.ProgressBar.SetValue(value)
	})
}

// Content returns the screen content
func (s *ProgressScreen) Content() fyne.CanvasObject {
	return s.content
}
