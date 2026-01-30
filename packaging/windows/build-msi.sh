#!/bin/bash
set -e

# Build Windows MSI installer for pvflasher using WiX Toolset
# Usage: ./build-msi.sh [version]
# NOTE: Requires WiX Toolset v3.11 or later (on Windows)

VERSION="${1:-1.0.0}"
APP_NAME="PvFlasher"

echo "Building Windows MSI installer: ${APP_NAME}-${VERSION}.msi"

# Check if running on Windows or have wine
if ! command -v candle.exe &>/dev/null && ! command -v wine &>/dev/null; then
    echo "Error: WiX Toolset not found"
    echo "Please install WiX Toolset from: https://wixtoolset.org/releases/"
    echo "Or run this script on Windows"
    exit 1
fi

# Navigate to project root
cd ../..

# Check if binary exists
if [ ! -f "build/bin/pvflasher.exe" ]; then
    echo "Error: Windows binary not found"
    echo "Build with: wails build -platform windows/amd64"
    exit 1
fi

# Copy binary to packaging directory
cp build/bin/pvflasher.exe packaging/windows/

# Update version in WXS file
sed "s/Version=\"1.0.0\"/Version=\"$VERSION\"/" packaging/windows/pvflasher.wxs >packaging/windows/pvflasher-versioned.wxs

# Compile WiX source
cd packaging/windows
candle.exe -dSourceDir=. pvflasher-versioned.wxs

# Link to create MSI
light.exe -ext WixUIExtension -out "${APP_NAME}-${VERSION}.msi" pvflasher-versioned.wixobj

# Move to release directory
mkdir -p ../../release/windows
mv "${APP_NAME}-${VERSION}.msi" "../../release/windows/"

# Cleanup
rm -f pvflasher-versioned.wxs pvflasher-versioned.wixobj pvflasher.exe

echo "MSI installer created: release/windows/${APP_NAME}-${VERSION}.msi"
echo "Install with: msiexec /i release/windows/${APP_NAME}-${VERSION}.msi"

# Alternative: Build with docker (cross-platform)
echo ""
echo "To build MSI on Linux, you can use Docker:"
echo "docker run -v \$(pwd):/work -w /work wixtoolset/wix:latest ..."
