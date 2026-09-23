package update

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aead.dev/minisign"
	"github.com/ulikunitz/xz"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"0.0.12", "v0.0.11", 1},
		{"v0.0.11", "0.0.11", 0},
		{"0.0.9", "0.0.10", -1},
		{"0.1.0", "0.0.99", 1},
		{"0.0.12-rc1", "0.0.12", -1},
		{"0.0.12", "0.0.12-rc1", 1},
	}
	for _, tt := range tests {
		if got := CompareVersions(tt.a, tt.b); got != tt.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIsReleaseVersion(t *testing.T) {
	for v, want := range map[string]bool{
		"v0.0.11": true, "0.0.11": true, "v0.1.0-rc1": true,
		"development": false, "v0.0.11-3-gabc1234": false, "v0.0.11-dev": false,
	} {
		if got := IsReleaseVersion(v); got != want {
			t.Errorf("IsReleaseVersion(%q) = %v, want %v", v, got, want)
		}
	}
}

func TestBuiltInPublicKey(t *testing.T) {
	pub, err := PublicKey()
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%X", pub.ID()); got != "B711C321FA6FF262" {
		t.Errorf("key ID = %s", got)
	}
}

// fixture serves a signed .tar.xz bundle containing a new pvflasher binary.
type fixture struct {
	srv     *httptest.Server
	bundle  []byte
	sig     string
	release *Release
	target  string
}

func newFixture(t *testing.T, signedFile, signedVersion string) *fixture {
	t.Helper()
	pub, priv, err := minisign.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	old := publicKey
	publicKey = func() (minisign.PublicKey, error) { return pub, nil }
	t.Cleanup(func() { publicKey = old })

	// pvflasher-linux-x86_64.tar.xz with usr/local/bin/pvflasher = "new binary"
	var buf bytes.Buffer
	xw, _ := xz.NewWriter(&buf)
	tw := tar.NewWriter(xw)
	body := []byte("new binary")
	tw.WriteHeader(&tar.Header{Name: "usr/local/bin/pvflasher", Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg})
	tw.Write(body)
	tw.Close()
	xw.Close()

	f := &fixture{bundle: buf.Bytes()}
	r := minisign.NewReader(bytes.NewReader(f.bundle))
	r.Read(make([]byte, len(f.bundle)+1))
	f.sig = base64.StdEncoding.EncodeToString(r.SignWithComments(priv,
		"timestamp:1 file:"+signedFile+" version:"+signedVersion, "test"))

	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Like GitHub's release download redirect, which 404s JSON requests.
		if strings.Contains(req.Header.Get("Accept"), "application/json") {
			http.NotFound(w, req)
			return
		}
		switch req.URL.Path {
		case "/latest.json":
			json.NewEncoder(w).Encode(Manifest{Version: "0.0.12", Platforms: map[string]Platform{
				"linux-x86_64-tar": {Signature: f.sig, URL: f.srv.URL + "/pvflasher-linux-x86_64.tar.xz"},
			}})
		case "/pvflasher-linux-x86_64.tar.xz":
			w.Write(f.bundle)
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(f.srv.Close)

	f.target = filepath.Join(t.TempDir(), "pvflasher")
	os.WriteFile(f.target, []byte("old binary"), 0o755)
	f.release = &Release{
		Version: "0.0.12",
		Asset:   Platform{Signature: f.sig, URL: f.srv.URL + "/pvflasher-linux-x86_64.tar.xz"},
		Install: Install{Kind: KindTarball, Path: f.target, CanSelfUpdate: true},
	}
	return f
}

func TestApplyReplacesBinary(t *testing.T) {
	f := newFixture(t, "pvflasher-linux-x86_64.tar.xz", "0.0.12")
	var progressed int64
	if err := Apply(t.Context(), f.release, func(done, _ int64) { progressed = done }); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(f.target)
	if string(got) != "new binary" {
		t.Errorf("target = %q, want the new binary", got)
	}
	if progressed != int64(len(f.bundle)) {
		t.Errorf("progress reported %d of %d bytes", progressed, len(f.bundle))
	}
	if info, _ := os.Stat(f.target); info.Mode().Perm()&0o100 == 0 {
		t.Error("replaced binary is not executable")
	}
}

func TestApplyRejectsTamperedBundle(t *testing.T) {
	f := newFixture(t, "pvflasher-linux-x86_64.tar.xz", "0.0.12")
	f.bundle[len(f.bundle)/2] ^= 0xff
	if err := Apply(t.Context(), f.release, nil); err == nil {
		t.Fatal("tampered bundle was installed")
	}
	if got, _ := os.ReadFile(f.target); string(got) != "old binary" {
		t.Errorf("target changed to %q", got)
	}
}

func TestApplyRejectsReplayedOldRelease(t *testing.T) {
	// A validly signed bundle from an older release served as the new one.
	f := newFixture(t, "pvflasher-linux-x86_64.tar.xz", "0.0.9")
	if err := Apply(t.Context(), f.release, nil); err == nil {
		t.Fatal("bundle signed for another version was installed")
	}
	if got, _ := os.ReadFile(f.target); string(got) != "old binary" {
		t.Errorf("target changed to %q", got)
	}
}

func TestCheckFindsUpdate(t *testing.T) {
	f := newFixture(t, "pvflasher-linux-x86_64.tar.xz", "0.0.12")
	old := ManifestURL
	ManifestURL = f.srv.URL + "/latest.json"
	t.Cleanup(func() { ManifestURL = old })

	st, err := Check(t.Context(), "v0.0.11")
	if err != nil {
		t.Fatal(err)
	}
	if st.Latest != "0.0.12" || st.Update == nil || st.Update.Version != "0.0.12" {
		t.Fatalf("Check = %+v, want an update to 0.0.12", st)
	}
	if st, _ := Check(t.Context(), "v0.0.12"); st.Update != nil || st.Latest != "0.0.12" {
		t.Errorf("up-to-date install: %+v", st)
	}
	if st, _ := Check(t.Context(), "development"); !st.DevBuild || st.Update != nil || st.Latest != "0.0.12" {
		t.Errorf("development build: %+v", st)
	}
}
