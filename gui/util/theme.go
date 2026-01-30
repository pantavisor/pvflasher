package util

import (
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemeMode represents the current theme mode
type ThemeMode int

const (
	ThemeModeLight ThemeMode = iota
	ThemeModeDark
)

// AppTheme implements fyne.Theme with custom colors and dark mode support
type AppTheme struct {
	mode ThemeMode
	mu   sync.RWMutex
}

var (
	currentTheme *AppTheme
	themeMu      sync.RWMutex
)

// NewAppTheme creates a new custom theme
func NewAppTheme() *AppTheme {
	return &AppTheme{
		mode: ThemeModeLight,
	}
}

// GetTheme returns the current app theme singleton
func GetTheme() *AppTheme {
	themeMu.RLock()
	if currentTheme != nil {
		themeMu.RUnlock()
		return currentTheme
	}
	themeMu.RUnlock()

	themeMu.Lock()
	defer themeMu.Unlock()
	if currentTheme == nil {
		currentTheme = NewAppTheme()
	}
	return currentTheme
}

// SetMode changes the theme mode (light/dark)
func (t *AppTheme) SetMode(mode ThemeMode) {
	t.mu.Lock()
	t.mode = mode
	t.mu.Unlock()
}

// Mode returns the current theme mode
func (t *AppTheme) Mode() ThemeMode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.mode
}

// IsDark returns true if dark mode is active
func (t *AppTheme) IsDark() bool {
	return t.Mode() == ThemeModeDark
}

// Toggle switches between light and dark mode
func (t *AppTheme) Toggle() {
	t.mu.Lock()
	if t.mode == ThemeModeLight {
		t.mode = ThemeModeDark
	} else {
		t.mode = ThemeModeLight
	}
	t.mu.Unlock()
}

// Color returns the color for the given theme color name
func (t *AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	t.mu.RLock()
	isDark := t.mode == ThemeModeDark
	t.mu.RUnlock()

	switch name {
	case theme.ColorNameBackground:
		if isDark {
			return ColorDarkBackground
		}
		return ColorBackground

	case theme.ColorNameButton:
		if isDark {
			return ColorDarkCardBg
		}
		return ColorPrimary

	case theme.ColorNameDisabledButton:
		// Lighter gray for disabled button background
		if isDark {
			return color.NRGBA{R: 0x4B, G: 0x50, B: 0x58, A: 0xFF}
		}
		return color.NRGBA{R: 0xE5, G: 0xE7, B: 0xEB, A: 0xFF}

	case theme.ColorNameDisabled:
		// Text color for disabled state - needs good contrast
		if isDark {
			return color.NRGBA{R: 0x9C, G: 0xA3, B: 0xAF, A: 0xFF}
		}
		return color.NRGBA{R: 0x6B, G: 0x72, B: 0x80, A: 0xFF}

	case theme.ColorNameError:
		return ColorError

	case theme.ColorNameFocus:
		return ColorPrimary

	case theme.ColorNameForeground:
		if isDark {
			return ColorDarkText
		}
		return ColorText

	case theme.ColorNameForegroundOnPrimary:
		// White text on primary (blue) background
		return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	case theme.ColorNameForegroundOnSuccess:
		// White text on success (green) background
		return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	case theme.ColorNameForegroundOnWarning:
		// Dark text on warning (amber) background for better contrast
		return color.NRGBA{R: 0x11, G: 0x18, B: 0x27, A: 0xFF}

	case theme.ColorNameForegroundOnError:
		// White text on error (red) background
		return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	case theme.ColorNameHover:
		if isDark {
			return color.NRGBA{R: 0x3D, G: 0x40, B: 0x47, A: 0xFF}
		}
		// Light blue hover that works well with dark text
		return color.NRGBA{R: 0xDB, G: 0xEA, B: 0xFE, A: 0xFF}

	case theme.ColorNameInputBackground:
		if isDark {
			return ColorDarkCardBg
		}
		return ColorCardBg

	case theme.ColorNameInputBorder:
		if isDark {
			return ColorDarkBorder
		}
		return ColorBorder

	case theme.ColorNameOverlayBackground:
		if isDark {
			return color.NRGBA{R: 0x1F, G: 0x21, B: 0x25, A: 0xFF}
		}
		return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	case theme.ColorNamePlaceHolder:
		if isDark {
			return ColorDarkTextSecondary
		}
		return ColorTextSecondary

	case theme.ColorNamePressed:
		if isDark {
			return color.NRGBA{R: 0x4D, G: 0x50, B: 0x57, A: 0xFF}
		}
		return color.NRGBA{R: 0xD1, G: 0xD5, B: 0xDB, A: 0xFF}

	case theme.ColorNamePrimary:
		return ColorPrimary

	case theme.ColorNameScrollBar:
		if isDark {
			return color.NRGBA{R: 0x5F, G: 0x62, B: 0x69, A: 0xFF}
		}
		return color.NRGBA{R: 0x9C, G: 0xA3, B: 0xAF, A: 0xFF}

	case theme.ColorNameSelection:
		if isDark {
			return color.NRGBA{R: 0x25, G: 0x63, B: 0xEB, A: 0x60}
		}
		return color.NRGBA{R: 0x25, G: 0x63, B: 0xEB, A: 0x40}

	case theme.ColorNameSeparator:
		if isDark {
			return ColorDarkBorder
		}
		return ColorBorder

	case theme.ColorNameShadow:
		if isDark {
			return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x66}
		}
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x33}

	case theme.ColorNameSuccess:
		return ColorSuccess

	case theme.ColorNameWarning:
		return ColorWarning

	case theme.ColorNameMenuBackground:
		if isDark {
			return ColorDarkCardBg
		}
		return ColorCardBg

	case theme.ColorNameHeaderBackground:
		return ColorHeaderBg
	}

	// Fall back to default theme for unhandled color names
	return theme.DefaultTheme().Color(name, variant)
}

// Font returns the font for the given text style
func (t *AppTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the icon resource for the given icon name
func (t *AppTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size for the given size name
func (t *AppTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 18
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameInputRadius:
		return 6
	case theme.SizeNameSelectionRadius:
		return 4
	}
	return theme.DefaultTheme().Size(name)
}

// CurrentTextColor returns the appropriate text color for the current theme mode
func CurrentTextColor() color.Color {
	if GetTheme().IsDark() {
		return ColorDarkText
	}
	return ColorText
}

// CurrentSecondaryTextColor returns the appropriate secondary text color
func CurrentSecondaryTextColor() color.Color {
	if GetTheme().IsDark() {
		return ColorDarkTextSecondary
	}
	return ColorTextSecondary
}

// CurrentBackgroundColor returns the appropriate background color
func CurrentBackgroundColor() color.Color {
	if GetTheme().IsDark() {
		return ColorDarkBackground
	}
	return ColorBackground
}

// CurrentCardBackground returns the appropriate card background color
func CurrentCardBackground() color.Color {
	if GetTheme().IsDark() {
		return ColorDarkCardBg
	}
	return ColorCardBg
}

// CurrentBorderColor returns the appropriate border color
func CurrentBorderColor() color.Color {
	if GetTheme().IsDark() {
		return ColorDarkBorder
	}
	return ColorBorder
}
