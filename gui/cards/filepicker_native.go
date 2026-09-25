//go:build darwin || windows

package cards

import (
	"errors"
	"os"
	"strings"

	"github.com/ncruces/zenity"
)

// pickImageNative shows the operating system's own Open panel (NSOpenPanel on
// macOS, the Win32 file dialog on Windows). Fyne only uses a native picker on
// Linux through the desktop portal, so these platforms would otherwise get
// Fyne's built-in dialog. It blocks until the user answers, so call it off
// the UI goroutine. ok is false when the native picker couldn't be shown and
// the caller should fall back to Fyne's dialog.
func pickImageNative() (path string, ok bool, err error) {
	patterns := make([]string, len(ImageExtensions))
	for i, ext := range ImageExtensions {
		patterns[i] = "*" + ext
	}
	opts := []zenity.Option{
		zenity.Title("Choose an OS image"),
		zenity.FileFilters{{Name: "Disk images (" + strings.Join(patterns, " ") + ")", Patterns: patterns, CaseFold: true}},
	}
	if home, herr := os.UserHomeDir(); herr == nil {
		opts = append(opts, zenity.Filename(home+string(os.PathSeparator)))
	}

	path, err = zenity.SelectFile(opts...)
	switch {
	case errors.Is(err, zenity.ErrCanceled):
		return "", true, nil
	case err != nil:
		return "", false, err
	}
	return path, true, nil
}
