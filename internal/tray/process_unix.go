//go:build !windows

package tray

import (
	"os/exec"
	"syscall"
)

func detach(cmd *exec.Cmd) {
	// Run the script in its own process group so it is not torn down with the
	// miner, and keep it from inheriting the miner's controlling terminal.
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
}