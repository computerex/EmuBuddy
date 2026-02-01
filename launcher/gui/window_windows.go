//go:build windows

package main

import (
	"strings"
	"syscall"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetCurrentProcessId = kernel32.NewProc("GetCurrentProcessId")
)

func isWindowFocused(windowTitle string) bool {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return false
	}

	// First check: Does the focused window belong to our process?
	var processId uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&processId)))
	
	currentPid, _, _ := procGetCurrentProcessId.Call()
	if processId == uint32(currentPid) {
		return true
	}

	// Fallback: Check window title (for compatibility)
	buf := make([]uint16, 256)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
	title := syscall.UTF16ToString(buf)

	return strings.Contains(title, windowTitle)
}
