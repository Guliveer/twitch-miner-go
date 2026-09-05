//go:build !windows

package tray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Logon autostart for non-Windows platforms.
//
//   - Linux/BSD: a freedesktop XDG autostart .desktop file under
//     $XDG_CONFIG_HOME/autostart (default ~/.config/autostart). The desktop
//     environment launches it at logon without an attached terminal, so the
//     miner starts in the background with only the tray icon.
//   - macOS: a LaunchAgent plist loaded via launchctl, which starts the binary
//     hidden in the background at logon.
const (
	autostartName = "twitch-miner-go"
	autostartDir  = "autostart"
)

func autostartDirPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, autostartDir)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", autostartDir)
}

func autostartFilePath() string {
	return filepath.Join(autostartDirPath(), autostartName+".desktop")
}

func launchAgentPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", "com.guliveer.twitchminer.go.plist")
}

// AutostartEnabled reports whether a logon autostart entry exists.
func AutostartEnabled() bool {
	if runtime.GOOS == "darwin" {
		return fileExists(launchAgentPath())
	}
	return fileExists(autostartFilePath())
}

// SetAutostart enables logon autostart, picking the platform's mechanism.
func SetAutostart(exePath string) error {
	if runtime.GOOS == "darwin" {
		return setAutostartDarwin(exePath)
	}
	return setAutostartLinux(exePath)
}

// ClearAutostart removes the logon autostart entry.
func ClearAutostart() error {
	if runtime.GOOS == "darwin" {
		_ = exec.Command("launchctl", "unload", "-w", launchAgentPath()).Run() //nolint:errcheck // best-effort unload
		if err := os.Remove(launchAgentPath()); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("autostart: removing LaunchAgent: %w", err)
		}
		return nil
	}
	if err := os.Remove(autostartFilePath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("autostart: removing .desktop: %w", err)
	}
	return nil
}

func SyncAutostart(enable bool, exePath string) error {
	if enable {
		return SetAutostart(exePath)
	}
	return ClearAutostart()
}

// setAutostartLinux writes a freedesktop autostart .desktop entry that starts
// the binary at logon. Dark=true keeps it out of desktop launch bars.
func setAutostartLinux(exePath string) error {
	if err := os.MkdirAll(autostartDirPath(), 0o755); err != nil {
		return fmt.Errorf("autostart: creating dir: %w", err)
	}
	exe, err := filepath.Abs(exePath)
	if err != nil {
		exe = exePath
	}
	content := strings.Join([]string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=Twitch Miner Go",
		"Comment=Twitch Channel Points Miner",
		"Exec=" + quote(exe) + " -no-console",
		"Terminal=false",
		"X-GNOME-Autostart-enabled=true",
		"",
	}, "\n")
	if err := os.WriteFile(autostartFilePath(), []byte(content), 0o644); err != nil {
		return fmt.Errorf("autostart: writing .desktop: %w", err)
	}
	return nil
}

func setAutostartDarwin(exePath string) error {
	plistPath := launchAgentPath()
	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return fmt.Errorf("autostart: creating LaunchAgents dir: %w", err)
	}
	exe, err := filepath.Abs(exePath)
	if err != nil {
		exe = exePath
	}
	esc := escapePlist(exe)
	content := strings.Join([]string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">`,
		`<plist version="1.0">`,
		`<dict>`,
		`  <key>Label</key>`,
		`  <string>com.guliveer.twitchminer.go</string>`,
		`  <key>ProgramArguments</key>`,
		`  <array>`,
		`    <string>` + esc + `</string>`,
		`    <string>-no-console</string>`,
		`  </array>`,
		`  <key>RunAtLoad</key>`,
		`  <true/>`,
		`  <key>KeepAlive</key>`,
		`  <false/>`,
		`</dict>`,
		`</plist>`,
		``,
	}, "\n")
	if err := os.WriteFile(plistPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("autostart: writing LaunchAgent plist: %w", err)
	}
	if err := exec.Command("launchctl", "load", "-w", plistPath).Run(); err != nil {
		return fmt.Errorf("autostart: loading LaunchAgent: %w", err)
	}
	return nil
}

func quote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

func escapePlist(s string) string {
	return strings.ReplaceAll(s, "&", "&amp;")
}