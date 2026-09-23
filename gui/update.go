package gui

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pvflasher/gui/util"
	"pvflasher/internal/update"
	"pvflasher/internal/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// updateCheckInterval limits automatic checks to one a day.
const updateCheckInterval = 24 * time.Hour

const releasesPage = "https://github.com/pantavisor/pvflasher/releases/latest"

// maybeCheckForUpdates runs the automatic background check at startup.
func (a *App) maybeCheckForUpdates() {
	update.CleanupPrevious()
	if !update.IsReleaseVersion(version.Version) {
		return
	}
	config, _ := util.LoadConfig()
	if config.DisableUpdateCheck || time.Since(config.LastUpdateCheck) < updateCheckInterval {
		return
	}
	go a.checkForUpdates(false)
}

// checkForUpdates queries the manifest. A manual check reports every
// outcome; the automatic one only surfaces an update the user hasn't skipped.
func (a *App) checkForUpdates(manual bool) {
	st, err := a.fetchUpdateStatus()
	fyne.Do(func() {
		switch {
		case err != nil:
			if manual {
				dialog.ShowError(err, a.window)
			}
		case st.Update == nil:
			if manual {
				dialog.ShowInformation("No Updates", "PvFlasher "+versionText()+" is the latest version.", a.window)
			}
		default:
			config, _ := util.LoadConfig()
			if manual || config.SkippedVersion != st.Update.Version {
				a.showUpdateDialog()
			}
		}
	})
}

// fetchUpdateStatus checks the latest release and records the result: an
// available update shows the footer link.
func (a *App) fetchUpdateStatus() (*update.Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	st, err := update.Check(ctx, version.Version)
	if err != nil {
		return nil, err
	}
	config, _ := util.LoadConfig()
	config.LastUpdateCheck = time.Now()
	_ = util.SaveConfig(config)

	fyne.Do(func() {
		a.pendingUpdate = st.Update
		if st.Update != nil {
			a.updateLink.SetText("Update available: v" + st.Update.Version)
			a.updateLink.Show()
		} else {
			a.updateLink.Hide()
		}
	})
	return st, nil
}

// updatesPanel is the Settings section showing the running and latest
// versions, with the update action when one is available.
func (a *App) updatesPanel(closeSettings func()) fyne.CanvasObject {
	current := widget.NewLabelWithStyle(versionText(), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	latest := widget.NewLabel("Checking…")
	state := widget.NewLabel("")
	state.Importance = widget.LowImportance
	state.SizeName = theme.SizeNameCaptionText
	state.Wrapping = fyne.TextWrapWord
	spinner := widget.NewActivity()
	spinner.Start()

	action := widget.NewButton("Check Again", nil)
	action.Disable()

	var refresh func()
	show := func(st *update.Status, err error) {
		spinner.Stop()
		spinner.Hide()
		action.Enable()
		action.Importance = widget.MediumImportance
		action.SetText("Check Again")
		action.OnTapped = refresh
		switch {
		case err != nil:
			latest.SetText("Unknown")
			state.SetText("Couldn't reach the update server. Check your connection.")
		case st.DevBuild:
			latest.SetText("v" + st.Latest)
			state.SetText("This is a development build, so it doesn't update.")
		case st.Update == nil:
			latest.SetText("v" + st.Latest)
			state.SetText("PvFlasher is up to date.")
		default:
			latest.SetText("v" + st.Latest)
			if st.Update.Install.CanSelfUpdate {
				state.SetText("An update is available.")
			} else {
				state.SetText("An update is available. This copy is managed by " + st.Update.Install.Reason + ".")
			}
			action.Importance = widget.HighImportance
			action.SetText("Update to v" + st.Update.Version + "…")
			action.OnTapped = func() {
				closeSettings()
				a.showUpdateDialog()
			}
		}
		action.Refresh()
	}
	refresh = func() {
		latest.SetText("Checking…")
		state.SetText("")
		spinner.Show()
		spinner.Start()
		action.Disable()
		go func() {
			st, err := a.fetchUpdateStatus()
			fyne.Do(func() { show(st, err) })
		}()
	}
	refresh()

	grid := widget.NewForm(
		widget.NewFormItem("Current version", current),
		widget.NewFormItem("Latest version", container.NewHBox(latest, spinner)),
	)
	return container.NewVBox(grid, state, container.NewHBox(action))
}

// showUpdateDialog presents the release notes and the install choice.
func (a *App) showUpdateDialog() {
	rel := a.pendingUpdate
	if rel == nil {
		return
	}

	intro := widget.NewLabel(fmt.Sprintf("PvFlasher %s is available. You have %s.", "v"+rel.Version, versionText()))
	intro.Wrapping = fyne.TextWrapWord

	notes := widget.NewRichTextFromMarkdown(rel.Notes)
	if rel.Notes == "" {
		notes = widget.NewRichTextFromMarkdown("See the release page for what's new.")
	}
	notes.Wrapping = fyne.TextWrapWord
	notesScroll := container.NewVScroll(notes)
	notesScroll.SetMinSize(fyne.NewSize(460, 180))

	content := container.NewVBox(intro, widget.NewSeparator(), notesScroll)
	d := dialog.NewCustomWithoutButtons("Update PvFlasher", content, a.window)

	later := widget.NewButton("Later", d.Hide)
	skip := widget.NewButton("Skip This Version", func() {
		config, _ := util.LoadConfig()
		config.SkippedVersion = rel.Version
		_ = util.SaveConfig(config)
		d.Hide()
	})

	var primary *widget.Button
	if rel.Install.CanSelfUpdate {
		primary = widget.NewButtonWithIcon("Install and Restart", theme.DownloadIcon(), func() {
			d.Hide()
			a.installUpdate(rel)
		})
		if a.isFlashing() {
			primary.Disable()
			intro.SetText(intro.Text + " Finish the current flash to install it.")
		}
	} else {
		// Package-managed or read-only installs are updated by other means.
		note := widget.NewLabel(fmt.Sprintf("This copy can't update itself (%s). Update it with your package manager or download the new version.", rel.Install.Reason))
		note.Wrapping = fyne.TextWrapWord
		note.Importance = widget.LowImportance
		content.Add(note)
		primary = widget.NewButtonWithIcon("Open Download Page", theme.ComputerIcon(), func() {
			_ = a.fyneApp.OpenURL(mustParseURL(releasesPage))
			d.Hide()
		})
	}
	primary.Importance = widget.HighImportance

	d.SetButtons([]fyne.CanvasObject{skip, later, primary})
	d.Resize(fyne.NewSize(520, 0))
	d.Show()
}

// installUpdate downloads, verifies and installs the update, then restarts.
func (a *App) installUpdate(rel *update.Release) {
	bar := widget.NewProgressBar()
	status := widget.NewLabel("Downloading…")
	ctx, cancel := context.WithCancel(context.Background())
	cancelBtn := widget.NewButton("Cancel", cancel)

	d := dialog.NewCustomWithoutButtons("Updating PvFlasher",
		container.NewVBox(status, util.MinWidth(400, bar)), a.window)
	d.SetButtons([]fyne.CanvasObject{cancelBtn})
	d.Show()

	go func() {
		defer cancel()
		err := update.Apply(ctx, rel, func(done, total int64) {
			fyne.Do(func() {
				if total > 0 {
					bar.SetValue(float64(done) / float64(total))
				}
				if total > 0 && done >= total {
					// Verifying and unpacking follow the download.
					status.SetText("Installing…")
					cancelBtn.Disable()
				} else {
					status.SetText(fmt.Sprintf("Downloading… %s of %s", util.FormatBytes(done), util.FormatBytes(total)))
				}
			})
		})
		fyne.Do(func() {
			d.Hide()
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				dialog.ShowError(fmt.Errorf("the update could not be installed: %w", err), a.window)
				return
			}
			if err := update.Relaunch(rel.Install); err != nil {
				dialog.ShowInformation("Update Installed",
					"PvFlasher v"+rel.Version+" is installed. Restart PvFlasher to use it.", a.window)
				return
			}
			a.fyneApp.Quit()
		})
	}()
}
