//go:build !ios

package main

import (
	"fmt"
	"os"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/tray"
)

func startTray(rootLog *logger.Logger, noTray bool, httpPort string, editorPort int, onExit func()) {
	if noTray || !tray.Available() {
		rootLog.Debug("System tray disabled", "no_tray", noTray, "session_available", tray.Available())
		return
	}

	exe, err := os.Executable()
	if err != nil {
		rootLog.Debug("Cannot resolve executable path, tray autostart disabled", "err", err)
	}

	tray.Run(tray.Options{
		Title:   "twitch-miner-go",
		ExePath: exe,
		OnLog:   rootLog.Info,
		Links: tray.Links{
			DashboardURL:    fmt.Sprintf("http://localhost:%s", httpPort),
			ConfigEditorURL: fmt.Sprintf("http://localhost:%d", editorPort),
		},
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