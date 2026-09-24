# Developer Guide

This guide covers how to set up your development environment, build the project, and contribute to **pvflasher**.

## 🛠️ Prerequisites

*   **Go** 1.24 or later.
*   **A C compiler** for cgo (the GUI uses [Fyne](https://fyne.io), which needs OpenGL/GLFW):
    *   **Linux**: `gcc`, plus `libgl1-mesa-dev xorg-dev` (Debian/Ubuntu) or the equivalent `mesa`/`libX*` development packages.
    *   **macOS**: Xcode Command Line Tools (`xcode-select --install`).
    *   **Windows**: a MinGW-w64 toolchain such as [MSYS2](https://www.msys2.org/).
*   Optional, for packaging: `fyne` and `fyne-cross` (`make deps` also installs `nfpm`), Docker for fyne-cross, `linuxdeploy` for AppImages, `nsis` and ImageMagick for the Windows installer.

## 🪟 Windows Development Notes

*   **Raw Disk I/O**: flashing needs Administrator rights; the app requests them with a UAC prompt when it starts writing.
*   **Console Output**: the GUI build uses `-H=windowsgui`. `cli/commands/console_windows.go` reattaches the console so CLI commands still print.

## 📂 Project Structure

*   `main.go`: entry point. With no arguments it starts the GUI, otherwise the CLI.
*   `gui/`: the Fyne desktop app.
    *   `app.go`, `flash.go`, `update.go`: window, flash flow, update checks.
    *   `cards/`: the Image, Target and Flash steps.
    *   `screens/`: progress, success and error screens.
    *   `pantavisor/`: Pantavisor release list and downloads.
    *   `util/`: theme, shared components, config.
*   `cli/commands/`: Cobra commands (`copy`, `list`, `create`, `verify`, `download`, `install`, `update`, `version`).
*   `pkg/flash/`: the flashing and verification engine (public, usable as a library).
*   `internal/`: private packages.
    *   `archive/`, `image/`: reading and decompressing images.
    *   `bmap/`: block-map parsing, creation and checksums.
    *   `device/`: device enumeration and the safe-target filter (`safety.go`).
    *   `platform/`: OS-specific raw I/O and privilege elevation.
    *   `update/`: the self-updater (manifest, signature checks, install).
    *   `version/`: the build version.
*   `tools/updater/`: signs release bundles and writes `latest.json` in CI.
*   `packaging/`: AppImage, DMG, Windows installer and Linux package files. See [packaging/README.md](../packaging/README.md).
*   `scripts/`: install scripts and `benchmark-flash.sh`.

## 🏗️ Building

### Basic Build

Build the unified binary (GUI + CLI) for your platform:

```bash
make build        # -> bin/pvflasher (bin/pvflasher.exe on Windows)
make run          # build and run the GUI
make install-local  # install to ~/.local/bin with a desktop entry
```

The version comes from `git describe`. Builds that aren't exactly a release tag report themselves as development builds and never self-update.

### Packaging

Release packages are built by GitHub Actions (`.github/workflows/release.yml`) when a `v*` tag is pushed; pushes to `main` build everything without publishing. You can run the same steps locally:

| Package | Command | Runs on |
| :--- | :--- | :--- |
| Linux binaries, AppImage, `.deb`, `.rpm`, Arch | `make release-ci-linux VERSION=vX.Y.Z` | Linux (Docker for fyne-cross) |
| Windows `.zip` | `make release-ci-windows VERSION=vX.Y.Z` | Linux (Docker for fyne-cross) |
| Windows installer | `packaging/windows/build-nsis.sh vX.Y.Z x86_64 path/to/pvflasher.exe release/windows` | Linux or Windows with NSIS |
| macOS `.app` / `.zip` | `make release-ci-darwin VERSION=vX.Y.Z` | macOS |
| macOS DMG | `packaging/dmg/build-dmg.sh path/to/pvflasher.app release/darwin/PvFlasher-vX.Y.Z-arm64.dmg` | macOS |

macOS builds target macOS 11 and later (`MACOS_MIN` in the Makefile); CI fails if the binary's minimum drifts. In CI the app and DMG are signed with the Developer ID certificate and notarized when these repository secrets are set: `APPLE_CERTIFICATE` (base64 `.p12`), `APPLE_CERTIFICATE_PASSWORD`, `APPLE_SIGNING_IDENTITY`, `APPLE_TEAM_ID`, `APPLE_ID` and `APPLE_APP_PASSWORD` (an app-specific password).

## 🔄 Releases and Self-Updates

To make a release, from an up-to-date `main`:

```bash
git-chglog --next-tag vX.Y.Z -o CHANGELOG.md
git commit -am "chore: update changelog for vX.Y.Z"
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin main vX.Y.Z
```

Wait for a release's workflow to finish before tagging the next one, so `releases/latest` (which the updater reads) points at the newest version.

Pushing a `v*` tag builds every package and creates the GitHub release. For stable tags the release job also runs `tools/updater`, which signs each update bundle (AppImage, `.tar.xz`, Windows and macOS `.zip`) with minisign and publishes `latest.json` in [Tauri's updater format](https://v2.tauri.app/plugin/updater/). The app reads it from `releases/latest/download/latest.json` (`internal/update`).

*   **Signing key**: the minisign private key is the `UPDATER_PRIVATE_KEY` repository secret (base64-encoded key file, like Tauri). The matching public key, ID `B711C321FA6FF262`, is compiled into `internal/update/pubkey.go`. Keep a backup of the private key: without it, existing installs can't receive updates.
*   **Downgrade protection**: each signature's trusted comment carries `file:` and `version:`, and the app refuses a bundle whose signed file name or version doesn't match the manifest.
*   **Testing a release**: point a build at another manifest with `PVFLASHER_UPDATE_MANIFEST=http://…/latest.json`; bundles must still be signed with the release key. Development builds (versions that aren't a plain tag) never update.
*   **Rotating the key**: ship a release whose `pubkey.go` holds the new key, signed with the old one; sign the releases after it with the new key.

## 🧪 Testing

Run the unit tests (the same set CI runs):

```bash
make test
```

## 🤝 Contributing

1.  Fork the repository.
2.  Create a feature branch (`git checkout -b feature/amazing-feature`).
3.  Commit your changes.
4.  Push to the branch.
5.  Open a Pull Request.

Please ensure `make test` passes before submitting.
