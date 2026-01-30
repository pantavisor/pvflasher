#!/bin/bash
set -e

# Build macOS DMG for pvflasher
# Usage: ./build-dmg.sh [version]
# NOTE: This script must be run on macOS

VERSION="${1:-1.0.0}"
APP_NAME="pvflasher"
DMG_NAME="${APP_NAME}-${VERSION}-macOS"

echo "Building macOS DMG: ${DMG_NAME}.dmg"

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "Error: This script must be run on macOS"
    exit 1
fi

# Check if binary exists
if [ ! -f "../../build/bin/${APP_NAME}.app" ]; then
    echo "Error: App bundle not found. Run 'wails build -platform darwin/universal' first."
    exit 1
fi

# Clean previous build
rm -rf "$DMG_NAME" "${DMG_NAME}.dmg"

# Create DMG staging directory
mkdir -p "$DMG_NAME"

# Copy app bundle
cp -R "../../build/bin/${APP_NAME}.app" "$DMG_NAME/"

# Create Applications symlink
ln -s /Applications "$DMG_NAME/Applications"

# Create DMG background (optional)
mkdir -p "$DMG_NAME/.background"
if [ -f "background.png" ]; then
    cp background.png "$DMG_NAME/.background/"
fi

# Create DMG
hdiutil create -volname "${APP_NAME}" \
    -srcfolder "$DMG_NAME" \
    -ov -format UDZO \
    "${DMG_NAME}.dmg"

# Move to release directory
mkdir -p ../../release/macos
mv "${DMG_NAME}.dmg" "../../release/macos/"

# Cleanup
rm -rf "$DMG_NAME"

echo "DMG created: release/macos/${DMG_NAME}.dmg"
echo "To install: Open the DMG and drag ${APP_NAME}.app to Applications"

# Optional: Notarize for distribution (requires Apple Developer account)
# echo "To notarize (optional):"
# echo "xcrun notarytool submit ${DMG_NAME}.dmg --keychain-profile \"AC_PASSWORD\" --wait"
# echo "xcrun stapler staple ${DMG_NAME}.dmg"
