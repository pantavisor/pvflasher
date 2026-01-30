#!/bin/bash
set -e

# Build Windows NSIS installer for pvflasher
# Usage: ./build-nsis.sh [version]
# NOTE: This script should be run on Windows or with cross-compilation setup

VERSION="${1:-1.0.0}"
APP_NAME="pvflasher"

echo "Building Windows NSIS installer: ${APP_NAME}-${VERSION}-Setup.exe"

# Navigate to project root
cd ../..

# Build with Wails (includes NSIS installer generation)
echo "Building Windows binary with NSIS installer..."
wails build -platform windows/amd64 -nsis

# The installer will be created in build/bin/

if [ -f "build/bin/${APP_NAME}-amd64-installer.exe" ]; then
    mv "build/bin/${APP_NAME}-amd64-installer.exe" "build/bin/${APP_NAME}-${VERSION}-Setup.exe"
    echo "Installer created: build/bin/${APP_NAME}-${VERSION}-Setup.exe"
else
    echo "Warning: Installer not found at expected location"
    echo "Check build/bin/ for output"
fi

echo "Installation note: Users will need Administrator privileges to install"
