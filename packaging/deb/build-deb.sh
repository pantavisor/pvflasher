#!/bin/bash
set -e

# Build Debian package for pvflasher (inside Docker container)
# Usage: ./build-deb.sh [version] [arch]

VERSION="${1:-1.0.0}"
ARCH="${2:-amd64}"

# Build Docker image
DOCKER_IMAGE="pvflasher-deb-builder"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "Building Docker image: $DOCKER_IMAGE"
docker build -t "$DOCKER_IMAGE" "$SCRIPT_DIR"

echo "Building .deb package for $ARCH with Docker..."

# Run the build inside Docker
docker run --rm \
  -v "$PROJECT_ROOT:/work" \
  "$DOCKER_IMAGE" \
  /work/packaging/deb/build-deb-inner.sh "$VERSION" "$ARCH"


