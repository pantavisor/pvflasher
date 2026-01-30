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
