package version

import (
	_ "embed"
	"strings"
)

// Version is the current version of the application. It is set at build time
// with -ldflags "-X pvflasher/internal/version.Version=vX.Y.Z", or else taken
// from version.txt, which CI stamps before building: fyne-cross drops the
// -X flag because `fyne package` passes its own -ldflags.
var Version = "development"

//go:embed version.txt
var stamped string

func init() {
	if v := strings.TrimSpace(stamped); Version == "development" && v != "" {
		Version = v
	}
}
