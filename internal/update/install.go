package update

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Kind is how this copy of pvflasher was installed, which decides the update
// bundle to download and how it is put in place.
type Kind string

const (
	KindAppImage Kind = "appimage" // Linux AppImage: replace the .AppImage file
	KindTarball  Kind = "tar"      // Linux binary from the .tar.xz or install.sh
	KindWindows  Kind = "zip"      // Windows pvflasher.exe from the .zip
	KindMacApp   Kind = "app"      // macOS pvflasher.app bundle from the .zip
	KindManaged  Kind = "managed"  // deb/rpm/pacman or not writable: notify only
)

// Install describes the running installation.
type Install struct {
	Kind Kind
	// Path is the file or bundle that gets replaced.
	Path string
	// CanSelfUpdate is false when the update has to go through a package
	// manager or needs rights the user doesn't have; Reason says why.
	CanSelfUpdate bool
	Reason        string
}

// PlatformKey is the manifest key for this install, e.g. "linux-x86_64-appimage".
func (i Install) PlatformKey() string {
	return runtime.GOOS + "-" + Arch() + "-" + string(i.Kind)
}

// DetectInstall inspects the running executable.
func DetectInstall() Install {
	if p := os.Getenv("APPIMAGE"); p != "" {
		return writable(Install{Kind: KindAppImage, Path: p})
	}
	exe, err := os.Executable()
	if err != nil {
		return Install{Kind: KindManaged, Reason: "cannot locate the running executable"}
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	switch runtime.GOOS {
	case "windows":
		return writable(Install{Kind: KindWindows, Path: exe})
	case "darwin":
		if i := strings.Index(exe, ".app/Contents/MacOS/"); i >= 0 {
			return writable(Install{Kind: KindMacApp, Path: exe[:i+len(".app")]})
		}
		return writable(Install{Kind: KindTarball, Path: exe})
	}

	// Files owned by the system package manager must be updated through it.
	for _, dir := range []string{"/usr/bin/", "/usr/sbin/", "/bin/", "/opt/"} {
		if strings.HasPrefix(exe, dir) {
			return Install{Kind: KindManaged, Path: exe, Reason: "installed by the system package manager"}
		}
	}
	return writable(Install{Kind: KindTarball, Path: exe})
}

// writable marks the install self-updatable when the user can replace it,
// which needs write access to the directory holding it.
func writable(i Install) Install {
	dir := filepath.Dir(i.Path)
	f, err := os.CreateTemp(dir, ".pvflasher-update-check-*")
	if err != nil {
		i.Kind = KindManaged
		i.Reason = "no permission to write to " + dir
		return i
	}
	name := f.Name()
	f.Close()
	os.Remove(name)
	i.CanSelfUpdate = true
	return i
}
