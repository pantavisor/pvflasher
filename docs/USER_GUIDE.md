# User Guide

This guide covers installing **pvflasher** and using it from the desktop app and the command line.

## 📥 Installing

Download the latest release from the [releases page](https://github.com/pantavisor/pvflasher/releases/latest):

| Platform | File | How to install |
| :--- | :--- | :--- |
| **macOS** | `PvFlasher-vX.Y.Z-arm64.dmg` (Apple Silicon) or `-x86_64.dmg` (Intel) | Open the DMG and drag **PvFlasher** into **Applications**. The app is signed and notarized by Apple. Requires macOS 11 or later. |
| **Windows** | `PvFlasher-Setup-vX.Y.Z-x86_64.exe` (or `-aarch64.exe`) | Run the installer. It installs for your user in `%LOCALAPPDATA%\Programs\PvFlasher` and needs no administrator rights. |
| **Linux** | `PvFlasher-vX.Y.Z-x86_64.AppImage` (or `-aarch64`) | Make it executable (`chmod +x`) and run it. |
| **Linux packages** | `.deb`, `.rpm`, `.pkg.tar.zst` | Install with your package manager (`apt`, `dnf`, `pacman`). |
| **Any** | `.tar.xz` / `.zip` archives, or the [install script](../README.md#quick-install) | Unpack and run the `pvflasher` binary. |

Flashing always needs administrator rights, whichever way you install. PvFlasher asks for them when it starts writing: a password prompt on Linux and macOS, a User Account Control prompt on Windows.

## 🖥️ Desktop App

![PvFlasher main window](../screenshot.png)

Run **PvFlasher** from your applications menu, or `pvflasher` with no arguments. The window has three steps.

### 1. Image

*   **Choose File…** opens a file dialog. You can also drag an image file onto the window. Supported: `.img`, `.iso`, `.wic`, and compressed `.gz`, `.bz2`, `.xz`, `.zst`, `.zip`, `.tar`, `.tgz`.
*   If a `.bmap` file sits next to the image (for example `system.img.bmap` for `system.img.gz`), PvFlasher uses it and only writes the blocks that hold data. The card shows whether a block map was found.
*   **Pantavisor Release…** opens the official release picker. Choose the channel, version and device, using the same names as [pantavisor.io/downloads](https://pantavisor.io/downloads/). The image is downloaded and checksum-verified when flashing starts.

### 2. Target

*   Pick the SD card or USB drive from the list. The list refreshes by itself when drives are plugged in or removed; the ↻ button refreshes it immediately.
*   To protect your computer, PvFlasher never offers internal disks, drives in use by the system (for example a USB disk mounted at `/home`) or empty card readers. **Why are N drives hidden?** lists them and gives the reason for each one.
*   A drive mounted by your desktop, such as a freshly inserted SD card, is shown with "mounted". It is unmounted before writing.

### 3. Flash

*   **Verify after writing** reads the written data back from the card and checks it. Keep it on unless you are in a hurry.
*   **Eject when finished** ejects the card when it's done, so it can be removed safely.
*   **Unmount without asking** skips the question about mounted volumes.
*   **Flash** asks you to confirm which drive will be erased, then asks for administrator rights and starts.

While flashing, PvFlasher shows the phase (downloading, writing, verifying, finishing), speed, time left and bytes written. **Cancel** asks before stopping, because an interrupted card won't boot until it is flashed again. When it's done you get the statistics (data written, duration, speed, method, verification) and a **View Log** button; **Flash Another** keeps the image and lets you pick the next card.

### Settings

The gear icon in the top-right corner opens Settings:

*   **Appearance**: follow the system's light or dark mode, or force Light or Dark.
*   **Updates**: your version, the latest release, and **Update to vX…** when there is one. See [Updates](#-updates).
*   **Help**: a link to the [troubleshooting guide](TROUBLESHOOTING.md).

## 🔄 Updates

PvFlasher checks for a new release once a day. When there is one, it shows **Update available: vX** in the bottom-right corner and a dialog with the release notes and three choices: **Install and Restart**, **Later** and **Skip This Version**. Clicking the version number in the corner opens the **Updates** section of Settings, where you can check again or turn automatic checks off.

Every update is checked against Pantacor's signing key before it is installed, and PvFlasher never updates while it is flashing.

| How PvFlasher was installed | What happens |
| :--- | :--- |
| macOS app in Applications | Updates itself and restarts. |
| Windows installer | Updates itself and restarts. |
| Linux AppImage, `.tar.xz` or install script | Updates itself and restarts. |
| `.deb`, `.rpm`, `pacman` | Shows the update; install it with your package manager. |
| macOS app run straight from Downloads | Shows the update; move PvFlasher to Applications to let it update itself. |

From the command line, use `pvflasher version` and `pvflasher update` (below).

## ⌨️ Command Line

Run `pvflasher <command> --help` for all flags.

### `pvflasher list`

Lists the block devices on the system with their size and mount points. Unlike the desktop app, it lists every disk, including internal ones, so double-check the device before writing to it.

```console
$ pvflasher list
Available devices:
- /dev/sda: GoPro Quik_Key (Removable)  [31914983424 bytes]
- /dev/sdb: ASMT ASM236X_NVME  [Mounted: /home/projects] [1024209543168 bytes]
```

### `pvflasher copy <image> <device>`

Writes an image to a device, using a `.bmap` file when one is found next to the image.

```bash
pvflasher copy ubuntu-24.04.img.xz /dev/sdX
pvflasher copy --bmap custom.bmap system.img /dev/sdX
```

| Flag | Meaning |
| :--- | :--- |
| `--bmap <path>` | Use this block map instead of looking for one next to the image. |
| `--no-verify` | Skip reading the data back after writing. |
| `--no-eject` | Don't eject the device when done. |
| `--force` | Write even if the device has mounted volumes. **Use with care.** |
| `--json` | Print progress and the result as JSON, for scripts. |

On Windows, devices are named `\\.\PhysicalDriveN`; use `pvflasher list` to find `N`. Run the terminal as Administrator.

### `pvflasher install` and `pvflasher download`

Interactive helpers for official Pantavisor images: pick the channel, version and device from a menu. `install` downloads and flashes (it accepts `--no-verify`, `--no-eject` and `--force`); `download` only saves the image (`-o` to choose where; the default is the cache).

### `pvflasher create <image>`

Creates a `.bmap` block map for a raw image, so future flashes only write the blocks that hold data. `-o` sets the output file (default `<image>.bmap`), `-b` the block size (default 4096).

### `pvflasher verify <device> <bmap-file>`

Checks a device's contents against a block map, for example to test a card that was flashed earlier.

### `pvflasher version` and `pvflasher update`

`pvflasher version` shows your version and the latest release. `pvflasher update` shows both, lists what's new and asks before installing; `--yes` installs without asking and `--check` only looks. Copies installed with a package manager are not modified.

```console
$ pvflasher update
Current version: v1.1.3
Latest version:  v1.2.0 (update available)

What's new in v1.2.0:
  ### Feature
  * add a macOS DMG and a Windows installer

Update to v1.2.0 now? [y/N]
```

## 💾 Pantavisor Image Cache

Downloaded Pantavisor images are kept so that flashing the same release again doesn't download it twice. Delete the files to free space or force a fresh download:

*   **Linux / macOS**: `~/.pvflasher/images/`
*   **Windows**: `%USERPROFILE%\.pvflasher\images\`

Settings are stored in `~/.pvflasher/config.json` (`%USERPROFILE%\.pvflasher\config.json` on Windows).
