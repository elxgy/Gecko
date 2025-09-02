package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func plainToAnsiIndex(ansiStr string, plainIndex int) int {
	plainPos := 0
	ansiPos := 0
	inEscape := false

	for ansiPos < len(ansiStr) && plainPos < plainIndex {
		if !inEscape && ansiStr[ansiPos] == '\x1b' {
			inEscape = true
		}

		if inEscape {
			if ansiStr[ansiPos] == 'm' {
				inEscape = false
			}
		} else {
			plainPos++
		}

		ansiPos++
	}

	return ansiPos
}

func stripAnsiCodes(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func (m *Model) adjustResultsOffset() {
	if m.findIndex < m.searchResultsOffset {
		m.searchResultsOffset = m.findIndex
	} else if m.findIndex >= m.searchResultsOffset+m.maxResultsDisplay {
		m.searchResultsOffset = m.findIndex - m.maxResultsDisplay + 1
	}
}

func (m *Model) jumpToCurrentResult() {
	if len(m.findResults) > 0 && m.findIndex >= 0 && m.findIndex < len(m.findResults) {
		m.textBuffer.SetCursor(m.findResults[m.findIndex])

		m.centerCursorOnScreen()

		searchQuery := m.lastSearchQuery
		if searchQuery != "" {
			endPos := Position{
				Line:   m.findResults[m.findIndex].Line,
				Column: m.findResults[m.findIndex].Column + len(searchQuery),
			}
			m.textBuffer.SetSelection(&Selection{
				Start: m.findResults[m.findIndex],
				End:   endPos,
			})
		}
	}
}

func (m *Model) ensureCursorVisible() {
	cursor := m.textBuffer.GetCursor()
	visibleLines := m.getVisibleLines()
	totalLines := len(m.textBuffer.GetLines())

	// Handle edge case: no visible lines or empty buffer
	if visibleLines <= 0 || totalLines == 0 {
		m.scrollOffset = 0
		return
	}

	// Ensure cursor line is within valid bounds
	cursorLine := clamp(cursor.Line, 0, max(totalLines-1, 0))

	// Adjust scroll offset to keep cursor visible
	if cursorLine < m.scrollOffset {
		// Cursor is above visible area - scroll up
		m.scrollOffset = cursorLine
	} else if cursorLine >= m.scrollOffset+visibleLines {
		// Cursor is below visible area - scroll down
		m.scrollOffset = cursorLine - visibleLines + 1
	}

	// Clamp scroll offset to valid range
	maxScrollOffset := max(totalLines-visibleLines, 0)
	m.scrollOffset = clamp(m.scrollOffset, 0, maxScrollOffset)
}

func (m *Model) centerCursorOnScreen() {
	cursor := m.textBuffer.GetCursor()
	visibleLines := m.getVisibleLines()
	totalLines := len(m.textBuffer.GetLines())

	targetOffset := cursor.Line - visibleLines/2
	if targetOffset < 0 {
		targetOffset = 0
	}

	maxOffset := totalLines - visibleLines
	if maxOffset < 0 {
		maxOffset = 0
	}
	if targetOffset > maxOffset {
		targetOffset = maxOffset
	}

	m.scrollOffset = targetOffset
}



func (m Model) saveFile() error {
	if m.filename == "" {
		return fmt.Errorf("no filename specified")
	}
	
	content := m.textBuffer.GetContent()
	
	// Create backup of original file if it exists
	if _, err := os.Stat(m.filename); err == nil {
		backupName := m.filename + ".bak"
		if backupErr := os.Rename(m.filename, backupName); backupErr != nil {
			// Log backup failure but continue with save
		}
	}
	
	// Write to temporary file first for atomic save
	tempFile := m.filename + ".tmp"
	if err := os.WriteFile(tempFile, []byte(content), FilePermissions); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tempFile, m.filename); err != nil {
		// Clean up temp file on failure
		os.Remove(tempFile)
		return fmt.Errorf("failed to save file: %w", err)
	}
	
	return nil
}

func copyToClipboard(text string) error {
	switch runtime.GOOS {
	case "linux":
		// Try xclip first, then xsel as fallback
		if err := tryClipboardCommand("xclip", []string{"-selection", "clipboard"}, text); err == nil {
			return nil
		}
		if err := tryClipboardCommand("xsel", []string{"--clipboard", "--input"}, text); err == nil {
			return nil
		}
		return fmt.Errorf("clipboard not available: install xclip or xsel")
	case "darwin":
		return tryClipboardCommand("pbcopy", []string{}, text)
	case "windows":
		return tryClipboardCommand("clip", []string{}, text)
	default:
		return fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}
}

func pasteFromClipboard() (string, error) {
	switch runtime.GOOS {
	case "linux":
		// Try xclip first, then xsel as fallback
		if output, err := tryClipboardOutput("xclip", []string{"-selection", "clipboard", "-o"}); err == nil {
			return output, nil
		}
		if output, err := tryClipboardOutput("xsel", []string{"--clipboard", "--output"}); err == nil {
			return output, nil
		}
		return "", fmt.Errorf("clipboard not available: install xclip or xsel")
	case "darwin":
		return tryClipboardOutput("pbpaste", []string{})
	case "windows":
		// Use PowerShell for reliable clipboard access on Windows
		return tryClipboardOutput("powershell", []string{"-command", "Get-Clipboard"})
	default:
		return "", fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}
}

func tryClipboardCommand(command string, args []string, input string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = strings.NewReader(input)
	return cmd.Run()
}

func tryClipboardOutput(command string, args []string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(output), "\r\n"), nil
}

func (m *Model) updateModified() {
	m.modified = m.textBuffer.GetContent() != m.originalText
	m.applySyntaxHighlighting()
}

func (m *Model) setMessage(msg string) {
	m.message = msg
	m.messageTime = time.Now()
}
