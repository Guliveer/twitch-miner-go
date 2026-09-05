//go:build !ios && !windows

package tray

// sessionAvailable reports whether a tray icon can be shown. On Linux, BSD
// and macOS the tray can always be attempted.
func sessionAvailable() bool {
	return true
}
