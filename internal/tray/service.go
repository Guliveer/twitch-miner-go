//go:build !ios

package tray

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ServiceAction is a subcommand passed to the platform service installer.
type ServiceAction string

const (
	ServiceInstall   ServiceAction = "install"
	ServiceUninstall ServiceAction = "uninstall"
	ServiceStart     ServiceAction = "start"
	ServiceStop      ServiceAction = "stop"
	ServiceRestart   ServiceAction = "restart"
	ServiceStatus    ServiceAction = "status"
)

// serviceScriptName returns the installer script we delegate to, or "" when
// the script cannot be found next to the running binary.
func serviceScriptName() string {
	name := "install-service.sh"
	if runtime.GOOS == "windows" {
		name = "install-service.bat"
	}

	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exe)

	// Try progressively higher directories so a binary placed in a bin/
	// subfolder still finds tools/ at the repo root.
	for i := 0; i < 4; i++ {
		candidate := filepath.Join(dir, "tools", name)
		if fileExists(candidate) {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// ServiceScriptAvailable reports whether the platform installer script was
// found next to the running binary.
func ServiceScriptAvailable() bool {
	return serviceScriptName() != ""
}

// RunServiceAction delegates a service subcommand to the platform installer.
// On Windows the script is launched in its own console window so any
// interactive prompts (e.g. during install) are visible; elevation is handled
// by the script itself (UAC / sudo).
func RunServiceAction(action ServiceAction, logf func(string, ...any)) {
	script := serviceScriptName()
	if script == "" {
		logf("service installer script not found next to the binary; run tools/install-service manually")
		return
	}

	logf("service: running %s %s", script, action)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// `cmd /c start` detaches the script into a fresh console window so the
		// miner's own console is not blocked and interactive prompts are usable.
		cmd = exec.Command("cmd", "/c", "start", "", script, string(action))
	default:
		cmd = exec.Command(script, string(action))
	}
	detach(cmd)
	if err := cmd.Start(); err != nil {
		logf("service: failed to launch installer: %v", err)
	}
}