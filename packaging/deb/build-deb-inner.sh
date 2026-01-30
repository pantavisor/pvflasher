#!/bin/bash
set -e

# Inner script that runs inside Docker
# Called from build-deb.sh with version as argument

VERSION="${1:-1.0.0}"
GOARCH="${2:-amd64}"

# Map Go architecture to Debian architecture
case "$GOARCH" in
"amd64") ARCH="amd64" ;;
"arm64") ARCH="arm64" ;;
*)
	echo "Unsupported architecture: $GOARCH"
	exit 1
	;;
esac

# Ensure version starts with a digit for Debian
if [[ ! "$VERSION" =~ ^[0-9] ]]; then
	# If version doesn't start with digit (e.g. "ae1da28"), prepend "0.0.0+"
	VERSION="0.0.0+$VERSION"
fi

PKG_NAME="pvflasher"
PKG_DIR="$PKG_NAME-$VERSION-$ARCH"

cd /work/packaging/deb

# Cleanup on exit
trap 'rm -rf "$PKG_DIR"' EXIT

echo "Building .deb package: ${PKG_NAME}_${VERSION}_${ARCH}.deb"

# Clean previous build
rm -rf "$PKG_DIR" "${PKG_NAME}_${VERSION}_${ARCH}.deb"

# Create package directory structure
mkdir -p "$PKG_DIR/DEBIAN"
mkdir -p "$PKG_DIR/usr/bin"
mkdir -p "$PKG_DIR/usr/share/applications"
mkdir -p "$PKG_DIR/usr/share/icons/hicolor/256x256/apps"
mkdir -p "$PKG_DIR/usr/share/doc/$PKG_NAME"

# Copy control file
cp control "$PKG_DIR/DEBIAN/control"

# Update version and architecture in control file
sed -i "s/Version: .*/Version: $VERSION/" "$PKG_DIR/DEBIAN/control"
sed -i "s/Architecture: .*/Architecture: $ARCH/" "$PKG_DIR/DEBIAN/control"

# Copy binary (must be built first)
BINARY_PATH="/work/fyne-cross/bin/linux-$GOARCH/pvflasher"
if [ ! -f "$BINARY_PATH" ]; then
	echo "Error: Binary not found at $BINARY_PATH"
	exit 1
fi
cp "$BINARY_PATH" "$PKG_DIR/usr/bin/"
chmod 755 "$PKG_DIR/usr/bin/pvflasher"

# Copy desktop file
cp ../pvflasher.desktop "$PKG_DIR/usr/share/applications/"

# Create a simple icon (you can replace this with a real icon)
if [ -f "/work/Icon.png" ]; then
	cp /work/Icon.png "$PKG_DIR/usr/share/icons/hicolor/256x256/apps/pvflasher.png"
elif [ -f "/work/icon.png" ]; then
	cp /work/icon.png "$PKG_DIR/usr/share/icons/hicolor/256x256/apps/pvflasher.png"
fi

# Create copyright file
cat >"$PKG_DIR/usr/share/doc/$PKG_NAME/copyright" <<EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: pvflasher
Source: https://github.com/pantacor/pvflasher

Files: *
Copyright: 2025 Sergio Marin
License: MIT
EOF

# Create changelog
cat >"$PKG_DIR/usr/share/doc/$PKG_NAME/changelog.Debian.gz" <<EOF
pvflasher ($VERSION) stable; urgency=medium

  * Initial release

 -- Sergio Marin <sergio.marin@pantacor.com>  $(date -R)
EOF
gzip -9 "$PKG_DIR/usr/share/doc/$PKG_NAME/changelog.Debian.gz"

# Set permissions
find "$PKG_DIR" -type d -exec chmod 755 {} \;
find "$PKG_DIR" -type f -exec chmod 644 {} \;
chmod 755 "$PKG_DIR/usr/bin/pvflasher"

# Build the package
dpkg-deb --build --root-owner-group "$PKG_DIR"

# Move to release directory
mkdir -p /work/release/linux
mv "${PKG_DIR}.deb" "/work/release/linux/${PKG_NAME}_${VERSION}_${ARCH}.deb"

# Cleanup
rm -rf "$PKG_DIR"

echo "Package created: /work/release/linux/${PKG_NAME}_${VERSION}_${ARCH}.deb"
echo "Install with: sudo dpkg -i release/linux/${PKG_NAME}_${VERSION}_${ARCH}.deb"
