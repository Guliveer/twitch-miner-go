//go:build windows

package telemetry

import "golang.org/x/sys/windows"

// applyInstanceFileAttrs marks .instance_id as hidden and read-only on
// Windows. Unlike the .NET/Unix convention, a leading dot in the filename
// does not hide the file here; it must be flagged explicitly.
func applyInstanceFileAttrs(path string) error {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return windows.SetFileAttributes(pathPtr, windows.FILE_ATTRIBUTE_HIDDEN|windows.FILE_ATTRIBUTE_READONLY)
}