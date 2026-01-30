#!/bin/bash
set -e

# Build AppImage for pvflasher (inside Docker container)
# Usage: ./build-appimage.sh [version] [arch]

VERSION="${1:-1.0.0}"
ARCH="${2:-amd64}"

# Build Docker image
DOCKER_IMAGE="pvflasher-appimage-builder"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "Building Docker image: $DOCKER_IMAGE"
docker build -t "$DOCKER_IMAGE" "$SCRIPT_DIR"

echo "Building AppImage for $ARCH with Docker..."

# Run the build inside Docker
docker run --rm \
  --device /dev/fuse \
  --cap-add SYS_ADMIN \
  --security-opt apparmor:unconfined \
  -v "$PROJECT_ROOT:/work" \
  "$DOCKER_IMAGE" \
  /work/packaging/appimage/build-appimage-inner.sh "$VERSION" "$ARCH"
