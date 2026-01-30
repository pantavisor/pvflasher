# Troubleshooting

## Permission Denied

**Symptom:**
`pvflasher` fails immediately with a "permission denied" error or asks for a password repeatedly.

**Solution:**
Writing to raw block devices requires root/administrator privileges.
*   **Linux**: The CLI attempts to auto-elevate using `sudo`. Ensure you have sudo access. If using the GUI, a polkit dialog should appear.
*   **Windows**: You **must** run the Command Prompt, PowerShell, or the GUI application as **Administrator**. Right-click the application or terminal and select "Run as administrator".

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
