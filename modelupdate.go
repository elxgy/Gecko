package main

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyHandler defines the signature for key handling functions
type KeyHandler func(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd)

// KeyMapping represents a key binding and its associated handler
type KeyMapping struct {
	Binding key.Binding
	Handler KeyHandler
}

// keyMappings provides a more efficient lookup for key handlers
var keyMappings = []KeyMapping{
	{keys.Quit, handleQuit},
	{keys.Save, handleSave},
	{keys.Help, handleHelp},
	{keys.GoToLine, handleGoToLine},
	{keys.Find, handleFind},
	{keys.FindNext, handleFindNext},
	{keys.FindPrev, handleFindPrev},
	{keys.Copy, handleCopy},
	{keys.Cut, handleCut},
	{keys.Paste, handlePaste},
	{keys.Undo, handleUndo},
	{keys.Redo, handleRedo},
	{keys.SelectAll, handleSelectAll},
	{keys.ShiftLeft, handleShiftLeft},
	{keys.ShiftRight, handleShiftRight},
	{keys.ShiftUp, handleShiftUp},
	{keys.ShiftDown, handleShiftDown},
	{keys.AltLeft, handleAltLeft},
	{keys.AltRight, handleAltRight},
}

// Update handles all incoming messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)
	case tea.KeyMsg:
		return m.handleKeyMessage(msg)
	case blinkMsg:
		return m.handleCursorBlink()
	}
	return m, nil
}

// handleWindowResize processes window resize events
func (m Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.recalculateLayout()
	m.postMovementUpdate()
	return m, nil
}

// handleKeyMessage processes keyboard input
func (m Model) handleKeyMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.minibufferType != MinibufferNone {
		return m.handleMinibufferInput(msg)
	}

	if handler := findKeyHandler(msg); handler != nil {
		return handler(m, msg)
	}

	return m.handleSpecialKeys(msg)
}

// handleCursorBlink toggles cursor visibility
func (m Model) handleCursorBlink() (tea.Model, tea.Cmd) {
	m.cursorVisible = !m.cursorVisible
	return m, blinkTick()
}

// findKeyHandler efficiently finds the appropriate handler for a key message
func findKeyHandler(msg tea.KeyMsg) KeyHandler {
	for _, mapping := range keyMappings {
		if key.Matches(msg, mapping.Binding) {
			return mapping.Handler
		}
	}
	return nil
}

// handleSpecialKeys processes special key combinations and navigation keys
func (m Model) handleSpecialKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlLeft:
		m.textBuffer.MoveToWordBoundary(false, false)
		m.postMovementUpdate()
		return m, nil
	case tea.KeyCtrlRight:
		m.textBuffer.MoveToWordBoundary(true, false)
		m.postMovementUpdate()
		return m, nil
	default:
		return m.handleNavigationKeys(msg)
	}
}

// handleNavigationKeys processes arrow keys and other navigation inputs
func (m Model) handleNavigationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var err error
	switch msg.Type {
	case tea.KeyLeft:
		err = m.textBuffer.MoveCursorDelta(0, -1, false)
	case tea.KeyRight:
		err = m.textBuffer.MoveCursorDelta(0, 1, false)
	case tea.KeyUp:
		err = m.textBuffer.MoveCursorDelta(-1, 0, false)
	case tea.KeyDown:
		err = m.textBuffer.MoveCursorDelta(1, 0, false)
	case tea.KeyHome:
		return m.handleHomeKey()
	case tea.KeyEnd:
		return m.handleEndKey()
	case tea.KeyPgUp, tea.KeyPgDown:
		return m.handlePageKeys(msg)
	default:
		return m.handleTextInput(msg)
	}

	if err != nil {
		return m, nil // Silently handle errors to maintain UI responsiveness
	}

	m.postMovementUpdate()
	return m, nil
}

// handleHomeKey moves cursor to beginning of current line
func (m Model) handleHomeKey() (tea.Model, tea.Cmd) {
	cursor := m.textBuffer.GetCursor()
	m.textBuffer.SetCursor(Position{Line: cursor.Line, Column: 0})
	m.postMovementUpdate()
	return m, nil
}

// handleEndKey moves cursor to end of current line
func (m Model) handleEndKey() (tea.Model, tea.Cmd) {
	cursor := m.textBuffer.GetCursor()
	lines := m.textBuffer.GetLines()
	if cursor.Line < len(lines) {
		m.textBuffer.SetCursor(Position{Line: cursor.Line, Column: len(lines[cursor.Line])})
	}
	m.postMovementUpdate()
	return m, nil
}

// handlePageKeys processes page up/down navigation
func (m Model) handlePageKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visible := m.getVisibleLines()
	var err error
	switch msg.Type {
	case tea.KeyPgUp:
		err = m.textBuffer.MoveCursorDelta(-visible, 0, false)
	case tea.KeyPgDown:
		err = m.textBuffer.MoveCursorDelta(visible, 0, false)
	}

	if err != nil {
		return m, nil // Silently handle errors to maintain UI responsiveness
	}

	m.postMovementUpdate()
	return m, nil
}

func (m *Model) postMovementUpdate() {
	m.ensureCursorVisible()
	m.updateWordBounds()
	m.viewportY = m.scrollOffset
	m.applySyntaxHighlighting()
}

func (m *Model) recalculateLayout() {
	m.ensureCursorVisible()

}

// handleTextInput processes text modification keys
func (m Model) handleTextInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var text string
	var shouldInsert bool

	switch msg.Type {
	case tea.KeyEnter:
		text, shouldInsert = "\n", true
	case tea.KeyTab:
		text, shouldInsert = "\t", true
	case tea.KeySpace:
		text, shouldInsert = " ", true
	case tea.KeyBackspace:
		return m.handleBackspace()
	case tea.KeyDelete:
		return m.handleDelete()
	case tea.KeyRunes:
		text, shouldInsert = m.buildPrintableText(msg.Runes)
	default:
		return m, nil
	}

	if shouldInsert && text != "" {
		return m.insertText(text)
	}
	return m, nil
}

// buildPrintableText creates a string from printable runes
func (m Model) buildPrintableText(runes []rune) (string, bool) {
	if len(runes) == 0 {
		return "", false
	}

	var b strings.Builder
	for _, r := range runes {
		if unicode.IsPrint(r) && r != '\u0000' {
			b.WriteRune(r)
		}
	}
	return b.String(), b.Len() > 0
}

// insertText inserts text and updates the model state
func (m Model) insertText(text string) (tea.Model, tea.Cmd) {
	if err := m.textBuffer.InsertText(text); err != nil {
		return m, nil // Silently handle errors to maintain UI responsiveness
	}

	m.invalidateHighlightCache()
	m.updateModified()
	m.postMovementUpdate()
	return m, nil
}

// handleBackspace processes backspace key
func (m Model) handleBackspace() (tea.Model, tea.Cmd) {
	if err := m.textBuffer.DeleteChar(true); err != nil {
		return m, nil
	}
	m.invalidateHighlightCache()
	m.updateModified()
	m.postMovementUpdate()
	return m, nil
}

// handleDelete processes delete key
func (m Model) handleDelete() (tea.Model, tea.Cmd) {
	if err := m.textBuffer.DeleteChar(false); err != nil {
		return m, nil
	}
	m.invalidateHighlightCache()
	m.updateModified()
	m.postMovementUpdate()
	return m, nil
}

func handleQuit(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func handleSave(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleSave()
}

func handleHelp(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.showHelp = !m.showHelp
	return m, nil
}

func handleGoToLine(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.minibufferType = MinibufferGoToLine
	m.minibufferInput = ""
	m.minibufferCursorPos = 0
	m.postMovementUpdate()
	return m, nil
}

func handleFind(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.minibufferType = MinibufferFind
	m.minibufferInput = ""
	m.minibufferCursorPos = 0
	return m, nil
}

func handleFindNext(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleFindNext()
}

func handleFindPrev(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleFindPrev()
}

func handleCopy(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleCopy()
}

func handleCut(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleCut()
}

func handlePaste(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handlePaste()
}

func handleUndo(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleUndo()
}

func handleRedo(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleRedo()
}

func handleSelectAll(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleSelectAll()
}

func (m *Model) updateWordBounds() {
	// Only update word bounds if cursor position has changed
	currentPos := m.textBuffer.GetCursor()
	if m.lastWordBoundsCursor.Line != currentPos.Line || m.lastWordBoundsCursor.Column != currentPos.Column {
		m.currentWordStart, m.currentWordEnd = m.textBuffer.GetWordBoundsAtCursor()
		m.lastWordBoundsCursor = currentPos
	}
}

// handleShiftLeft extends selection leftward
func handleShiftLeft(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleShiftMovement(0, -1)
}

// handleShiftRight extends selection rightward
func handleShiftRight(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleShiftMovement(0, 1)
}

// handleShiftUp extends selection upward
func handleShiftUp(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleShiftMovement(-1, 0)
}

// handleShiftDown extends selection downward
func handleShiftDown(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleShiftMovement(1, 0)
}

// handleShiftMovement is a common handler for all shift+arrow key combinations
func (m Model) handleShiftMovement(lineDelta, columnDelta int) (tea.Model, tea.Cmd) {
	if err := m.textBuffer.MoveCursorDelta(lineDelta, columnDelta, true); err != nil {
		return m, nil
	}

	m.postMovementUpdate()
	return m, nil
}

// handleAltLeft moves cursor to previous word boundary
func handleAltLeft(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleWordBoundaryMovement(false)
}

// handleAltRight moves cursor to next word boundary
func handleAltRight(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleWordBoundaryMovement(true)
}

// handleWordBoundaryMovement moves cursor to word boundaries
func (m Model) handleWordBoundaryMovement(forward bool) (tea.Model, tea.Cmd) {
	m.textBuffer.MoveToWordBoundary(forward, false)
	m.postMovementUpdate()
	return m, nil
}
