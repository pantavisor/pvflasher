# Troubleshooting

## Permission Denied

**Symptom:**
`pvflasher` fails immediately with a "permission denied" error or asks for a password repeatedly.

**Solution:**
Writing to raw block devices requires root/administrator privileges.
*   **Linux**: The desktop app asks for your password through polkit when flashing starts; if no polkit agent is running it asks in its own dialog instead (sudo). The CLI auto-elevates with `sudo`. Make sure your user may use `sudo`.
*   **macOS**: The desktop app asks for an administrator password when flashing starts. Your account must be an administrator.
*   **Windows**: The desktop app shows a User Account Control prompt when flashing starts; accept it. For the CLI, run Command Prompt or PowerShell as **Administrator**.

## My Drive Isn't in the Target List

**Symptom:**
The SD card or USB drive you want to flash doesn't appear in the desktop app.

**Solution:**
The desktop app hides drives it considers unsafe to erase. Click **Why are N drives hidden?** under the list to see each hidden drive and the reason:
*   **Internal disk**: built-in disks are never offered.
*   **In use by the system**: the drive has a volume mounted outside the usual removable-media folders (for example `/`, `/home` or `/mnt/data`). Unmount it if it really is the target.
*   **No media inserted**: an empty card reader, or a card that was ejected. Re-insert the card.

The list refreshes by itself; click ↻ to refresh right away. The CLI (`pvflasher list` and `pvflasher copy`) doesn't hide anything, so double-check the device path there.

## macOS: "PvFlasher" Can't Be Opened

**Symptom:**
macOS says the app "cannot be opened", "is damaged" or "Not Opened: Apple could not verify…", or nothing happens when you open it.

**Solution:**
*   Use the **DMG** from the releases page (v1.2.0 or later); the app and DMG are signed and notarized by Apple. Builds before v1.1.3 were not notarized: allow them once in **System Settings → Privacy & Security → Open Anyway**, or better, update.
*   PvFlasher needs **macOS 11 (Big Sur) or later**. Builds before v1.1.2 wrongly required a very recent macOS and fail to open with error `-10825`; download the current release.

## Updates Don't Install

**Symptom:**
PvFlasher shows an update but offers **Open Download Page** instead of **Install and Restart**, or `pvflasher update` says "this copy can't update itself".

**Solution:**
The app can only replace itself when it can write to its own folder:
*   **Installed with a package manager** (`.deb`, `.rpm`, `pacman`): update with the package manager, or install the new package from the releases page.
*   **macOS app run from Downloads**: macOS runs it from a temporary read-only copy. Move **PvFlasher** into **Applications** and open it from there.
*   **Installed in a system folder** (for example `/usr/local/bin` or `C:\Program Files`): reinstall with the Windows installer, the AppImage or the install script, which install per user.

If the update check itself fails ("Couldn't reach the update server"), check your internet connection or proxy: updates come from `github.com`. Updates are also turned off in development builds.

## Image Type Not Recognized

**Symptom:**
The application fails to recognize the image file or doesn't show it in the file dialog.

**Solution:**
pvflasher supports `.img`, `.iso`, and `.wic` files, as well as compressed versions of these. Ensure your file has one of these extensions. If it's a multi-file archive like `.tar`, ensure the image is inside.

## Device Busy / Resource Busy

**Symptom:**
Error "device or resource busy" or "Text file busy" when trying to flash.

**Solution:**
The device is likely mounted or in use by another application.
1.  **Unmount**: Ensure all partitions on the target drive are unmounted. `pvflasher` attempts to open devices with exclusive access.
    *   Linux: `umount /dev/sdX1`
2.  **Close Apps**: Close file managers or other disk utilities that might be scanning the drive.
3.  **Force**: Use the `--force` flag in the CLI if you are sure you want to overwrite a mounted device (not recommended).

## "No such file or directory" for Bmap

**Symptom:**
You are trying to use bmap optimization but it fails to find the file.

**Solution:**
`pvflasher` looks for a `.bmap` file with the same base name as your image.
*   Image: `system.img.gz` -> Expected Bmap: `system.img.bmap`
*   Ensure the `.bmap` file exists in the same directory.
*   Alternatively, specify it manually: `pvflasher copy --bmap /path/to/file.bmap ...`

## Slow Flashing Speed

**Symptom:**
Flashing is slower than expected.

**Cause:**
*   **No Bmap**: If no `.bmap` file is present, `pvflasher` writes every block, including empty space (unless sparse detection works).
*   **USB 2.0**: Check if you are using a USB 2.0 port or a slow USB drive.
*   **Verification**: The verification step reads back the written data, effectively doubling the time. Use `--no-verify` to skip (at your own risk).

## Verify Failed

**Symptom:**
Checksum mismatch error at the end of the process.

**Solution:**
1.  **Bad Cable/Port**: Try a different USB port or cable.
2.  **Failing Drive**: The USB drive might be failing (bad sectors).
3.  **RAM Issues**: Unstable system RAM can cause data corruption during compression/decompression.

## Pantavisor Download Failed

**Symptom:**
Error "download failed" when trying to flash a Pantavisor image.

**Solution:**
1.  **Internet Connection**: Ensure you have a stable internet connection.
2.  **Disk Space**: Ensure you have enough free space in your temporary directory/cache to store the downloaded image.
3.  **Firewall/Proxy**: Check if your network blocks access to AWS S3 (where releases are hosted).
4.  **Corrupt Cache**: If a previous download was interrupted, the cached file might be corrupt. Try cleaning the cache directory:
    *   Linux/macOS: `rm -rf ~/.pvflasher/images/*`
    *   Windows: `del /q %USERPROFILE%\.pvflasher\images\*`

## Blurry UI on Chromebook / ChromeOS Linux

**Symptom:**
The UI looks blurry or scaled incorrectly when running `pvflasher` inside the Linux environment on ChromeOS.

**Solution:**
1.  Right-click the `pvflasher` icon in the shelf.
2.  Select `Use low density` or disable display scaling for the app.
3.  Restart the application.

This forces the app to render at native resolution instead of the blurry scaled mode.

## Interface Too Small After Fixing Blur

**Symptom:**
After disabling ChromeOS scaling, the application becomes sharp but the text and controls are too small.

**Solution:**
Launch the app with an explicit Fyne scale factor:

```bash
FYNE_SCALE=1.5 ./pvflasher
```

Adjust the value to match your display density.
