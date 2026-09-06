//go:build !windows

package tray

// HideConsole is a no-op on platforms other than Windows, where the process
// is not attached to a Win32 console. The tray icon and desktop shell are the
// only UI on those platforms, so there is nothing to hide.
func HideConsole() {}

func showConsole() {}

func isConsoleHidden() bool { return false }
