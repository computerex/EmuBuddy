//go:build !windows

package main

// isWindowFocused checks if a window with the given title is focused.
// On non-Windows platforms, we always return true since we can't easily
// check window focus without platform-specific code.
func isWindowFocused(windowTitle string) bool {
	return true
}
