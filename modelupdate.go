package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		//==================== MINIBUFFER INPUT ====================

		if m.minibufferType != MinibufferNone {
			return m.handleMinibufferInput(msg)
		}

		//==================== APPLICATION CONTROL KEYS ====================

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Save):
			return m.handleSave()

		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, keys.GoToLine):
			m.minibufferType = MinibufferGoToLine
			m.minibufferInput = ""
			m.minibufferCursorPos = 0
			return m, nil

		case key.Matches(msg, keys.Find):
			m.minibufferType = MinibufferFind
			m.minibufferInput = ""
			m.minibufferCursorPos = 0
			return m, nil

		case key.Matches(msg, keys.FindNext):
			return m.handleFindNext()

		case key.Matches(msg, keys.FindPrev):
			return m.handleFindPrev()
		}

		//==================== CLIPBOARD & EDITING OPERATIONS ====================

		switch {
		case key.Matches(msg, keys.Copy):
			return m.handleCopy()

		case key.Matches(msg, keys.Cut):
			return m.handleCut()

		case key.Matches(msg, keys.Paste):
			return m.handlePaste()

		case key.Matches(msg, keys.Undo):
			return m.handleUndo()

		case key.Matches(msg, keys.Redo):
			return m.handleRedo()

		case key.Matches(msg, keys.SelectAll):
			return m.handleSelectAll()
		}

		//==================== SELECTION NAVIGATION (SHIFT + ARROWS) ====================

		switch {
		case key.Matches(msg, keys.ShiftLeft):
			m.textBuffer.MoveCursor(0, -1, true)
			m.ensureCursorVisible()

		case key.Matches(msg, keys.ShiftRight):
			m.textBuffer.MoveCursor(0, 1, true)
			m.ensureCursorVisible()

		case key.Matches(msg, keys.ShiftUp):
			m.textBuffer.MoveCursor(-1, 0, true)
			m.ensureCursorVisible()

		case key.Matches(msg, keys.ShiftDown):
			m.textBuffer.MoveCursor(1, 0, true)
			m.ensureCursorVisible()
		}

		//==================== WORD NAVIGATION (CTRL + ARROWS) ====================

		switch {
		case msg.Type == tea.KeyCtrlLeft:
			m.textBuffer.MoveToWordBoundary(false, false)
			m.ensureCursorVisible()

		case msg.Type == tea.KeyCtrlRight:
			m.textBuffer.MoveToWordBoundary(true, false)
			m.ensureCursorVisible()
		}

		//==================== WORD SELECTION (ALT + ARROWS) ====================

		switch {
		case key.Matches(msg, keys.AltLeft):
			m.textBuffer.MoveToWordBoundary(false, true)
			m.ensureCursorVisible()
			return m, nil

		case key.Matches(msg, keys.AltRight):
			m.textBuffer.MoveToWordBoundary(true, true)
			m.ensureCursorVisible()
			return m, nil
		}

		//==================== BASIC ARROW NAVIGATION ====================

		switch msg.Type {
		case tea.KeyLeft:
			m.textBuffer.MoveCursor(0, -1, false)
			m.ensureCursorVisible()

		case tea.KeyRight:
			m.textBuffer.MoveCursor(0, 1, false)
			m.ensureCursorVisible()

		case tea.KeyUp:
			m.textBuffer.MoveCursor(-1, 0, false)
			m.ensureCursorVisible()

		case tea.KeyDown:
			m.textBuffer.MoveCursor(1, 0, false)
			m.ensureCursorVisible()
		}

		//==================== LINE NAVIGATION (HOME/END) ====================

		switch msg.Type {
		case tea.KeyHome:
			cursor := m.textBuffer.GetCursor()
			m.textBuffer.SetCursor(Position{Line: cursor.Line, Column: 0})
			m.ensureCursorVisible()

		case tea.KeyEnd:
			cursor := m.textBuffer.GetCursor()
			lines := m.textBuffer.GetLines()
			if cursor.Line < len(lines) {
				m.textBuffer.SetCursor(Position{Line: cursor.Line, Column: len(lines[cursor.Line])})
			}
			m.ensureCursorVisible()
		}

		//==================== PAGE NAVIGATION (PGUP/PGDN) ====================

		switch msg.Type {
		case tea.KeyPgUp:
			visible := m.getVisibleLines()
			m.textBuffer.MoveCursor(-visible, 0, false)
			m.ensureCursorVisible()

		case tea.KeyPgDown:
			visible := m.getVisibleLines()
			m.textBuffer.MoveCursor(visible, 0, false)
			m.ensureCursorVisible()
		}

		//==================== TEXT MODIFICATION KEYS ====================

		switch msg.Type {
		case tea.KeyEnter:
			m.textBuffer.InsertText("\n")
			m.updateModified()
			m.ensureCursorVisible()

		case tea.KeyBackspace:
			m.textBuffer.DeleteChar(true)
			m.updateModified()
			m.ensureCursorVisible()

		case tea.KeyDelete:
			m.textBuffer.DeleteChar(false)
			m.updateModified()
			m.ensureCursorVisible()

		case tea.KeyTab:
			m.textBuffer.InsertText("\t")
			m.updateModified()
			m.ensureCursorVisible()

		case tea.KeySpace:
			m.textBuffer.InsertText(" ")
			m.updateModified()
			m.ensureCursorVisible()

		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				text := string(msg.Runes)
				m.textBuffer.InsertText(text)
				m.updateModified()
				m.ensureCursorVisible()
			}
		}
	}
	return m, nil
}
