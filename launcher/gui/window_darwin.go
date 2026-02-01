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
	focusCheckDone  chan struct{}
)

func init() {
	focusCheckDone = make(chan struct{})
	go focusChecker("EmuBuddy")
}

// isWindowFocused returns the cached focus status.
// The actual check runs in a background goroutine to avoid blocking the controller loop.
func isWindowFocused(windowTitle string) bool {
	focusCacheMutex.RLock()
	defer focusCacheMutex.RUnlock()
	return focusCache
}

// focusChecker runs in a background goroutine, checking focus every 500ms
func focusChecker(windowTitle string) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			focused := checkWindowFocus(windowTitle)
			focusCacheMutex.Lock()
			focusCache = focused
			focusCacheMutex.Unlock()
		case <-focusCheckDone:
			return
		}
	}
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
