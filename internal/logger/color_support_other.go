//go:build !windows

package logger

import "os"

// colorSupportedPlatform checks whether stdout is a terminal on Unix-like
// systems. If it is, ANSI escapes are assumed to be supported.
func colorSupportedPlatform() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
