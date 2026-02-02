#!/bin/bash
set -e

# Configuration
PROJECT_NAME="pvflasher"
BINARY_NAME="pvflasher"
REPO_URL="https://gitlab.com/pantacor/pvflasher"
API_URL="https://gitlab.com/api/v4/projects/pantacor%2Fpvflasher"

# Detect Architecture
ARCH=$(uname -m)
case "$ARCH" in
x86_64) FILE_ARCH="x86_64" ;;
aarch64 | arm64) FILE_ARCH="aarch64" ;;
*)
	echo "Unsupported architecture: $ARCH"
	exit 1
	;;
esac

# Support specific version
VERSION=$1

# Directories
BIN_DIR="$HOME/.local/bin"
APP_DIR="$HOME/.local/share/applications"
ICON_DIR="$HOME/.local/share/icons/hicolor/256x256/apps"

echo "Installing $PROJECT_NAME..."

# Create directories
mkdir -p "$BIN_DIR"
mkdir -p "$APP_DIR"
mkdir -p "$ICON_DIR"

# Get version if not specified
if [ -z "$VERSION" ]; then
	echo "Fetching latest release information..."
	VERSION=$(curl -s "$API_URL/releases" | grep -oP '"tag_name":"\K[^"]+' | head -1)
fi

if [ -z "$VERSION" ]; then
	echo "Error: Could not determine version. Please check $REPO_URL"
	exit 1
fi

echo "Installing version: $VERSION ($ARCH)"

# Construct download URL for AppImage
# Filename pattern: PvFlasher-x86_64.AppImage or PvFlasher-aarch64.AppImage
APPIMAGE_URL="$API_URL/packages/generic/$PROJECT_NAME/$VERSION/PvFlasher-$FILE_ARCH.AppImage"

echo "Downloading AppImage from $APPIMAGE_URL..."
curl -fL -o "$BIN_DIR/$BINARY_NAME" "$APPIMAGE_URL"
chmod +x "$BIN_DIR/$BINARY_NAME"

# Download Icon
echo "Downloading icon..."
curl -sL -o "$ICON_DIR/$BINARY_NAME.png" "$REPO_URL/-/raw/main/Icon.png"

# Create Desktop Entry
echo "Creating desktop entry..."
cat >"$APP_DIR/$BINARY_NAME.desktop" <<EOF
[Desktop Entry]
Name=PvFlasher
Comment=Cross-platform USB Image Flasher
Exec=$BIN_DIR/$BINARY_NAME
Icon=$BINARY_NAME
Terminal=false
Type=Application
Categories=System;Utility;
Keywords=usb;flash;image;disk;bmap;
StartupNotify=true
EOF

echo "Installation complete!"
echo "You can now run '$BINARY_NAME' from your terminal or application menu."
echo "Note: Make sure $BIN_DIR is in your PATH."
