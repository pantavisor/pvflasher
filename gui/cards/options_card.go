package cards

import (
	"pvflasher/gui/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
}

// NewOptionsCard creates a new flash options card
func NewOptionsCard(callbacks OptionsCardCallbacks) *OptionsCard {
	return &OptionsCard{
		callbacks: callbacks,
	}
}

// Build constructs and returns the card UI
func (c *OptionsCard) Build() fyne.CanvasObject {
	stepLabel := util.StepLabel("STEP 3")
	titleLabel := util.SubHeadingLabel("Flash Options & Summary")

	// Modern styled checkboxes with improved spacing
	c.ForceCheck = widget.NewCheck("Force write (ignore mount warnings)", func(b bool) {
		if c.callbacks.OnForceChanged != nil {
			c.callbacks.OnForceChanged(b)
		}
	})

	c.VerifyCheck = widget.NewCheck("Validate image after write", func(b bool) {
		if c.callbacks.OnVerifyChanged != nil {
			c.callbacks.OnVerifyChanged(b)
		}
	})
	c.VerifyCheck.SetChecked(true)

	c.EjectCheck = widget.NewCheck("Eject device after completion", func(b bool) {
		if c.callbacks.OnEjectChanged != nil {
			c.callbacks.OnEjectChanged(b)
		}
	})
	c.EjectCheck.SetChecked(true)

	c.FlashButton = util.PrimaryActionButton("Start Flash", func() {
		if c.callbacks.OnStartFlash != nil {
			c.callbacks.OnStartFlash()
		}
	})
	c.FlashButton.Disable()

	header := container.NewVBox(
		stepLabel,
		util.SectionSpacer(6),
		titleLabel,
		util.SectionSpacer(8),
	)

	contentBox := container.NewVBox(
		util.InstructionLabel("Options:"),
		util.SectionSpacer(16),
		c.ForceCheck,
		util.SectionSpacer(12),
		c.VerifyCheck,
		util.SectionSpacer(12),
		c.EjectCheck,
	)

	// Use border to place button at bottom with full width
	cardContent := container.NewBorder(
		header,        // top
		c.FlashButton, // bottom (button with full width)
		nil,           // left
		nil,           // right
		contentBox,    // center
	)

	return util.StyledCardWithBorder(cardContent)
}

// SetFlashEnabled enables or disables the flash button
func (c *OptionsCard) SetFlashEnabled(enabled bool) {
	if enabled {
		c.FlashButton.Enable()
	} else {
		c.FlashButton.Disable()
	}
}
