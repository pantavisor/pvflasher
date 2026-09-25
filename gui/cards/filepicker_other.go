//go:build !darwin && !windows

package cards

// pickImageNative is not needed on Linux: with the flatpak build tag Fyne's
// file dialog already uses the desktop portal (the GNOME/KDE picker).
func pickImageNative() (path string, ok bool, err error) {
	return "", false, nil
}
