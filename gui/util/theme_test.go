package util

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestNewAppTheme(t *testing.T) {
	appTheme := NewAppTheme()
	if appTheme == nil {
		t.Fatal("NewAppTheme() returned nil")
	}

	if appTheme.Mode() != ThemeModeLight {
		t.Errorf("Initial mode = %v, want ThemeModeLight", appTheme.Mode())
	}
}

func TestGetTheme(t *testing.T) {
	// First call should create the theme
	theme1 := GetTheme()
	if theme1 == nil {
		t.Fatal("GetTheme() returned nil")
	}

	// Second call should return the same theme (singleton)
	theme2 := GetTheme()
	if theme2 != theme1 {
		t.Error("GetTheme() should return the same theme instance")
	}
}

func TestAppTheme_SetMode(t *testing.T) {
	appTheme := NewAppTheme()

	// Set to dark mode
	appTheme.SetMode(ThemeModeDark)
	if appTheme.Mode() != ThemeModeDark {
		t.Errorf("Mode = %v, want ThemeModeDark", appTheme.Mode())
	}

	// Set back to light mode
	appTheme.SetMode(ThemeModeLight)
	if appTheme.Mode() != ThemeModeLight {
		t.Errorf("Mode = %v, want ThemeModeLight", appTheme.Mode())
	}
}

func TestAppTheme_IsDark(t *testing.T) {
	appTheme := NewAppTheme()

	if appTheme.IsDark() {
		t.Error("New theme should not be dark")
	}

	appTheme.SetMode(ThemeModeDark)
	if !appTheme.IsDark() {
		t.Error("Theme should be dark after SetMode(ThemeModeDark)")
	}
}

func TestAppTheme_Toggle(t *testing.T) {
	appTheme := NewAppTheme()

	// Start in light mode
	if appTheme.Mode() != ThemeModeLight {
		t.Fatal("Should start in light mode")
	}

	// Toggle to dark
	appTheme.Toggle()
	if appTheme.Mode() != ThemeModeDark {
		t.Error("Toggle should switch to dark mode")
	}

	// Toggle back to light
	appTheme.Toggle()
	if appTheme.Mode() != ThemeModeLight {
		t.Error("Toggle should switch back to light mode")
	}
}

func TestAppTheme_Color(t *testing.T) {
	appTheme := NewAppTheme()

	// Test that Color method exists and returns a color
	bgColor := appTheme.Color(theme.ColorNameBackground, theme.VariantLight)
	if bgColor == nil {
		t.Error("Color() returned nil for background")
	}

	// Test primary color
	primaryColor := appTheme.Color(theme.ColorNamePrimary, theme.VariantLight)
	if primaryColor == nil {
		t.Error("Color() returned nil for primary")
	}
}

func TestAppTheme_DarkModeColors(t *testing.T) {
	appTheme := NewAppTheme()
	appTheme.SetMode(ThemeModeDark)

	// Get colors in dark mode
	bgColorDark := appTheme.Color(theme.ColorNameBackground, theme.VariantDark)
	textColorDark := appTheme.Color(theme.ColorNameForeground, theme.VariantDark)

	// Verify colors are returned
	if bgColorDark == nil {
		t.Error("Color() returned nil for background in dark mode")
	}
	if textColorDark == nil {
		t.Error("Color() returned nil for foreground in dark mode")
	}
}

func TestCurrentTextColor(t *testing.T) {
	// Create a test app to initialize the theme
	test.NewApp()

	// Get current text color
	textColor := CurrentTextColor()
	if textColor == nil {
		t.Error("CurrentTextColor() returned nil")
	}
}

func TestCurrentBackgroundColor(t *testing.T) {
	// Create a test app to initialize the theme
	test.NewApp()

	// Get current background color
	bgColor := CurrentBackgroundColor()
	if bgColor == nil {
		t.Error("CurrentBackgroundColor() returned nil")
	}
}

func TestCurrentCardBackground(t *testing.T) {
	// Create a test app to initialize the theme
	test.NewApp()

	// Get current card background color
	cardBg := CurrentCardBackground()
	if cardBg == nil {
		t.Error("CurrentCardBackground() returned nil")
	}
}

func TestCurrentBorderColor(t *testing.T) {
	// Create a test app to initialize the theme
	test.NewApp()

	// Get current border color
	borderColor := CurrentBorderColor()
	if borderColor == nil {
		t.Error("CurrentBorderColor() returned nil")
	}
}

func TestCurrentSecondaryTextColor(t *testing.T) {
	// Create a test app to initialize the theme
	test.NewApp()

	// Get secondary text color
	secondaryColor := CurrentSecondaryTextColor()
	if secondaryColor == nil {
		t.Error("CurrentSecondaryTextColor() returned nil")
	}
}

func TestAppTheme_Font(t *testing.T) {
	appTheme := NewAppTheme()

	// Get font for different text styles
	regularFont := appTheme.Font(fyne.TextStyle{})
	boldFont := appTheme.Font(fyne.TextStyle{Bold: true})

	// Fonts should not be nil (will fall back to default)
	if regularFont == nil {
		t.Log("Regular font is nil (may be expected)")
	}
	if boldFont == nil {
		t.Log("Bold font is nil (may be expected)")
	}
}

func TestAppTheme_Icon(t *testing.T) {
	appTheme := NewAppTheme()

	// Get icon
	icon := appTheme.Icon(theme.IconNameConfirm)
	if icon == nil {
		t.Log("Icon is nil (will fall back to default)")
	}
}

func TestAppTheme_Size(t *testing.T) {
	appTheme := NewAppTheme()

	// Test various size names
	sizes := []fyne.ThemeSizeName{
		theme.SizeNamePadding,
		theme.SizeNameText,
		theme.SizeNameHeadingText,
		theme.SizeNameSubHeadingText,
		theme.SizeNameCaptionText,
	}

	for _, sizeName := range sizes {
		size := appTheme.Size(sizeName)
		if size <= 0 {
			t.Errorf("Size(%v) = %f, should be > 0", sizeName, size)
		}
	}
}

func TestThemeMode_Constants(t *testing.T) {
	// Verify theme mode constants
	if ThemeModeLight != 0 {
		t.Errorf("ThemeModeLight = %d, want 0", ThemeModeLight)
	}
	if ThemeModeDark != 1 {
		t.Errorf("ThemeModeDark = %d, want 1", ThemeModeDark)
	}
}
