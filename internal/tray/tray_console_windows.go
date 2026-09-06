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

	// On Windows 11, GetConsoleWindow returns the PseudoConsoleWindow host
	// that Windows Terminal attaches to, not the CASCADIA window the user
	// sees. Hiding the pseudo window only minimizes the terminal, and
	// hiding its owner directly does not stick unless the terminal is the
	// foreground window first. Resolve the owner, focus it, then hide the
	// window that actually took the focus.
	owner := console
	if getWindow := findProc("GetWindow", "user32.dll"); getWindow != nil {
		// GW_OWNER = 4.
		if w, _, _ := getWindow.Call(console, 4); w != 0 {
			owner = w
		}
	}

	if setForeground := findProc("SetForegroundWindow", "user32.dll"); setForeground != nil {
		_, _, _ = setForeground.Call(owner)
	}

	target := owner
	if getForeground := findProc("GetForegroundWindow", "user32.dll"); getForeground != nil {
		if fg, _, _ := getForeground.Call(); fg != 0 && (fg == owner || fg == console) {
			target = fg
		}
	}

	// SW_HIDE = 0. Only windows related to this console are hidden; the
	// foreground guard above prevents hiding another app's window when
	// SetForegroundWindow was denied by the foreground lock.
	if showWindow := findProc("ShowWindow", "user32.dll"); showWindow != nil {
		_, _, _ = showWindow.Call(target, 0)
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
