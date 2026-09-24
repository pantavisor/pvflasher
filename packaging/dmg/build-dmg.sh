#!/bin/bash
# Build a drag-to-Applications DMG for an existing PvFlasher.app (macOS only).
# Usage: build-dmg.sh <path/to/pvflasher.app> <output.dmg>
#
# Sign, notarize and staple the .app before calling this; the release
# workflow then signs and notarizes the DMG itself.
set -euo pipefail

APP="$1"
OUT="$2"

[[ "$OSTYPE" == darwin* ]] || { echo "build-dmg.sh must run on macOS" >&2; exit 1; }
[ -d "$APP" ] || { echo "app bundle not found: $APP" >&2; exit 1; }

STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT

# ditto keeps signatures, stapled tickets and extended attributes intact.
ditto "$APP" "$STAGE/PvFlasher.app"
ln -s /Applications "$STAGE/Applications"

mkdir -p "$(dirname "$OUT")"
rm -f "$OUT"
hdiutil create -volname "PvFlasher" -srcfolder "$STAGE" -fs HFS+ -format UDZO -ov "$OUT"
echo "DMG created: $OUT"
