//go:build windows

package tray

import (
	"golang.org/x/sys/windows"
)

// sessionAvailable reports whether the process is running in an interactive
// session that can show a tray icon. A Windows service (or any process in
// session 0) has no desktop access, so the tray icon cannot be displayed
// there; the process still runs normally without it.
func sessionAvailable() bool {
	var sid uint32
	if err := windows.ProcessIdToSessionId(windows.GetCurrentProcessId(), &sid); err != nil {
		// Unknown session: assume a tray can be shown.
		return true
	}
	// Session 0 hosts services and is not interactive.
	return sid != 0
}
