package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Styles are now defined in styles.go

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
	return tea.Tick(time.Millisecond*TickIntervalMs, func(t time.Time) tea.Msg {
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

	fmt.Print(ClearScreen)
}
