# Developer Guide

This guide covers how to set up your development environment, build the project, and contribute to **pvflasher**.

## 🛠️ Prerequisites

*   **Go**: Version 1.21 or later.
*   **Node.js**: Version 16 or later (for the GUI frontend).
*   **Wails**: The Wails v2 CLI tool.
    ```bash
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    ```
*   **Build Tools**:
    *   **Linux**: `build-essential`, `libgtk-3-dev`, `libwebkit2gtk-4.0-dev`
    *   **Windows**: MinGW or TDM-GCC (required for CGO if building certain platform-specific components, though standard Go compiler suffices for the core).

## 🪟 Windows Development Notes

*   **Raw Disk I/O**: Testing disk I/O on Windows requires running the development build with Administrator privileges.
*   **Console Output**: The GUI build uses `-H=windowsgui`. For CLI output during development, the `cli/commands/console_windows.go` logic handles console attachment.

## 📂 Project Structure

*   `cmd/pvflasher`: Main entry point. Dispatches to CLI or GUI mode based on arguments.
*   `internal/`: Core logic (private packages).
    *   `bmap/`: XML parsing and generation.
    *   `image/`: Image reading and decompression.
    *   `device/`: Device enumeration.
    *   `flash/`: Flashing and verification engine.
    *   `platform/`: OS-specific I/O and privilege escalation.
*   `gui/`: Wails application.
    *   `frontend/`: React + TypeScript source code.
*   `cli/`: Cobra-based CLI command definitions.

## 🏗️ Building

We use a `Makefile` to manage builds and packaging.

### Basic Build

Build the unified binary (GUI + CLI) for your current platform:

```bash
make build
```

The output binary will be in `bin/pvflasher` (or `pvflasher.exe`).

### Development Mode

Run the app in "dev" mode with hot-reloading for the frontend:

```bash
make run
```

### Packaging

To create distributable packages:

*   **Debian/Ubuntu (.deb)**:
    ```bash
    make package-deb
    ```
*   **RedHat/Fedora (.rpm)**:
    ```bash
    make package-rpm
    ```
*   **Universal Linux (.AppImage)**:
    ```bash
    make package-appimage
    ```
*   **Windows Installer (.msi)**:
    ```bash
    make package-windows
    ```
*   **macOS (.dmg)**:
    *(Must be run on macOS)*
    ```bash
    make package-dmg
    ```

## 🔄 Releases and Self-Updates

Pushing a `v*` tag builds every package and creates the GitHub release. For stable tags the release job also runs `tools/updater`, which signs each update bundle (AppImage, `.tar.xz`, Windows and macOS `.zip`) with minisign and publishes `latest.json` in [Tauri's updater format](https://v2.tauri.app/plugin/updater/). The app reads it from `releases/latest/download/latest.json` (`internal/update`).

*   **Signing key**: the minisign private key is the `UPDATER_PRIVATE_KEY` repository secret (base64-encoded key file, like Tauri). The matching public key, ID `B711C321FA6FF262`, is compiled into `internal/update/pubkey.go`. Keep a backup of the private key: without it, existing installs can't receive updates.
*   **Downgrade protection**: each signature's trusted comment carries `file:` and `version:`, and the app refuses a bundle whose signed file name or version doesn't match the manifest.
*   **Testing a release**: point a build at another manifest with `PVFLASHER_UPDATE_MANIFEST=http://…/latest.json`; bundles must still be signed with the release key. Development builds (versions that aren't a plain tag) never update.
*   **Rotating the key**: ship a release whose `pubkey.go` holds the new key, signed with the old one; sign the releases after it with the new key.

## 🧪 Testing

Run all unit tests in `internal/` and `pkg/`:

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
