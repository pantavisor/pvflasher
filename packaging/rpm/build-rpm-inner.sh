#!/bin/bash
set -e

# Inner script that runs inside Docker
# Called from build-rpm.sh with version as argument

VERSION="${1:-1.0.0}"
GOARCH="${2:-amd64}"

# Map Go architecture to RPM architecture
case "$GOARCH" in
"amd64") RPM_ARCH="x86_64" ;;
"arm64") RPM_ARCH="aarch64" ;;
*)
	echo "Unsupported architecture: $GOARCH"
	exit 1
	;;
esac

PKG_NAME="pvflasher"
SPEC_FILE="pvflasher.spec"

cd /work/packaging/rpm

echo "Building RPM package: ${PKG_NAME}-${VERSION} for ${RPM_ARCH}"

# Create RPM build directories
mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

# Check if binary exists
BINARY_PATH="/work/fyne-cross/bin/linux-$GOARCH/pvflasher"
if [ ! -f "$BINARY_PATH" ]; then
	echo "Error: Binary not found at $BINARY_PATH"
	exit 1
fi

# Parse VERSION to handle hyphens (RPM doesn't allow hyphens in Version field)
# Convert "0e8b0ac-dev" to Version: "0e8b0ac" and Release: "0.dev"
RPM_VERSION="$VERSION"
RPM_RELEASE="1"

if [[ "$VERSION" == *"-"* ]]; then
	# Split on the last hyphen
	RPM_VERSION="${VERSION%-*}"
	SUFFIX="${VERSION##*-}"
	RPM_RELEASE="0.${SUFFIX}"
fi

# Create source tarball
TEMP_DIR=$(mktemp -d)
SRC_DIR="$TEMP_DIR/$PKG_NAME-$RPM_VERSION"
mkdir -p "$SRC_DIR"

cp "$BINARY_PATH" "$SRC_DIR/"
cp ../pvflasher.desktop "$SRC_DIR/"
if [ -f "/work/Icon.png" ]; then
	cp /work/Icon.png "$SRC_DIR/appicon.png"
elif [ -f "/work/icon.png" ]; then
	cp /work/icon.png "$SRC_DIR/appicon.png"
else
	# Create placeholder icon
	touch "$SRC_DIR/appicon.png"
fi

# Create tarball with the RPM_VERSION (without hyphens)
cd "$TEMP_DIR"
tar czf "${PKG_NAME}-${RPM_VERSION}.tar.gz" "${PKG_NAME}-${RPM_VERSION}"
mv "${PKG_NAME}-${RPM_VERSION}.tar.gz" ~/rpmbuild/SOURCES/

# Copy spec file and update version
cd /work/packaging/rpm
cp "$SPEC_FILE" ~/rpmbuild/SPECS/
sed -i "s/Version:.*/Version:        $RPM_VERSION/" ~/rpmbuild/SPECS/$SPEC_FILE
sed -i "s/Release:.*/Release:        $RPM_RELEASE%{?dist}/" ~/rpmbuild/SPECS/$SPEC_FILE

# Build RPM
rpmbuild --target "$RPM_ARCH" -bb ~/rpmbuild/SPECS/$SPEC_FILE

# Copy RPM to release directory
mkdir -p /work/release/linux
find ~/rpmbuild/RPMS -name "${PKG_NAME}-${RPM_VERSION}*.rpm" -exec cp {} /work/release/linux/ \;

# Cleanup
rm -rf "$TEMP_DIR"

RPM_FILE=$(find /work/release/linux -name "${PKG_NAME}-${RPM_VERSION}*${RPM_ARCH}*.rpm" -printf "%f\n" | head -1)
echo "Package created: /work/release/linux/${RPM_FILE}"
echo "Install with: sudo dnf install release/linux/${RPM_FILE}"
echo "         or: sudo rpm -i release/linux/${RPM_FILE}"
