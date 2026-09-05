//go:build !windows

package telemetry

import "os"

// applyInstanceFileAttrs reinforces read-only on .instance_id. The leading
// dot in the filename already hides it on Unix file systems; the explicit
// chmod ensures the read-only attribute survives regardless of how the file
// was first created.
func applyInstanceFileAttrs(path string) error {
	return os.Chmod(path, 0o444)
}