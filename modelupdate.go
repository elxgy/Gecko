package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type keyHandler func(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd)

var keyHandlers = map[string]keyHandler{
	"quit":       func(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) { return m, tea.Quit },
	"save":       func(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) { return m.handleSave() },
	"help":       handleHelp,
	"goto":       handleGoToLine,
	"find":       handleFind,
	"findNext":   func(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) { return m.handleFindNext() },
	"findPrev":   func(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) { return m.handleFindPrev() },
	"copy":       func(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) { return m.handleCopy() },
	"cut":        func(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) { return m.handleCut() },
	"paste":      func(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) { return m.handlePaste() },
	"undo":       func(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) { return m.handleUndo() },
	"redo":       func(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) { return m.handleRedo() },
	"selectAll":  func(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) { return m.handleSelectAll() },
	"shiftLeft":  handleShiftLeft,
	"shiftRight": handleShiftRight,
	"shiftUp":    handleShiftUp,
	"shiftDown":  handleShiftDown,
	"altLeft":    handleAltLeft,
	"altRight":   handleAltRight,
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Enforce minimum terminal dimensions to prevent UI breakage
		m.width = max(msg.Width, MinTerminalWidth)
		m.height = max(msg.Height, MinTerminalHeight)
		m.ensureCursorVisible()
		return m, nil

	case tea.KeyMsg:
		if m.minibufferType != MinibufferNone {
			return m.handleMinibufferInput(msg)
		}

		if handler := matchKeyHandler(msg); handler != nil {
			return handler(m, msg)
		}

		return handleSpecialKeys(m, msg)
	}

	return m, nil
}

func matchKeyHandler(msg tea.KeyMsg) keyHandler {
	if key.Matches(msg, keys.Quit) {
		return keyHandlers["quit"]
	}
	if key.Matches(msg, keys.Save) {
		return keyHandlers["save"]
	}
	if key.Matches(msg, keys.Help) {
		return keyHandlers["help"]
	}
	if key.Matches(msg, keys.GoToLine) {
		return keyHandlers["goto"]
	}
	if key.Matches(msg, keys.Find) {
		return keyHandlers["find"]
	}
	if key.Matches(msg, keys.FindNext) {
		return keyHandlers["findNext"]
	}
	if key.Matches(msg, keys.FindPrev) {
		return keyHandlers["findPrev"]
	}
	if key.Matches(msg, keys.Copy) {
		return keyHandlers["copy"]
	}
	if key.Matches(msg, keys.Cut) {
		return keyHandlers["cut"]
	}
	if key.Matches(msg, keys.Paste) {
		return keyHandlers["paste"]
	}
	if key.Matches(msg, keys.Undo) {
		return keyHandlers["undo"]
	}
	if key.Matches(msg, keys.Redo) {
		return keyHandlers["redo"]
	}
	if key.Matches(msg, keys.SelectAll) {
		return keyHandlers["selectAll"]
	}
	if key.Matches(msg, keys.ShiftLeft) {
		return keyHandlers["shiftLeft"]
	}
	if key.Matches(msg, keys.ShiftRight) {
		return keyHandlers["shiftRight"]
	}
	if key.Matches(msg, keys.ShiftUp) {
		return keyHandlers["shiftUp"]
	}
	if key.Matches(msg, keys.ShiftDown) {
		return keyHandlers["shiftDown"]
	}
	if key.Matches(msg, keys.AltLeft) {
		return keyHandlers["altLeft"]
	}
	if key.Matches(msg, keys.AltRight) {
		return keyHandlers["altRight"]
	}

	return nil
}

func handleSpecialKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlLeft {
		m.textBuffer.MoveToWordBoundary(false, false)
		m.ensureCursorVisible()
		return m, nil
	}
	if msg.Type == tea.KeyCtrlRight {
		m.textBuffer.MoveToWordBoundary(true, false)
		m.ensureCursorVisible()
		return m, nil
	}

	if msg.Type == tea.KeyLeft {
		m.textBuffer.MoveCursor(0, -1, false)
		m.ensureCursorVisible()
		return m, nil
	}
	if msg.Type == tea.KeyRight {
		m.textBuffer.MoveCursor(0, 1, false)
		m.ensureCursorVisible()
		return m, nil
	}
	if msg.Type == tea.KeyUp {
		m.textBuffer.MoveCursor(-1, 0, false)
		m.ensureCursorVisible()
		return m, nil
	}
	if msg.Type == tea.KeyDown {
		m.textBuffer.MoveCursor(1, 0, false)
		m.ensureCursorVisible()
		return m, nil
	}

	if msg.Type == tea.KeyHome {
		cursor := m.textBuffer.GetCursor()
		m.textBuffer.SetCursor(Position{Line: cursor.Line, Column: 0})
		m.ensureCursorVisible()
		return m, nil
	}
	if msg.Type == tea.KeyEnd {
		cursor := m.textBuffer.GetCursor()
		lines := m.textBuffer.GetLines()
		if cursor.Line < len(lines) {
			m.textBuffer.SetCursor(Position{Line: cursor.Line, Column: len(lines[cursor.Line])})
		}
		m.ensureCursorVisible()
		return m, nil
	}

	if msg.Type == tea.KeyPgUp {
		visible := m.getVisibleLines()
		m.textBuffer.MoveCursor(-visible, 0, false)
		m.ensureCursorVisible()
		return m, nil
	}
	if msg.Type == tea.KeyPgDown {
		visible := m.getVisibleLines()
		m.textBuffer.MoveCursor(visible, 0, false)
		m.ensureCursorVisible()
		return m, nil
	}

	return handleTextModification(m, msg)
}

func handleTextModification(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cursorLine := m.textBuffer.GetCursorLine()
	
	switch msg.Type {
	case tea.KeyEnter:
		m.textBuffer.InsertText("\n")
		m.updateModified()
		m.markLinesDirty(cursorLine, cursorLine+1)
		m.ensureCursorVisible()

	case tea.KeyBackspace:
		m.textBuffer.DeleteChar(true)
		m.updateModified()
		m.markLinesDirty(max(0, cursorLine-1), cursorLine)
		m.ensureCursorVisible()

	case tea.KeyDelete:
		m.textBuffer.DeleteChar(false)
		m.updateModified()
		m.markLinesDirty(cursorLine, cursorLine+1)
		m.ensureCursorVisible()

	case tea.KeyTab:
		m.textBuffer.InsertText("\t")
		m.updateModified()
		m.markLinesDirty(cursorLine, cursorLine)
		m.ensureCursorVisible()

	case tea.KeySpace:
		m.textBuffer.InsertText(" ")
		m.updateModified()
		m.markLinesDirty(cursorLine, cursorLine)
		m.ensureCursorVisible()

	case tea.KeyRunes:
		if len(msg.Runes) > 0 {
			text := string(msg.Runes)
			m.textBuffer.InsertText(text)
			m.updateModified()
			m.markLinesDirty(cursorLine, cursorLine)
			m.ensureCursorVisible()
		}
	}

	// Apply incremental highlighting for better performance
	m.applyIncrementalHighlighting()

	return m, nil
}

func handleHelp(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.showHelp = !m.showHelp
	return m, nil
}

func handleGoToLine(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.minibufferType = MinibufferGoToLine
	m.minibufferInput = ""
	m.minibufferCursorPos = 0
	return m, nil
}

func handleFind(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.minibufferType = MinibufferFind
	m.minibufferInput = ""
	m.minibufferCursorPos = 0
	return m, nil
}

func handleShiftLeft(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.textBuffer.MoveCursor(0, -1, true)
	m.ensureCursorVisible()
	return m, nil
}

func handleShiftRight(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.textBuffer.MoveCursor(0, 1, true)
	m.ensureCursorVisible()
	return m, nil
}

func handleShiftUp(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.textBuffer.MoveCursor(-1, 0, true)
	m.ensureCursorVisible()
	return m, nil
}

func handleShiftDown(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.textBuffer.MoveCursor(1, 0, true)
	m.ensureCursorVisible()
	return m, nil
}

func handleAltLeft(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.textBuffer.MoveToWordBoundary(false, true)
	m.ensureCursorVisible()
	return m, nil
}

func handleAltRight(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.textBuffer.MoveToWordBoundary(true, true)
	m.ensureCursorVisible()
	return m, nil
}
