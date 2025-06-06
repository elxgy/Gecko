package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#6f7cbf")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	modifiedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#6f7cbf")).
			Foreground(lipgloss.Color("196")).
			Bold(true)

	editorStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#6f7cbf")).
			Padding(0, 1)

	lineNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(4).
			Align(lipgloss.Right)

	selectedTextStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#a600a0")).
				Foreground(lipgloss.Color("#f8f8f2"))

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#282a36"))

	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6f7cbf")).
			AlignHorizontal(lipgloss.Center).
			Padding(1, 2).
			Margin(1, 0)

	helpTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6f7cbf")).
			Bold(true).
			Underline(true)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4"))
)

type Model struct {
	textBuffer          *TextBuffer
	filename            string
	modified            bool
	originalText        string
	width               int
	height              int
	showHelp            bool
	lastSaved           time.Time
	message             string
	messageTime         time.Time
	clipboard           string
	scrollOffset        int
	minibufferType      MinibufferType
	minibufferInput     string
	minibufferCursorPos int
	findResults         []Position
	findIndex           int
	lastSearchQuery     string
	searchResultsOffset int 
	maxResultsDisplay   int 
}

func NewModel(filename string) Model {
	var content string
	var originalText string

	if filename != "" {
		if data, err := os.ReadFile(filename); err == nil {
			content = string(data)
			originalText = content
		}
	}

	textBuffer := NewTextBuffer(content)

	return Model{
		textBuffer:        textBuffer,
		filename:          filename,
		originalText:      originalText,
		modified:          false,
		findResults:       []Position{},
		findIndex:         -1,
		maxResultsDisplay: 8,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{}}
	})
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
		// Set cursor position first
		m.textBuffer.SetCursor(m.findResults[m.findIndex])
		
		// Center the result on screen
		m.centerCursorOnScreen()
		
		// Set up text selection
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

func (m Model) handleSave() (tea.Model, tea.Cmd) {
	if m.filename != "" {
		err := m.saveFile()
		if err == nil {
			m.modified = false
			m.originalText = m.textBuffer.GetContent()
			m.lastSaved = time.Now()
			m.setMessage("File saved successfully")
		} else {
			m.setMessage(fmt.Sprintf("Error saving file: %v", err))
		}
	} else {
		m.setMessage("No filename specified")
	}
	return m, nil
}

func (m Model) handleCopy() (tea.Model, tea.Cmd) {
	if m.textBuffer.HasSelection() {
		text := m.textBuffer.GetSelectedText()
		m.clipboard = text
		if err := copyToClipboard(text); err != nil {
			m.setMessage("Copied to internal clipboard")
		} else {
			m.setMessage("Copied to system clipboard")
		}
	} else {
		m.setMessage("No text selected")
	}
	return m, nil
}

func (m Model) handleCut() (tea.Model, tea.Cmd) {
	if m.textBuffer.HasSelection() {
		text := m.textBuffer.GetSelectedText()
		m.clipboard = text
		m.textBuffer.DeleteSelection()
		m.updateModified()
		if err := copyToClipboard(text); err != nil {
			m.setMessage("Cut to internal clipboard")
		} else {
			m.setMessage("Cut to system clipboard")
		}
	} else {
		m.setMessage("No text selected")
	}
	return m, nil
}

func (m Model) handlePaste() (tea.Model, tea.Cmd) {
	if text, err := pasteFromClipboard(); err == nil && text != "" {
		m.textBuffer.InsertText(text)
		m.updateModified()
		m.setMessage("Pasted from system clipboard")
	} else if m.clipboard != "" {
		m.textBuffer.InsertText(m.clipboard)
		m.updateModified()
		m.setMessage("Pasted from internal clipboard")
	} else {
		m.setMessage("Nothing to paste")
	}
	return m, nil
}

func (m Model) handleUndo() (tea.Model, tea.Cmd) {
	if m.textBuffer.Undo() {
		m.updateModified()
		m.ensureCursorVisible()
	} else {
		m.setMessage("Nothing to undo")
	}
	return m, nil
}

func (m Model) handleRedo() (tea.Model, tea.Cmd) {
	if m.textBuffer.Redo() {
		m.updateModified()
		m.ensureCursorVisible()
	} else {
		m.setMessage("Nothing to redo")
	}
	return m, nil
}

func (m Model) handleSelectAll() (tea.Model, tea.Cmd) {
	m.textBuffer.SelectAll()
	return m, nil
}

func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	return m.renderEditor()
}

func (m Model) handleFindNext() (tea.Model, tea.Cmd) {
	if len(m.findResults) == 0 {
		m.setMessage("No search results available")
		return m, nil
	}

	m.findIndex = (m.findIndex + 1) % len(m.findResults)
	m.jumpToCurrentResult()
	m.setSearchMessage()
	return m, nil
}

func (m Model) handleFindPrev() (tea.Model, tea.Cmd) {
	if len(m.findResults) == 0 {
		m.setMessage("No search results available")
		return m, nil
	}

	m.findIndex--
	if m.findIndex < 0 {
		m.findIndex = len(m.findResults) - 1
	}
	
	m.jumpToCurrentResult()
	m.setSearchMessage()
	return m, nil
}

func (m *Model) setSearchMessage() {
	if len(m.findResults) > 0 && m.findIndex >= 0 && m.findIndex < len(m.findResults) {
		currentResult := m.findResults[m.findIndex]
		lines := m.textBuffer.GetLines()
		var linePreview string
		if currentResult.Line < len(lines) {
			line := lines[currentResult.Line]
			start := max(0, currentResult.Column-10)
			end := min(len(line), currentResult.Column+len(m.lastSearchQuery)+10)
			linePreview = line[start:end]
			if start > 0 {
				linePreview = "..." + linePreview
			}
			if end < len(line) {
				linePreview = linePreview + "..."
			}
		}
		
		m.setMessage(fmt.Sprintf("Match %d/%d at line %d: %s", 
			m.findIndex+1, 
			len(m.findResults), 
			currentResult.Line+1,
			linePreview))
	}
}

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


func (m Model) renderEditor() string {
	lines := m.textBuffer.GetLines()
	cursor := m.textBuffer.GetCursor()
	selection := m.textBuffer.GetSelection()

	minibufferHeight := m.getMinibufferHeight()
	visibleLines := m.height - 5 - minibufferHeight
	
	startLine := m.scrollOffset
	endLine := startLine + visibleLines

	if endLine > len(lines) {
		endLine = len(lines) 
	}

	var content strings.Builder

	for i := 0; i < visibleLines; i++ {
		actualLineIndex := startLine + i
		lineNum := lineNumberStyle.Render(fmt.Sprintf("%4d", actualLineIndex+1))

		line := ""
		if actualLineIndex < len(lines) {
			line = lines[actualLineIndex]
		}

		renderedLine := ""
		if actualLineIndex < len(lines) {
			renderedLine = m.renderLineWithSelection(line, actualLineIndex, cursor, selection)
		} else {
			renderedLine = ""
		}
		if actualLineIndex == cursor.Line {
			renderedLine = cursorLineStyle.Render(renderedLine)
		}

		lineContent := lineNum + " " + renderedLine
		paddedLine := m.padLineToWidth(lineContent)
		content.WriteString(paddedLine + "\n")
	}

	editorContent := content.String()
	editor := editorStyle.
		Width(m.width - 2).
		Height(visibleLines + 2). 
		Render(editorContent)

	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Center, editor, statusBar)
}

func (m Model) renderLineWithSelection(line string, lineIndex int, cursor Position, selection *Selection) string {
	if selection == nil {
		if lineIndex == cursor.Line {
			if cursor.Column >= len(line) {
				return line + "█"
			} else {
				return line[:cursor.Column] + "█" + line[cursor.Column+1:]
			}
		}
		return line
	}

	start, end := m.normalizeSelection(selection)

	if lineIndex < start.Line || lineIndex > end.Line {
		if lineIndex == cursor.Line {
			if cursor.Column >= len(line) {
				return line + "█"
			} else {
				return line[:cursor.Column] + "█" + line[cursor.Column+1:]
			}
		}
		return line
	}

	var result strings.Builder

	if lineIndex == start.Line && lineIndex == end.Line {
		result.WriteString(line[:start.Column])
		selectedText := line[start.Column:end.Column]
		result.WriteString(selectedTextStyle.Render(selectedText))
		result.WriteString(line[end.Column:])
	} else if lineIndex == start.Line {
		result.WriteString(line[:start.Column])
		selectedText := line[start.Column:]
		result.WriteString(selectedTextStyle.Render(selectedText))
	} else if lineIndex == end.Line {
		selectedText := line[:end.Column]
		result.WriteString(selectedTextStyle.Render(selectedText))
		result.WriteString(line[end.Column:])
	} else {
		result.WriteString(selectedTextStyle.Render(line))
	}

	return result.String()
}

func (m Model) renderStatusBar() string {
	if m.minibufferType != MinibufferNone {
		return m.renderMinibuffer()
	}

	cursor := m.textBuffer.GetCursor()
	lines := m.textBuffer.GetLines()

	var modifiedIndicator string
	if m.modified {
		modifiedIndicator = modifiedStyle.Render(" *")
	}

	filename := m.filename
	if filename == "" {
		filename = "<untitled>"
	}

	position := fmt.Sprintf("Line %d, Column %d", cursor.Line+1, cursor.Column+1)
	totalLines := fmt.Sprintf("Total: %d lines", len(lines))

	left := fmt.Sprintf("%s%s", filename, modifiedIndicator)
	center := position
	right := totalLines

	if m.message != "" && time.Since(m.messageTime) < 3*time.Second {
		center = m.message
	}

	statusWidth := m.width - 4
	centerStart := (statusWidth - len(center)) / 2
	rightStart := statusWidth - len(right)

	var status strings.Builder
	status.WriteString(left)

	for i := len(left); i < centerStart; i++ {
		status.WriteString(" ")
	}
	status.WriteString(center)
	for i := len(left) + len(center) + (centerStart - len(left)); i < rightStart; i++ {
		status.WriteString(" ")
	}
	status.WriteString(right)

	currentLen := len(status.String())
	if currentLen < statusWidth {
		status.WriteString(strings.Repeat(" ", statusWidth-currentLen))
	}

	return statusBarStyle.Width(m.width - 2).Render(status.String())
}

func (m Model) renderHelp() string {
	var help strings.Builder

	help.WriteString(helpTitleStyle.Render("Text Editor Help"))
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
		{"Home/End", "Move to line start/end (or first/last result)"},
		{"PgUp/PgDn", "Move by page"},
		{"Enter", "Confirm action / Go to selected result"},
		{"Escape", "Cancel dialog input"},
		{"Ctrl+H", "Toggle this help"},
	}

	for _, item := range helpItems {
		help.WriteString(helpKeyStyle.Render(fmt.Sprintf("%-15s", item.key)))
		help.WriteString(" ")
		help.WriteString(helpDescStyle.Render(item.desc))
		help.WriteString("\n")
	}

	help.WriteString("\n")
	help.WriteString(helpStyle.Render("Press Ctrl+H again to close help"))

	return helpBoxStyle.Render(help.String())
}

func (m *Model) updateModified() {
	m.modified = m.textBuffer.GetContent() != m.originalText
}

func (m *Model) setMessage(msg string) {
	m.message = msg
	m.messageTime = time.Now()
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

func stripAnsiCodes(s string) string {
	result := ""
	inEscape := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			continue
		}

		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}

		result += string(s[i])
	}

	return result
}

func (m Model) getVisibleLines() int {
	return m.height - 5
}

func (m *Model) ensureCursorVisible() {
	cursor := m.textBuffer.GetCursor()
	visibleLines := m.getVisibleLines()

	if cursor.Line < m.scrollOffset {
		m.scrollOffset = cursor.Line
	} else if cursor.Line >= m.scrollOffset+visibleLines {
		m.scrollOffset = cursor.Line - visibleLines + 1
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m *Model) centerCursorOnScreen() {
	cursor := m.textBuffer.GetCursor()
	visibleLines := m.getVisibleLines()
	totalLines := len(m.textBuffer.GetLines())
	
	// Try to center the cursor line on screen
	targetOffset := cursor.Line - visibleLines/2
	
	// Clamp the offset to valid bounds
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

func main() {
	var filename string
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	model := NewModel(filename)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}