package main

import (
	"os"
	"runtime"
	"strings"
	"time"
)

// ============================================================================
// Mathematical Utilities
// ============================================================================

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

// ============================================================================
// ANSI Code Processing
// ============================================================================

func plainToAnsiIndex(ansiStr string, plainIndex int) int {
	if plainIndex <= 0 {
		return 0
	}

	plainPos := 0
	ansiPos := 0

	for ansiPos < len(ansiStr) {
		if isAnsiEscapeStart(ansiStr, ansiPos) {
			ansiPos = skipAnsiSequence(ansiStr, ansiPos)
		} else {
			if plainPos == plainIndex {
				return ansiPos
			}
			plainPos++
			ansiPos++
		}
	}

	return ansiPos
}

func isAnsiEscapeStart(s string, pos int) bool {
	return pos < len(s) && s[pos] == '\x1b'
}

func skipAnsiSequence(s string, pos int) int {
	pos++ // Skip the escape character
	for pos < len(s) && s[pos] != 'm' {
		pos++
	}
	if pos < len(s) {
		pos++ // Skip the 'm'
	}
	return pos
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

// ============================================================================
// Search and Find Operations
// ============================================================================

func (m *Model) adjustResultsOffset() {
	if m.findIndex < m.searchResultsOffset {
		m.searchResultsOffset = m.findIndex
	} else if m.findIndex >= m.searchResultsOffset+m.maxResultsDisplay {
		m.searchResultsOffset = m.findIndex - m.maxResultsDisplay + 1
	}
}

func (m *Model) jumpToCurrentResult() {
	if !m.hasValidFindResult() {
		return
	}

	m.textBuffer.SetCursor(m.findResults[m.findIndex])
	m.centerCursorOnScreen()
	m.postMovementUpdate()
	m.selectCurrentSearchResult()
}

func (m *Model) hasValidFindResult() bool {
	return len(m.findResults) > 0 && m.findIndex >= 0 && m.findIndex < len(m.findResults)
}

func (m *Model) selectCurrentSearchResult() {
	searchQuery := m.lastSearchQuery
	if searchQuery == "" {
		return
	}

	endPos := Position{
		Line:   m.findResults[m.findIndex].Line,
		Column: m.findResults[m.findIndex].Column + len(searchQuery),
	}
	m.textBuffer.SetSelection(&Selection{
		Start: m.findResults[m.findIndex],
		End:   endPos,
	})
}

// ============================================================================
// Cursor and Viewport Management
// ============================================================================

func (m *Model) ensureCursorVisible() {
	cursor := m.textBuffer.GetCursor()
	visibleLines := m.getVisibleLines()

	if visibleLines <= 0 {
		m.scrollOffset = 0
		return
	}

	m.adjustVerticalScroll(cursor, visibleLines)
	m.adjustHorizontalScroll(cursor)
}

func (m *Model) adjustVerticalScroll(cursor Position, visibleLines int) {
	if cursor.Line < m.scrollOffset {
		m.scrollOffset = cursor.Line
	} else if cursor.Line >= m.scrollOffset+visibleLines {
		m.scrollOffset = cursor.Line - visibleLines + 1
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m *Model) adjustHorizontalScroll(cursor Position) {
	visibleContentWidth := m.calculateVisibleContentWidth()
	m.horizontalOffset = max(0, cursor.Column-visibleContentWidth/2)

	if cursor.Column < m.horizontalOffset {
		m.horizontalOffset = cursor.Column
	}

	if m.horizontalOffset < 0 {
		m.horizontalOffset = 0
	}
}

func (m *Model) calculateVisibleContentWidth() int {
	visibleContentWidth := m.width - 9 // borders 2 + padding 2 + lineNum 4 + space 1
	if visibleContentWidth < 1 {
		return 1
	}
	return visibleContentWidth
}

func (m *Model) centerCursorOnScreen() {
	cursor := m.textBuffer.GetCursor()
	visibleLines := m.getVisibleLines()
	totalLines := len(m.textBuffer.GetLines())

	targetOffset := m.calculateCenterOffset(cursor.Line, visibleLines)
	maxOffset := m.calculateMaxOffset(totalLines, visibleLines)
	m.scrollOffset = clamp(targetOffset, 0, maxOffset)
}

func (m *Model) calculateCenterOffset(cursorLine, visibleLines int) int {
	return cursorLine - visibleLines/2
}

func (m *Model) calculateMaxOffset(totalLines, visibleLines int) int {
	maxOffset := totalLines - visibleLines
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
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

// ============================================================================
// Text Processing and Line Ending Utilities
// ============================================================================

// normalizeLineEndings converts different line ending formats to Unix format (LF)
func normalizeLineEndings(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	return content
}

// convertLineEndingsForOS converts line endings to the appropriate format for the current OS
func convertLineEndingsForOS(content string) string {
	switch runtime.GOOS {
	case "windows":
		return strings.ReplaceAll(content, "\n", "\r\n")
	default:
		return content
	}
}

// ============================================================================
// File Operations
// ============================================================================

func (m Model) saveFile() error {
	content := m.textBuffer.GetContent()
	content = convertLineEndingsForOS(content)
	return os.WriteFile(m.filename, []byte(content), 0644)
}

// ============================================================================
// Model State Management
// ============================================================================

func (m *Model) updateModified() {
	m.modified = m.textBuffer.GetContent() != m.originalText
	m.applySyntaxHighlighting()
}

func (m *Model) setMessage(msg string) {
	m.message = msg
	m.messageTime = time.Now()
}
