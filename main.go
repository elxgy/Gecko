package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SearchMatch struct {
	StartPos int
	EndPos   int
	Line     int
	Col      int
	LineText string
}

type Editor struct {
	app           *tview.Application
	textArea      *tview.TextArea
	statusBar     *tview.TextView
	mainLayout    *tview.Flex
	searchLayout  *tview.Flex
	rootLayout    *tview.Flex
	filename      string
	modified      bool
	lastLine      int
	lastCol       int
	searchBar     *tview.InputField
	searchResults *tview.List
	searchMatches []SearchMatch
	currentMatch  int
	searchActive  bool
	savedCursor   struct {
		row, col int
	}
}

func NewEditor() *Editor {
	app := tview.NewApplication()

	textArea := tview.NewTextArea().
		SetPlaceholder("Start typing... Press Ctrl+G for help").
		SetWrap(false)

	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[white]No file | Press Ctrl+G for help")

	searchBar := tview.NewInputField().
		SetLabel("Search: ").
		SetFieldBackgroundColor(tcell.ColorBlack)

	searchResults := tview.NewList()
	searchResults.SetTitle(" Search Results ").
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textArea, 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	searchLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(searchBar, 1, 0, false).
		AddItem(searchResults, 0, 1, false)

	rootLayout := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(mainLayout, 0, 2, true).
		AddItem(searchLayout, 0, 1, false)

	editor := &Editor{
		app:           app,
		textArea:      textArea,
		statusBar:     statusBar,
		mainLayout:    mainLayout,
		searchLayout:  searchLayout,
		rootLayout:    rootLayout,
		filename:      "",
		modified:      false,
		lastLine:      -1,
		lastCol:       -1,
		searchBar:     searchBar,
		searchResults: searchResults,
	}

	app.SetRoot(rootLayout, true)
	editor.setupKeyBindings()
	editor.setupCursorTracking()
	editor.setupSearchComponents()

	rootLayout.ResizeItem(searchLayout, 0, 0)

	return editor
}

func (e *Editor) setupKeyBindings() {
	e.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if e.searchActive {
			return event
		}
		return e.handleKeyPress(event)
	})

	e.textArea.SetChangedFunc(func() {
		e.modified = true
		e.updateStatusBar()
	})
}

func (e *Editor) setupCursorTracking() {
	e.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		line, col, _, _ := e.textArea.GetCursor()

		if line != e.lastLine || col != e.lastCol {
			e.lastLine = line
			e.lastCol = col
			e.updateStatusBar()
		}
		return false
	})
}

func (e *Editor) setupSearchComponents() {
	e.setupSearchBar()
	e.setupSearchResults()
}

func (e *Editor) setupSearchBar() {
	e.searchBar.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			e.cancelSearch()
			return nil
		case tcell.KeyEnter:
			e.acceptSearch()
			return nil
		case tcell.KeyUp:
			e.navigateMatches(true)
			return nil
		case tcell.KeyDown:
			e.navigateMatches(false)
			return nil
		case tcell.KeyCtrlG:
			e.cancelSearch()
			return nil
		case tcell.KeyTab:
			e.app.SetFocus(e.searchResults)
			return nil
		}
		return event
	})

	e.searchBar.SetChangedFunc(func(query string) {
		e.updateSearch(query)
	})
}

func (e *Editor) setupSearchResults() {
	e.searchResults.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			e.cancelSearch()
			return nil
		case tcell.KeyEnter:
			currentItem := e.searchResults.GetCurrentItem()
			if currentItem >= 0 && currentItem < len(e.searchMatches) {
				e.jumpToMatch(currentItem)
				e.acceptSearch()
			}
			return nil
		case tcell.KeyTab:
			e.app.SetFocus(e.searchBar)
			return nil
		case tcell.KeyUp, tcell.KeyDown:
			return event
		}
		return event
	})

	e.searchResults.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(e.searchMatches) {
			e.previewMatch(index)
		}
	})

	e.searchResults.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(e.searchMatches) {
			e.jumpToMatch(index)
			e.acceptSearch()
		}
	})
}

func (e *Editor) startSearch() {
	e.savedCursor.row, e.savedCursor.col, _, _ = e.textArea.GetCursor()

	e.rootLayout.ResizeItem(e.searchLayout, 40, 0)

	e.searchBar.SetText("")
	e.app.SetFocus(e.searchBar)
	e.searchActive = true
	e.searchMatches = nil
	e.currentMatch = -1
	e.searchResults.Clear()

}

func (e *Editor) cancelSearch() {
	e.rootLayout.ResizeItem(e.searchLayout, 0, 0)

	e.app.SetFocus(e.textArea)
	e.searchActive = false
	e.clearHighlights()

	currentLine, _, _, _ := e.textArea.GetCursor()
	if abs(currentLine-e.savedCursor.row) > 5 {
		e.textArea.SetOffset(e.savedCursor.row, e.savedCursor.col)
	}

	e.searchResults.Clear()
}

func (e *Editor) acceptSearch() {
	e.rootLayout.ResizeItem(e.searchLayout, 0, 0)

	e.app.SetFocus(e.textArea)
	e.searchActive = false
}

func (e *Editor) updateSearch(query string) {
	if query == "" {
		e.clearHighlights()
		e.searchMatches = nil
		e.currentMatch = -1
		e.searchBar.SetLabel("Search: ")
		e.searchResults.Clear()
		return
	}

	fullText := e.textArea.GetText()
	e.searchMatches = e.findAllMatches(fullText, query)

	if len(e.searchMatches) == 0 {
		e.searchBar.SetLabel("Search (0 matches): ")
		e.clearHighlights()
		e.currentMatch = -1
		e.searchResults.Clear()
		return
	}

	e.updateSearchResultsList(query)

	currentLine, _, _, _ := e.textArea.GetCursor()
	closestMatch := e.findClosestMatch(currentLine)

	e.currentMatch = closestMatch
	e.searchResults.SetCurrentItem(closestMatch)

	e.setHighlightsOnly(closestMatch)

	e.searchBar.SetLabel(fmt.Sprintf("Search (%d matches): ", len(e.searchMatches)))
}

func (e *Editor) findAllMatches(text, pattern string) []SearchMatch {
	var matches []SearchMatch
	lines := strings.Split(text, "\n")
	start := 0
	patternLen := len(pattern)

	if patternLen == 0 {
		return matches
	}

	for lineNum, line := range lines {
		lineStart := start
		for {
			pos := strings.Index(text[start:], pattern)
			if pos == -1 {
				break
			}

			absStart := start + pos
			absEnd := absStart + patternLen

			if absStart >= lineStart && absStart < lineStart+len(line)+1 {
				col := absStart - lineStart

				contextStart := 0
				contextEnd := len(line)
				if contextEnd > 80 {
					contextStart = max(0, col-20)
					contextEnd = min(len(line), col+60)
				}

				lineText := line[contextStart:contextEnd]
				if contextStart > 0 {
					lineText = "..." + lineText
				}
				if contextEnd < len(line) {
					lineText = lineText + "..."
				}

				matches = append(matches, SearchMatch{
					StartPos: absStart,
					EndPos:   absEnd,
					Line:     lineNum + 1,
					Col:      col + 1,
					LineText: lineText,
				})
			}

			start = absStart + 1
		}

		start = lineStart + len(line) + 1
	}
	return matches
}

func (e *Editor) updateSearchResultsList(query string) {
	e.searchResults.Clear()

	for _, match := range e.searchMatches {
		mainText := fmt.Sprintf("Line %d:%d", match.Line, match.Col)
		secondaryText := match.LineText
		e.searchResults.AddItem(mainText, secondaryText, 0, nil)
	}
}

func (e *Editor) previewMatch(index int) {
	if index < 0 || index >= len(e.searchMatches) {
		return
	}

	match := e.searchMatches[index]
	currentLine, _, _, _ := e.textArea.GetCursor()

	distance := abs(match.Line - 1 - currentLine)

	if distance > 20 {
		e.smartScrollToLine(match.Line - 1)
	}

	e.setHighlights(index)
	e.currentMatch = index
}

//Dont uncomment!!!!!, its shit but may be useful'

// func (e *Editor) gentleScrollToLine(line int) {
// 	text := e.textArea.GetText()
// 	lines := strings.Split(text, "\n")
// 	totalLines := len(lines)

// 	if line < 0 {
// 		line = 0
// 	}
// 	if line >= totalLines {
// 		line = totalLines - 1
// 	}

// 	currentLine, currentCol, _, _ := e.textArea.GetCursor()

// 	if abs(currentLine-line) > 10 {
// 		preservedCol := currentCol
// 		if preservedCol > len(lines[line]) {
// 			preservedCol = len(lines[line])
// 		}

// 		e.textArea.SetOffset(line, preservedCol)
// 	}
// }

func (e *Editor) jumpToMatch(index int) {
	if index < 0 || index >= len(e.searchMatches) {
		return
	}

	match := e.searchMatches[index]

	e.smartScrollToLine(match.Line - 1)

	e.textArea.SetOffset(match.Line-1, match.Col-1)
	e.textArea.Select(match.StartPos, match.EndPos)

	e.currentMatch = index
}

func (e *Editor) smartScrollToLine(line int) {
	text := e.textArea.GetText()
	lines := strings.Split(text, "\n")
	totalLines := len(lines)

	if line < 0 {
		line = 0
	}
	if line >= totalLines {
		line = totalLines - 1
	}

	_, _, _, height := e.textArea.GetInnerRect()
	visibleLines := height
	if visibleLines <= 0 {
		visibleLines = 20
	}

	contextOffset := visibleLines / 4
	scrollTarget := line - contextOffset

	if scrollTarget < 0 {
		scrollTarget = 0
	}

	maxScroll := totalLines - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scrollTarget > maxScroll {
		scrollTarget = maxScroll
	}

	e.textArea.SetOffset(scrollTarget, 0)

	go func() {
		time.Sleep(5 * time.Millisecond)
		e.app.QueueUpdateDraw(func() {
			e.textArea.SetOffset(line, 0)
		})
	}()
}

func (e *Editor) navigateMatches(prev bool) {
	if len(e.searchMatches) == 0 {
		return
	}

	if prev {
		e.currentMatch--
		if e.currentMatch < 0 {
			e.currentMatch = len(e.searchMatches) - 1
		}
	} else {
		e.currentMatch++
		if e.currentMatch >= len(e.searchMatches) {
			e.currentMatch = 0
		}
	}

	e.previewMatch(e.currentMatch)
	e.searchResults.SetCurrentItem(e.currentMatch)
	e.searchBar.SetLabel(fmt.Sprintf("Search (%d/%d): ", e.currentMatch+1, len(e.searchMatches)))
}

func (e *Editor) absoluteToRowCol(absPos int) (int, int) {
	text := e.textArea.GetText()
	if absPos > len(text) {
		absPos = len(text)
	}

	row := 0
	col := 0

	for i := 0; i < absPos && i < len(text); i++ {
		if text[i] == '\n' {
			row++
			col = 0
		} else {
			col++
		}
	}

	return row, col
}

func (e *Editor) calculateAbsolutePosition(row, col int) int {
	text := e.textArea.GetText()
	pos := 0
	currentRow := 0

	for i, char := range text {
		if currentRow == row {
			if col <= 0 {
				return i
			}
			col--
		}
		if char == '\n' {
			currentRow++
		}
		pos = i + 1
	}

	return pos
}

func (e *Editor) setHighlights(matchIndex int) {
	if matchIndex < 0 || matchIndex >= len(e.searchMatches) {
		return
	}

	match := e.searchMatches[matchIndex]
	e.textArea.Select(match.StartPos, match.EndPos)
}

func (e *Editor) clearHighlights() {
	e.textArea.Select(-1, -1)
}

func (e *Editor) findClosestMatch(currentLine int) int {
	if len(e.searchMatches) == 0 {
		return -1
	}

	closestIndex := 0
	minDistance := abs(e.searchMatches[0].Line - 1 - currentLine)

	for i, match := range e.searchMatches {
		distance := abs(match.Line - 1 - currentLine)
		if distance < minDistance {
			minDistance = distance
			closestIndex = i
		}
	}

	return closestIndex
}

func (e *Editor) setHighlightsOnly(matchIndex int) {
	if matchIndex < 0 || matchIndex >= len(e.searchMatches) {
		return
	}

	match := e.searchMatches[matchIndex]
	currentLine, _, _, _ := e.textArea.GetCursor()

	if abs(match.Line-1-currentLine) <= 50 {
		e.textArea.Select(match.StartPos, match.EndPos)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (e *Editor) handleKeyPress(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlS:
		e.saveFile()
		return nil
	case tcell.KeyCtrlO:
		e.showOpenDialog()
		return nil
	case tcell.KeyCtrlQ:
		if e.modified {
			e.showQuitDialog()
		} else {
			e.app.Stop()
		}
		return nil
	case tcell.KeyCtrlG:
		e.showHelp()
		return nil
	case tcell.KeyCtrlF:
		e.startSearch()
		return nil
	case tcell.KeyCtrlA:
		e.textArea.Select(0, len(e.textArea.GetText()))
		return nil
	case tcell.KeyCtrlC, tcell.KeyCtrlX, tcell.KeyCtrlV:
		return event
	}

	return event
}

func (e *Editor) saveFile() {
	if e.filename == "" {
		e.showSaveAsDialog()
		return
	}

	content := e.textArea.GetText()
	err := os.WriteFile(e.filename, []byte(content), 0644)
	if err != nil {
		e.showError(fmt.Sprintf("Error saving file: %v", err))
		return
	}

	e.modified = false
	e.updateStatusBar()
	e.showMessage("File saved successfully")
}

func (e *Editor) showSaveAsDialog() {
	e.filename = "untitled.txt"
	e.saveFile()
}

func (e *Editor) showOpenDialog() {
	e.showMessage("Open dialog - TODO: Implement file browser")
}

func (e *Editor) showFindDialog() {
	e.startSearch()
}

func (e *Editor) showQuitDialog() {
	modal := tview.NewModal().
		SetText("File has been modified. Save before quitting?").
		AddButtons([]string{"Save & Quit", "Quit without saving", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonIndex {
			case 0:
				e.saveFile()
				e.app.Stop()
			case 1:
				e.app.Stop()
			case 2:
				e.app.SetRoot(e.rootLayout, true)
			}
		})

	e.app.SetRoot(modal, true)
}

func (e *Editor) showMessage(message string) {
	e.statusBar.SetText(fmt.Sprintf("[green]%s[white]", message))
	go func() {
		time.Sleep(2 * time.Second)
		e.updateStatusBar()
	}()
}

func (e *Editor) showError(message string) {
	e.statusBar.SetText(fmt.Sprintf("[red]%s[white]", message))
	go func() {
		time.Sleep(3 * time.Second)
		e.updateStatusBar()
	}()
}

func (e *Editor) updateStatusBar() {
	modifiedIndicator := ""
	if e.modified {
		modifiedIndicator = " [red]●[white]"
	}

	filename := e.filename
	if filename == "" {
		filename = "No file"
	}

	text := e.textArea.GetText()
	lines := len(strings.Split(text, "\n"))

	cursorLine := e.lastLine + 1
	cursorCol := e.lastCol + 1

	status := fmt.Sprintf("%s%s | %d:%d | %d lines | Ctrl+G for help",
		filename,
		modifiedIndicator,
		cursorLine,
		cursorCol,
		lines)

	e.statusBar.SetText(status)
}

func (e *Editor) showHelp() {
	helpText := `
Gecko Editor

Keyboard Shortcuts:
  Ctrl+S - Save file
  Ctrl+O - Open file
  Ctrl+Q - Quit
  Ctrl+G - Show this help
  Ctrl+F - Find text (opens search panel)
  Ctrl+A - Select all
  Ctrl+C - Copy
  Ctrl+X - Cut  
  Ctrl+V - Paste

Navigation:
  Arrow Keys - Move cursor
  Home/End - Start/End of line
  Ctrl+Home - Start of file
  Ctrl+End - End of file
  Page Up/Down - Scroll page

Search Panel (Ctrl+F):
  Type to search in real-time
  Tab - Switch between search box and results
  Enter - Jump to selected result and close panel
  Esc - Cancel search and close panel
  Arrow Keys - Navigate results list (auto-scroll preview)
  Up/Down in search box - Navigate matches

Press OK to continue...
`

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			e.app.SetRoot(e.rootLayout, true)
		})

	e.app.SetRoot(modal, true)
	e.app.SetFocus(modal)
}

func (e *Editor) LoadFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	e.filename = filename
	e.textArea.SetText(string(content), false)
	e.modified = false
	e.lastLine = 0
	e.lastCol = 0
	e.updateStatusBar()

	return nil
}

func (e *Editor) SaveFile(filename string) error {
	content := e.textArea.GetText()
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return err
	}

	e.filename = filename
	e.modified = false
	e.updateStatusBar()

	return nil
}

func (e *Editor) Run() error {
	return e.app.Run()
}

func main() {
	editor := NewEditor()

	if len(os.Args) > 1 {
		filename := os.Args[1]
		if err := editor.LoadFile(filename); err != nil {
			log.Printf("Could not load file %s: %v", filename, err)
		}
	}

	if err := editor.Run(); err != nil {
		log.Fatal(err)
	}
}
