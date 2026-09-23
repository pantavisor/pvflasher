package update

import (
	"archive/tar"
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"aead.dev/minisign"
	"github.com/ulikunitz/xz"
)

// ProgressFunc receives download progress; total is -1 when unknown.
type ProgressFunc func(done, total int64)

// Apply downloads the release bundle, verifies its signature and replaces the
// running installation. It does not restart the app; call Relaunch for that.
func Apply(ctx context.Context, rel *Release, progress ProgressFunc) error {
	if rel == nil || !rel.Install.CanSelfUpdate {
		return errors.New("this installation cannot update itself")
	}
	pub, err := publicKey()
	if err != nil {
		return err
	}
	sig, err := decodeSignature(rel.Asset.Signature)
	if err != nil {
		return err
	}

	// Download next to the target so the final rename stays on one filesystem.
	dir := filepath.Dir(rel.Install.Path)
	tmp, err := os.CreateTemp(dir, ".pvflasher-update-*")
	if err != nil {
		return fmt.Errorf("preparing download: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	verifier, err := download(ctx, rel.Asset.URL, tmp, progress)
	tmp.Close()
	if err != nil {
		return err
	}
	if !verifier.Verify(pub, sig) {
		return errors.New("the update's signature is not valid; it was not installed")
	}
	if err := checkTrustedComment(sig, path.Base(rel.Asset.URL), rel.Version); err != nil {
		return err
	}

	switch rel.Install.Kind {
	case KindAppImage:
		return replaceFile(tmpName, rel.Install.Path)
	case KindTarball:
		return extractAndReplace(tmpName, rel.Install.Path, extractFromTarXz)
	case KindWindows:
		return extractAndReplace(tmpName, rel.Install.Path, extractFromZip)
	case KindMacApp:
		return replaceMacApp(tmpName, rel.Install.Path)
	}
	return fmt.Errorf("unsupported install kind %q", rel.Install.Kind)
}

func download(ctx context.Context, url string, w io.Writer, progress ProgressFunc) (*minisign.Reader, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{} // no overall timeout: bundles are large; ctx cancels
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading update: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading update: server returned %s", resp.Status)
	}

	// The minisign reader hashes the stream so the signature is checked
	// without holding the bundle in memory.
	verifier := minisign.NewReader(resp.Body)
	var done int64
	buf := make([]byte, 256<<10)
	for {
		n, rerr := verifier.Read(buf)
		if n > 0 {
			if _, err := w.Write(buf[:n]); err != nil {
				return nil, fmt.Errorf("saving update: %w", err)
			}
			done += int64(n)
			if progress != nil {
				progress(done, resp.ContentLength)
			}
		}
		if rerr == io.EOF {
			return verifier, nil
		}
		if rerr != nil {
			return nil, fmt.Errorf("downloading update: %w", rerr)
		}
	}
}

// checkTrustedComment makes sure the signature was made for this exact file
// and version, so an old signed bundle can't be replayed as an update.
func checkTrustedComment(sig []byte, file, version string) error {
	var s minisign.Signature
	if err := s.UnmarshalText(sig); err != nil {
		return fmt.Errorf("reading signature: %w", err)
	}
	fields := map[string]string{}
	for _, f := range strings.Fields(s.TrustedComment) {
		if k, v, ok := strings.Cut(f, ":"); ok {
			fields[k] = v
		}
	}
	if fields["file"] != file || strings.TrimPrefix(fields["version"], "v") != strings.TrimPrefix(version, "v") {
		return fmt.Errorf("the update's signature is for %s %s, not %s %s; it was not installed",
			fields["file"], fields["version"], file, version)
	}
	return nil
}

// replaceFile atomically puts src in place of dst, keeping it executable.
func replaceFile(src, dst string) error {
	if err := os.Chmod(src, 0o755); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		// A running .exe can't be overwritten, but it can be renamed away.
		old := dst + ".old"
		os.Remove(old)
		if err := os.Rename(dst, old); err != nil {
			return fmt.Errorf("replacing %s: %w", dst, err)
		}
		if err := os.Rename(src, dst); err != nil {
			os.Rename(old, dst)
			return fmt.Errorf("replacing %s: %w", dst, err)
		}
		return nil
	}
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("replacing %s: %w", dst, err)
	}
	return nil
}

type extractFunc func(archive string, w io.Writer) error

func extractAndReplace(archive, dst string, extract extractFunc) error {
	out, err := os.CreateTemp(filepath.Dir(dst), ".pvflasher-update-bin-*")
	if err != nil {
		return err
	}
	name := out.Name()
	defer os.Remove(name)
	err = extract(archive, out)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	return replaceFile(name, dst)
}

// extractFromTarXz copies the pvflasher binary out of the Linux .tar.xz.
func extractFromTarXz(archive string, w io.Writer) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	xr, err := xz.NewReader(f)
	if err != nil {
		return fmt.Errorf("reading update archive: %w", err)
	}
	tr := tar.NewReader(xr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return errors.New("update archive has no pvflasher binary")
		}
		if err != nil {
			return fmt.Errorf("reading update archive: %w", err)
		}
		if h.Typeflag == tar.TypeReg && path.Base(h.Name) == "pvflasher" && path.Base(path.Dir(h.Name)) == "bin" {
			_, err = io.Copy(w, tr)
			return err
		}
	}
}

// extractFromZip copies pvflasher.exe out of the Windows .zip.
func extractFromZip(archive string, w io.Writer) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return fmt.Errorf("reading update archive: %w", err)
	}
	defer zr.Close()
	for _, f := range zr.File {
		if strings.EqualFold(path.Base(f.Name), "pvflasher.exe") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			_, err = io.Copy(w, rc)
			return err
		}
	}
	return errors.New("update archive has no pvflasher.exe")
}

// replaceMacApp unpacks pvflasher.app from the macOS .zip and swaps bundles.
func replaceMacApp(archive, bundle string) error {
	stage, err := os.MkdirTemp(filepath.Dir(bundle), ".pvflasher-update-app-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)

	zr, err := zip.OpenReader(archive)
	if err != nil {
		return fmt.Errorf("reading update archive: %w", err)
	}
	defer zr.Close()
	newApp := ""
	for _, f := range zr.File {
		i := strings.Index(f.Name, ".app/")
		if i < 0 {
			continue
		}
		rel := f.Name[strings.LastIndex(f.Name[:i], "/")+1:] // "pvflasher.app/…"
		target := filepath.Join(stage, filepath.FromSlash(rel))
		if !strings.HasPrefix(target, stage+string(os.PathSeparator)) {
			return fmt.Errorf("update archive has an invalid path: %s", f.Name)
		}
		newApp = filepath.Join(stage, rel[:strings.Index(rel, ".app/")+len(".app")])
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := writeZipFile(f, target); err != nil {
			return err
		}
	}
	if newApp == "" {
		return errors.New("update archive has no .app bundle")
	}

	old := bundle + ".old"
	os.RemoveAll(old)
	if err := os.Rename(bundle, old); err != nil {
		return fmt.Errorf("replacing %s: %w", bundle, err)
	}
	if err := os.Rename(newApp, bundle); err != nil {
		os.Rename(old, bundle)
		return fmt.Errorf("replacing %s: %w", bundle, err)
	}
	os.RemoveAll(old)
	return nil
}

func writeZipFile(f *zip.File, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	mode := f.Mode().Perm()
	if mode == 0 {
		mode = 0o644
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, rc); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// Relaunch starts the freshly installed version. The caller should quit
// right after it returns.
func Relaunch(inst Install) error {
	var cmd *exec.Cmd
	switch inst.Kind {
	case KindMacApp:
		cmd = exec.Command("open", "-n", inst.Path)
	default:
		cmd = exec.Command(inst.Path)
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	markInheritedFDsCloseOnExec()
	return cmd.Start()
}

// CleanupPrevious removes what a previous update left behind (the renamed
// Windows executable, which could not be deleted while it was running).
func CleanupPrevious() {
	if runtime.GOOS != "windows" {
		return
	}
	if exe, err := os.Executable(); err == nil {
		os.Remove(exe + ".old")
	}
}
