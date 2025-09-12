//go:build windows

package main

import (
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

// enableWindowsANSI enables ANSI escape sequence processing on Windows terminals
func enableWindowsANSI() {
	// Enable ANSI escape sequences on Windows 10+
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")

	// Get stdout handle
	handle, _, _ := getStdHandle.Call(uintptr(^uint32(10) + 1)) // STD_OUTPUT_HANDLE = -11

	// Get current console mode
	var mode uint32
	getConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))

	// Enable ANSI escape sequences (ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004)
	mode |= 0x0004
	setConsoleMode.Call(handle, uintptr(mode))
}

// ensureUTF8Output ensures proper UTF-8 output on Windows terminals
func ensureUTF8Output() {
	// Set console output code page to UTF-8 (65001)
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	setConsoleOutputCP.Call(uintptr(65001))
}

// copyToClipboard copies text to the Windows clipboard
func copyToClipboard(text string) error {
	cmd := exec.Command("powershell", "-command", "Set-Clipboard -Value $args[0]", text)
	return cmd.Run()
}

// pasteFromClipboard pastes text from the Windows clipboard
func pasteFromClipboard() (string, error) {
	cmd := exec.Command("powershell", "-command", "Get-Clipboard")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Remove trailing CRLF that PowerShell adds
	result := strings.TrimSuffix(string(output), "\r\n")
	return result, nil
}
