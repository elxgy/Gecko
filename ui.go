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

	editorContent := editorStyle.
		Width(m.width - 2).
		Height(visibleLines + 2).
		Render(content)

	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Center, editorContent, statusBar)
}

func (m Model) renderVisibleLines(lines []string, startLine, endLine int, cursor Position, selection *Selection, visibleLines int) string {
	var content strings.Builder

	for i := 0; i < visibleLines; i++ {
		actualLineIndex := startLine + i

		lineNum := lineNumberStyle.Render(fmt.Sprintf("%4d", actualLineIndex+1))

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
		renderedLine = cursorLineStyle.Render(renderedLine)
	}

	return renderedLine
}

func (m Model) renderLineWithSelection(line string, lineIndex int, cursor Position, selection *Selection) string {
	plainLine := stripAnsiCodes(line)
	plainLen := len(plainLine)

	if selection == nil && lineIndex == cursor.Line {
		cursorCol := clamp(cursor.Column, 0, plainLen)
		cursorIndex := plainToAnsiIndex(line, cursorCol)
		return line[:cursorIndex] + "█" + line[cursorIndex:]
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

	return before + selectedTextStyle.Render(stripAnsiCodes(selected)) + after
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
	contentWidth := m.width

	leftStyle := lipgloss.NewStyle().Width(contentWidth / 3).Align(lipgloss.Left).Background(lipgloss.Color("#6f7cbf"))
	centerStyle := lipgloss.NewStyle().Width(contentWidth / 3).Align(lipgloss.Center).Background(lipgloss.Color("#6f7cbf"))
	rightStyle := lipgloss.NewStyle().Width(contentWidth / 3).Align(lipgloss.Right).Background(lipgloss.Color("#6f7cbf"))

	leftRendered := leftStyle.Render(left)
	centerRendered := centerStyle.Render(center)
	rightRendered := rightStyle.Render(right)

	statusContent := lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, centerRendered, rightRendered)

	return statusBarStyle.Render(statusContent)
}

func (m Model) renderHelp() string {
	const keyColumnWidth = 18
	maxWidth := m.width * 40 / 100
	if maxWidth < 50 {
		maxWidth = 50
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
	availableWidth := m.width - 4

	cleanLine := stripAnsiCodes(line)
	currentLength := len(cleanLine)

	if currentLength < availableWidth {
		padding := strings.Repeat(" ", availableWidth-currentLength)
		return line + padding
	}

	return line
}

func (m Model) getVisibleLines() int {
	minibufferHeight := m.getMinibufferHeight()
	visibleLines := m.height - 4 - minibufferHeight
	if visibleLines < 0 {
		return 0
	}
	return visibleLines
}
