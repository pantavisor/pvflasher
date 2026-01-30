#!/bin/bash
set -e

# Build AppImage for pvflasher on the host (no docker run)
# Usage: ./build-appimage-host.sh [version] [arch]

VERSION="${1:-1.0.0}"
GOARCH="${2:-amd64}"

# Map Go architecture to AppImage architecture
case "$GOARCH" in
"amd64") APPIMAGE_ARCH="x86_64" ;;
"arm64") APPIMAGE_ARCH="aarch64" ;;
*)
	echo "Unsupported architecture: $GOARCH"
	exit 1
	;;
esac

# Ensure tools are available
if ! command -v linuxdeploy &>/dev/null; then
	echo "Error: linuxdeploy not found in PATH"
	exit 1
fi

PKG_NAME="PvFlasher"
APP_DIR="${PKG_NAME}.AppDir"

# Navigate to project root
cd "$(dirname "$0")/.."

echo "Building AppImage on host: ${PKG_NAME}-${VERSION}-${APPIMAGE_ARCH}.AppImage"

# Check if binary was built by fyne-cross
BINARY_PATH="fyne-cross/bin/linux-$GOARCH/pvflasher"
if [ ! -f "$BINARY_PATH" ]; then
	echo "Error: Binary not found at $BINARY_PATH. Did you run 'make package-linux-$GOARCH'?"
	exit 1
fi

# Clean previous build
rm -rf "packaging/appimage/$APP_DIR"

# Create AppDir structure
mkdir -p "packaging/appimage/$APP_DIR/usr/bin"
mkdir -p "packaging/appimage/$APP_DIR/usr/share/applications"
mkdir -p "packaging/appimage/$APP_DIR/usr/share/icons/hicolor/512x512/apps"

# Copy binary
cp "$BINARY_PATH" "packaging/appimage/$APP_DIR/usr/bin/pvflasher"
chmod +x "packaging/appimage/$APP_DIR/usr/bin/pvflasher"

# Create a desktop file for AppImage
cat >"packaging/appimage/$APP_DIR/pvflasher.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=PvFlasher
Comment=Cross-platform USB Image Flasher
Exec=pvflasher
Icon=pvflasher
Terminal=false
Categories=Utility;
EOF

cp "packaging/appimage/$APP_DIR/pvflasher.desktop" "packaging/appimage/$APP_DIR/usr/share/applications/"

# Copy and resize icon
if [ -f "Icon.png" ]; then
	convert "Icon.png" -resize 512x512 "packaging/appimage/$APP_DIR/pvflasher.png"
	cp "packaging/appimage/$APP_DIR/pvflasher.png" "packaging/appimage/$APP_DIR/usr/share/icons/hicolor/512x512/apps/pvflasher.png"
fi

# Create AppRun script
cat >"packaging/appimage/$APP_DIR/AppRun" <<'APPRUN_EOF'
#!/bin/bash
SELF=$(readlink -f "$0")
HERE=${SELF%/*}
export PATH="${HERE}/usr/bin:${PATH}"
export LD_LIBRARY_PATH="${HERE}/usr/lib:${HERE}/usr/lib/x86_64-linux-gnu:${LD_LIBRARY_PATH}"
export XDG_DATA_DIRS="${HERE}/usr/share:${XDG_DATA_DIRS:-/usr/local/share:/usr/share}"
exec "${HERE}/usr/bin/pvflasher" "$@"
APPRUN_EOF

chmod +x "packaging/appimage/$APP_DIR/AppRun"

# Use linuxdeploy
echo "Bundling dependencies with linuxdeploy..."
export NO_STRIP=true
export APPIMAGE_EXTRACT_AND_RUN=1

# Run linuxdeploy from the packaging/appimage directory
cd packaging/appimage
ARCH=$APPIMAGE_ARCH linuxdeploy --appdir="$APP_DIR" \
	--executable="$APP_DIR/usr/bin/pvflasher" \
	--desktop-file="$APP_DIR/usr/share/applications/pvflasher.desktop" \
	--icon-file="$APP_DIR/usr/share/icons/hicolor/512x512/apps/pvflasher.png" \
	--output=appimage

# Find and rename the generated AppImage
GENERATED_APPIMAGE=$(find . -maxdepth 1 -name "*.AppImage" -print -quit)
if [ -f "$GENERATED_APPIMAGE" ]; then
	mv "$GENERATED_APPIMAGE" "../../release/linux/${PKG_NAME}-${VERSION}-${APPIMAGE_ARCH}.AppImage"
	echo "AppImage created: release/linux/${PKG_NAME}-${VERSION}-${APPIMAGE_ARCH}.AppImage"
else
	echo "Error: Failed to generate AppImage"
	exit 1
fi

# Cleanup
rm -rf "$APP_DIR"
