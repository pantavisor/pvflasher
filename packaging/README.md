# PVFlasher Packaging Guide

This directory contains packaging configurations and build scripts for creating distributable packages of PVFlasher for different platforms.

## Table of Contents

- [Quick Start](#quick-start)
- [Linux Packages](#linux-packages)
  - [Debian/Ubuntu (.deb)](#debianubuntu-deb)
  - [Fedora/RHEL (.rpm)](#fedorarhel-rpm)
  - [AppImage (Universal)](#appimage-universal)
- [macOS Package](#macos-package)
- [Windows Installers](#windows-installers)
- [Troubleshooting](#troubleshooting)

## Quick Start

Build packages using the Makefile from the project root:

```bash
# Build for current platform
make package-all VERSION=1.0.0

# Build specific package type
make package-deb VERSION=1.0.0
make package-rpm VERSION=1.0.0
make package-appimage VERSION=1.0.0

# Build all Linux packages
make package-linux VERSION=1.0.0
```

**All package artifacts are placed in the `release/` directory:**
- `release/linux/` - .deb, .rpm, .AppImage
- `release/macos/` - .dmg
- `release/windows/` - .exe, .msi

This directory is in `.gitignore` to keep your repository clean.

## Prerequisites

### All Platforms
- Go 1.19 or later
- Wails v2.11.0 or later
- Node.js and npm

### Linux
- For .deb: `dpkg`, `dpkg-deb`
- For .rpm: `rpm-build`, `rpmbuild`
- For AppImage: `wget` (to download appimagetool)

```bash
# Debian/Ubuntu
sudo apt-get install dpkg-dev rpm

# Fedora/RHEL
sudo dnf install rpm-build dpkg
```

### macOS
- Xcode Command Line Tools
- Running on macOS (required for DMG creation)

### Windows
- For NSIS: Bundled with Wails
- For MSI: WiX Toolset v3.11+ (optional)

## Linux Packages

### Debian/Ubuntu (.deb)

Creates a standard Debian package that can be installed with `dpkg` or `apt`.

**Build:**
```bash
make package-deb VERSION=1.0.0
```

**Manual build:**
```bash
cd packaging/deb
./build-deb.sh 1.0.0
```

**Install:**
```bash
sudo dpkg -i release/linux/pvflasher_1.0.0_amd64.deb
sudo apt-get install -f  # Fix dependencies if needed
```

**Uninstall:**
```bash
sudo apt-get remove pvflasher
```

**Package contents:**
- Binary: `/usr/bin/pvflasher`
- Desktop entry: `/usr/share/applications/pvflasher.desktop`
- Icon: `/usr/share/icons/hicolor/256x256/apps/pvflasher.png`

### Fedora/RHEL (.rpm)

Creates an RPM package for Red Hat-based distributions.

**Build:**
```bash
make package-rpm VERSION=1.0.0
```

**Manual build:**
```bash
cd packaging/rpm
./build-rpm.sh 1.0.0
```

**Install:**
```bash
sudo dnf install release/linux/pvflasher-1.0.0-1.*.rpm
# or
sudo rpm -i release/linux/pvflasher-1.0.0-1.*.rpm
```

**Uninstall:**
```bash
sudo dnf remove pvflasher
```

### AppImage (Universal)

Creates a self-contained AppImage that runs on most Linux distributions without installation.

**Build:**
```bash
make package-appimage VERSION=1.0.0
```

**Manual build:**
```bash
cd packaging/appimage
./build-appimage.sh 1.0.0
```

**Run:**
```bash
chmod +x release/linux/PVFlasher-1.0.0-x86_64.AppImage
./release/linux/PVFlasher-1.0.0-x86_64.AppImage
```

**Benefits:**
- No installation required
- Runs on any Linux distribution
- Portable - can run from USB drive
- No root privileges needed

## macOS Package

Creates a DMG disk image containing the application bundle.

**Important:** DMG creation must be performed on macOS.

**Build:**
```bash
# On macOS:
make package-dmg VERSION=1.0.0
```

**Manual build:**
```bash
# First build the macOS binary
wails build -platform darwin/universal -o PVFlasher

cd packaging/dmg
./build-dmg.sh 1.0.0
```

**Install:**
1. Open the DMG file
2. Drag PVFlasher.app to Applications folder
3. Eject the DMG

**Uninstall:**
```bash
rm -rf /Applications/PVFlasher.app
```

**Note about code signing:**
- For distribution, the app should be code signed
- Requires an Apple Developer account
- See commented sections in `build-dmg.sh` for notarization

## Windows Installers

### NSIS Installer (Recommended)

Creates a standard Windows installer using NSIS (bundled with Wails).

**Build:**
```bash
make package-windows VERSION=1.0.0
```

**Manual build:**
```bash
wails build -platform windows/amd64 -nsis -o pvflasher
```

**Features:**
- Standard Windows installer UI
- Start Menu shortcuts
- Uninstaller
- Administrator privileges prompt

**Install:**
Double-click `PVFlasher-1.0.0-Setup.exe` and follow prompts.

### MSI Installer (Alternative)

Creates an MSI installer using WiX Toolset.

**Prerequisites:**
- WiX Toolset v3.11 or later

**Build:**
```bash
# First build Windows binary
wails build -platform windows/amd64 -o pvflasher

# Then create MSI
cd packaging/windows
./build-msi.sh 1.0.0
```

**Install:**
```cmd
msiexec /i PVFlasher-1.0.0.msi
```

**Features:**
- Enterprise deployment support
- Group Policy installation
- Desktop and Start Menu shortcuts

## Cross-Platform Building

### From Linux

**Build Linux packages:**
```bash
make package-linux VERSION=1.0.0
```

**Build Windows installer:**
```bash
make package-windows VERSION=1.0.0
```

**Note:** macOS packages must be built on macOS.

### From macOS

All packages can be built from macOS:

```bash
# Linux packages (using Docker recommended)
make package-linux VERSION=1.0.0

# macOS package
make package-dmg VERSION=1.0.0

# Windows installer
make package-windows VERSION=1.0.0
```

### From Windows

Windows can build Windows packages natively. For Linux/macOS, use WSL or Docker.

## Directory Structure

```
packaging/
├── README.md              # This file
├── pvflasher.desktop        # Linux desktop entry
├── deb/                   # Debian packaging
│   ├── control           # Package metadata
│   └── build-deb.sh      # Build script
├── rpm/                   # RPM packaging
│   ├── pvflasher.spec      # RPM spec file
│   └── build-rpm.sh      # Build script
├── appimage/              # AppImage packaging
│   └── build-appimage.sh # Build script
├── dmg/                   # macOS packaging
│   └── build-dmg.sh      # Build script
└── windows/               # Windows packaging
    ├── build-nsis.sh     # NSIS build script
    ├── build-msi.sh      # MSI build script
    └── pvflasher.wxs       # WiX configuration
```

## Release Directory Structure

All built packages are organized in the `release/` directory:

```
release/
├── linux/
│   ├── pvflasher_1.0.0_amd64.deb
│   ├── pvflasher-1.0.0-1.x86_64.rpm
│   └── PVFlasher-1.0.0-x86_64.AppImage
├── macos/
│   └── PVFlasher-1.0.0-macOS.dmg
└── windows/
    ├── PVFlasher-1.0.0-Setup.exe
    └── PVFlasher-1.0.0.msi
```

This directory is included in `.gitignore` to keep binary packages out of version control.

## Package Metadata

All packages include:
- **Name:** PvFlasher
- **Description:** Cross-platform USB Image pvflasher
- **Author:** Sergio Marin
- **Version:** Specified during build
- **License:** MIT
- **Homepage:** https://github.com/pantacor/pvflasher

## Troubleshooting

### Binary not found error

**Problem:** "Binary not found. Run 'make build' first."

**Solution:**
```bash
cd /path/to/pvflasher
make build
```

### Permission denied

**Problem:** Cannot execute build scripts

**Solution:**
```bash
chmod +x packaging/*/build-*.sh
```

### Missing dependencies (Linux)

**Problem:** Package build fails due to missing tools

**Solution:**
```bash
# Debian/Ubuntu
sudo apt-get install dpkg-dev rpm build-essential

# Fedora/RHEL
sudo dnf install rpm-build dpkg-dev
```

### AppImage: appimagetool not found

**Solution:** The script will automatically download appimagetool. Ensure you have internet connectivity.

### macOS: Command not found (hdiutil)

**Problem:** Not running on macOS

**Solution:** DMG creation must be performed on macOS. Use a Mac or macOS VM.

### Windows: NSIS not found

**Solution:** NSIS is bundled with Wails. Ensure Wails v2.11.0+ is installed:
```bash
wails version
```

## Version Management

To release a new version:

1. Update version in `wails.json`
2. Update version in packaging control files if hardcoded
3. Build packages with new version:
   ```bash
   make package-all VERSION=x.y.z
   ```

## CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Build Packages

on:
  release:
    types: [created]

jobs:
  build-linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install dependencies
        run: sudo apt-get install -y dpkg-dev rpm
      - name: Build packages
        run: make package-linux VERSION=${{ github.ref_name }}
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: linux-packages
          path: packaging/**/*.{deb,rpm,AppImage}
```

## Support

For issues or questions about packaging:
- Open an issue: https://github.com/pantacor/pvflasher/issues
- Check documentation: https://wails.io/docs/guides/packaging

## License

The packaging scripts and configurations are licensed under the same terms as the PVFlasher application (MIT License).
