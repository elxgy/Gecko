package main

import (
	"fmt"
	"os"
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
	highlighter         *Highlighter
	highlightedContent  []string
}

type SelectionInfo struct {
	hasSelection bool
	startCol     int
	endCol       int
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
	model := Model{
		scrollOffset:      0,
		textBuffer:        textBuffer,
		filename:          filename,
		originalText:      originalText,
		modified:          false,
		findResults:       []Position{},
		findIndex:         -1,
		maxResultsDisplay: 8,
		highlighter:       NewHighlighter(filename),
	}

	model.applySyntaxHighlighting()
	model.ensureCursorVisible()
	return model
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{}}
	})
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

	fmt.Print("\033[2J\033[H")
}