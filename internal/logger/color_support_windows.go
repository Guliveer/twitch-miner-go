package logger

import (
	"os"
	"syscall"
	"unsafe"
)

const enableVirtualTerminalProcessing = 0x0004

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode      = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode      = kernel32.NewProc("SetConsoleMode")
)

// colorSupportedPlatform tries to enable ANSI escape processing on the
// Windows console. Returns true only when the console handle is valid and
// the ENABLE_VIRTUAL_TERMINAL_PROCESSING flag was accepted by the kernel.
func colorSupportedPlatform() bool {
	handle := syscall.Handle(os.Stdout.Fd())

	var mode uint32
	r, _, _ := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if r == 0 {
		return false
	}

	if mode&enableVirtualTerminalProcessing != 0 {
		return true
	}

	r, _, _ = procSetConsoleMode.Call(uintptr(handle), uintptr(mode|enableVirtualTerminalProcessing))
	return r != 0
}
