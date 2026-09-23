package cards

import (
	"pvflasher/gui/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// OptionsCardCallbacks defines callbacks for options card events
type OptionsCardCallbacks struct {
	OnForceChanged  func(checked bool)
	OnVerifyChanged func(checked bool)
	OnEjectChanged  func(checked bool)
	OnStartFlash    func()
}

// OptionsCard represents the flash options card
type OptionsCard struct {
	callbacks OptionsCardCallbacks

	// Widgets
	ForceCheck  *widget.Check
	VerifyCheck *widget.Check
	EjectCheck  *widget.Check
	FlashButton *widget.Button
	hint        *widget.Label
}

// NewOptionsCard creates a new flash options card
func NewOptionsCard(callbacks OptionsCardCallbacks) *OptionsCard {
	return &OptionsCard{
		callbacks: callbacks,
	}
}

// Build constructs and returns the card UI
func (c *OptionsCard) Build() fyne.CanvasObject {
	c.VerifyCheck = widget.NewCheck("Verify after writing", func(b bool) {
		if c.callbacks.OnVerifyChanged != nil {
			c.callbacks.OnVerifyChanged(b)
		}
	})
	c.VerifyCheck.SetChecked(true)

	c.EjectCheck = widget.NewCheck("Eject when finished", func(b bool) {
		if c.callbacks.OnEjectChanged != nil {
			c.callbacks.OnEjectChanged(b)
		}
	})
	c.EjectCheck.SetChecked(true)

	c.ForceCheck = widget.NewCheck("Unmount without asking", func(b bool) {
		if c.callbacks.OnForceChanged != nil {
			c.callbacks.OnForceChanged(b)
		}
	})

	c.FlashButton = widget.NewButtonWithIcon("Flash", theme.UploadIcon(), func() {
		if c.callbacks.OnStartFlash != nil {
			c.callbacks.OnStartFlash()
		}
	})
	c.FlashButton.Importance = widget.HighImportance
	c.FlashButton.Disable()

	c.hint = widget.NewLabel("")
	c.hint.Alignment = fyne.TextAlignCenter
	c.hint.Importance = widget.LowImportance
	c.hint.SizeName = theme.SizeNameCaptionText
	c.SetFlashEnabled(false)

	options := container.NewVBox(c.VerifyCheck, c.EjectCheck, c.ForceCheck)

	return util.NewSurface(container.NewBorder(
		util.StepHeader("3", "Flash"),
		container.NewVBox(c.hint, util.TallButton(c.FlashButton, 44)),
		nil, nil,
		container.NewCenter(options),
	))
}

// SetFlashEnabled enables or disables the flash button
func (c *OptionsCard) SetFlashEnabled(enabled bool) {
	if enabled {
		c.FlashButton.Enable()
		c.hint.SetText("Ready to flash")
	} else {
		c.FlashButton.Disable()
		c.hint.SetText("Choose an image and a target first")
	}
}
