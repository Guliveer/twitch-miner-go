//go:build !ios

package main

import (
	"fmt"
	"os"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/tray"
)

func startTray(rootLog *logger.Logger, noTray bool, httpPort string, editorPort int, configEditorEnabled bool, onExit func()) {
	if noTray || !tray.Available() {
		rootLog.Debug("System tray disabled", "no_tray", noTray, "session_available", tray.Available())
		return
	}

	exe, err := os.Executable()
	if err != nil {
		rootLog.Debug("Cannot resolve executable path, tray autostart disabled", "err", err)
	}

	// The config editor is not started in DB mode, so no link points at it —
	// an empty URL makes the tray hide the matching menu item.
	links := tray.Links{
		DashboardURL: fmt.Sprintf("http://localhost:%s", httpPort),
	}
	if configEditorEnabled {
		links.ConfigEditorURL = fmt.Sprintf("http://localhost:%d", editorPort)
	}

	tray.Run(tray.Options{
		Title:   "twitch-miner-go",
		ExePath: exe,
		OnLog:   rootLog.Info,
		Links:   links,
		OnExit: func() {
			rootLog.Info("Tray exit requested, shutting down")
			onExit()
		},
	})
	rootLog.Info("🖥️  System tray icon available — right-click for menu")
}

func applyNoConsole(noConsole bool) {
	if noConsole {
		tray.HideConsole()
	}
}
