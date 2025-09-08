package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	textBuffer           *TextBuffer
	filename             string
	modified             bool
	originalText         string
	width                int
	height               int
	showHelp             bool
	lastSaved            time.Time
	message              string
	messageTime          time.Time
	clipboard            string
	scrollOffset         int
	horizontalOffset     int
	minibufferType       MinibufferType
	minibufferInput      string
	minibufferCursorPos  int
	findResults          []Position
	findIndex            int
	lastSearchQuery      string
	searchResultsOffset  int
	maxResultsDisplay    int
	highlighter          *Highlighter
	highlightedContent   []string
	currentWordStart     int
	currentWordEnd       int
	cursorVisible        bool
	lastWordBoundsCursor Position
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
		scrollOffset:         0,
		horizontalOffset:     0,
		textBuffer:           textBuffer,
		filename:             filename,
		originalText:         originalText,
		modified:             false,
		findResults:          []Position{},
		findIndex:            -1,
		maxResultsDisplay:    8,
		highlighter:          NewHighlighter(filename),
		currentWordStart:     -1,
		currentWordEnd:       -1,
		lastWordBoundsCursor: Position{Line: -1, Column: -1},
	}

	model.applySyntaxHighlighting()
	model.ensureCursorVisible()
	model.updateWordBounds()
	return model
}

type blinkMsg time.Time

func blinkTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return blinkMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return blinkTick()
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
