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
		if m.minibufferType != MinibufferNone {
			return m.handleMinibufferInput(msg)
		}

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Save):
			return m.handleSave()

		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

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

		case key.Matches(msg, keys.ShiftLeft):
			m.textBuffer.MoveCursor(0, -1, true)
			m.updateModified()
			m.ensureCursorVisible()

		case key.Matches(msg, keys.ShiftRight):
			m.textBuffer.MoveCursor(0, 1, true)
			m.updateModified()
			m.ensureCursorVisible()

		case key.Matches(msg, keys.ShiftUp):
			m.textBuffer.MoveCursor(-1, 0, true)
			m.updateModified()
			m.ensureCursorVisible()

		case key.Matches(msg, keys.ShiftDown):
			m.textBuffer.MoveCursor(1, 0, true)
			m.updateModified()
			m.ensureCursorVisible()

		case msg.Type == tea.KeyCtrlLeft:
			m.textBuffer.MoveToWordBoundary(false, false)
			m.ensureCursorVisible()

		case msg.Type == tea.KeyCtrlRight:
			m.textBuffer.MoveToWordBoundary(true, false)
			m.ensureCursorVisible()

		case msg.Type == tea.KeyCtrlShiftLeft:
			m.textBuffer.MoveToWordBoundary(false, true)
			m.ensureCursorVisible()

		case msg.Type == tea.KeyCtrlShiftRight:
			m.textBuffer.MoveToWordBoundary(true, true)
			m.ensureCursorVisible()

		case msg.Type == tea.KeyLeft:
			m.textBuffer.MoveCursor(0, -1, false)
			m.ensureCursorVisible()

		case msg.Type == tea.KeyRight:
			m.textBuffer.MoveCursor(0, 1, false)
			m.ensureCursorVisible()

		case msg.Type == tea.KeyUp:
			m.textBuffer.MoveCursor(-1, 0, false)
			m.ensureCursorVisible()

		case msg.Type == tea.KeyDown:
			m.textBuffer.MoveCursor(1, 0, false)
			m.ensureCursorVisible()

		case msg.Type == tea.KeyHome:
			cursor := m.textBuffer.GetCursor()
			m.textBuffer.SetCursor(Position{Line: cursor.Line, Column: 0})
			m.ensureCursorVisible()

		case msg.Type == tea.KeyEnd:
			cursor := m.textBuffer.GetCursor()
			lines := m.textBuffer.GetLines()
			if cursor.Line < len(lines) {
				m.textBuffer.SetCursor(Position{Line: cursor.Line, Column: len(lines[cursor.Line])})
			}
			m.ensureCursorVisible()

		case msg.Type == tea.KeyPgUp:
			visible := m.getVisibleLines()
			m.textBuffer.MoveCursor(-visible, 0, false)
			m.ensureCursorVisible()

		case msg.Type == tea.KeyPgDown:
			visible := m.getVisibleLines()
			m.textBuffer.MoveCursor(visible, 0, false)
			m.ensureCursorVisible()

		case msg.Type == tea.KeyEnter:
			m.textBuffer.InsertText("\n")
			m.updateModified()
			m.ensureCursorVisible()

		case msg.Type == tea.KeyBackspace:
			m.textBuffer.DeleteChar(true)
			m.updateModified()
			m.ensureCursorVisible()

		case msg.Type == tea.KeyDelete:
			m.textBuffer.DeleteChar(false)
			m.updateModified()
			m.ensureCursorVisible()

		case msg.Type == tea.KeyTab:
			m.textBuffer.InsertText("\t")
			m.updateModified()
			m.ensureCursorVisible()

		case msg.Type == tea.KeyRunes:
			if len(msg.Runes) > 0 {
				m.textBuffer.InsertText(string(msg.Runes))
				m.updateModified()
				m.ensureCursorVisible()
			}
		}
	}
	return m, nil
}
