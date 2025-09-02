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
			lipgloss.WithWhitespaceForeground(lipgloss.Color(ColorWhitespace)),
		)
	}
	return baseView
}

func (m Model) renderEditor() string {
	lines := m.textBuffer.GetLines()
	cursor := m.textBuffer.GetCursor()
	selection := m.textBuffer.GetSelection()

	if len(m.highlightedContent) == len(lines) {
		lines = m.highlightedContent
	}

	visibleLines := m.getVisibleLines()
	startLine := m.scrollOffset
	endLine := min(startLine+visibleLines, len(lines))

	content := m.renderVisibleLines(lines, startLine, endLine, cursor, selection, visibleLines)

	// Calculate responsive editor dimensions
	editorWidth := m.calculateEditorWidth()
	editorHeight := m.calculateEditorHeight(visibleLines)
	
	editorContent := EditorStyle.
		Width(editorWidth).
		Height(editorHeight).
		Render(content)

	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Center, editorContent, statusBar)
}

func (m Model) renderVisibleLines(lines []string, startLine, endLine int, cursor Position, selection *Selection, visibleLines int) string {
	var content strings.Builder

	for i := 0; i < visibleLines; i++ {
		actualLineIndex := startLine + i

		lineNum := LineNumberStyle.Render(fmt.Sprintf("%4d", actualLineIndex+1))

		renderedLine := m.getRenderedLine(lines, actualLineIndex, cursor, selection)

		lineContent := lineNum + " " + renderedLine
		paddedLine := m.padLineToWidth(lineContent)
		content.WriteString(paddedLine + "\n")
	}

	return content.String()
}

func (m Model) getRenderedLine(lines []string, lineIndex int, cursor Position, selection *Selection) string {
	if lineIndex >= len(lines) {
		return ""
	}

	line := lines[lineIndex]

	renderedLine := m.renderLineWithSelection(line, lineIndex, cursor, selection)

	if lineIndex == cursor.Line {
		renderedLine = CursorLineStyle.Render(renderedLine)
	}

	return renderedLine
}

func (m Model) renderLineWithSelection(line string, lineIndex int, cursor Position, selection *Selection) string {
	plainLine := stripAnsiCodes(line)
	plainLen := len(plainLine)

	if selection == nil && lineIndex == cursor.Line {
		// Highlight current word if no selection
		wordStart, wordEnd := m.textBuffer.GetCurrentWordBounds()
		if wordStart.Line == lineIndex && wordStart.Column != wordEnd.Column {
			// Render word highlighting
			startIndex := plainToAnsiIndex(line, wordStart.Column)
			endIndex := plainToAnsiIndex(line, wordEnd.Column)
			cursorCol := clamp(cursor.Column, 0, plainLen)
			cursorIndex := plainToAnsiIndex(line, cursorCol)
			
			before := line[:startIndex]
			highlightedWord := line[startIndex:endIndex]
			after := line[endIndex:]
			
			// Insert cursor within the highlighted word
			if cursorIndex >= startIndex && cursorIndex <= endIndex {
				relativeCursorPos := cursorIndex - startIndex
				highlightedWithCursor := highlightedWord[:relativeCursorPos] + "█" + highlightedWord[relativeCursorPos:]
				return before + CurrentWordStyle.Render(stripAnsiCodes(highlightedWithCursor)) + after
			} else {
				// Cursor outside word, render normally with cursor
				result := before + CurrentWordStyle.Render(stripAnsiCodes(highlightedWord)) + after
				if cursorIndex < startIndex {
					return result[:cursorIndex] + "█" + result[cursorIndex:]
				} else {
					return result[:cursorIndex] + "█" + result[cursorIndex:]
				}
			}
		} else {
			// No word to highlight, just show cursor
			cursorCol := clamp(cursor.Column, 0, plainLen)
			cursorIndex := plainToAnsiIndex(line, cursorCol)
			return line[:cursorIndex] + "█" + line[cursorIndex:]
		}
	}

	selectionInfo := m.getSelectionInfo(lineIndex, selection, plainLen)
	if !selectionInfo.hasSelection {
		return line
	}

	startCol := selectionInfo.startCol
	endCol := selectionInfo.endCol

	if startCol > endCol {
		startCol, endCol = endCol, startCol
	}

	startIndex := plainToAnsiIndex(line, startCol)
	endIndex := plainToAnsiIndex(line, endCol)

	before := line[:startIndex]
	selected := line[startIndex:endIndex]
	after := line[endIndex:]

	return before + SelectedTextStyle.Render(stripAnsiCodes(selected)) + after
}

func (m Model) getSelectionInfo(lineIndex int, selection *Selection, plainLen int) SelectionInfo {
	if selection == nil {
		return SelectionInfo{hasSelection: false}
	}

	// Use TextBuffer's normalizeSelection method
	start, end := m.textBuffer.normalizeSelection()

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
		return ModifiedStyle.Render(filename)
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
	// Handle very narrow terminals
	if m.width < MinTerminalWidth {
		return m.formatCompactStatusBar(left, center, right)
	}

	contentWidth := max(m.width, MinTerminalWidth)
	
	// Calculate section width with minimum constraints
	sectionWidth := max(contentWidth / StatusBarSections, 8) // Minimum 8 chars per section
	
	// Adjust content if sections are too wide
	leftTrimmed, centerTrimmed, rightTrimmed := m.adjustStatusBarSections(left, center, right, sectionWidth)
	
	leftStyle := lipgloss.NewStyle().Width(sectionWidth).Align(lipgloss.Left).Background(lipgloss.Color(ColorPrimary))
	centerStyle := lipgloss.NewStyle().Width(sectionWidth).Align(lipgloss.Center).Background(lipgloss.Color(ColorPrimary))
	rightStyle := lipgloss.NewStyle().Width(sectionWidth).Align(lipgloss.Right).Background(lipgloss.Color(ColorPrimary))

	leftRendered := leftStyle.Render(leftTrimmed)
	centerRendered := centerStyle.Render(centerTrimmed)
	rightRendered := rightStyle.Render(rightTrimmed)

	statusContent := lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, centerRendered, rightRendered)

	return StatusBarStyle.Render(statusContent)
}

func (m Model) renderHelp() string {
	// Handle very small terminals gracefully
	if m.width < MinTerminalWidth || m.height < MinTerminalHeight {
		return HelpBoxStyle.Render("Terminal too small for help. Press Ctrl+H to close.")
	}
	
	const keyColumnWidth = KeyColumnWidth
	maxWidth := m.width * HelpMaxWidthPercent / 100
	if maxWidth < HelpMinWidth {
		maxWidth = HelpMinWidth
	}
	
	// Ensure minimum content width
	maxWidth = max(maxWidth, MinContentWidth + 10)

	contentWidth := maxWidth - 3
	descWidth := max(contentWidth - keyColumnWidth - 1, 10) // Minimum description width
	var help strings.Builder

	title := HelpTitleStyle.Render("Text Editor Help")
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
		key := HelpKeyStyle.Render(item.key)
		key = lipgloss.NewStyle().Width(keyColumnWidth).Render(key)

		desc := lipgloss.NewStyle().
			Width(descWidth).
			MaxWidth(descWidth).
			Render(HelpDescStyle.Render(item.desc))

		help.WriteString(key + " " + desc + "\n")
	}

	help.WriteString("\n")
	footer := HelpStyle.Render("Press Ctrl+H again to close help")
	help.WriteString(lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, footer))

	return HelpBoxStyle.Render(help.String())
}

func (m Model) padLineToWidth(line string) string {
	// Calculate available width more precisely
	availableWidth := m.calculateContentWidth()

	cleanLine := stripAnsiCodes(line)
	currentLength := len(cleanLine)

	if currentLength < availableWidth {
		padding := strings.Repeat(" ", availableWidth-currentLength)
		return line + padding
	}

	// Handle line truncation for small terminals
	if currentLength > availableWidth {
		if availableWidth > 3 {
			// Safely truncate with ellipsis
			truncateAt := max(0, availableWidth-3)
			if truncateAt < len(line) {
				return line[:truncateAt] + "..."
			}
		}
		// Fallback for very narrow terminals
		if availableWidth > 0 && availableWidth < len(line) {
			return line[:availableWidth]
		}
	}

	return line
}

func (m Model) getVisibleLines() int {
	minibufferHeight := m.getMinibufferHeight()
	// Account for editor border (2), status bar (1), and padding (1)
	reservedHeight := 4 + minibufferHeight
	visibleLines := max(m.height - reservedHeight, 1)
	
	// Ensure we have at least 1 visible line even in very small terminals
	if visibleLines < 1 {
		return 1
	}
	return visibleLines
}

// calculateEditorWidth computes the responsive editor width
func (m Model) calculateEditorWidth() int {
	// Ensure minimum editor width, accounting for borders and padding
	minWidth := max(MinContentWidth, MinTerminalWidth-4)
	calculatedWidth := max(m.width-2, minWidth)
	return calculatedWidth
}

// calculateEditorHeight computes the responsive editor height
func (m Model) calculateEditorHeight(visibleLines int) int {
	// Add border space (2) to visible lines
	return max(visibleLines+2, 3)
}

// calculateContentWidth computes available width for content rendering
func (m Model) calculateContentWidth() int {
	// Account for editor borders (2) and line number space (estimated 4)
	reservedWidth := 6
	availableWidth := max(m.width-reservedWidth, MinContentWidth)
	return availableWidth
}



// adjustStatusBarSections truncates status bar sections to fit available width
func (m Model) adjustStatusBarSections(left, center, right string, sectionWidth int) (string, string, string) {
	// Truncate each section to fit within sectionWidth
	leftTrimmed := m.truncateToWidth(left, sectionWidth)
	centerTrimmed := m.truncateToWidth(center, sectionWidth)
	rightTrimmed := m.truncateToWidth(right, sectionWidth)
	
	return leftTrimmed, centerTrimmed, rightTrimmed
}

// truncateToWidth truncates text to fit within specified width
func (m Model) truncateToWidth(text string, width int) string {
	if len(text) <= width {
		return text
	}
	
	if width <= 3 {
		return strings.Repeat(".", max(width, 0))
	}
	
	return text[:width-3] + "..."
}

// formatCompactStatusBar creates a minimal status bar for very narrow terminals (3-section version)
func (m Model) formatCompactStatusBar(left, center, right string) string {
	// Show only essential info in compact format
	compactInfo := fmt.Sprintf(" %s ", center) // Show center info (cursor position)
	if len(compactInfo) > m.width {
		// Fallback for extremely narrow terminals
		return strings.Repeat(" ", max(m.width, 1))
	}
	
	padding := strings.Repeat(" ", max(0, m.width-len(compactInfo)))
	return compactInfo + padding
}
