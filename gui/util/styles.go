package util

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Color definitions for professional UI - Modern palette with high contrast
var (
	ColorPrimary       = color.NRGBA{R: 0x25, G: 0x63, B: 0xEB, A: 0xFF} // Vibrant blue
	ColorSuccess       = color.NRGBA{R: 0x05, G: 0x9E, B: 0x69, A: 0xFF} // Deeper teal green
	ColorWarning       = color.NRGBA{R: 0xD9, G: 0x77, B: 0x06, A: 0xFF} // Darker amber
	ColorError         = color.NRGBA{R: 0xDC, G: 0x26, B: 0x26, A: 0xFF} // Darker red
	ColorBackground    = color.NRGBA{R: 0xF3, G: 0xF4, B: 0xF6, A: 0xFF} // Subtle light gray
	ColorBorder        = color.NRGBA{R: 0xD1, G: 0xD5, B: 0xDB, A: 0xFF} // Slightly darker border
	ColorText          = color.NRGBA{R: 0x11, G: 0x18, B: 0x27, A: 0xFF} // Near-black text
	ColorTextSecondary = color.NRGBA{R: 0x37, G: 0x41, B: 0x51, A: 0xFF} // Darker gray - improved contrast
	ColorCardBg        = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF} // White
	ColorHeaderBg      = color.NRGBA{R: 0x1F, G: 0x2E, B: 0x3E, A: 0xFF} // Dark header
	ColorDanger        = color.NRGBA{R: 0xB9, G: 0x1C, B: 0x1C, A: 0xFF} // Deep red for system drives
	ColorDisabled      = color.NRGBA{R: 0x9C, G: 0xA3, B: 0xAF, A: 0xFF} // Gray for disabled state

	// Dark theme colors
	ColorDarkBackground    = color.NRGBA{R: 0x1F, G: 0x21, B: 0x25, A: 0xFF} // Dark background
	ColorDarkCardBg        = color.NRGBA{R: 0x2D, G: 0x2F, B: 0x34, A: 0xFF} // Dark card background
	ColorDarkText          = color.NRGBA{R: 0xF3, G: 0xF4, B: 0xF6, A: 0xFF} // Light text for dark mode
	ColorDarkTextSecondary = color.NRGBA{R: 0xA1, G: 0xA1, B: 0xAA, A: 0xFF} // Lighter secondary text
	ColorDarkBorder        = color.NRGBA{R: 0x3F, G: 0x42, B: 0x49, A: 0xFF} // Subtle dark border
)

const (
	// Spacing constants
	SpacingXS = 4.0
	SpacingS  = 8.0
	SpacingM  = 12.0
	SpacingL  = 16.0
	SpacingXL = 24.0

	// Border radius for modern look
	BorderRadiusCard   = 12.0
	BorderRadiusButton = 8.0

	// Shadow size
	ShadowSize = 4.0
)

// StyledCard creates a professional-looking card with border and padding
func StyledCard(title, description string, content fyne.CanvasObject) *widget.Card {
	card := widget.NewCard(title, description, content)
	return card
}

// StyledButton creates a professional button with styling
func StyledButton(label string, callback func()) *widget.Button {
	btn := widget.NewButton(label, callback)
	btn.Importance = widget.HighImportance
	return btn
}

// PrimaryButton creates a primary action button (blue)
func PrimaryButton(label string, callback func()) *widget.Button {
	btn := widget.NewButton(label, callback)
	btn.Importance = widget.HighImportance
	return btn
}

// DangerButton creates a danger/destructive action button (red)
func DangerButton(label string, callback func()) *widget.Button {
	btn := widget.NewButton(label, callback)
	btn.Importance = widget.HighImportance
	return btn
}

// SuccessButton creates a success action button (green)
func SuccessButton(label string, callback func()) *widget.Button {
	btn := widget.NewButton(label, callback)
	btn.Importance = widget.HighImportance
	return btn
}

// WarningButton creates a warning action button (orange)
func WarningButton(label string, callback func()) *widget.Button {
	btn := widget.NewButton(label, callback)
	btn.Importance = widget.LowImportance
	return btn
}

// HeadingLabel creates a large heading label with theme-aware color
func HeadingLabel(text string) fyne.CanvasObject {
	label := canvas.NewText(text, CurrentTextColor())
	label.TextSize = 20
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter
	return label
}

// SubHeadingLabel creates a medium heading label with proper styling
func SubHeadingLabel(text string) fyne.CanvasObject {
	label := canvas.NewText(text, CurrentTextColor())
	label.TextSize = 16
	label.TextStyle = fyne.TextStyle{Bold: true}
	return label
}

// StepLabel creates a step indicator label (STEP 1, STEP 2, etc.)
func StepLabel(text string) fyne.CanvasObject {
	label := canvas.NewText(text, CurrentSecondaryTextColor())
	label.TextSize = 12
	label.TextStyle = fyne.TextStyle{Bold: true}
	return label
}

// InstructionLabel creates an instruction label with proper contrast and wrapping
func InstructionLabel(text string) fyne.CanvasObject {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

// DescriptionLabel creates a smaller, secondary text label with wrapping
func DescriptionLabel(text string) fyne.CanvasObject {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

// DarkLabel creates a label with theme-aware text color
func DarkLabel(text string) fyne.CanvasObject {
	label := canvas.NewText(text, CurrentTextColor())
	label.TextSize = 12
	return label
}

// StyledProgressBar creates a professional progress bar with custom container
func StyledProgressBar() *widget.ProgressBar {
	progressBar := widget.NewProgressBar()
	return progressBar
}

// CardWithBorder creates a card-like container with a visible border
func CardWithBorder(title string, content fyne.CanvasObject) fyne.CanvasObject {
	// Create title
	titleLabel := canvas.NewText(title, ColorText)
	titleLabel.TextSize = 14
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Create border line under title
	borderLine := canvas.NewLine(ColorBorder)
	borderLine.StrokeWidth = 1

	// Create content container with padding
	contentBox := container.NewVBox(
		content,
	)

	// Combine title and content
	cardContent := container.NewVBox(
		titleLabel,
		borderLine,
		containerWithPadding(contentBox, 12),
	)

	return cardContent
}

// containerWithPadding wraps a container with padding
func containerWithPadding(content fyne.CanvasObject, padding float32) fyne.CanvasObject {
	return container.NewVBox(
		container.NewPadded(content),
	)
}

// SectionSpacer creates vertical spacing between sections
func SectionSpacer(height float32) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(0, height))
	return spacer
}

// StatusBadge creates a colored status indicator badge
func StatusBadge(text string, statusColor color.Color) fyne.CanvasObject {
	badge := canvas.NewText(text, statusColor)
	badge.TextSize = 12
	return badge
}

// InfoBox creates an information box with icon and text (with text wrapping)
func InfoBox(icon string, title string, message string) fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(icon+" "+title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	messageLabel := widget.NewLabel(message)
	messageLabel.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		titleLabel,
		messageLabel,
	)
}

// WarningBox creates a warning info box with amber title (with text wrapping)
func WarningBox(title string, message string) fyne.CanvasObject {
	titleLabel := canvas.NewText("⚠️ "+title, ColorWarning)
	titleLabel.TextSize = 13
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	messageLabel := widget.NewLabel(message)
	messageLabel.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		titleLabel,
		messageLabel,
	)
}

// ErrorBox creates an error info box with red title (with text wrapping)
func ErrorBox(title string, message string) fyne.CanvasObject {
	titleLabel := canvas.NewText(title, ColorError)
	titleLabel.TextSize = 13
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	messageLabel := widget.NewLabel(message)
	messageLabel.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		titleLabel,
		messageLabel,
	)
}

// SuccessBox creates a success info box
func SuccessBox(title string, message string) fyne.CanvasObject {
	return InfoBox("✅", title, message)
}

// StyledContainer creates a container with consistent spacing
func StyledContainer(objects ...fyne.CanvasObject) *fyne.Container {
	return container.NewVBox(objects...)
}

// HorizontalSpacer creates horizontal spacing
func HorizontalSpacer(width float32) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(width, 0))
	return spacer
}

// StyledHBox creates a styled horizontal box with spacing
func StyledHBox(objects ...fyne.CanvasObject) *fyne.Container {
	return container.NewHBox(objects...)
}

// CreateTitleBar creates a professional title bar
func CreateTitleBar(title string) fyne.CanvasObject {
	return CreateTitleBarWithAction(title, nil)
}

// CreateTitleBarWithAction creates a title bar with an optional action widget (e.g., theme toggle)
func CreateTitleBarWithAction(title string, action fyne.CanvasObject) fyne.CanvasObject {
	// Title text using theme-aware color
	heading := canvas.NewText(title, CurrentTextColor())
	heading.TextSize = 24
	heading.TextStyle = fyne.TextStyle{Bold: true}
	heading.Alignment = fyne.TextAlignCenter

	// Create centered heading
	centeredHeading := container.NewCenter(heading)

	// Create header content with optional action on the right
	var headerContent fyne.CanvasObject
	if action != nil {
		// Use Border layout: centered heading in middle, action on right
		headerContent = container.NewBorder(
			nil, nil,
			nil,                         // left - empty
			container.NewPadded(action), // right - theme toggle
			centeredHeading,             // center - title
		)
	} else {
		headerContent = centeredHeading
	}

	// Add vertical padding
	return container.NewPadded(headerContent)
}

// ColoredText creates text with custom color
func ColoredText(text string, textColor color.Color) *canvas.Text {
	t := canvas.NewText(text, textColor)
	return t
}

// LargeButton creates a larger, more prominent button with better styling
func LargeButton(label string, callback func()) *widget.Button {
	btn := widget.NewButton(label, callback)
	btn.Importance = widget.HighImportance
	// Note: Fyne will render this with system theming
	// The visual prominence comes from Importance flag
	return btn
}

// IconLabel creates a label with icon prefix
func IconLabel(icon string, text string) fyne.CanvasObject {
	label := canvas.NewText(icon+" "+text, CurrentTextColor())
	label.TextSize = 12
	return label
}

// ThemedCheck creates a checkbox with a properly themed label
// This works around Fyne's Check widget not always respecting custom theme colors
func ThemedCheck(label string, changed func(bool)) *fyne.Container {
	check := widget.NewCheck("", changed)
	labelWidget := widget.NewLabel(label)
	labelWidget.Wrapping = fyne.TextWrapWord

	return container.NewBorder(nil, nil, check, nil, labelWidget)
}

// ThemedCheckWithState creates a themed checkbox with initial state
func ThemedCheckWithState(label string, checked bool, changed func(bool)) (*widget.Check, *fyne.Container) {
	check := widget.NewCheck("", changed)
	check.Checked = checked
	labelWidget := widget.NewLabel(label)
	labelWidget.Wrapping = fyne.TextWrapWord

	return check, container.NewBorder(nil, nil, check, nil, labelWidget)
}

// ThemeToggleButton creates a button to toggle between light and dark mode
func ThemeToggleButton(app fyne.App, refreshCallback func()) *widget.Button {
	btn := widget.NewButton("", nil)

	updateButton := func() {
		if GetTheme().IsDark() {
			btn.SetText("☀️ Light")
		} else {
			btn.SetText("🌙 Dark")
		}
	}

	btn.OnTapped = func() {
		GetTheme().Toggle()
		updateButton()
		// Re-apply theme to trigger UI refresh
		app.Settings().SetTheme(GetTheme())

		// Save preference
		config, err := LoadConfig()
		if err == nil {
			if GetTheme().IsDark() {
				config.Theme = "dark"
			} else {
				config.Theme = "light"
			}
			SaveConfig(config)
		}

		if refreshCallback != nil {
			refreshCallback()
		}
	}

	updateButton()
	btn.Importance = widget.LowImportance
	return btn
}

// StyledCardWithBorder creates a modern card with shadow, rounded corners effect
// Uses theme-aware colors for proper light/dark mode support
func StyledCardWithBorder(content fyne.CanvasObject) fyne.CanvasObject {
	// Create background using theme-aware color
	bg := canvas.NewRectangle(CurrentCardBackground())

	// Create subtle border using theme-aware color
	borderColor := CurrentBorderColor()

	// Create the main content with padding
	paddedContent := container.NewPadded(content)

	// Create thin border rectangles
	topBorder := canvas.NewRectangle(borderColor)
	topBorder.SetMinSize(fyne.NewSize(0, 1))
	bottomBorder := canvas.NewRectangle(borderColor)
	bottomBorder.SetMinSize(fyne.NewSize(0, 1))
	leftBorder := canvas.NewRectangle(borderColor)
	leftBorder.SetMinSize(fyne.NewSize(1, 0))
	rightBorder := canvas.NewRectangle(borderColor)
	rightBorder.SetMinSize(fyne.NewSize(1, 0))

	// Stack: background + border + content
	cardWithBorder := container.NewStack(
		bg,
		container.NewBorder(
			topBorder,
			bottomBorder,
			leftBorder,
			rightBorder,
			paddedContent,
		),
	)

	return cardWithBorder
}

// StyledCardWithBorderAndPadding creates a styled card with custom padding on all sides
func StyledCardWithBorderAndPadding(content fyne.CanvasObject, padding float32) fyne.CanvasObject {
	// Create background rectangle using theme-aware color
	bg := canvas.NewRectangle(CurrentCardBackground())

	// Create spacers for padding on all sides
	topSpacer := canvas.NewRectangle(color.Transparent)
	topSpacer.SetMinSize(fyne.NewSize(0, padding))
	bottomSpacer := canvas.NewRectangle(color.Transparent)
	bottomSpacer.SetMinSize(fyne.NewSize(0, padding))
	leftSpacer := canvas.NewRectangle(color.Transparent)
	leftSpacer.SetMinSize(fyne.NewSize(padding, 0))
	rightSpacer := canvas.NewRectangle(color.Transparent)
	rightSpacer.SetMinSize(fyne.NewSize(padding, 0))

	// Create the content container with padding on all sides
	innerContent := container.NewBorder(topSpacer, bottomSpacer, leftSpacer, rightSpacer, content)

	// Combine background and content
	cardContainer := container.NewStack(
		bg,
		innerContent,
	)

	return cardContainer
}

// ColoredLabel is a custom widget that displays text with explicit color control
// It supports dynamic text updates via SetText() and text wrapping
type ColoredLabel struct {
	widget.BaseWidget
	Text     string // Exported Text field for compatibility
	color    color.Color
	bold     bool
	size     float32
	maxWidth float32
	label    *widget.Label
}

// NewColoredLabel creates a new colored label with dynamic update support
// Uses theme-aware colors - pass nil to use current theme text color
func NewColoredLabel(text string, textColor color.Color) *ColoredLabel {
	if textColor == nil {
		textColor = CurrentTextColor()
	}
	cl := &ColoredLabel{
		Text:     text,
		color:    textColor,
		size:     12,
		bold:     false,
		maxWidth: 240,
	}
	cl.ExtendBaseWidget(cl)
	return cl
}

// NewColoredLabelBold creates a new bold colored label
// Uses theme-aware colors - pass nil to use current theme text color
func NewColoredLabelBold(text string, textColor color.Color) *ColoredLabel {
	if textColor == nil {
		textColor = CurrentTextColor()
	}
	cl := &ColoredLabel{
		Text:     text,
		color:    textColor,
		size:     12,
		bold:     true,
		maxWidth: 240,
	}
	cl.ExtendBaseWidget(cl)
	return cl
}

// NewThemedLabel creates a label that automatically uses theme-aware text color
func NewThemedLabel(text string) *ColoredLabel {
	return NewColoredLabel(text, nil)
}

// NewThemedLabelBold creates a bold label that automatically uses theme-aware text color
func NewThemedLabelBold(text string) *ColoredLabel {
	return NewColoredLabelBold(text, nil)
}

// SetText updates the label text
func (cl *ColoredLabel) SetText(text string) {
	cl.Text = text
	if cl.label != nil {
		cl.label.SetText(text)
	}
	cl.Refresh()
}

// SetColor updates the label color (note: widget.Label uses theme colors, this is kept for API compatibility)
func (cl *ColoredLabel) SetColor(c color.Color) {
	cl.color = c
	cl.Refresh()
}

// CreateRenderer creates the renderer for this widget
func (cl *ColoredLabel) CreateRenderer() fyne.WidgetRenderer {
	lbl := widget.NewLabel(cl.Text)
	lbl.Wrapping = fyne.TextWrapBreak
	if cl.bold {
		lbl.TextStyle = fyne.TextStyle{Bold: true}
	}
	cl.label = lbl

	return &coloredLabelRenderer{label: cl, wrappedLabel: lbl}
}

// coloredLabelRenderer is a custom renderer for ColoredLabel
type coloredLabelRenderer struct {
	label        *ColoredLabel
	wrappedLabel *widget.Label
}

func (r *coloredLabelRenderer) Destroy() {}

func (r *coloredLabelRenderer) Layout(size fyne.Size) {
	r.wrappedLabel.Resize(size)
}

func (r *coloredLabelRenderer) MinSize() fyne.Size {
	min := r.wrappedLabel.MinSize()
	// Cap width to allow wrapping within card constraints
	if min.Width > r.label.maxWidth {
		min.Width = r.label.maxWidth
	}
	// Ensure minimum height
	if min.Height < r.label.size+4 {
		min.Height = r.label.size + 4
	}
	return min
}

func (r *coloredLabelRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.wrappedLabel}
}

func (r *coloredLabelRenderer) Refresh() {
	r.wrappedLabel.SetText(r.label.Text)
	if r.label.bold {
		r.wrappedLabel.TextStyle = fyne.TextStyle{Bold: true}
	}
	r.wrappedLabel.Wrapping = fyne.TextWrapBreak
	r.wrappedLabel.Refresh()
}

// MinSize returns the minimum size for the widget
func (cl *ColoredLabel) MinSize() fyne.Size {
	if cl.label != nil {
		min := cl.label.MinSize()
		// Cap width to allow wrapping within card constraints
		if min.Width > cl.maxWidth {
			min.Width = cl.maxWidth
		}
		if min.Height < cl.size+4 {
			min.Height = cl.size + 4
		}
		return min
	}
	return fyne.NewSize(50, cl.size+4)
}
