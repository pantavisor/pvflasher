// Command updater signs release bundles and writes the latest.json manifest
// read by pvflasher's self-updater (see internal/update).
//
//	UPDATER_PRIVATE_KEY=<base64 minisign key> go run ./tools/updater \
//	    --version v0.0.12 --base-url https://github.com/pantavisor/pvflasher/releases/download/v0.0.12 \
//	    --changelog CHANGELOG.md --out release/latest.json release/linux/* release/windows/* release/darwin/*
//
// Files that are not update bundles are ignored, so passing every release
// asset is fine. Each bundle also gets a <name>.sig file next to it.
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"aead.dev/minisign"

	"pvflasher/internal/update"
)

// bundle patterns map release asset names to manifest platform keys. The
// first key is pvflasher's own ("<os>-<arch>-<kind>"); the plain
// "<os>-<arch>" key keeps the manifest readable by Tauri-style clients.
var bundles = []struct {
	pattern *regexp.Regexp
	keys    func(arch string) []string
}{
	{regexp.MustCompile(`^PvFlasher-v[^-]+-(x86_64|aarch64)\.AppImage$`),
		func(a string) []string { return []string{"linux-" + a + "-appimage", "linux-" + a} }},
	{regexp.MustCompile(`^pvflasher-linux-(x86_64|aarch64)\.tar\.xz$`),
		func(a string) []string { return []string{"linux-" + a + "-tar"} }},
	{regexp.MustCompile(`^pvflasher-windows-(x86_64|aarch64)\.zip$`),
		func(a string) []string { return []string{"windows-" + a + "-zip", "windows-" + a} }},
	{regexp.MustCompile(`^pvflasher-darwin-(amd64|arm64)\.zip$`),
		func(a string) []string {
			a = map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[a]
			return []string{"darwin-" + a + "-app", "darwin-" + a}
		}},
}

func main() {
	version := flag.String("version", "", "release version, e.g. v0.0.12")
	baseURL := flag.String("base-url", "", "URL the release assets are downloaded from")
	changelog := flag.String("changelog", "", "CHANGELOG.md to take the release notes from")
	out := flag.String("out", "latest.json", "manifest to write")
	flag.Parse()
	if *version == "" || *baseURL == "" {
		fail("--version and --base-url are required")
	}

	key, err := loadKey(os.Getenv("UPDATER_PRIVATE_KEY"))
	if err != nil {
		fail(err.Error())
	}
	pub, _ := key.Public().(minisign.PublicKey)
	if want, err := update.PublicKey(); err != nil || want.ID() != pub.ID() {
		fail(fmt.Sprintf("UPDATER_PRIVATE_KEY (key %X) does not match the public key built into pvflasher", pub.ID()))
	}

	m := update.Manifest{
		Version:   strings.TrimPrefix(*version, "v"),
		PubDate:   time.Now().UTC().Format(time.RFC3339),
		Platforms: map[string]update.Platform{},
	}
	if *changelog != "" {
		m.Notes = releaseNotes(*changelog, *version)
	}

	for _, file := range flag.Args() {
		name := filepath.Base(file)
		for _, b := range bundles {
			sub := b.pattern.FindStringSubmatch(name)
			if sub == nil {
				continue
			}
			sig, err := sign(key, file, name, m.Version)
			if err != nil {
				fail(err.Error())
			}
			encoded := base64.StdEncoding.EncodeToString(sig)
			if err := os.WriteFile(file+".sig", []byte(encoded), 0o644); err != nil {
				fail(err.Error())
			}
			for _, k := range b.keys(sub[1]) {
				m.Platforms[k] = update.Platform{Signature: encoded, URL: strings.TrimSuffix(*baseURL, "/") + "/" + name}
			}
			fmt.Printf("signed %s\n", name)
		}
	}
	if len(m.Platforms) == 0 {
		fail("no update bundles among the given files")
	}

	data, _ := json.MarshalIndent(m, "", "  ")
	if err := os.WriteFile(*out, append(data, '\n'), 0o644); err != nil {
		fail(err.Error())
	}
	keys := make([]string, 0, len(m.Platforms))
	for k := range m.Platforms {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Printf("wrote %s for %s: %s\n", *out, m.Version, strings.Join(keys, ", "))
}

// loadKey accepts the base64-encoded minisign key file (Tauri's format) or the
// key file itself.
func loadKey(s string) (minisign.PrivateKey, error) {
	var key minisign.PrivateKey
	s = strings.TrimSpace(s)
	if s == "" {
		return key, fmt.Errorf("UPDATER_PRIVATE_KEY is not set")
	}
	text := []byte(s)
	if !strings.Contains(s, "\n") {
		if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
			text = decoded
		}
	}
	lines := strings.Split(strings.TrimSpace(string(text)), "\n")
	if err := key.UnmarshalText([]byte(lines[len(lines)-1])); err != nil {
		return key, fmt.Errorf("reading UPDATER_PRIVATE_KEY: %w", err)
	}
	return key, nil
}

// sign produces a prehashed minisign signature whose trusted comment binds
// the file name and version (checked by the updater).
func sign(key minisign.PrivateKey, file, name, version string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := minisign.NewReader(f)
	if _, err := io.Copy(io.Discard, r); err != nil {
		return nil, err
	}
	trusted := fmt.Sprintf("timestamp:%d file:%s version:%s", time.Now().Unix(), name, version)
	untrusted := "signature from pvflasher updater key " + strings.ToUpper(strconv.FormatUint(key.ID(), 16))
	return r.SignWithComments(key, trusted, untrusted), nil
}

// releaseNotes extracts the version's section from a git-chglog CHANGELOG.
func releaseNotes(path, version string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	start := bytes.Index(data, []byte(`<a name="`+version+`"></a>`))
	if start < 0 {
		return ""
	}
	section := data[start:]
	if end := bytes.Index(section[1:], []byte(`<a name="`)); end >= 0 {
		section = section[:end+1]
	}
	var lines []string
	for _, l := range strings.Split(string(section), "\n") {
		// Keep the bullets and group headings, drop the anchor, title and date.
		if strings.HasPrefix(l, "<a ") || strings.HasPrefix(l, "## ") || strings.HasPrefix(l, "> ") {
			continue
		}
		lines = append(lines, l)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "updater:", msg)
	os.Exit(1)
}
