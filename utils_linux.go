//go:build linux

package main

import (
	"os/exec"
	"strings"
)

// enableWindowsANSI is a no-op on Linux systems
func enableWindowsANSI() {
	// No-op on Linux - ANSI escape sequences are supported by default
}

// ensureUTF8Output is a no-op on Linux systems
func ensureUTF8Output() {
	// No-op on Linux - UTF-8 is typically the default encoding
}

// copyToClipboard copies text to the Linux clipboard using xclip
func copyToClipboard(text string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// pasteFromClipboard pastes text from the Linux clipboard using xclip
func pasteFromClipboard() (string, error) {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-o")
	output, err := cmd.Output()
	return string(output), err
}