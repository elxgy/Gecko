package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) handleSave() (tea.Model, tea.Cmd) {
	if m.filename != "" {
		err := m.saveFile()
		if err == nil {
			m.modified = false
			m.originalText = m.textBuffer.GetContent()
			m.lastSaved = time.Now()
			m.setMessage(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess)).Render("File saved successfully"))
		} else {
			m.setMessage(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Render(fmt.Sprintf("Error saving file: %v", err)))
		}
	} else {
		m.setMessage(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render("No filename specified"))
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
	cursorLine := m.textBuffer.GetCursorLine()
	if text, err := pasteFromClipboard(); err == nil && text != "" {
		m.textBuffer.InsertText(text)
		m.updateModified()
		// Mark lines dirty based on pasted content
		newLines := len(strings.Split(text, "\n")) - 1
		m.markLinesDirty(cursorLine, cursorLine+newLines)
		m.applyIncrementalHighlighting()
		m.setMessage("Pasted from system clipboard")
	} else if m.clipboard != "" {
		m.textBuffer.InsertText(m.clipboard)
		m.updateModified()
		// Mark lines dirty based on pasted content
		newLines := len(strings.Split(m.clipboard, "\n")) - 1
		m.markLinesDirty(cursorLine, cursorLine+newLines)
		m.applyIncrementalHighlighting()
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
		// Re-highlight everything after undo since we don't know what changed
		m.applySyntaxHighlighting()
	} else {
		m.setMessage("Nothing to undo")
	}
	return m, nil
}

func (m Model) handleRedo() (tea.Model, tea.Cmd) {
	if m.textBuffer.Redo() {
		m.updateModified()
		m.ensureCursorVisible()
		// Re-highlight everything after redo since we don't know what changed
		m.applySyntaxHighlighting()
	} else {
		m.setMessage("Nothing to redo")
	}
	return m, nil
}

func (m Model) handleSelectAll() (tea.Model, tea.Cmd) {
	m.textBuffer.SelectAll()
	return m, nil
}

func (m Model) handleFindNext() (tea.Model, tea.Cmd) {
	if len(m.findResults) > 0 {
		if m.findIndex < len(m.findResults)-1 {
			m.findIndex++
		} else {
			m.findIndex = 0
		}
		m.adjustResultsOffset()
		m.jumpToCurrentResult()
		m.setSearchMessage()
	} else if m.lastSearchQuery != "" {
		m.findResults = m.textBuffer.FindText(m.lastSearchQuery, false)
		if len(m.findResults) > 0 {
			m.findIndex = 0
			m.adjustResultsOffset()
			m.jumpToCurrentResult()
			m.setSearchMessage()
		} else {
			m.setMessage("No matches found")
		}
	} else {
		m.setMessage("No search query")
	}
	return m, nil
}

func (m Model) handleFindPrev() (tea.Model, tea.Cmd) {
	if len(m.findResults) > 0 {
		if m.findIndex > 0 {
			m.findIndex--
		} else {
			m.findIndex = len(m.findResults) - 1
		}
		m.adjustResultsOffset()
		m.jumpToCurrentResult()
		m.setSearchMessage()
	} else if m.lastSearchQuery != "" {
		m.findResults = m.textBuffer.FindText(m.lastSearchQuery, false)
		if len(m.findResults) > 0 {
			m.findIndex = 0
			m.adjustResultsOffset()
			m.jumpToCurrentResult()
			m.setSearchMessage()
		} else {
			m.setMessage("No matches found")
		}
	} else {
		m.setMessage("No search query")
	}
	return m, nil
}

func (m *Model) setSearchMessage() {
	if len(m.findResults) > 0 {
		m.setMessage(fmt.Sprintf("Match %d of %d for \"%s\"", m.findIndex+1, len(m.findResults), m.lastSearchQuery))
	}
}
