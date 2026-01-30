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

	// Use themed checkboxes with separate labels for better visibility
	var forceContainer, verifyContainer, ejectContainer *fyne.Container

	c.ForceCheck, forceContainer = util.ThemedCheckWithState(
		"Force write (ignore mount warnings)",
		false,
		func(b bool) {
			if c.callbacks.OnForceChanged != nil {
				c.callbacks.OnForceChanged(b)
			}
		},
	)

	c.VerifyCheck, verifyContainer = util.ThemedCheckWithState(
		"Validate image after write",
		true,
		func(b bool) {
			if c.callbacks.OnVerifyChanged != nil {
				c.callbacks.OnVerifyChanged(b)
			}
		},
	)

	c.EjectCheck, ejectContainer = util.ThemedCheckWithState(
		"Eject device after completion",
		true,
		func(b bool) {
			if c.callbacks.OnEjectChanged != nil {
				c.callbacks.OnEjectChanged(b)
			}
		},
	)

	c.FlashButton = util.LargeButton("START FLASH", func() {
		if c.callbacks.OnStartFlash != nil {
			c.callbacks.OnStartFlash()
		}
	})
	c.FlashButton.Importance = widget.HighImportance
	c.FlashButton.Disable()

	header := container.NewVBox(
		stepLabel,
		util.SectionSpacer(4),
		titleLabel,
		util.SectionSpacer(4),
	)

	contentBox := container.NewVBox(
		util.InstructionLabel("Options:"),
		util.SectionSpacer(8),
		forceContainer,
		util.SectionSpacer(8),
		verifyContainer,
		util.SectionSpacer(8),
		ejectContainer,
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
