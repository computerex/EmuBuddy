//go:build darwin

package main

import (
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	focusCache      bool
	focusCacheMutex sync.RWMutex
	lastFocusCheck  time.Time
	focusCacheTTL   = 100 * time.Millisecond // Check focus 10 times per second instead of 60
)

// isWindowFocused checks if a window with the given title is focused.
// Uses AppleScript to get the frontmost application window title on macOS.
// Caches the result to avoid spawning osascript processes too frequently.
func isWindowFocused(windowTitle string) bool {
	// Check if we have a recent cached value
	focusCacheMutex.RLock()
	if time.Since(lastFocusCheck) < focusCacheTTL {
		cached := focusCache
		focusCacheMutex.RUnlock()
		return cached
	}
	focusCacheMutex.RUnlock()

	// Need to check focus
	focused := checkWindowFocus(windowTitle)

	// Update cache
	focusCacheMutex.Lock()
	focusCache = focused
	lastFocusCheck = time.Now()
	focusCacheMutex.Unlock()

	return focused
}

func checkWindowFocus(windowTitle string) bool {
	// AppleScript to get the frontmost window title
	script := `
		tell application "System Events"
			set frontApp to first application process whose frontmost is true
			set appName to name of frontApp
			try
				tell frontApp
					set windowTitle to name of front window
				end tell
				return appName & " - " & windowTitle
			on error
				return appName
			end try
		end tell
	`

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err == nil {
		title := strings.TrimSpace(string(output))
		return strings.Contains(title, windowTitle)
	}

	// If we can't detect, assume focused to not block input
	return true
}
