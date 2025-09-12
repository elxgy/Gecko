package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleSave() (tea.Model, tea.Cmd) {
	if m.filename != "" {
		err := m.saveFile()
		if err == nil {
			m.modified = false
			m.originalText = m.textBuffer.GetContent()
			m.lastSaved = time.Now()
			m.setMessage(flashSuccessStyle.Render("File saved successfully"))
		} else {
			m.setMessage(flashErrorStyle.Render(fmt.Sprintf("Error saving file: %v", err)))
		}
	} else {
		m.setMessage(flashWarningStyle.Render("No filename specified"))
	}
	return m, nil
}

func (m Model) handleCopy() (tea.Model, tea.Cmd) {
	if !m.textBuffer.HasSelection() {
		m.setMessage("No text selected")
		return m, nil
	}

	text := m.textBuffer.GetSelectedText()
	m.performClipboardOperation(text, "Copied", false)
	return m, nil
}

func (m Model) handleCut() (tea.Model, tea.Cmd) {
	if !m.textBuffer.HasSelection() {
		m.setMessage("No text selected")
		return m, nil
	}

	text := m.textBuffer.GetSelectedText()
	m.textBuffer.DeleteSelection()
	m.updateModified()
	m.performClipboardOperation(text, "Cut", false)
	return m, nil
}

// performClipboardOperation handles the common clipboard logic for copy/cut operations
func (m *Model) performClipboardOperation(text, operation string, isInternalOnly bool) {
	m.clipboard = text

	if isInternalOnly {
		m.setMessage(fmt.Sprintf("%s to internal clipboard", operation))
		return
	}

	if err := copyToClipboard(text); err != nil {
		m.setMessage(fmt.Sprintf("%s to internal clipboard", operation))
	} else {
		m.setMessage(fmt.Sprintf("%s to system clipboard", operation))
	}
}

func (m Model) handlePaste() (tea.Model, tea.Cmd) {
	if text, err := pasteFromClipboard(); err == nil && text != "" {
		if err := m.textBuffer.InsertText(text); err != nil {
			m.setMessage("Error pasting text")
			return m, nil
		}
		m.updateModified()
		m.setMessage("Pasted from system clipboard")
	} else if m.clipboard != "" {
		if err := m.textBuffer.InsertText(m.clipboard); err != nil {
			m.setMessage("Error pasting text")
			return m, nil
		}
		m.updateModified()
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
	} else {
		m.setMessage("Nothing to undo")
	}
	return m, nil
}

func (m Model) handleRedo() (tea.Model, tea.Cmd) {
	if m.textBuffer.Redo() {
		m.updateModified()
		m.ensureCursorVisible()
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
	return m.performFindOperation(true), nil
}

func (m Model) handleFindPrev() (tea.Model, tea.Cmd) {
	return m.performFindOperation(false), nil
}

// performFindOperation handles the common logic for find next/previous operations
func (m Model) performFindOperation(isNext bool) Model {
	if len(m.findResults) > 0 {
		m.navigateToResult(isNext)
		m.updateSearchDisplay()
		return m
	}

	if m.lastSearchQuery == "" {
		m.setMessage("No search query")
		return m
	}

	return m.performNewSearch()
}

// navigateToResult moves to the next or previous search result
func (m *Model) navigateToResult(isNext bool) {
	if isNext {
		if m.findIndex < len(m.findResults)-1 {
			m.findIndex++
		} else {
			m.findIndex = 0
		}
	} else {
		if m.findIndex > 0 {
			m.findIndex--
		} else {
			m.findIndex = len(m.findResults) - 1
		}
	}
}

// performNewSearch executes a new search with the last query
func (m Model) performNewSearch() Model {
	m.findResults = m.textBuffer.FindText(m.lastSearchQuery, false)
	if len(m.findResults) > 0 {
		m.findIndex = 0
		m.updateSearchDisplay()
	} else {
		m.setMessage("No matches found")
	}
	return m
}

// updateSearchDisplay updates the display after a successful search navigation
func (m *Model) updateSearchDisplay() {
	m.adjustResultsOffset()
	m.jumpToCurrentResult()
	m.setSearchMessage()
}

func (m *Model) setSearchMessage() {
	if len(m.findResults) > 0 {
		m.setMessage(fmt.Sprintf("Match %d of %d for \"%s\"", m.findIndex+1, len(m.findResults), m.lastSearchQuery))
	}
}
