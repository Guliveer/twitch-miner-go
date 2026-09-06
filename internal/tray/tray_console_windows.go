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
	// GetConsoleWindow is not surfaced in golang.org/x/sys/windows, so the
	// export is resolved directly. Windows 11 builds (10.0.26200+) moved it
	// from user32.dll to kernel32.dll; a missing export means no console —
	// never a panic.
	proc := findProc("GetConsoleWindow", "kernel32.dll", "user32.dll")
	if proc == nil {
		return
	}
	console, _, _ := proc.Call()
	if console == 0 {
		// No console attached; nothing to hide.
		return
	}
	// SW_HIDE = 0.
	if showWindow := findProc("ShowWindow", "user32.dll"); showWindow != nil {
		_, _, _ = showWindow.Call(console, 0)
	}
}

// findProc resolves an export from the first DLL that provides it. It returns
// nil when the export is unavailable on this Windows build, so callers never
// hit the LazyProc panic path.
func findProc(name string, dlls ...string) *windows.LazyProc {
	for _, dllName := range dlls {
		proc := windows.NewLazySystemDLL(dllName).NewProc(name)
		if proc.Find() == nil {
			return proc
		}
	}
	return nil
}
