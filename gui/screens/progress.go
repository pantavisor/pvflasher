package screens

import (
	"strings"
	"time"

	"pvflasher/gui/util"
	"pvflasher/pkg/flash"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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
	InfiniteBar *widget.ProgressBarInfinite // shown during byte-less phases (e.g. syncing)
	PhaseLabel  *widget.Label
	JobLabel    *widget.Label
	SpeedLabel  *widget.Label
	BytesLabel  *widget.Label
	CancelBtn   *widget.Button

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
	s.PhaseLabel = widget.NewLabelWithStyle("Preparing…", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	s.PhaseLabel.SizeName = theme.SizeNameHeadingText

	s.JobLabel = widget.NewLabel("")
	s.JobLabel.Importance = widget.LowImportance
	s.JobLabel.Truncation = fyne.TextTruncateEllipsis

	s.ProgressBar = widget.NewProgressBar()
	s.InfiniteBar = widget.NewProgressBarInfinite()
	s.InfiniteBar.Stop()
	s.InfiniteBar.Hide()

	s.SpeedLabel = widget.NewLabel("")
	s.BytesLabel = widget.NewLabel("")
	s.BytesLabel.Alignment = fyne.TextAlignTrailing

	s.CancelBtn = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		if s.callbacks.OnCancel != nil {
			s.callbacks.OnCancel()
		}
	})

	_, keepNotice := util.Notice(theme.NewPrimaryThemedResource(theme.InfoIcon()),
		"Don't remove the target or close PvFlasher until flashing has finished.")

	panel := util.NewSurface(container.NewVBox(
		s.PhaseLabel,
		s.JobLabel,
		util.SectionSpacer(8),
		container.NewStack(s.ProgressBar, s.InfiniteBar),
		container.NewGridWithColumns(2, s.SpeedLabel, s.BytesLabel),
		util.SectionSpacer(4),
		keepNotice,
	))

	s.content = container.NewCenter(util.MinWidth(560, container.NewVBox(
		panel,
		util.SectionSpacer(8),
		container.NewHBox(layoutSpacer(), s.CancelBtn),
	)))
	return s.content
}

// Reset prepares the screen for a new job.
func (s *ProgressScreen) Reset(image, target string) {
	fyne.Do(func() {
		s.PhaseLabel.SetText("Preparing…")
		s.JobLabel.SetText(image + "  →  " + target)
		s.ProgressBar.SetValue(0)
		s.InfiniteBar.Stop()
		s.InfiniteBar.Hide()
		s.ProgressBar.Show()
		s.SpeedLabel.SetText("")
		s.BytesLabel.SetText("")
		s.CancelBtn.Enable()
		s.CancelBtn.SetText("Cancel")
	})
}

// SetCancelling shows that a cancel request is in flight.
func (s *ProgressScreen) SetCancelling() {
	fyne.Do(func() {
		s.CancelBtn.Disable()
		s.CancelBtn.SetText("Cancelling…")
	})
}

// phaseTitle maps flasher phase identifiers to user-facing text.
func phaseTitle(phase string) string {
	switch phase {
	case "starting":
		return "Preparing…"
	case "writing":
		return "Writing image…"
	case "verifying":
		return "Verifying…"
	case "syncing":
		return "Finishing write…"
	case "ejecting":
		return "Ejecting…"
	}
	if phase == "" {
		return "Working…"
	}
	return strings.ToUpper(phase[:1]) + phase[1:]
}

// indeterminatePhase reports the trailing, byte-less phases where a determinate
// bar/speed would sit static (syncing flushes the page cache to the device in
// one blocking call; ejecting is instant). These get the animated bar. The
// writing/verifying phases have a known total and keep the normal bar.
func indeterminatePhase(phase string) bool {
	switch phase {
	case "syncing", "ejecting":
		return true
	}
	return false
}

// UpdateProgress updates the progress display
func (s *ProgressScreen) UpdateProgress(p flash.Progress) {
	fyne.Do(func() {
		s.PhaseLabel.SetText(phaseTitle(p.Phase))

		if indeterminatePhase(p.Phase) {
			// Animate so it's clearly alive even though there are no byte updates.
			if !s.InfiniteBar.Visible() {
				s.ProgressBar.Hide()
				s.InfiniteBar.Show()
				s.InfiniteBar.Start()
			}
			if p.Phase == "syncing" {
				s.SpeedLabel.SetText("Flushing buffers to the device — this can take a while")
			} else {
				s.SpeedLabel.SetText("")
			}
			s.BytesLabel.SetText(util.FormatBytes(p.BytesProcessed) + " written")
			return
		}

		// Determinate phase (starting/writing/verifying).
		if !s.InfiniteBar.Hidden {
			s.InfiniteBar.Stop()
			s.InfiniteBar.Hide()
			s.ProgressBar.Show()
		}

		if p.BytesTotal > 0 {
			s.ProgressBar.SetValue(float64(p.BytesProcessed) / float64(p.BytesTotal))
		} else {
			s.ProgressBar.SetValue(p.Percentage / 100.0)
		}

		speed := util.FormatSpeed(p.Speed)
		if p.BytesTotal > 0 && p.Speed > 0 && p.BytesProcessed < p.BytesTotal {
			eta := time.Duration(float64(p.BytesTotal-p.BytesProcessed)/p.Speed) * time.Second
			speed += " · " + util.FormatDuration(eta.Round(time.Second)) + " left"
		}
		s.SpeedLabel.SetText(speed)

		if p.BytesTotal > 0 {
			s.BytesLabel.SetText(util.FormatBytes(p.BytesProcessed) + " of " + util.FormatBytes(p.BytesTotal))
		} else if p.SourceTotal > 0 {
			s.BytesLabel.SetText(util.FormatBytes(p.BytesProcessed) + " written · " + util.FormatBytes(p.SourceRead) + " of " + util.FormatBytes(p.SourceTotal) + " read")
		} else {
			s.BytesLabel.SetText(util.FormatBytes(p.BytesProcessed) + " written")
		}
	})
}

// SetPhase updates just the phase label
func (s *ProgressScreen) SetPhase(phase string) {
	fyne.Do(func() {
		s.PhaseLabel.SetText(phase)
	})
}

// SetDetail updates the secondary status line (e.g. download speed).
func (s *ProgressScreen) SetDetail(detail string) {
	fyne.Do(func() {
		s.SpeedLabel.SetText(detail)
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
