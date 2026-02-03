#!/bin/bash
set -e

# Inner script that runs inside Docker
# Called from build-appimage.sh with version as argument

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

PKG_NAME="PvFlasher"
APP_DIR="${PKG_NAME}.AppDir"

cd /work

# Cleanup on exit
trap 'rm -rf "packaging/appimage/$APP_DIR"' EXIT

echo "Building AppImage: ${PKG_NAME}-${VERSION}-${APPIMAGE_ARCH}.AppImage"

cd packaging/appimage

# Check if binary was built by fyne-cross
BINARY_PATH="/work/fyne-cross/bin/linux-$GOARCH/pvflasher"
if [ ! -f "$BINARY_PATH" ]; then
	echo "Error: Binary not found at $BINARY_PATH"
	exit 1
fi

# Clean previous build
rm -rf "$APP_DIR" "${PKG_NAME}-${VERSION}-${APPIMAGE_ARCH}.AppImage"

# Create AppDir structure
mkdir -p "$APP_DIR/usr/bin"
mkdir -p "$APP_DIR/usr/share/applications"
mkdir -p "$APP_DIR/usr/share/icons/hicolor/512x512/apps"

# Copy binary
cp "$BINARY_PATH" "$APP_DIR/usr/bin/"
chmod +x "$APP_DIR/usr/bin/pvflasher"

# Create a desktop file for AppImage
cat >"$APP_DIR/pvflasher.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=PvFlasher
Exec=pvflasher
Icon=pvflasher
Categories=Utility;
EOF

cp "$APP_DIR/pvflasher.desktop" "$APP_DIR/usr/share/applications/"

# Copy and resize icon to valid resolution (512x512)
ICON_FILE=""
if [ -f "/work/Icon.png" ]; then
	ICON_FILE="/work/Icon.png"
elif [ -f "/work/icon.png" ]; then
	ICON_FILE="/work/icon.png"
fi

if [ -n "$ICON_FILE" ]; then
	# Resize to 512x512 (a valid resolution for linuxdeploy)
	convert "$ICON_FILE" -resize 512x512 "$APP_DIR/pvflasher.png"
	cp "$APP_DIR/pvflasher.png" "$APP_DIR/usr/share/icons/hicolor/512x512/apps/pvflasher.png"
else
	echo "Warning: No icon found"
fi

# Create a simple AppRun script (Fyne doesn't need GNOME schemas like GTK)
if [ "$APPIMAGE_ARCH" = "x86_64" ]; then
	LIBS_ARCH="x86_64-linux-gnu"
elif [ "$APPIMAGE_ARCH" = "aarch64" ]; then
	LIBS_ARCH="aarch64-linux-gnu"
else
	LIBS_ARCH="x86_64-linux-gnu"
fi

cat >"$APP_DIR/AppRun" <<APPRUN_EOF
#!/bin/bash
SELF=\$(readlink -f "\$0")
HERE=\${SELF%/*}

# Set up library paths
export PATH="\${HERE}/usr/bin:\${PATH}"
export LD_LIBRARY_PATH="\${HERE}/usr/lib:\${HERE}/usr/lib/${LIBS_ARCH}:\${LD_LIBRARY_PATH}"

# Set up XDG paths for data files
export XDG_DATA_DIRS="\${HERE}/usr/share:\${XDG_DATA_DIRS:-/usr/local/share:/usr/share}"

# Run the application
exec "\${HERE}/usr/bin/pvflasher" "\$@"
APPRUN_EOF

chmod +x "$APP_DIR/AppRun"

# Use linuxdeploy to bundle all dependencies
echo "Bundling dependencies with linuxdeploy..."
export NO_STRIP=true
ARCH=$APPIMAGE_ARCH linuxdeploy --appdir="$APP_DIR" \
	--executable="$APP_DIR/usr/bin/pvflasher" \
	--desktop-file="$APP_DIR/usr/share/applications/pvflasher.desktop" \
	--icon-file="$APP_DIR/usr/share/icons/hicolor/512x512/apps/pvflasher.png" \
	--output=appimage

# The linuxdeploy tool creates the AppImage directly with the name pattern:
# PvFlasher-<version>-<arch>.AppImage
# We need to ensure we grab the right one
GENERATED_APPIMAGE="${PKG_NAME}-${APPIMAGE_ARCH}.AppImage"
if [ ! -f "$GENERATED_APPIMAGE" ]; then
	# linuxdeploy might put version in name if using special plugins, but usually it's just name-arch.AppImage
	# unless $VERSION is passed to linuxdeploy (which we didn't).
	# Wait, linuxdeploy doesn't take version arg directly for filename unless env var VERSION is set.
	# Let's try to find what it created.
	GENERATED_APPIMAGE=$(find . -maxdepth 1 -name "*.AppImage" -print -quit)
fi

# Rename and move
if [ -f "$GENERATED_APPIMAGE" ]; then
	mv "$GENERATED_APPIMAGE" "${PKG_NAME}-${VERSION}-${APPIMAGE_ARCH}.AppImage"
	# Move to release directory
	mkdir -p /work/release/linux
	mv "${PKG_NAME}-${VERSION}-${APPIMAGE_ARCH}.AppImage" "/work/release/linux/"
	echo "AppImage created: /work/release/linux/${PKG_NAME}-${VERSION}-${APPIMAGE_ARCH}.AppImage"
else
	echo "Error: Failed to generate AppImage"
	ls -l
	exit 1
fi

# Cleanup
rm -rf "$APP_DIR"

echo "Run with: ./release/linux/${PKG_NAME}-${VERSION}-${APPIMAGE_ARCH}.AppImage"
