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
)

func isWindowFocused(windowTitle string) bool {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return false
	}

	buf := make([]uint16, 256)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
	title := syscall.UTF16ToString(buf)

	return strings.Contains(title, windowTitle)
}
