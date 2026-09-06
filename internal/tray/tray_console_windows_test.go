//go:build windows

package tray

import "testing"

func TestGetConsoleWindowResolves(t *testing.T) {
	// Windows 11 builds (10.0.26200+) moved GetConsoleWindow from user32.dll to
	// kernel32.dll; resolving it only from user32.dll used to panic (regression
	// guard for the HideConsole crash).
	proc := findProc("GetConsoleWindow", "kernel32.dll", "user32.dll")
	if proc == nil {
		t.Fatal("GetConsoleWindow not found in kernel32.dll or user32.dll")
	}
}

func TestShowWindowResolves(t *testing.T) {
	proc := findProc("ShowWindow", "user32.dll")
	if proc == nil {
		t.Fatal("ShowWindow not found in user32.dll")
	}
}
