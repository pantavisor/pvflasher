package util

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestStyledButton(t *testing.T) {
	test.NewApp()

	called := false
	btn := StyledButton("Test", func() {
		called = true
	})

	if btn == nil {
		t.Fatal("StyledButton returned nil")
	}

	if btn.Text != "Test" {
		t.Errorf("Button text = %q, want %q", btn.Text, "Test")
	}

	// Simulate click
	btn.OnTapped()
	if !called {
		t.Error("Button callback was not called")
	}
}

func TestPrimaryButton(t *testing.T) {
	test.NewApp()

	btn := PrimaryButton("Primary", func() {})
	if btn == nil {
		t.Fatal("PrimaryButton returned nil")
	}

	if btn.Text != "Primary" {
		t.Errorf("Button text = %q, want %q", btn.Text, "Primary")
	}

	if btn.Importance != widget.HighImportance {
		t.Error("Primary button should have HighImportance")
	}
}

func TestDangerButton(t *testing.T) {
	test.NewApp()

	btn := DangerButton("Danger", func() {})
	if btn == nil {
		t.Fatal("DangerButton returned nil")
	}

	if btn.Text != "Danger" {
		t.Errorf("Button text = %q, want %q", btn.Text, "Danger")
	}
}

func TestSuccessButton(t *testing.T) {
	test.NewApp()

	btn := SuccessButton("Success", func() {})
	if btn == nil {
		t.Fatal("SuccessButton returned nil")
	}

	if btn.Text != "Success" {
		t.Errorf("Button text = %q, want %q", btn.Text, "Success")
	}
}

func TestWarningButton(t *testing.T) {
	test.NewApp()

	btn := WarningButton("Warning", func() {})
	if btn == nil {
		t.Fatal("WarningButton returned nil")
	}

	if btn.Text != "Warning" {
		t.Errorf("Button text = %q, want %q", btn.Text, "Warning")
	}
}

func TestStyledCard(t *testing.T) {
	test.NewApp()

	content := widget.NewLabel("Content")
	card := StyledCard("Title", "Description", content)

	if card == nil {
		t.Fatal("StyledCard returned nil")
	}
}

func TestInstructionLabel(t *testing.T) {
	test.NewApp()

	label := InstructionLabel("Test instruction")
	if label == nil {
		t.Fatal("InstructionLabel returned nil")
	}
}

func TestDescriptionLabel(t *testing.T) {
	test.NewApp()

	label := DescriptionLabel("Test description")
	if label == nil {
		t.Fatal("DescriptionLabel returned nil")
	}
}

func TestHeadingLabel(t *testing.T) {
	test.NewApp()

	label := HeadingLabel("Test heading")
	if label == nil {
		t.Fatal("HeadingLabel returned nil")
	}
}

func TestSubHeadingLabel(t *testing.T) {
	test.NewApp()

	label := SubHeadingLabel("Test subheading")
	if label == nil {
		t.Fatal("SubHeadingLabel returned nil")
	}
}

func TestStepLabel(t *testing.T) {
	test.NewApp()

	label := StepLabel("STEP 1")
	if label == nil {
		t.Fatal("StepLabel returned nil")
	}
}

func TestStyledProgressBar(t *testing.T) {
	test.NewApp()

	bar := StyledProgressBar()
	if bar == nil {
		t.Fatal("StyledProgressBar returned nil")
	}
}

func TestSectionSpacer(t *testing.T) {
	test.NewApp()

	spacer := SectionSpacer(20)
	if spacer == nil {
		t.Fatal("SectionSpacer returned nil")
	}
}

func TestHorizontalSpacer(t *testing.T) {
	test.NewApp()

	spacer := HorizontalSpacer(20)
	if spacer == nil {
		t.Fatal("HorizontalSpacer returned nil")
	}
}

func TestStyledContainer(t *testing.T) {
	test.NewApp()

	container := StyledContainer()
	if container == nil {
		t.Fatal("StyledContainer returned nil")
	}
}

func TestStyledHBox(t *testing.T) {
	test.NewApp()

	hbox := StyledHBox()
	if hbox == nil {
		t.Fatal("StyledHBox returned nil")
	}
}

func TestCreateTitleBar(t *testing.T) {
	test.NewApp()

	titleBar := CreateTitleBar("Test Title")
	if titleBar == nil {
		t.Fatal("CreateTitleBar returned nil")
	}
}

func TestColoredText(t *testing.T) {
	test.NewApp()

	text := ColoredText("Test", ColorPrimary)
	if text == nil {
		t.Fatal("ColoredText returned nil")
	}

	if text.Text != "Test" {
		t.Errorf("Text = %q, want %q", text.Text, "Test")
	}
}

func TestLargeButton(t *testing.T) {
	test.NewApp()

	btn := LargeButton("Large", func() {})
	if btn == nil {
		t.Fatal("LargeButton returned nil")
	}
}

func TestIconLabel(t *testing.T) {
	test.NewApp()

	label := IconLabel("📁", "Files")
	if label == nil {
		t.Fatal("IconLabel returned nil")
	}
}

func TestStyledCardWithBorder(t *testing.T) {
	test.NewApp()

	content := widget.NewLabel("Content")
	card := StyledCardWithBorder(content)
	if card == nil {
		t.Fatal("StyledCardWithBorder returned nil")
	}
}

func TestStyledCardWithBorderAndPadding(t *testing.T) {
	test.NewApp()

	content := widget.NewLabel("Content")
	card := StyledCardWithBorderAndPadding(content, 16)
	if card == nil {
		t.Fatal("StyledCardWithBorderAndPadding returned nil")
	}
}

func TestNewColoredLabel(t *testing.T) {
	test.NewApp()

	label := NewColoredLabel("Test", ColorPrimary)
	if label == nil {
		t.Fatal("NewColoredLabel returned nil")
	}

	if label.Text != "Test" {
		t.Errorf("Label text = %q, want %q", label.Text, "Test")
	}
}

func TestNewColoredLabelBold(t *testing.T) {
	test.NewApp()

	label := NewColoredLabelBold("Bold Test", ColorPrimary)
	if label == nil {
		t.Fatal("NewColoredLabelBold returned nil")
	}
}

func TestNewThemedLabel(t *testing.T) {
	test.NewApp()

	label := NewThemedLabel("Themed")
	if label == nil {
		t.Fatal("NewThemedLabel returned nil")
	}
}

func TestNewThemedLabelBold(t *testing.T) {
	test.NewApp()

	label := NewThemedLabelBold("Bold Themed")
	if label == nil {
		t.Fatal("NewThemedLabelBold returned nil")
	}
}

func TestColoredLabel_SetText(t *testing.T) {
	test.NewApp()

	label := NewColoredLabel("Original", nil)
	label.SetText("Updated")

	if label.Text != "Updated" {
		t.Errorf("Text = %q, want %q", label.Text, "Updated")
	}
}

func TestCardWithBorder(t *testing.T) {
	test.NewApp()

	content := widget.NewLabel("Content")
	card := CardWithBorder("Title", content)
	if card == nil {
		t.Fatal("CardWithBorder returned nil")
	}
}

func TestStatusBadge(t *testing.T) {
	test.NewApp()

	badge := StatusBadge("Active", ColorSuccess)
	if badge == nil {
		t.Fatal("StatusBadge returned nil")
	}
}

func TestInfoBox(t *testing.T) {
	test.NewApp()

	infoBox := InfoBox("ℹ️", "Title", "Message")
	if infoBox == nil {
		t.Fatal("InfoBox returned nil")
	}
}

func TestWarningBox(t *testing.T) {
	test.NewApp()

	warningBox := WarningBox("Warning", "Be careful")
	if warningBox == nil {
		t.Fatal("WarningBox returned nil")
	}
}

func TestErrorBox(t *testing.T) {
	test.NewApp()

	errorBox := ErrorBox("Error", "Something went wrong")
	if errorBox == nil {
		t.Fatal("ErrorBox returned nil")
	}
}

func TestSuccessBox(t *testing.T) {
	test.NewApp()

	successBox := SuccessBox("Success", "Operation completed")
	if successBox == nil {
		t.Fatal("SuccessBox returned nil")
	}
}

func TestDarkLabel(t *testing.T) {
	test.NewApp()

	label := DarkLabel("Dark text")
	if label == nil {
		t.Fatal("DarkLabel returned nil")
	}
}

func TestThemedCheck(t *testing.T) {
	test.NewApp()

	checkContainer := ThemedCheck("Option", func(bool) {})
	if checkContainer == nil {
		t.Fatal("ThemedCheck returned nil")
	}
}

func TestThemedCheckWithState(t *testing.T) {
	test.NewApp()

	check, container := ThemedCheckWithState("Option", true, func(bool) {})
	if check == nil {
		t.Fatal("ThemedCheckWithState returned nil check")
	}
	if container == nil {
		t.Fatal("ThemedCheckWithState returned nil container")
	}
}

func TestColorDefinitions(t *testing.T) {
	// Test that colors are defined (they're color.NRGBA values)
	// Just verify these compile by accessing them
	_ = ColorPrimary
	_ = ColorSuccess
	_ = ColorWarning
	_ = ColorError
	_ = ColorBackground
	_ = ColorBorder
	_ = ColorText
	_ = ColorTextSecondary
	_ = ColorCardBg
	_ = ColorHeaderBg
	_ = ColorDanger
	_ = ColorDisabled

	// Verify colors have non-zero alpha
	if ColorPrimary.A == 0 {
		t.Error("ColorPrimary should have non-zero alpha")
	}
}

func TestSpacingConstants(t *testing.T) {
	// Test spacing constants
	if SpacingXS != 4.0 {
		t.Errorf("SpacingXS = %f, want 4.0", SpacingXS)
	}
	if SpacingS != 8.0 {
		t.Errorf("SpacingS = %f, want 8.0", SpacingS)
	}
	if SpacingM != 12.0 {
		t.Errorf("SpacingM = %f, want 12.0", SpacingM)
	}
	if SpacingL != 16.0 {
		t.Errorf("SpacingL = %f, want 16.0", SpacingL)
	}
	if SpacingXL != 24.0 {
		t.Errorf("SpacingXL = %f, want 24.0", SpacingXL)
	}
}

func TestBorderRadiusConstants(t *testing.T) {
	if BorderRadiusCard != 12.0 {
		t.Errorf("BorderRadiusCard = %f, want 12.0", BorderRadiusCard)
	}
	if BorderRadiusButton != 8.0 {
		t.Errorf("BorderRadiusButton = %f, want 8.0", BorderRadiusButton)
	}
}

func TestShadowSize(t *testing.T) {
	if ShadowSize != 4.0 {
		t.Errorf("ShadowSize = %f, want 4.0", ShadowSize)
	}
}
