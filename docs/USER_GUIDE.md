# User Guide

This guide provides detailed instructions for using **pvflasher** via both the CLI and GUI.

## 🖥️ CLI Usage

The Command Line Interface (CLI) is designed for efficiency and automation.

### `pvflasher list`

Lists all available block devices on the system. It filters for removable devices where possible and indicates if a device is currently mounted.

**Example:**
```bash
$ pvflasher list
Available devices:
- /dev/sdb: SanDisk Ultra 16GB (Removable) [15931539456 bytes]
- /dev/sdc: Generic Flash Disk (Removable) [Mounted: /media/user/USB] [8053063680 bytes]
```

---

### `pvflasher copy`

Writes an image file to a target device.

**Syntax:**
```bash
pvflasher copy [flags] <image_path> <device_path>
```

**Arguments:**
*   `<image_path>`: Path to the source image (supports raw `.img`, `.iso`, `.wic` or compressed `.gz`, `.xz`, `.bz2`, `.zst`, `.zip`).
*   `<device_path>`: Path to the target block device (e.g., `/dev/sdX` on Linux, `\\.\PhysicalDriveN` on Windows).

### Windows Considerations

When using **pvflasher** on Windows, keep the following in mind:

1.  **Administrator Privileges**: Writing to physical drives requires elevated rights. Always run your terminal (for CLI) or the application (for GUI) as Administrator.
2.  **Device Paths**: Devices are identified using the `\\.\PhysicalDriveN` syntax. Use `pvflasher list` to find the correct index `N`.
3.  **Volume Dismounting**: pvflasher automatically attempts to dismount all volumes on the target disk before flashing to ensure exclusive access.

**Flags:**
*   `--bmap <path>`: Explicitly specify the path to a `.bmap` file. If not provided, pvflasher attempts to find a file with the same name as the image (e.g., `image.img.bmap` for `image.img.gz`).
*   `--force`: Allow writing to mounted devices or devices that appear to be system drives. **Use with caution.**
*   `--no-verify`: Skip the checksum verification step after flashing. Faster, but less safe.
*   `--no-eject`: Do not eject/unmount the device after flashing completes.
*   `--json`: Output progress and result in JSON format (useful for wrapping pvflasher in other tools).

**Examples:**

*   **Standard Flash (Auto-detect bmap):**
    ```bash
    pvflasher copy ubuntu-22.04.img.gz /dev/sdb
    ```

*   **Flash with Explicit Bmap:**
    ```bash
    pvflasher copy --bmap custom.bmap system.img /dev/sdc
    ```

*   **Flash Raw (No Bmap):**
    If no bmap is found or provided, pvflasher will perform a standard raw copy (dd-style), skipping empty blocks if sparse file detection is successful.

---


### `pvflasher create`

Generates a `.bmap` file from an existing sparse image file. This is useful if you have a raw image and want to benefit from faster flashing in the future.

**Syntax:**
```bash
pvflasher create [flags] <image_path>
```

**Flags:**
*   `-o, --output <path>`: Output filename for the bmap. Defaults to `<image_path>.bmap`.

**Example:**
```bash
pvflasher create my-backup.img
```

---


### `pvflasher verify`

Verifies the content of a device against a bmap file to ensure data integrity.

**Syntax:**
```bash
pvflasher verify [flags] <device_path>
```

**Flags:**
*   `--bmap <path>`: Path to the bmap file to verify against.

---


## 🖥️ GUI Usage

The Graphical User Interface provides a simple 3-step process.

1.  **Launch**: Run the `pvflasher` executable (or `pvflasher-gui` if built separately).
2.  **Select Image**:
    *   **Local File**: Click the "Select Image" area or drag and drop your image file. pvflasher will automatically look for a corresponding `.bmap` file.
    *   **Pantavisor**: Switch to the "Pantavisor" tab to browse official releases. Select the channel, version, and target device. The image will be downloaded automatically when you start the flash process.
3.  **Select Target**:
    *   Choose your USB drive from the dropdown list.
    *   The list automatically refreshes when devices are plugged/unplugged.
4.  **Flash**:
    *   Click "Flash".
    *   If you selected a Pantavisor image, it will be downloaded first.
    *   You will be prompted for your password (sudo/admin) to authorize the write operation.
    *   Watch the progress bar as it goes through Reading/Downloading -> Writing -> Verifying phases.

### Pantavisor Features in GUI

The Pantavisor tab allows you to:
*   **Auto-fetch**: Connects to Pantavisor CI to get the latest available images.
*   **Caching**: Downloaded images are cached locally to speed up future flashes.
*   **SHA256 Validation**: Automatically verifies the integrity of the downloaded image before flashing.

#### Managing the Image Cache

Downloaded Pantavisor images are stored in a local cache directory. If you want to free up disk space or force a fresh download, you can manually clean this directory:

*   **Linux/macOS**: `~/.pvflasher/images/`
*   **Windows**: `%USERPROFILE%\.pvflasher\images\`

To clean the cache, simply delete the files inside these directories.

### Settings
(Coming Soon)
*   **Verification**: Toggle on/off.
*   **Dark Mode**: Switch themes.
