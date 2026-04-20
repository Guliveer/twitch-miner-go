package logger

import "os"

// ColorSupported returns true when the console is likely to render ANSI escape
// sequences correctly. It respects the NO_COLOR convention and checks whether
// stdout is a terminal. On Windows the platform-specific init function also
// tries to enable virtual-terminal processing; if that fails the function
// returns false.
func ColorSupported() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return colorSupportedPlatform()
}
