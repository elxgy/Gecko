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

	if len(m.highlightedContent) == len(lines) {
		lines = m.highlightedContent
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
	innerWidth := m.width - 4 // borders and padding
	if innerWidth < 1 {
		innerWidth = 1
	}
	for i := 0; i < visibleLines; i++ {
		actualLineIndex := startLine + i
		lineNum := lineNumberStyle.Render(fmt.Sprintf("%4d", actualLineIndex+1))
		renderedLine := m.getRenderedLine(lines, actualLineIndex, cursor, selection)
		// Apply horizontal offset
		plainLine := stripAnsiCodes(renderedLine)
		if m.horizontalOffset > 0 && len(plainLine) > m.horizontalOffset {
			renderedLine = renderedLine[plainToAnsiIndex(renderedLine, m.horizontalOffset):]
		}
		lineContent := lineNum + " " + renderedLine
		paddedLine := lipgloss.NewStyle().Width(innerWidth).Render(lineContent)
		contentLines = append(contentLines, paddedLine)
	}
	return strings.Join(contentLines, "\n")
}

func (m Model) getRenderedLine(lines []string, lineIndex int, cursor Position, selection *Selection) string {
	if lineIndex >= len(lines) {
		return ""
	}

	line := lines[lineIndex]

	renderedLine := m.renderLineWithSelection(line, lineIndex, cursor, selection)

	if lineIndex == cursor.Line {
		renderedLine = cursorLineStyle.Render(renderedLine)
	}

	return renderedLine
}

func (m Model) renderLineWithSelection(line string, lineIndex int, cursor Position, selection *Selection) string {
	plainLine := stripAnsiCodes(line)
	plainLen := len(plainLine)

	// Apply word highlight if on cursor line and valid bounds
	if lineIndex == cursor.Line && m.currentWordStart >= 0 && m.currentWordEnd > m.currentWordStart {
		startCol := clamp(m.currentWordStart, 0, plainLen)
		endCol := clamp(m.currentWordEnd, 0, plainLen)
		if startCol < endCol {
			startIndex := plainToAnsiIndex(line, startCol)
			endIndex := plainToAnsiIndex(line, endCol)
			before := line[:startIndex]
			word := line[startIndex:endIndex]
			after := line[endIndex:]
			line = before + wordHighlightStyle.Render(word) + after
		}
	}

	// Apply selection if present
	selectionInfo := m.getSelectionInfo(lineIndex, selection, plainLen)
	if selectionInfo.hasSelection {
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
		line = before + selectedTextStyle.Render(selected) + after
	}

	// Always render cursor on cursor line
	if lineIndex == cursor.Line {
		cursorCol := clamp(cursor.Column, 0, plainLen)
		cursorIndex := plainToAnsiIndex(line, cursorCol)
		var charLen int
		var cursorCharPlain string
		if cursorCol < plainLen {
			r := rune(plainLine[cursorCol])
			charEndIndex := plainToAnsiIndex(line, cursorCol+1)
			charLen = charEndIndex - cursorIndex
			cursorCharPlain = string(r)
		} else {
			cursorIndex = len(line)
			charLen = 0
			cursorCharPlain = " "
		}
		cursorStart := "\x1b[47;30m"
		cursorReset := "\x1b[49;39m"
		styledCursor := cursorStart + cursorCharPlain + cursorReset
		line = line[:cursorIndex] + styledCursor + line[cursorIndex+charLen:]
	}

	return line
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
	// The status bar itself has horizontal padding of 1 on each side, so
	// the inner content area is two columns narrower than the full terminal width.
	// Account for status bar's 1-cell padding on both sides
	contentWidth := m.width - 2 // total width - padding(2)

	// Reserve at least 3 columns for spacing between sections
	if contentWidth < 30 {
		// Extremely narrow, just concatenate with single spaces, then hard-set width for consistency.
		compact := fmt.Sprintf("%s %s %s", left, center, right)
		rendered := lipgloss.NewStyle().Width(contentWidth).Background(lipgloss.Color("#6f7cbf")).Render(compact)
		return statusBarStyle.Render(rendered)
	}

	// Determine widths: give left and right their needed size first, let center take the rest.
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

	// Calculate remaining width for center ensuring it never becomes negative.
	centerWidth := contentWidth - leftRequired - rightRequired
	if centerWidth < 5 {
		// Reclaim space from the sides so we always have at least 5 cols for center.
		deficit := 5 - centerWidth
		reclaim := (deficit + 1) / 2 // round up

		if leftRequired > reclaim {
			leftRequired -= reclaim
		}
		if rightRequired > reclaim {
			rightRequired -= reclaim
		}
		centerWidth = 5
	}

	// Safety clamp in case rounding pushed a side below 1.
	if leftRequired < 1 {
		leftRequired = 1
	}
	if rightRequired < 1 {
		rightRequired = 1
	}

	leftStyle := lipgloss.NewStyle().Width(leftRequired).Align(lipgloss.Left).Background(lipgloss.Color("#6f7cbf"))
	centerStyle := lipgloss.NewStyle().Width(centerWidth).Align(lipgloss.Center).Background(lipgloss.Color("#6f7cbf"))
	rightStyle := lipgloss.NewStyle().Width(rightRequired).Align(lipgloss.Right).Background(lipgloss.Color("#6f7cbf"))

	statusContent := lipgloss.JoinHorizontal(lipgloss.Top,
		leftStyle.Render(left),
		centerStyle.Render(center),
		rightStyle.Render(right),
	)

	// Ensure the inner bar always occupies the exact content width.
	statusContent = lipgloss.NewStyle().Width(contentWidth).Render(statusContent)

	return statusBarStyle.Render(statusContent)
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
