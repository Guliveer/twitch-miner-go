//go:build windows

package tray

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// Registry paths for a per-user, logon-start autostart entry.
const (
	autostartRunKey = `Software\Microsoft\Windows\CurrentVersion\Run`
	autostartName   = "twitch-miner-go"
)

// autostartCommand builds the command line stored in the Run key. The miner is
// launched with -no-console so it starts hidden (only the tray icon remains),
// matching the "start quietly on logon" behaviour.
func autostartCommand(exePath string) string {
	// Quote defensively in case the executable path contains spaces.
	exe := exePath
	if q, err := filepath.Abs(exePath); err == nil {
		exe = q
	}
	return `"` + exe + `" -no-console`
}

// AutostartEnabled reports whether a logon autostart entry is present for this
// program in the current user's Run key. Missing/closed key or a false value
// is treated as disabled.
func AutostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, autostartRunKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close() //nolint:errcheck // registry key; cleanup on exit
	_, _, err = k.GetStringValue(autostartName)
	return err == nil
}

// SetAutostart enables logon autostart by writing the current user's Run key.
// The entry launches exePath hidden (no console, tray icon remains).
func SetAutostart(exePath string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, autostartRunKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("autostart: opening Run key: %w", err)
	}
	defer k.Close() //nolint:errcheck // registry key; cleanup on exit
	if err := k.SetStringValue(autostartName, autostartCommand(exePath)); err != nil {
		return fmt.Errorf("autostart: setting Run value: %w", err)
	}
	return nil
}

// ClearAutostart removes the logon autostart entry, if present.
func ClearAutostart() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, autostartRunKey, registry.SET_VALUE)
	if err != nil {
		// Key missing means nothing to clear.
		return nil
	}
	defer k.Close() //nolint:errcheck // registry key; cleanup on exit
	if err := k.DeleteValue(autostartName); err != nil {
		// Value missing is not an error worth surfacing.
		return nil
	}
	return nil
}

func SyncAutostart(enable bool, exePath string) error {
	if enable {
		return SetAutostart(exePath)
	}
	return ClearAutostart()
}