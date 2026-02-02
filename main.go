package main

import (
	"os"
	"runtime"

	"pvflasher/cli/commands"
	"pvflasher/gui"
)

func main() {
	if len(os.Args) == 1 {
		if !hasDisplay() {
			println("Error: No display detected. The GUI requires an X11 or Wayland session.")
			println("For CLI usage, use subcommands (e.g., 'pvflasher list' or 'pvflasher --help').")
			os.Exit(1)
		}
		runGUI()
	} else {
		commands.Execute()
	}
}

func hasDisplay() bool {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return true
	}
	// Support both X11 and Wayland (with wayland build tag, Fyne supports native Wayland)
	return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}

func runGUI() {
	app := gui.NewApp()
	app.Run()
}
