package gui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"pvflasher/gui/pantavisor"
	"pvflasher/gui/util"
	"pvflasher/internal/device"
	"pvflasher/pkg/flash"
	"pvflasher/internal/platform"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// startFlash begins the flash operation
func (a *App) startFlash() error {
	a.mu.Lock()
	selectedDevice := a.selectedDevice
	a.mu.Unlock()

	// Check if device is mounted and needs confirmation
	mountPoints := a.getDeviceMountPoints(selectedDevice)
	if len(mountPoints) > 0 && !a.forceChecked {
		// Show confirmation dialog for mounted device
		mountList := strings.Join(mountPoints, ", ")
		dialog.ShowConfirm(
			"⚠️ Device is Mounted",
			fmt.Sprintf("The selected device has mounted volumes:\n%s\n\nThese volumes will be unmounted before flashing. All data on this device will be permanently erased.\n\nDo you want to continue?", mountList),
			func(confirmed bool) {
				if confirmed {
					// Set force flag and proceed
					a.mu.Lock()
					a.forceChecked = true
					a.mu.Unlock()
					a.proceedWithFlash()
				}
			},
			a.window,
		)
		return nil
	}

	a.proceedWithFlash()
	return nil
}

// getDeviceMountPoints returns the mount points for a device
func (a *App) getDeviceMountPoints(devicePath string) []string {
	mgr := device.NewManager()
	devices, err := mgr.List()
	if err != nil {
		return nil
	}

	// Normalize the device path for comparison
	normalizedPath := strings.ToUpper(strings.TrimPrefix(devicePath, `\\.\`))

	for _, d := range devices {
		normalizedName := strings.ToUpper(d.Name)
		if normalizedName == normalizedPath {
			return d.MountPoints
		}
	}
	return nil
}

// proceedWithFlash continues with the flash operation after any confirmations
func (a *App) proceedWithFlash() {
	a.mu.Lock()
	needsDownload := a.selectedImage == "" && a.selectedRel != nil
	rel := a.selectedRel
	a.mu.Unlock()

	a.showProgressScreen()

	if needsDownload {
		// Download the image first
		go a.downloadThenFlash(rel)
	} else {
		a.mu.Lock()
		if a.bmapPath == "" {
			a.bmapPath = a.CheckBmap(a.selectedImage)
		}
		a.mu.Unlock()

		if platform.IsRoot() {
			go a.runInProcessFlash()
		} else {
			go a.runElevatedFlash()
		}
	}
}

// downloadThenFlash downloads the Pantavisor image then starts flashing
func (a *App) downloadThenFlash(rel *pantavisor.DeviceRelease) {
	// Update UI to show download phase
	if a.progressScreen != nil {
		a.progressScreen.SetPhase("Downloading image...")
		a.progressScreen.SetProgress(0)
	}

	// Get cached image path
	cachedPath, err := pantavisor.GetCachedImagePath(rel.FullImage.URL)
	if err != nil {
		a.handleFlashError(fmt.Errorf("failed to get cache path: %w", err))
		return
	}

	// Check if image is already cached and valid
	if pantavisor.ValidateCachedFile(cachedPath, rel.FullImage.SHA256) {
		if a.progressScreen != nil {
			a.progressScreen.SetPhase("Using cached image...")
		}
	} else {
		// Download the image
		err = pantavisor.DownloadFileWithSHA(rel.FullImage.URL, cachedPath, rel.FullImage.SHA256, func(p pantavisor.DownloadProgress) {
			fyne.Do(func() {
				if a.progressScreen != nil {
					a.progressScreen.SetProgress(p.Percentage / 100.0)
					if p.Phase == "validating" {
						a.progressScreen.SetPhase("Validating checksum...")
					} else {
						speedStr := util.FormatSpeed(p.Speed)
						a.progressScreen.SetPhase(fmt.Sprintf("Downloading... %.1f%% (%s)", p.Percentage, speedStr))
					}
				}
			})
		})

		if err != nil {
			errMsg := err.Error()
			if _, ok := err.(*pantavisor.SHA256MismatchError); ok {
				errMsg = "Checksum validation failed - file may be corrupted"
			}
			a.handleFlashError(fmt.Errorf("download failed: %s", errMsg))
			return
		}
	}

	// Set the downloaded image path
	a.mu.Lock()
	a.selectedImage = cachedPath
	a.bmapPath = a.CheckBmap(cachedPath)
	a.mu.Unlock()

	// Reset progress for flash phase
	if a.progressScreen != nil {
		fyne.Do(func() {
			a.progressScreen.SetProgress(0)
			a.progressScreen.SetPhase("Starting flash...")
		})
	}

	// Now start the actual flash
	if platform.IsRoot() {
		a.runInProcessFlash()
	} else {
		a.runElevatedFlash()
	}
}

// runInProcessFlash executes flash in the current process
func (a *App) runInProcessFlash() {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	opts := flash.Options{
		ImagePath:  a.selectedImage,
		DevicePath: a.selectedDevice,
		BmapPath:   a.bmapPath,
		Force:      a.forceChecked,
		NoVerify:   !a.verifyChecked,
		NoEject:    !a.ejectChecked,
		ProgressCb: func(p flash.Progress) {
			a.updateProgressUI(p)
		},
	}

	f := flash.NewFlasher(opts)
	result, err := f.Flash(ctx)

	if err != nil {
		a.handleFlashError(err)
	} else {
		a.handleFlashSuccess(result)
	}
}

// buildFlashArgs constructs the command-line arguments for the flash subprocess
func (a *App) buildFlashArgs() []string {
	args := []string{"copy", a.selectedImage, a.selectedDevice, "--json"}
	if a.bmapPath != "" {
		args = append(args, "--bmap", a.bmapPath)
	}
	if a.forceChecked {
		args = append(args, "--force")
	}
	if !a.verifyChecked {
		args = append(args, "--no-verify")
	}
	if !a.ejectChecked {
		args = append(args, "--no-eject")
	}
	return args
}

// runElevatedFlash executes flash as elevated subprocess
func (a *App) runElevatedFlash() {
	args := a.buildFlashArgs()

	if runtime.GOOS == "darwin" {
		a.runElevatedFlashDarwin(args)
		return
	}

	elevator := platform.NewElevator()
	cmd, err := elevator.ElevateCommand(args...)
	if err != nil {
		a.handleFlashError(err)
		return
	}

	a.cmd = cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.handleFlashError(err)
		return
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Start()
	if err != nil {
		// pkexec not found or failed to start — fall back to sudo
		a.runSudoFlashWithPassword(args)
		return
	}

	a.lastLogs = []string{}
	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		a.handleElevatedOutputLine(scanner.Text())
	}

	err = cmd.Wait()
	if err == nil {
		a.handleFlashSuccess(a.findFlashResultInLogs())
		return
	}

	stderrStr := strings.TrimSpace(stderr.String())

	// Check if pkexec failed due to authentication issues (no agent, user cancelled, etc.)
	// Exit 126 = user dismissed auth dialog, Exit 127 = no auth agent found
	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	isPkexecAuthFailure := exitCode == 126 || exitCode == 127 ||
		strings.Contains(stderrStr, "authentication agent") ||
		strings.Contains(stderrStr, "Not authorized") ||
		strings.Contains(stderrStr, "dismissed")

	if isPkexecAuthFailure {
		// Fall back to sudo with password prompt
		a.runSudoFlashWithPassword(args)
		return
	}

	if stderrStr != "" {
		a.lastLogs = append(a.lastLogs, "STDERR: "+stderrStr)
		a.handleFlashError(fmt.Errorf("%s", stderrStr))
	} else {
		a.handleFlashError(fmt.Errorf("operation failed: %v", err))
	}
}

// runSudoFlashWithPassword prompts for password via Fyne dialog and runs sudo -S
func (a *App) runSudoFlashWithPassword(args []string) {
	passwordChan := make(chan string, 1)

	fyne.Do(func() {
		entry := widget.NewPasswordEntry()
		entry.SetPlaceHolder("Enter your password")
		dialog.ShowForm("Administrator Password Required", "OK", "Cancel",
			[]*widget.FormItem{
				widget.NewFormItem("Password", entry),
			},
			func(confirmed bool) {
				if confirmed {
					passwordChan <- entry.Text
				} else {
					passwordChan <- ""
				}
			},
			a.window,
		)
	})

	password := <-passwordChan
	if password == "" {
		a.handleFlashError(fmt.Errorf("authentication cancelled"))
		return
	}

	exe, err := os.Executable()
	if err != nil {
		a.handleFlashError(err)
		return
	}
	if appimagePath := os.Getenv("APPIMAGE"); appimagePath != "" {
		exe = appimagePath
	}

	fullArgs := append([]string{"-S", exe}, args...)
	cmd := exec.Command("sudo", fullArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.handleFlashError(err)
		return
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Pipe password to sudo's stdin
	cmd.Stdin = strings.NewReader(password + "\n")

	a.cmd = cmd

	err = cmd.Start()
	if err != nil {
		a.handleFlashError(err)
		return
	}

	a.lastLogs = []string{}
	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		a.handleElevatedOutputLine(scanner.Text())
	}

	err = cmd.Wait()
	if err == nil {
		a.handleFlashSuccess(a.findFlashResultInLogs())
		return
	}

	stderrStr := stderr.String()
	if stderrStr != "" {
		a.lastLogs = append(a.lastLogs, "STDERR: "+stderrStr)
		a.handleFlashError(fmt.Errorf("%s", stderrStr))
	} else {
		a.handleFlashError(fmt.Errorf("operation failed: %v", err))
	}
}

func (a *App) runElevatedFlashDarwin(args []string) {
	logFile, err := os.CreateTemp("", "pvflasher-darwin-elevated-*.log")
	if err != nil {
		a.handleFlashError(err)
		return
	}
	logPath := logFile.Name()
	logFile.Close()
	defer os.Remove(logPath)

	cmd, err := darwinElevatedCommand(args, logPath)
	if err != nil {
		a.handleFlashError(err)
		return
	}

	a.cmd = cmd
	if err := cmd.Start(); err != nil {
		a.handleFlashError(err)
		return
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	processedLines := 0
	for {
		select {
		case err := <-done:
			processedLines = a.processElevatedLogFile(logPath, processedLines)
			if err == nil {
				a.handleFlashSuccess(a.findFlashResultInLogs())
				return
			}

			if msg := a.lastNonJSONLogLine(); msg != "" {
				a.handleFlashError(fmt.Errorf("%s", msg))
			} else {
				a.handleFlashError(fmt.Errorf("operation failed: %v", err))
			}
			return
		case <-ticker.C:
			processedLines = a.processElevatedLogFile(logPath, processedLines)
		}
	}
}

func darwinElevatedCommand(args []string, logPath string) (*exec.Cmd, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	parts := []string{shellQuoteForShell(exe)}
	for _, arg := range args {
		parts = append(parts, shellQuoteForShell(arg))
	}

	commandLine := strings.Join(parts, " ") + " > " + shellQuoteForShell(logPath) + " 2>&1"
	script := fmt.Sprintf("do shell script \"%s\" with administrator privileges", escapeAppleScriptString(commandLine))
	return exec.Command("osascript", "-e", script), nil
}

func (a *App) processElevatedLogFile(path string, processedLines int) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return processedLines
	}

	content := string(data)
	if content == "" {
		return processedLines
	}

	lines := strings.Split(content, "\n")
	if !strings.HasSuffix(content, "\n") {
		lines = lines[:len(lines)-1]
	}

	if processedLines > len(lines) {
		processedLines = 0
	}

	for _, line := range lines[processedLines:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		a.handleElevatedOutputLine(line)
	}

	return len(lines)
}

func (a *App) handleElevatedOutputLine(line string) {
	a.lastLogs = append(a.lastLogs, line)
	if len(a.lastLogs) > 100 {
		a.lastLogs = a.lastLogs[1:]
	}

	var p struct {
		Phase      string  `json:"phase"`
		Processed  int64   `json:"processed"`
		Total      int64   `json:"total"`
		Percentage float64 `json:"percentage"`
		Speed      float64 `json:"speed"`
	}
	if err := json.Unmarshal([]byte(line), &p); err == nil {
		progress := flash.Progress{
			Phase:          p.Phase,
			BytesProcessed: p.Processed,
			BytesTotal:     p.Total,
			Percentage:     p.Percentage,
			Speed:          p.Speed,
		}
		a.updateProgressUI(progress)
	}
}

func (a *App) findFlashResultInLogs() *flash.FlashResult {
	for i := len(a.lastLogs) - 1; i >= 0; i-- {
		var r flash.FlashResult
		if err := json.Unmarshal([]byte(a.lastLogs[i]), &r); err == nil {
			if r.BytesWritten > 0 || r.Duration > 0 {
				return &r
			}
		}
	}
	return nil
}

func (a *App) lastNonJSONLogLine() string {
	for i := len(a.lastLogs) - 1; i >= 0; i-- {
		line := strings.TrimSpace(a.lastLogs[i])
		if line == "" {
			continue
		}

		var js json.RawMessage
		if json.Unmarshal([]byte(line), &js) == nil {
			continue
		}
		return line
	}
	return ""
}

func shellQuoteForShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func escapeAppleScriptString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}

// updateProgressUI sends progress updates to the channel for thread-safe UI updates
func (a *App) updateProgressUI(p flash.Progress) {
	a.mu.Lock()
	progressChan := a.progressChan
	a.mu.Unlock()

	// Send to channel; use non-blocking send to avoid blocking the flash operation
	if progressChan != nil {
		select {
		case progressChan <- p:
		default:
			// Channel full; skip this update to avoid blocking
		}
	}
}

// progressListener listens for progress updates and periodically updates UI widgets
func (a *App) progressListener() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case p, ok := <-a.progressChan:
			if !ok {
				// Channel closed, exit listener
				return
			}
			// Store progress data thread-safely
			a.progressMu.Lock()
			a.lastProgress = p
			a.progressChanged = true
			a.progressMu.Unlock()

		case <-ticker.C:
			// Check if progress changed and schedule widget update
			a.progressMu.Lock()
			changed := a.progressChanged
			if changed {
				a.progressChanged = false
				progress := a.lastProgress
				a.progressMu.Unlock()

				// Update the progress screen
				if a.progressScreen != nil {
					a.progressScreen.UpdateProgress(progress)
				}
			} else {
				a.progressMu.Unlock()
			}
		}
	}
}

// handleFlashSuccess handles successful flash completion
func (a *App) handleFlashSuccess(result *flash.FlashResult) {
	a.mu.Lock()
	a.lastResult = result
	// Close progress channel to signal listener to stop
	if a.progressChan != nil {
		close(a.progressChan)
		a.progressChan = nil
	}
	a.mu.Unlock()

	// Update success screen with stats
	if a.successScreen != nil {
		a.successScreen.UpdateStats(result)
	}

	a.showSuccessScreen()
}

// handleFlashError handles flash errors
func (a *App) handleFlashError(err error) {
	a.mu.Lock()
	a.lastError = err.Error()
	// Close progress channel to signal listener to stop
	if a.progressChan != nil {
		close(a.progressChan)
		a.progressChan = nil
	}
	a.mu.Unlock()

	// Update error screen with message
	if a.errorScreen != nil {
		a.errorScreen.SetError(a.lastError)
	}

	a.showErrorScreen()
}

// CancelFlash cancels an ongoing operation
func (a *App) CancelFlash() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}
	if a.cmd != nil && a.cmd.Process != nil {
		a.cmd.Process.Kill()
	}
}
