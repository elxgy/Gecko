package main

import (
	"os"
	"os/exec"
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
	if plainIndex <= 0 {
		return 0
	}

	plainPos := 0
	ansiPos := 0

	for ansiPos < len(ansiStr) {
		if ansiPos < len(ansiStr) && ansiStr[ansiPos] == '\x1b' {
			// Skip ANSI escape sequence
			ansiPos++ // Skip the escape character
			for ansiPos < len(ansiStr) && ansiStr[ansiPos] != 'm' {
				ansiPos++
			}
			if ansiPos < len(ansiStr) {
				ansiPos++ // Skip the 'm'
			}
		} else {
			if plainPos == plainIndex {
				return ansiPos
			}
			plainPos++
			ansiPos++
		}
	}

	// If we've reached the end, return the final position
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

	if visibleLines <= 0 {
		m.scrollOffset = 0
		return
	}

	if cursor.Line < m.scrollOffset {
		m.scrollOffset = cursor.Line
	} else if cursor.Line >= m.scrollOffset+visibleLines {
		m.scrollOffset = cursor.Line - visibleLines + 1
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	// Horizontal scrolling
	visibleContentWidth := m.width - 9 // borders 2 + padding 2 + lineNum 4 + space 1
	if visibleContentWidth < 1 {
		visibleContentWidth = 1
	}

	m.horizontalOffset = max(0, cursor.Column - visibleContentWidth/2)

	if cursor.Column < m.horizontalOffset {
		m.horizontalOffset = cursor.Column
	}

	if m.horizontalOffset < 0 {
		m.horizontalOffset = 0
	}
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

func (m Model) normalizeSelection(selection *Selection) (Position, Position) {
	if selection == nil {
		cursor := m.textBuffer.GetCursor()
		return cursor, cursor
	}

	start, end := selection.Start, selection.End

	if start.Line > end.Line || (start.Line == end.Line && start.Column > end.Column) {
		start, end = end, start
	}

	return start, end
}

func (m Model) saveFile() error {
	content := m.textBuffer.GetContent()
	return os.WriteFile(m.filename, []byte(content), 0644)
}

func copyToClipboard(text string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func pasteFromClipboard() (string, error) {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-o")
	output, err := cmd.Output()
	return string(output), err
}

func (m *Model) updateModified() {
	m.modified = m.textBuffer.GetContent() != m.originalText
	m.applySyntaxHighlighting()
}

func (m *Model) setMessage(msg string) {
	m.message = msg
	m.messageTime = time.Now()
}