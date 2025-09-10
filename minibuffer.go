package main

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)


type MinibufferType int

const (
	MinibufferNone MinibufferType = iota
	MinibufferGoToLine
	MinibufferFind
	MinibufferFindResults
)

func (m Model) getMinibufferHeight() int {
	switch m.minibufferType {
	case MinibufferNone:
		return 1
	case MinibufferGoToLine, MinibufferFind:
		return 1
	case MinibufferFindResults:
		resultsCount := len(m.findResults)
		if resultsCount > m.maxResultsDisplay {
			resultsCount = m.maxResultsDisplay
		}
		return 3 + resultsCount
	default:
		return 1
	}
}

func (m Model) handleMinibufferInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		return handleEscapeKey(m)
	case tea.KeyEnter:
		return handleEnterKey(m)
	case tea.KeyUp, tea.KeyDown:
		return handleNavigationKeys(m, msg)
	case tea.KeyBackspace, tea.KeyDelete, tea.KeyLeft, tea.KeyRight, tea.KeyHome, tea.KeyEnd:
		return handleEditingKeys(m, msg)
	case tea.KeyRunes:
		return handleTextInput(m, msg)
	}

	return m, nil
}

func handleEscapeKey(m Model) (tea.Model, tea.Cmd) {
	m.minibufferType = MinibufferNone
	m.minibufferInput = ""
	m.minibufferCursorPos = 0
	m.searchResultsOffset = 0
	m.textBuffer.ClearSelection()
	return m, nil
}

func handleEnterKey(m Model) (tea.Model, tea.Cmd) {
	switch m.minibufferType {
	case MinibufferGoToLine:
		return handleGoToLineEnter(m)
	case MinibufferFind:
		return handleFindEnter(m)
	case MinibufferFindResults:
		return handleFindResultsEnter(m)
	}
	return m, nil
}

func handleGoToLineEnter(m Model) (tea.Model, tea.Cmd) {
	if line, err := strconv.Atoi(m.minibufferInput); err == nil && line > 0 {
		totalLines := len(m.textBuffer.GetLines())
		if line <= totalLines {
			m.textBuffer.GoToLine(line - 1)
			m.ensureCursorVisible()
		} else {
			m.setMessage(fmt.Sprintf("Line %d is beyond end of file (total: %d lines)", line, totalLines))
		}
	} else {
		m.setMessage("Invalid line number")
	}
	m.minibufferType = MinibufferNone
	m.minibufferInput = ""
	m.minibufferCursorPos = 0
	return m, nil
}

func handleFindEnter(m Model) (tea.Model, tea.Cmd) {
	if m.minibufferInput != "" {
		m.lastSearchQuery = m.minibufferInput
		m.findResults = m.textBuffer.FindText(m.minibufferInput, false)
		if len(m.findResults) > 0 {
			m.findIndex = 0
			m.searchResultsOffset = 0
			m.minibufferType = MinibufferFindResults
			m.jumpToCurrentResult()
		} else {
			m.setMessage("No matches found")
			m.minibufferType = MinibufferNone
			m.minibufferInput = ""
			m.minibufferCursorPos = 0
		}
	} else {
		m.minibufferType = MinibufferNone
		m.minibufferInput = ""
		m.minibufferCursorPos = 0
	}
	return m, nil
}

func handleFindResultsEnter(m Model) (tea.Model, tea.Cmd) {
	if len(m.findResults) > 0 && m.findIndex >= 0 {
		m.jumpToCurrentResult()
	}
	m.minibufferType = MinibufferNone
	m.searchResultsOffset = 0
	return m, nil
}

func handleNavigationKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.minibufferType == MinibufferFindResults {
		if msg.Type == tea.KeyUp {
			if m.findIndex > 0 {
				m.findIndex--
			} else {
				m.findIndex = len(m.findResults) - 1
			}
		} else if msg.Type == tea.KeyDown {
			if m.findIndex < len(m.findResults)-1 {
				m.findIndex++
			} else {
				m.findIndex = 0
			}
		}
		m.adjustResultsOffset()
		m.jumpToCurrentResult()
	}
	return m, nil
}

func handleEditingKeys(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.minibufferType != MinibufferFind && m.minibufferType != MinibufferGoToLine {
		return m, nil
	}

	switch msg.Type {
	case tea.KeyBackspace:
		if m.minibufferCursorPos > 0 {
			m.minibufferInput = m.minibufferInput[:m.minibufferCursorPos-1] + m.minibufferInput[m.minibufferCursorPos:]
			m.minibufferCursorPos--
		}
	case tea.KeyDelete:
		if m.minibufferCursorPos < len(m.minibufferInput) {
			m.minibufferInput = m.minibufferInput[:m.minibufferCursorPos] + m.minibufferInput[m.minibufferCursorPos+1:]
		}
	case tea.KeyLeft:
		if m.minibufferCursorPos > 0 {
			m.minibufferCursorPos--
		}
	case tea.KeyRight:
		if m.minibufferCursorPos < len(m.minibufferInput) {
			m.minibufferCursorPos++
		}
	case tea.KeyHome:
		m.minibufferCursorPos = 0
	case tea.KeyEnd:
		m.minibufferCursorPos = len(m.minibufferInput)
	}

	return m, nil
}

func handleTextInput(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if (m.minibufferType == MinibufferFind || m.minibufferType == MinibufferGoToLine) && len(msg.Runes) > 0 {
		char := string(msg.Runes)
		m.minibufferInput = m.minibufferInput[:m.minibufferCursorPos] + char + m.minibufferInput[m.minibufferCursorPos:]
		m.minibufferCursorPos++
	}
	return m, nil
}

func (m Model) renderMinibuffer() string {
	switch m.minibufferType {
	case MinibufferGoToLine:
		return m.renderGoToLineMinibuffer()
	case MinibufferFind:
		return m.renderFindMinibuffer()
	case MinibufferFindResults:
		return m.renderFindResultsMinibuffer()
	}
	return ""
}

func (m Model) renderGoToLineMinibuffer() string {
	prompt := "Go to line: "

	var inputDisplay strings.Builder
	for i, char := range m.minibufferInput {
		if i == m.minibufferCursorPos {
			inputDisplay.WriteString(minibufferCursorStyle.Render(string(char)))
		} else {
			inputDisplay.WriteString(string(char))
		}
	}

	if m.minibufferCursorPos >= len(m.minibufferInput) {
		inputDisplay.WriteString(minibufferCursorStyle.Render(" "))
	}

	content := minibufferPromptStyle.Render(prompt) + minibufferInputStyle.Render(inputDisplay.String())

	minibufferWidth := m.width - 4
	currentLen := len(prompt) + len(m.minibufferInput)
	if currentLen < minibufferWidth {
		padding := strings.Repeat(" ", minibufferWidth-currentLen)
		content += padding
	}

	return minibufferStyle.Width(m.width - 2).Render(content)
}

func (m Model) renderFindMinibuffer() string {
	prompt := "Find: "

	var inputDisplay strings.Builder
	for i, char := range m.minibufferInput {
		if i == m.minibufferCursorPos {
			inputDisplay.WriteString(minibufferCursorStyle.Render(string(char)))
		} else {
			inputDisplay.WriteString(string(char))
		}
	}

	if m.minibufferCursorPos >= len(m.minibufferInput) {
		inputDisplay.WriteString(minibufferCursorStyle.Render(" "))
	}

	content := minibufferPromptStyle.Render(prompt) + minibufferInputStyle.Render(inputDisplay.String())

	minibufferWidth := m.width - 4
	currentLen := len(prompt) + len(m.minibufferInput)
	if currentLen < minibufferWidth {
		padding := strings.Repeat(" ", minibufferWidth-currentLen)
		content += padding
	}

	return minibufferStyle.Width(m.width - 2).Render(content)
}

func (m Model) renderFindResultsMinibuffer() string {
	var lines []string

	header := fmt.Sprintf("Search results for '%s' (%d matches):",
		m.lastSearchQuery, len(m.findResults))
	lines = append(lines, minibufferPromptStyle.Render(header))

	textLines := m.textBuffer.GetLines()
	start := m.searchResultsOffset
	end := start + m.maxResultsDisplay
	if end > len(m.findResults) {
		end = len(m.findResults)
	}

	for i := start; i < end; i++ {
		result := m.findResults[i]
		isSelected := i == m.findIndex

		var linePreview string
		if result.Line < len(textLines) {
			line := textLines[result.Line]

			contextStart := max(0, result.Column-30)
			contextEnd := min(len(line), result.Column+len(m.lastSearchQuery)+30)

			linePreview = line[contextStart:contextEnd]

			if contextStart > 0 {
				linePreview = "..." + linePreview
			}
			if contextEnd < len(line) {
				linePreview = linePreview + "..."
			}

			if len(linePreview) > 70 {
				linePreview = linePreview[:67] + "..."
			}

			linePreview = strings.ReplaceAll(linePreview, "\t", "    ")
		} else {
			linePreview = "<end of file>"
		}

		resultText := fmt.Sprintf("  %4d:%-4d  %s",
			result.Line+1, result.Column+1, linePreview)

		if isSelected {
			resultText = searchResultSelectedStyle.Render(resultText)
		} else {
			resultText = searchResultNormalStyle.Render(resultText)
		}

		lines = append(lines, resultText)
	}

	hint := "Esc: cancel"
	lines = append(lines, helpStyle.Render(hint))

	content := strings.Join(lines, "\n")

	return minibufferStyle.
		Width(m.width - 2).
		Render(content)
}