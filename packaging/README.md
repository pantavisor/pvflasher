# Packaging

How PvFlasher's release packages are built. Every package below is produced by GitHub Actions (`.github/workflows/release.yml`) when a `v*` tag is pushed and uploaded to the GitHub release; pushes to `main` build them without publishing. The commands here are for building or debugging a package locally.

| Package | Files in the release | Built by |
| :--- | :--- | :--- |
| Linux binary | `pvflasher-linux-{x86_64,aarch64}.tar.xz` | `make release-ci-linux` (fyne-cross) |
| AppImage | `PvFlasher-vX.Y.Z-{x86_64,aarch64}.AppImage` | `scripts/build-appimage-host.sh` via `make release-ci-linux` |
| Debian / Fedora / Arch | `.deb`, `.rpm`, `.pkg.tar.zst` | `nfpm` with `nfpm.yaml` (`make package-deb-amd64`, `package-rpm-amd64`, `package-archlinux-amd64`) |
| Windows zip | `pvflasher-windows-{x86_64,aarch64}.zip` | `make release-ci-windows` (fyne-cross) |
| Windows installer | `PvFlasher-Setup-vX.Y.Z-{x86_64,aarch64}.exe` | `windows/build-nsis.sh` + `windows/pvflasher.nsi` |
| macOS app | `pvflasher-darwin-{amd64,arm64}.zip` | `make release-ci-darwin` (on macOS) |
| macOS disk image | `PvFlasher-vX.Y.Z-{x86_64,arm64}.dmg` | `dmg/build-dmg.sh` (on macOS) |
| Update manifest | `latest.json`, `*.sig` | `tools/updater` (see the [Developer Guide](../docs/DEVELOPER_GUIDE.md#-releases-and-self-updates)) |

The `.tar.xz`, AppImage, Windows `.zip` and macOS `.zip` files are also the self-updater's bundles, so their names must not change without updating `tools/updater`.

## Linux

```bash
make release-ci-linux VERSION=vX.Y.Z    # binaries, AppImages and archives in release/linux/
make package-deb-amd64 VERSION=vX.Y.Z   # or package-rpm-amd64, package-archlinux-amd64
```

Needs Docker (fyne-cross), `linuxdeploy` for the AppImage (downloaded automatically if missing), ImageMagick, and `nfpm` for the distribution packages (`make deps`). The desktop entries are `pvflasher.desktop` (system install) and `pvflasher-local.desktop` (`make install-local`); both set `StartupWMClass=PvFlasher` so desktops match the window to its launcher.

## Windows installer

```bash
packaging/windows/build-nsis.sh vX.Y.Z x86_64 path/to/pvflasher.exe release/windows
```

Needs `makensis` (NSIS) and ImageMagick, and runs on Linux as well as Windows. The installer:

*   installs per user into `%LOCALAPPDATA%\Programs\PvFlasher`, with no administrator rights (PvFlasher asks for them when it flashes), so the self-updater can replace `pvflasher.exe`;
*   adds a Start Menu shortcut, an optional desktop shortcut and an Apps & Features entry with an uninstaller;
*   keeps the user's settings and image cache (`%USERPROFILE%\.pvflasher`) on uninstall.

The installer isn't code-signed yet, so Windows SmartScreen may warn on first run.

`windows/build-msi.sh` and `windows/pvflasher.wxs` are a legacy WiX MSI setup from an earlier version of the app and are not used by the release.

## macOS

```bash
make release-ci-darwin VERSION=vX.Y.Z                 # pvflasher.app for amd64 and arm64
packaging/dmg/build-dmg.sh release/darwin/pvflasher-darwin-arm64/pvflasher.app \
    release/darwin/PvFlasher-vX.Y.Z-arm64.dmg
```

Runs on macOS only. Apps target macOS 11 and later (`MACOS_MIN` in the Makefile). Locally built apps are only ad hoc signed; in CI, when the `APPLE_*` secrets are set, the workflow signs the app with the Developer ID certificate, notarizes and staples it, then builds the DMG and signs, notarizes and staples that too. Sign and notarize the app before building the DMG, because the DMG copies the app as it is.

## Output

All build output goes to `release/` (git-ignored): `release/linux/`, `release/windows/` and `release/darwin/`.
