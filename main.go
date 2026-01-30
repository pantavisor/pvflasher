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
			println("If you are in a terminal, use subcommands (e.g., 'pvflasher list' or 'pvflasher --help').")
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
	return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}

func runGUI() {
	app := gui.NewApp()
	app.Run()
}
