package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	baseView := m.renderEditor()
	if m.showHelp {
		helpView := m.renderHelp()
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			helpView,
			lipgloss.WithWhitespaceChars(""),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
		)
	}
	return baseView
}

func (m Model) renderEditor() string {
	lines := m.textBuffer.GetLines()
	cursor := m.textBuffer.GetCursor()
	selection := m.textBuffer.GetSelection()

	// Use highlighted lines if available and properly sized
	if len(m.highlightedLines) == len(lines) {
		lines = m.highlightedLines
	}

	visibleLines := m.getVisibleLines()
	startLine := m.scrollOffset
	endLine := min(startLine+visibleLines, len(lines))

	content := m.renderVisibleLines(lines, startLine, endLine, cursor, selection, visibleLines)

	statusBar := m.renderStatusBar()

	// Ensure editor and status bar share exact width constraints
	return lipgloss.JoinVertical(lipgloss.Left,
		editorStyle.Render(content),
		statusBar,
	)
}

func (m Model) renderVisibleLines(lines []string, startLine, endLine int, cursor Position, selection *Selection, visibleLines int) string {
	var contentLines []string
	innerWidth := m.calculateInnerWidth()

	for i := 0; i < visibleLines; i++ {
		actualLineIndex := startLine + i
		lineContent := m.renderSingleVisibleLine(lines, actualLineIndex, cursor, selection, innerWidth)
		contentLines = append(contentLines, lineContent)
	}
	return strings.Join(contentLines, "\n")
}

// calculateInnerWidth calculates the available width for content
func (m Model) calculateInnerWidth() int {
	innerWidth := m.width - 4 // borders and padding
	if innerWidth < 1 {
		innerWidth = 1
	}
	return innerWidth
}

// renderSingleVisibleLine renders a single line with line number and styling
func (m Model) renderSingleVisibleLine(lines []string, lineIndex int, cursor Position, selection *Selection, innerWidth int) string {
	lineNum := m.formatLineNumber(lineIndex)
	renderedLine := m.getRenderedLine(lines, lineIndex, cursor, selection)
	renderedLine = m.applyHorizontalOffset(renderedLine)
	renderedLine = m.applyCursorLineStyling(renderedLine, lineIndex, cursor)

	lineContent := lineNum + " " + renderedLine
	return lipgloss.NewStyle().Width(innerWidth).Render(lineContent)
}

// formatLineNumber formats the line number with consistent styling
func (m Model) formatLineNumber(lineIndex int) string {
	return lineNumberStyle.Render(fmt.Sprintf("%4d", lineIndex+1))
}

// applyHorizontalOffset applies horizontal scrolling offset to the line
func (m Model) applyHorizontalOffset(renderedLine string) string {
	if m.horizontalOffset <= 0 {
		return renderedLine
	}

	plainLine := stripAnsiCodes(renderedLine)
	if len(plainLine) > m.horizontalOffset {
		return renderedLine[plainToAnsiIndex(renderedLine, m.horizontalOffset):]
	}
	return renderedLine
}

// applyCursorLineStyling applies cursor line styling to the text content
func (m Model) applyCursorLineStyling(renderedLine string, lineIndex int, cursor Position) string {
	if lineIndex == cursor.Line {
		return cursorLineStyle.Render(renderedLine)
	}
	return renderedLine
}

func (m Model) getRenderedLine(lines []string, lineIndex int, cursor Position, selection *Selection) string {
	if lineIndex >= len(lines) {
		return ""
	}

	line := lines[lineIndex]

	renderedLine := m.renderLineWithSelection(line, lineIndex, cursor, selection)

	// Cursor line styling is now handled in renderVisibleLines to avoid affecting line numbers

	return renderedLine
}

func (m Model) renderLineWithSelection(line string, lineIndex int, cursor Position, selection *Selection) string {
	originalPlainLine := stripAnsiCodes(line)
	plainLen := len(originalPlainLine)

	// Apply word highlighting first
	line = m.applyWordHighlightIfNeeded(line, lineIndex, cursor)

	// Apply selection highlighting second
	line = m.applySelection(line, lineIndex, plainLen, selection)

	// Apply cursor last, but preserve underlying styling when invisible
	line = m.applyCursorIfNeeded(line, lineIndex, cursor, originalPlainLine, plainLen)

	return line
}

// applyWordHighlightIfNeeded applies word highlighting if on cursor line with valid bounds
func (m Model) applyWordHighlightIfNeeded(line string, lineIndex int, cursor Position) string {
	if lineIndex == cursor.Line && m.hasValidWordBounds() {
		return m.applyWordHighlight(line, m.currentWordStart, m.currentWordEnd)
	}
	return line
}

// hasValidWordBounds checks if current word bounds are valid for highlighting
func (m Model) hasValidWordBounds() bool {
	return m.currentWordStart >= 0 && m.currentWordEnd > m.currentWordStart
}

// applyCursorIfNeeded applies cursor styling if on cursor line
func (m Model) applyCursorIfNeeded(line string, lineIndex int, cursor Position, originalPlainLine string, plainLen int) string {
	if lineIndex == cursor.Line {
		return m.applyCursor(line, cursor.Column, originalPlainLine, plainLen)
	}
	return line
}

func (m Model) applyWordHighlight(highlighted string, start, end int) string {
	// Check for invalid bounds (returned when no word should be highlighted)
	if start == -1 || end == -1 || start < 0 || end < 0 || start >= end {
		return highlighted
	}
	startByte := plainToAnsiIndex(highlighted, start)
	endByte := plainToAnsiIndex(highlighted, end)
	if startByte >= endByte {
		return highlighted
	}
	prefix := highlighted[:startByte]
	middle := highlighted[startByte:endByte]
	suffix := highlighted[endByte:]
	styledMiddle := wordHighlightStyle.Render(middle)
	return prefix + styledMiddle + suffix
}

func (m Model) applySelection(line string, lineIndex int, plainLen int, selection *Selection) string {
	selectionInfo := m.getSelectionInfo(lineIndex, selection, plainLen)
	if !selectionInfo.hasSelection {
		return line
	}

	// Handle empty selected lines
	if plainLen == 0 {
		return selectedTextStyle.Render(" ")
	}

	startCol, endCol := m.normalizeSelectionBounds(selectionInfo.startCol, selectionInfo.endCol, plainLen)
	if startCol >= endCol {
		return line
	}

	return m.renderSelectedLine(line, startCol, endCol)
}

// normalizeSelectionBounds ensures selection bounds are valid and properly ordered
func (m Model) normalizeSelectionBounds(startCol, endCol, plainLen int) (int, int) {
	if startCol > endCol {
		startCol, endCol = endCol, startCol
	}

	if startCol < 0 {
		startCol = 0
	}
	if endCol > plainLen {
		endCol = plainLen
	}

	return startCol, endCol
}

// renderSelectedLine renders a line with selection styling applied
func (m Model) renderSelectedLine(line string, startCol, endCol int) string {
	startIndex := plainToAnsiIndex(line, startCol)
	endIndex := plainToAnsiIndex(line, endCol)

	// Validate byte positions
	if !m.isValidByteRange(startIndex, endIndex, len(line)) {
		return line
	}

	before := line[:startIndex]
	selected := line[startIndex:endIndex]
	after := line[endIndex:]

	// Apply selection styling
	plainSelected := stripAnsiCodes(selected)
	styledSelected := selectedTextStyle.Render(plainSelected)

	return before + styledSelected + after
}

// isValidByteRange checks if byte indices are valid for the given line length
func (m Model) isValidByteRange(startIndex, endIndex, lineLen int) bool {
	return startIndex >= 0 && endIndex >= 0 &&
		startIndex < lineLen && endIndex <= lineLen &&
		startIndex < endIndex
}

func (m Model) applyCursor(line string, cursorCol int, plainLine string, plainLen int) string {
	// When cursor is invisible, return the line unchanged to preserve all styling
	if !m.cursorVisible {
		return line
	}

	cursorCol = clamp(cursorCol, 0, plainLen)
	cursorIndex := plainToAnsiIndex(line, cursorCol)
	var charLen int
	var cursorChar string

	if cursorCol < plainLen {
		charEndIndex := plainToAnsiIndex(line, cursorCol+1)
		charLen = charEndIndex - cursorIndex
		// Extract the existing styled character to preserve word highlighting
		cursorChar = line[cursorIndex:charEndIndex]
	} else {
		cursorIndex = len(line)
		charLen = 0
		cursorChar = " "
	}

	// Strip any existing ANSI codes from the character and apply cursor styling
	// This preserves the character but applies cursor background/foreground
	plainChar := stripAnsiCodes(cursorChar)
	styledCursor := cursorStyle.Render(plainChar)

	return line[:cursorIndex] + styledCursor + line[cursorIndex+charLen:]
}

func (m Model) getSelectionInfo(lineIndex int, selection *Selection, plainLen int) SelectionInfo {
	if selection == nil {
		return SelectionInfo{hasSelection: false}
	}

	start, end := m.normalizeSelection(selection)

	if lineIndex < start.Line || lineIndex > end.Line {
		return SelectionInfo{hasSelection: false}
	}

	startCol := 0
	if lineIndex == start.Line {
		startCol = clamp(start.Column, 0, plainLen)
	}

	endCol := plainLen
	if lineIndex == end.Line {
		endCol = clamp(end.Column, 0, plainLen)
	}

	return SelectionInfo{
		hasSelection: true,
		startCol:     startCol,
		endCol:       endCol,
	}
}

func (m Model) renderStatusBar() string {
	if m.minibufferType != MinibufferNone {
		return m.renderMinibuffer()
	}

	leftSection := m.getStatusBarFilename()
	centerSection := m.getStatusBarCenterInfo()
	rightSection := m.getStatusBarRightInfo()

	return m.formatStatusBar(leftSection, centerSection, rightSection)
}

func (m Model) getStatusBarFilename() string {
	filename := m.filename
	if filename == "" {
		filename = "<untitled>"
	}

	if m.modified {
		return modifiedStyle.Render(filename)
	}
	return filename
}

func (m Model) getStatusBarCenterInfo() string {
	cursor := m.textBuffer.GetCursor()

	if m.message != "" && time.Since(m.messageTime) < 3*time.Second {
		return m.message
	}

	return fmt.Sprintf("Line %d, Column %d", cursor.Line+1, cursor.Column+1)
}

func (m Model) getStatusBarRightInfo() string {
	lines := m.textBuffer.GetLines()
	return fmt.Sprintf("Total: %d lines", len(lines))
}

func (m Model) formatStatusBar(left, center, right string) string {
	contentWidth := m.width - 2 // total width - padding(2)

	// Handle extremely narrow terminals
	if contentWidth < 30 {
		return m.renderCompactStatusBar(left, center, right, contentWidth)
	}

	return m.renderFullStatusBar(left, center, right, contentWidth)
}

// renderCompactStatusBar renders a compact status bar for narrow terminals
func (m Model) renderCompactStatusBar(left, center, right string, contentWidth int) string {
	compact := fmt.Sprintf("%s %s %s", left, center, right)
	rendered := lipgloss.NewStyle().Width(contentWidth).Background(lipgloss.Color("#6f7cbf")).Render(compact)
	return statusBarStyle.Render(rendered)
}

// renderFullStatusBar renders a full status bar with proper section allocation
func (m Model) renderFullStatusBar(left, center, right string, contentWidth int) string {
	leftRequired, rightRequired := m.calculateSectionWidths(left, right, contentWidth)
	centerWidth := m.calculateCenterWidth(contentWidth, leftRequired, rightRequired)
	leftRequired, rightRequired = m.ensureMinimumWidths(leftRequired, rightRequired)

	statusContent := m.createStatusContent(left, center, right, leftRequired, centerWidth, rightRequired)
	statusContent = lipgloss.NewStyle().Width(contentWidth).Render(statusContent)

	return statusBarStyle.Render(statusContent)
}

// calculateSectionWidths calculates the required widths for left and right sections
func (m Model) calculateSectionWidths(left, right string, contentWidth int) (int, int) {
	leftRequired := lipgloss.Width(stripAnsiCodes(left)) + 1 // +1 padding
	rightRequired := lipgloss.Width(stripAnsiCodes(right)) + 1

	// Ensure we never allocate more than 40% for either side to keep center visible
	maxSide := int(float64(contentWidth) * 0.4)
	if leftRequired > maxSide {
		leftRequired = maxSide
	}
	if rightRequired > maxSide {
		rightRequired = maxSide
	}

	return leftRequired, rightRequired
}

// calculateCenterWidth calculates the center width and adjusts side widths if needed
func (m Model) calculateCenterWidth(contentWidth, leftRequired, rightRequired int) int {
	centerWidth := contentWidth - leftRequired - rightRequired
	if centerWidth < 5 {
		// Ensure center always has at least 5 columns
		centerWidth = 5
	}
	return centerWidth
}

// ensureMinimumWidths ensures side sections have minimum width of 1
func (m Model) ensureMinimumWidths(leftRequired, rightRequired int) (int, int) {
	if leftRequired < 1 {
		leftRequired = 1
	}
	if rightRequired < 1 {
		rightRequired = 1
	}
	return leftRequired, rightRequired
}

// createStatusContent creates the styled status bar content
func (m Model) createStatusContent(left, center, right string, leftWidth, centerWidth, rightWidth int) string {
	backgroundColor := lipgloss.Color("#6f7cbf")

	leftStyle := lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Left).Background(backgroundColor)
	centerStyle := lipgloss.NewStyle().Width(centerWidth).Align(lipgloss.Center).Background(backgroundColor)
	rightStyle := lipgloss.NewStyle().Width(rightWidth).Align(lipgloss.Right).Background(backgroundColor)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftStyle.Render(left),
		centerStyle.Render(center),
		rightStyle.Render(right),
	)
}

func (m Model) renderHelp() string {
	const keyColumnWidth = 18
	// Choose help box width based on current terminal size.
	// Use 60% of width when wide enough, but never exceed terminal minus 4 cols.
	maxWidth := m.width * 60 / 100
	if maxWidth > m.width-4 {
		maxWidth = m.width - 4
	}
	// Ensure a sane minimum so content is readable.
	if maxWidth < 40 {
		maxWidth = m.width - 4 // in very narrow terminals just use almost full width
	}

	contentWidth := maxWidth - 3
	descWidth := contentWidth - keyColumnWidth - 1
	var help strings.Builder

	title := helpTitleStyle.Render("Text Editor Help")
	help.WriteString(lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, title))
	help.WriteString("\n\n")

	helpItems := []struct {
		key  string
		desc string
	}{
		{"Ctrl+S", "Save file"},
		{"Ctrl+Q", "Quit"},
		{"Ctrl+C", "Copy selected text"},
		{"Ctrl+X", "Cut selected text"},
		{"Ctrl+V", "Paste text"},
		{"Ctrl+Z", "Undo"},
		{"Ctrl+Y", "Redo"},
		{"Ctrl+A", "Select all"},
		{"Ctrl+F", "Find text (shows results list)"},
		{"↑/↓", "Navigate through search results"},
		{"Ctrl+N", "Find next occurrence"},
		{"Ctrl+L", "Find previous occurrence"},
		{"Ctrl+G", "Go to line"},
		{"Shift+Arrow", "Select text"},
		{"Ctrl+Arrow", "Move by word"},
		{"Alt+Arrow", "Select by word"},
		{"Home/End", "Move to line start/end"},
		{"PgUp/PgDn", "Move by page"},
		{"Enter", "Confirm action"},
		{"Escape", "Cancel dialog input"},
		{"Ctrl+H", "Toggle this help"},
	}

	for _, item := range helpItems {
		key := helpKeyStyle.Render(item.key)
		key = lipgloss.NewStyle().Width(keyColumnWidth).Render(key)

		desc := lipgloss.NewStyle().
			Width(descWidth).
			MaxWidth(descWidth).
			Render(helpDescStyle.Render(item.desc))

		help.WriteString(key + " " + desc + "\n")
	}

	help.WriteString("\n")
	footer := helpStyle.Render("Press Ctrl+H again to close help")
	help.WriteString(lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, footer))

	return helpBoxStyle.Render(help.String())
}

func (m Model) padLineToWidth(line string) string {
	// Ensure the inner content strictly fits within the editor window so the right border is always visible.
	innerWidth := m.width - 4 // 1 char left border + 1 left padding + 1 right padding + 1 right border
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Lipgloss will pad or truncate as necessary based on the supplied width, which keeps ANSI width calculations accurate.
	return lipgloss.NewStyle().Width(innerWidth).Render(line)
}
func (m Model) getVisibleLines() int {
	minibufferHeight := m.getMinibufferHeight()
	visibleLines := m.height - 3 - minibufferHeight
	if visibleLines < 0 {
		return 0
	}
	return visibleLines
}
