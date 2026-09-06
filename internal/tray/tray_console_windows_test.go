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

func TestWindowResolutionProcsResolve(t *testing.T) {
	// HideConsole resolves the real terminal window (Windows Terminal vs
	// legacy conhost) through these exports; each must exist on a desktop
	// Windows build for the hide to work.
	for _, name := range []string{"GetWindow", "SetForegroundWindow", "GetForegroundWindow"} {
		if proc := findProc(name, "user32.dll"); proc == nil {
			t.Fatalf("%s not found in user32.dll", name)
		}
	}
}

func TestConsoleToggleStateStartsVisible(t *testing.T) {
	// The tray toggle must start in the "visible" state so the menu item
	// shows "Hide Terminal" on a fresh start. This test never calls
	// HideConsole(): the test binary may share the developer's terminal, and
	// hiding it would be destructive.
	if isConsoleHidden() {
		t.Fatal("console toggle starts hidden; menu would show 'Show Terminal'")
	}
}

func TestShowConsoleWithoutHideIsNoop(t *testing.T) {
	// showConsole with nothing hidden must not touch any window. It also
	// must not panic or flip the state, so a stray menu click before any
	// HideConsole is harmless.
	showConsole()
	if isConsoleHidden() {
		t.Fatal("showConsole flipped the state to hidden")
	}
}
