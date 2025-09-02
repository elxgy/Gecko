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
	lastHighlightedHash string
	dirtyLines          map[int]bool
	highlightingEnabled bool
}

type SelectionInfo struct {
	hasSelection bool
	startCol     int
	endCol       int
}

func NewModel(filename string) Model {
	var content string
	var originalText string
	var loadError string

	if filename != "" {
		data, err := os.ReadFile(filename)
		if err != nil {
			// Handle different types of file errors
			if os.IsNotExist(err) {
				loadError = fmt.Sprintf("File not found: %s", filename)
			} else if os.IsPermission(err) {
				loadError = fmt.Sprintf("Permission denied: %s", filename)
			} else {
				loadError = fmt.Sprintf("Error reading file: %v", err)
			}
			// Create empty buffer for new file
			content = ""
			originalText = ""
		} else {
			content = string(data)
			originalText = content
		}
	}

	textBuffer := NewTextBuffer(content)
	model := Model{
		scrollOffset:        0,
		textBuffer:          textBuffer,
		filename:            filename,
		originalText:        originalText,
		modified:            false,
		findResults:         []Position{},
		findIndex:           -1,
		maxResultsDisplay:   8,
		highlighter:         NewHighlighter(filename),
		dirtyLines:          make(map[int]bool),
		highlightingEnabled: true,
	}

	// Set load error message if there was one
	if loadError != "" {
		model.setMessage(loadError)
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
