//go:build windows

package tray

import (
	"golang.org/x/sys/windows"
)

// HideConsole hides the console window associated with the current process.
// It is a no-op when the process has no attached console (e.g. a GUI
// subsystem build). On Windows the miner is a console application, so hiding
// the console lets it keep running in the background with only the tray icon.
func HideConsole() {
	// GetConsoleWindow is not surfaced in golang.org/x/sys/windows, so we
	// resolve the USER32 entry points directly.
	user32 := windows.NewLazySystemDLL("user32.dll")
	procGetConsoleWindow := user32.NewProc("GetConsoleWindow")
	procShowWindow := user32.NewProc("ShowWindow")

	console, _, _ := procGetConsoleWindow.Call()
	if console == 0 {
		// No console attached; nothing to hide.
		return
	}
	// SW_HIDE = 0.
	_, _, _ = procShowWindow.Call(console, 0)
}