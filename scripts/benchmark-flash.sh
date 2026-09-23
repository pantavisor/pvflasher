#!/usr/bin/env bash
# Benchmark pvflasher against other flashing tools on the same image and drive.
#
# Every run starts from a cold page cache and is timed until the tool exits,
# i.e. until the data is on the device (each tool syncs before exiting).
#
# usage: sudo scripts/benchmark-flash.sh --device /dev/sdX --model "<model>" \
#            --image image.wic.bz2 [--bmap image.wic.bmap] [--pvflasher ./pvflasher] \
#            [--balena path/to/balena] [--only REGEX] [--reps N] [--out results.csv]
#
# --only runs just the tools whose name matches REGEX and appends to --out.
#
# --model must match the drive's model as reported by lsblk; it guards against
# typos in --device, because EVERY RUN ERASES THE DEVICE.
set -euo pipefail
export LC_ALL=C # decimal points in timings, whatever the user locale

DEVICE="" MODEL="" IMAGE="" BMAP="" PVFLASHER="pvflasher" BALENA="" ONLY="" REPS=1 OUT="flash-benchmark.csv"
while [ $# -gt 0 ]; do
	case "$1" in
	--device) DEVICE=$2; shift 2 ;;
	--model) MODEL=$2; shift 2 ;;
	--image) IMAGE=$2; shift 2 ;;
	--bmap) BMAP=$2; shift 2 ;;
	--pvflasher) PVFLASHER=$2; shift 2 ;;
	--balena) BALENA=$2; shift 2 ;;
	--only) ONLY=$2; shift 2 ;;
	--reps) REPS=$2; shift 2 ;;
	--out) OUT=$2; shift 2 ;;
	*) echo "unknown option: $1" >&2; exit 2 ;;
	esac
done

die() { echo "error: $*" >&2; exit 1; }
[ "$(id -u)" = 0 ] || die "run as root (sudo)"
[ -b "$DEVICE" ] || die "--device must be a block device"
[ -f "$IMAGE" ] || die "--image not found"
[ -n "$BMAP" ] || BMAP="${IMAGE%.*}.bmap"
[ -f "$BMAP" ] || die "bmap not found: $BMAP"

actual_model=$(lsblk -dno MODEL "$DEVICE" | xargs)
[ "$actual_model" = "$MODEL" ] || die "$DEVICE model is '$actual_model', expected '$MODEL'"
[ "$(lsblk -dno RM "$DEVICE" | xargs)" = 1 ] || [ "$(lsblk -dno TRAN "$DEVICE" | xargs)" = usb ] ||
	die "$DEVICE is not a removable/USB drive"
if lsblk -no MOUNTPOINTS "$DEVICE" | grep -qv '^$'; then
	die "$DEVICE has mounted partitions; unmount them first"
fi

case "$IMAGE" in
*.bz2) CAT=bzcat ;; *.gz) CAT=zcat ;; *.xz) CAT=xzcat ;; *.zst) CAT=zstdcat ;; *) CAT=cat ;;
esac

# sudo resets PATH, so also look where the invoking user installs tools.
if [ -z "$BALENA" ]; then
	user_home=$(getent passwd "${SUDO_USER:-root}" | cut -d: -f6)
	for c in "$(command -v balena || true)" "$user_home/.local/balena-cli/balena" "$user_home/.local/bin/balena"; do
		[ -n "$c" ] && [ -x "$c" ] && { BALENA=$c; break; }
	done
fi

# Tools are optional: missing ones are skipped (and reported).
declare -a NAMES CMDS
add() {
	if [ -z "$ONLY" ] || [[ "$1" =~ $ONLY ]]; then NAMES+=("$1"); CMDS+=("$2"); fi
}
skip() { echo "# skipped: $1 (not found)" >&2; }
add "dd (bs=4M, fsync)" "$CAT '$IMAGE' | dd of='$DEVICE' bs=4M iflag=fullblock oflag=direct conv=fsync status=none"
command -v bmaptool >/dev/null || skip bmaptool
command -v bmaptool >/dev/null &&
	add "bmaptool $(bmaptool --version 2>&1 | awk '{print $2}')" "bmaptool -q copy --bmap '$BMAP' '$IMAGE' '$DEVICE'"
command -v "$PVFLASHER" >/dev/null || skip pvflasher
command -v "$PVFLASHER" >/dev/null && {
	add "pvflasher (no verify)" "'$PVFLASHER' copy '$IMAGE' '$DEVICE' --bmap '$BMAP' --force --no-verify --no-eject"
	add "pvflasher (verify)" "'$PVFLASHER' copy '$IMAGE' '$DEVICE' --bmap '$BMAP' --force --no-eject"
}
command -v rpi-imager >/dev/null || skip rpi-imager
command -v rpi-imager >/dev/null && {
	add "Raspberry Pi Imager (no verify)" "rpi-imager --cli --disable-verify --disable-eject '$IMAGE' '$DEVICE'"
	add "Raspberry Pi Imager (verify)" "rpi-imager --cli --disable-eject '$IMAGE' '$DEVICE'"
}
if [ -n "$BALENA" ]; then
	add "balenaEtcher engine (etcher-sdk via balena CLI)" "'$BALENA' local flash '$IMAGE' --drive '$DEVICE' --yes"
else
	skip "balena CLI"
fi

image_size=$(stat -c %s "$IMAGE")
# Only numeric tag values: the bmap header comment also mentions the tags.
bmap_value() { sed -nE "s|^[[:space:]]*<$1>[[:space:]]*([0-9]+)[[:space:]]*</$1>.*|\1|p" "$BMAP" | head -1; }
raw_size=$(bmap_value ImageSize)
mapped=$(bmap_value MappedBlocksCount)
block=$(bmap_value BlockSize)

[ -n "$ONLY" ] && [ -f "$OUT" ] || echo "tool,rep,seconds,exit" >"$OUT"
[ ${#NAMES[@]} -gt 0 ] || die "no tools selected"
{
	echo "# image: $(basename "$IMAGE") ($image_size bytes compressed, $raw_size raw, $((mapped * block)) mapped)"
	echo "# device: $DEVICE $actual_model ($(lsblk -dno SIZE "$DEVICE" | xargs), $(lsblk -dno TRAN "$DEVICE" | xargs))"
	echo "# host: $(uname -sr), $(nproc) CPUs, $(date -Iseconds)"
} | tee -a "${OUT%.csv}.txt"

for rep in $(seq 1 "$REPS"); do
	for i in "${!NAMES[@]}"; do
		name=${NAMES[$i]}
		# A tool may have ejected (powered off) the card; wait for it to come back.
		if [ "$(lsblk -dno MODEL "$DEVICE" 2>/dev/null | xargs)" != "$MODEL" ]; then
			echo "  $DEVICE ($MODEL) is gone — re-insert the card to continue…"
			until [ "$(lsblk -dno MODEL "$DEVICE" 2>/dev/null | xargs)" = "$MODEL" ]; do sleep 1; done
			sleep 3
		fi
		# Stray mounts from udisks auto-mounting the freshly written partitions.
		for p in $(lsblk -lno PATH,MOUNTPOINTS "$DEVICE" | awk 'NF>1{print $1}'); do umount "$p" || true; done
		sync
		echo 3 >/proc/sys/vm/drop_caches
		sleep 2
		printf '%-34s rep %d … ' "$name" "$rep"
		start=$(date +%s.%N)
		set +e
		bash -c "${CMDS[$i]}" >"/tmp/flash-bench-$i.log" 2>&1
		rc=$?
		set -e
		secs=$(echo "$(date +%s.%N) - $start" | bc)
		printf '%6.1f s (exit %d)\n' "$secs" "$rc"
		echo "\"$name\",$rep,$secs,$rc" >>"$OUT"
	done
done
echo "results: $OUT"
