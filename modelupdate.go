package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type keyHandler func(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd)

var keyHandlers = map[string]keyHandler{
	"quit":       handleQuit,
	"save":       handleSave,
	"help":       handleHelp,
	"goto":       handleGoToLine,
	"find":       handleFind,
	"findNext":   handleFindNext,
	"findPrev":   handleFindPrev,
	"copy":       handleCopy,
	"cut":        handleCut,
	"paste":      handlePaste,
	"undo":       handleUndo,
	"redo":       handleRedo,
	"selectAll":  handleSelectAll,
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
		m.width = msg.Width
		m.height = msg.Height
		m.postMovementUpdate()
		return m, nil

	case tea.KeyMsg:
		if m.minibufferType != MinibufferNone {
			return m.handleMinibufferInput(msg)
		}

		if handler := matchKeyHandler(msg); handler != nil {
			return handler(m, msg)
		}

		return handleSpecialKeys(m, msg)
	case blinkMsg:
		m.cursorVisible = !m.cursorVisible
		return m, blinkTick()
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
		m.postMovementUpdate()
		return m, nil
	}
	if msg.Type == tea.KeyCtrlRight {
		m.textBuffer.MoveToWordBoundary(true, false)
		m.postMovementUpdate()
		return m, nil
	}

	return handleArrowKeys(m, msg)
}

func handleArrowKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyLeft:
		m.textBuffer.MoveCursor(0, -1, false)
	case tea.KeyRight:
		m.textBuffer.MoveCursor(0, 1, false)
	case tea.KeyUp:
		m.textBuffer.MoveCursor(-1, 0, false)
	case tea.KeyDown:
		m.textBuffer.MoveCursor(1, 0, false)
	default:
		return handleHomeEndKeys(m, msg)
	}
	m.postMovementUpdate()
	return m, nil
}

func handleHomeEndKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyHome:
		cursor := m.textBuffer.GetCursor()
		m.textBuffer.SetCursor(Position{Line: cursor.Line, Column: 0})
	case tea.KeyEnd:
		cursor := m.textBuffer.GetCursor()
		lines := m.textBuffer.GetLines()
		if cursor.Line < len(lines) {
			m.textBuffer.SetCursor(Position{Line: cursor.Line, Column: len(lines[cursor.Line])})
		}
	default:
		return handlePageKeys(m, msg)
	}
	m.postMovementUpdate()
	return m, nil
}

func handlePageKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visible := m.getVisibleLines()
	switch msg.Type {
	case tea.KeyPgUp:
		m.textBuffer.MoveCursor(-visible, 0, false)
	case tea.KeyPgDown:
		m.textBuffer.MoveCursor(visible, 0, false)
	default:
		return handleTextModification(m, msg)
	}
	m.postMovementUpdate()
	return m, nil
}

func (m *Model) postMovementUpdate() {
	m.ensureCursorVisible()
	m.updateWordBounds()
}

func handleTextModification(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.textBuffer.InsertText("\n")
	case tea.KeyBackspace:
		m.textBuffer.DeleteChar(true)
	case tea.KeyDelete:
		m.textBuffer.DeleteChar(false)
	case tea.KeyTab:
		m.textBuffer.InsertText("\t")
	case tea.KeySpace:
		m.textBuffer.InsertText(" ")
	case tea.KeyRunes:
		if len(msg.Runes) > 0 {
			var b strings.Builder
			for _, r := range msg.Runes {
				if unicode.IsPrint(r) && r != '\u0000' {
					b.WriteRune(r)
				}
			}
			if b.Len() > 0 {
				m.textBuffer.InsertText(b.String())
			}
		}
	default:
		return m, nil
	}
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
	currentPos := m.textBuffer.cursor
	if m.lastWordBoundsCursor.Line != currentPos.Line || m.lastWordBoundsCursor.Column != currentPos.Column {
		m.currentWordStart, m.currentWordEnd = m.textBuffer.GetWordBoundsAtCursor()
		m.lastWordBoundsCursor = currentPos
	}
}

func handleShiftLeft(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
    m.textBuffer.MoveCursor(0, -1, true)
    m.postMovementUpdate()
    selText := m.textBuffer.GetSelectedText()
    m.message = fmt.Sprintf("Selected %d characters", len(selText))
    m.messageTime = time.Now()
    return m, nil
}

func handleShiftRight(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
    m.textBuffer.MoveCursor(0, 1, true)
    m.postMovementUpdate()
    selText := m.textBuffer.GetSelectedText()
    m.message = fmt.Sprintf("Selected %d characters", len(selText))
    m.messageTime = time.Now()
    return m, nil
}

func handleShiftUp(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
    m.textBuffer.MoveCursor(-1, 0, true)
    m.postMovementUpdate()
    selText := m.textBuffer.GetSelectedText()
    m.message = fmt.Sprintf("Selected %d characters", len(selText))
    m.messageTime = time.Now()
    return m, nil
}

func handleShiftDown(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
    m.textBuffer.MoveCursor(1, 0, true)
    m.postMovementUpdate()
    selText := m.textBuffer.GetSelectedText()
    m.message = fmt.Sprintf("Selected %d characters", len(selText))
    m.messageTime = time.Now()
    return m, nil
}

func handleAltLeft(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
    m.textBuffer.MoveToWordBoundary(false, true)
    m.postMovementUpdate()
    return m, nil
}

func handleAltRight(m Model, _ tea.KeyMsg) (tea.Model, tea.Cmd) {
    m.textBuffer.MoveToWordBoundary(true, true)
    m.postMovementUpdate()
    return m, nil
}