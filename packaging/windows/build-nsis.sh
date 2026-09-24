#!/bin/bash
# Build the PvFlasher Windows installer with NSIS (runs on Linux or Windows).
# Usage: build-nsis.sh <version> <arch: x86_64|aarch64> <path/to/pvflasher.exe> <output-dir>
set -euo pipefail

VERSION="${1#v}"
ARCH="$2"
EXE="$(realpath "$3")"
OUTDIR="$(realpath -m "$4")"

# VIProductVersion needs a plain X.Y.Z; development builds become 0.0.0.
NUMERIC=$(echo "$VERSION" | sed -nE 's/^([0-9]+\.[0-9]+\.[0-9]+).*/\1/p')
NUMERIC="${NUMERIC:-0.0.0}"

HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/../.." && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# Windows icon from the app icon (multi-size .ico).
convert "$ROOT/Icon.png" -define icon:auto-resize=256,64,48,32,16 "$WORK/pvflasher.ico"

mkdir -p "$OUTDIR"
OUTFILE="$OUTDIR/PvFlasher-Setup-v${VERSION}-${ARCH}.exe"
makensis -V2 \
	-DVERSION="$NUMERIC" \
	-DARCH="$ARCH" \
	-DEXE="$EXE" \
	-DICON="$WORK/pvflasher.ico" \
	-DOUTFILE="$OUTFILE" \
	"$HERE/pvflasher.nsi"
echo "Installer created: $OUTFILE"
