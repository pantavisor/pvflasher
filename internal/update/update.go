// Package update implements a Tauri-style self-updater.
//
// Every release publishes a latest.json manifest listing, per platform, the
// download URL of an update bundle and its minisign signature (both in the
// same shapes Tauri uses). The app fetches the manifest, downloads the bundle
// for its platform and install kind, verifies the signature against the public
// key compiled into the binary, and only then replaces itself.
//
// Unlike Tauri, the signed trusted comment also binds the file name and the
// version, so a tampered manifest cannot point at an older (validly signed)
// release to downgrade users.
package update

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ManifestURL is where the latest release's manifest is published. The
// PVFLASHER_UPDATE_MANIFEST environment variable overrides it, e.g. to test a
// release from a staging server; bundles must still carry a valid signature.
var ManifestURL = manifestURL()

func manifestURL() string {
	if u := os.Getenv("PVFLASHER_UPDATE_MANIFEST"); u != "" {
		return u
	}
	return "https://github.com/pantavisor/pvflasher/releases/latest/download/latest.json"
}

// Manifest is the latest.json document, in Tauri's format.
type Manifest struct {
	Version   string              `json:"version"`
	Notes     string              `json:"notes"`
	PubDate   string              `json:"pub_date"`
	Platforms map[string]Platform `json:"platforms"`
}

// Platform is one downloadable update bundle.
type Platform struct {
	// Signature is the base64-encoded minisign signature file, as in Tauri.
	Signature string `json:"signature"`
	URL       string `json:"url"`
}

// Release is an available update for this installation.
type Release struct {
	Version string
	Notes   string
	PubDate time.Time
	Asset   Platform
	Install Install
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

// FetchManifest downloads and parses the manifest.
func FetchManifest(ctx context.Context, url string) (*Manifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checking for updates: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checking for updates: server returned %s", resp.Status)
	}
	var m Manifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("reading update manifest: %w", err)
	}
	return &m, nil
}

// Status is the result of an update check.
type Status struct {
	Current string   // running version
	Latest  string   // newest published version, without "v"
	Update  *Release // nil when up to date or when this build can't update
	// DevBuild is set for development builds, which never update.
	DevBuild bool
}

// Check compares the running version with the latest release. Installs that
// cannot update themselves still get Status.Update (with
// Install.CanSelfUpdate false) so the user can be told about it.
func Check(ctx context.Context, current string) (*Status, error) {
	m, err := FetchManifest(ctx, ManifestURL)
	if err != nil {
		return nil, err
	}
	st := &Status{
		Current:  current,
		Latest:   strings.TrimPrefix(m.Version, "v"),
		DevBuild: !IsReleaseVersion(current),
	}
	if st.DevBuild || CompareVersions(m.Version, current) <= 0 {
		return st, nil
	}
	inst := DetectInstall()
	rel := &Release{Version: st.Latest, Notes: m.Notes, Install: inst}
	rel.PubDate, _ = time.Parse(time.RFC3339, m.PubDate)
	if inst.CanSelfUpdate {
		asset, ok := m.Platforms[inst.PlatformKey()]
		if !ok {
			// The release has no bundle for this platform; fall back to a notice.
			rel.Install.CanSelfUpdate = false
			rel.Install.Reason = "no update bundle for " + inst.PlatformKey()
		}
		rel.Asset = asset
	}
	st.Update = rel
	return st, nil
}

// Arch returns the architecture in Tauri's naming.
func Arch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	case "386":
		return "i686"
	case "arm":
		return "armv7"
	}
	return runtime.GOARCH
}

var releaseVersion = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-(rc|beta|alpha)[.\d]*)?$`)

// IsReleaseVersion reports whether v is a tagged release rather than a
// development or locally modified build (e.g. "development", "v0.0.11-3-gabc").
func IsReleaseVersion(v string) bool {
	return releaseVersion.MatchString(v)
}

// CompareVersions compares two semantic versions ("v" prefix optional),
// returning -1, 0 or 1. A pre-release sorts before its release.
func CompareVersions(a, b string) int {
	coreA, preA, _ := strings.Cut(strings.TrimPrefix(a, "v"), "-")
	coreB, preB, _ := strings.Cut(strings.TrimPrefix(b, "v"), "-")
	pa, pb := strings.Split(coreA, "."), strings.Split(coreB, ".")
	for i := 0; i < 3; i++ {
		var na, nb int
		if i < len(pa) {
			na, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			nb, _ = strconv.Atoi(pb[i])
		}
		if na != nb {
			if na < nb {
				return -1
			}
			return 1
		}
	}
	switch {
	case preA == preB:
		return 0
	case preA == "":
		return 1
	case preB == "":
		return -1
	case preA < preB:
		return -1
	}
	return 1
}

// decodeSignature accepts the base64 form used in latest.json as well as a
// raw minisign signature file.
func decodeSignature(sig string) ([]byte, error) {
	if strings.HasPrefix(sig, "untrusted comment:") {
		return []byte(sig), nil
	}
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(sig))
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding: %w", err)
	}
	return b, nil
}
