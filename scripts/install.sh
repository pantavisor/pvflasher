#!/bin/bash
set -e

# Configuration
PROJECT_NAME="pvflasher"
BINARY_NAME="pvflasher"
REPO_URL="https://github.com/pantavisor/pvflasher"
API_URL="https://api.github.com/repos/pantavisor/pvflasher"

# Detect OS
OS=$(uname -s)
case "$OS" in
Linux | Darwin) ;;
*)
	echo "Unsupported OS: $OS"
	exit 1
	;;
esac

# Support specific version
VERSION=$1

echo "Installing $PROJECT_NAME on $OS..."

# Get version if not specified
if [ -z "$VERSION" ]; then
	echo "Fetching latest release information..."
	VERSION=$(curl -s "$API_URL/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')
fi

if [ -z "$VERSION" ]; then
	echo "Error: Could not determine version. Please check $REPO_URL"
	exit 1
fi

echo "Installing version: $VERSION"

if [ "$OS" = "Linux" ]; then
	# --- Linux: native package if possible, AppImage fallback ---
	ARCH=$(uname -m)
	case "$ARCH" in
	x86_64) FILE_ARCH="x86_64"; DEB_ARCH="amd64" ;;
	aarch64 | arm64) FILE_ARCH="aarch64"; DEB_ARCH="arm64" ;;
	*)
		echo "Unsupported architecture: $ARCH"
		exit 1
		;;
	esac

	# Native package filenames omit the 'v' prefix (e.g. 0.0.2, not v0.0.2)
	PKG_VERSION="${VERSION#v}"

	# Detect package manager (order matters: pacman before apt/dnf)
	if command -v pacman &>/dev/null; then
		PKG_MGR="pacman"
	elif command -v apt &>/dev/null; then
		PKG_MGR="apt"
	elif command -v dnf &>/dev/null; then
		PKG_MGR="dnf"
	elif command -v yum &>/dev/null; then
		PKG_MGR="yum"
	else
		PKG_MGR=""
	fi

	if [ "$PKG_MGR" = "pacman" ]; then
		PKG_FILE="pvflasher-${PKG_VERSION}-1-${FILE_ARCH}.pkg.tar.zst"
		echo "Detected Arch Linux (pacman). Downloading $PKG_FILE..."
		TMP_FILE=$(mktemp /tmp/pvflasher-XXXXXX.pkg.tar.zst)
		curl -fL -o "$TMP_FILE" "$REPO_URL/releases/download/$VERSION/$PKG_FILE"
		echo "Installing (requires sudo)..."
		sudo pacman -U --noconfirm "$TMP_FILE"
		rm -f "$TMP_FILE"

	elif [ "$PKG_MGR" = "apt" ]; then
		PKG_FILE="pvflasher_${PKG_VERSION}_${DEB_ARCH}.deb"
		echo "Detected Debian/Ubuntu (apt). Downloading $PKG_FILE..."
		TMP_FILE=$(mktemp /tmp/pvflasher-XXXXXX.deb)
		curl -fL -o "$TMP_FILE" "$REPO_URL/releases/download/$VERSION/$PKG_FILE"
		echo "Installing (requires sudo)..."
		sudo apt install -y "$TMP_FILE"
		rm -f "$TMP_FILE"

	elif [ "$PKG_MGR" = "dnf" ] || [ "$PKG_MGR" = "yum" ]; then
		PKG_FILE="pvflasher-${PKG_VERSION}-1.${FILE_ARCH}.rpm"
		echo "Detected RPM-based distro ($PKG_MGR). Downloading $PKG_FILE..."
		TMP_FILE=$(mktemp /tmp/pvflasher-XXXXXX.rpm)
		curl -fL -o "$TMP_FILE" "$REPO_URL/releases/download/$VERSION/$PKG_FILE"
		echo "Installing (requires sudo)..."
		sudo $PKG_MGR install -y "$TMP_FILE"
		rm -f "$TMP_FILE"

	else
		# Fallback: AppImage works on any Linux without a package manager
		echo "No supported package manager found. Falling back to AppImage..."
		BIN_DIR="$HOME/.local/bin"
		APP_DIR="$HOME/.local/share/applications"
		ICON_DIR="$HOME/.local/share/icons/hicolor/256x256/apps"
		mkdir -p "$BIN_DIR" "$APP_DIR" "$ICON_DIR"

		APPIMAGE_URL="$REPO_URL/releases/download/$VERSION/PvFlasher-$VERSION-$FILE_ARCH.AppImage"
		echo "Downloading AppImage..."
		curl -fL -o "$BIN_DIR/$BINARY_NAME" "$APPIMAGE_URL"
		chmod +x "$BIN_DIR/$BINARY_NAME"

		echo "Downloading icon..."
		curl -sL -o "$ICON_DIR/$BINARY_NAME.png" "https://raw.githubusercontent.com/pantavisor/pvflasher/main/Icon.png"

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

		echo ""
		echo "Installation complete!"
		echo "You can now run '$BINARY_NAME' from your terminal or application menu."
		echo "Note: Make sure $BIN_DIR is in your PATH."
		exit 0
	fi

	echo ""
	echo "Installation complete!"
	echo "You can now run '$BINARY_NAME' from your terminal or application menu."

elif [ "$OS" = "Darwin" ]; then
	# --- macOS: zip install to /usr/local/bin ---
	ARCH=$(uname -m)
	case "$ARCH" in
	x86_64) FILE_ARCH="amd64" ;;
	arm64) FILE_ARCH="arm64" ;;
	*)
		echo "Unsupported architecture: $ARCH"
		exit 1
		;;
	esac

	INSTALL_DIR="/usr/local/bin"
	ZIP_URL="$REPO_URL/releases/download/$VERSION/pvflasher-darwin-$FILE_ARCH.zip"
	ZIP_FILE=$(mktemp /tmp/pvflasher-XXXXXX.zip)
	EXTRACT_DIR=$(mktemp -d /tmp/pvflasher-XXXXXX)

	echo "Downloading..."
	curl -fL -o "$ZIP_FILE" "$ZIP_URL"

	echo "Extracting..."
	unzip -o "$ZIP_FILE" -d "$EXTRACT_DIR"

	echo "Installing to $INSTALL_DIR (requires sudo)..."
	sudo cp "$EXTRACT_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
	sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"

	rm -f "$ZIP_FILE"
	rm -rf "$EXTRACT_DIR"

	echo ""
	echo "Installation complete!"
	echo "You can now run '$BINARY_NAME' from your terminal."
fi
