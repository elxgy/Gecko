package main

import (
	"strings"
	"unicode"
)

type Position struct {
	Line   int
	Column int
}

type Selection struct {
	Start Position
	End   Position
}

type TextBuffer struct {
	lines                   []string
	cursor                  Position
	selection               *Selection
	history                 []TextState
	historyIndex            int
	maxHistory              int
	selectAllOriginalCursor *Position
}

type TextState struct {
	Lines  []string
	Cursor Position
}

func NewTextBuffer(content string) *TextBuffer {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	tb := &TextBuffer{
		lines:      lines,
		cursor:     Position{Line: 0, Column: 0},
		maxHistory: 100,
	}

	tb.saveState()
	return tb
}

func (tb *TextBuffer) GetContent() string {
	return strings.Join(tb.lines, "\n")
}

func (tb *TextBuffer) GetLines() []string {
	return tb.lines
}

func (tb *TextBuffer) GetCursor() Position {
	return tb.cursor
}

func (tb *TextBuffer) GetSelection() *Selection {
	return tb.selection
}

func (tb *TextBuffer) HasSelection() bool {
	return tb.selection != nil
}

func (tb *TextBuffer) SetSelection(selection *Selection) {
	tb.selection = selection
}

func (tb *TextBuffer) ClearSelection() {
	tb.selection = nil
}

func (tb *TextBuffer) SetCursor(pos Position) {
	tb.cursor = tb.clampPosition(pos)
}

func (tb *TextBuffer) restoreSelectAllCursor() {
	if tb.selectAllOriginalCursor != nil {
		tb.cursor = *tb.selectAllOriginalCursor
		tb.selectAllOriginalCursor = nil
	}
}

func (tb *TextBuffer) MoveCursor(deltaLine, deltaColumn int, extend bool) {
	if !extend && tb.selectAllOriginalCursor != nil {
		tb.restoreSelectAllCursor()
		tb.selection = nil
		return
	}

	if tb.selectAllOriginalCursor != nil {
		tb.selectAllOriginalCursor = nil
	}

	if extend && tb.selection == nil {
		tb.selection = &Selection{
			Start: tb.cursor,
			End:   tb.cursor,
		}
	}

	newPos := Position{
		Line:   tb.cursor.Line + deltaLine,
		Column: tb.cursor.Column + deltaColumn,
	}

	tb.cursor = tb.clampPosition(newPos)

	if extend && tb.selection != nil {
		tb.selection.End = tb.cursor
	} else if !extend {
		tb.selection = nil
	}
}

func (tb *TextBuffer) MoveToWordBoundary(forward bool, extend bool) {
	if !extend && tb.selectAllOriginalCursor != nil {
		tb.restoreSelectAllCursor()
		tb.selection = nil
		return
	}

	if tb.selectAllOriginalCursor != nil {
		tb.selectAllOriginalCursor = nil
	}

	if extend && tb.selection == nil {
		tb.selection = &Selection{
			Start: tb.cursor,
			End:   tb.cursor,
		}
	}

	newPos := tb.cursor

	if forward {
		newPos = tb.findNextWordBoundary(newPos)
	} else {
		newPos = tb.findPrevWordBoundary(newPos)
	}

	tb.cursor = newPos

	if extend && tb.selection != nil {
		tb.selection.End = tb.cursor
	} else if !extend {
		tb.selection = nil
	}
}

func (tb *TextBuffer) SelectAll() {
	tb.selectAllOriginalCursor = &Position{
		Line:   tb.cursor.Line,
		Column: tb.cursor.Column,
	}

	tb.selection = &Selection{
		Start: Position{Line: 0, Column: 0},
		End:   Position{Line: len(tb.lines) - 1, Column: len(tb.lines[len(tb.lines)-1])},
	}
	tb.cursor = tb.selection.End
}

func (tb *TextBuffer) SelectLine() {
	if tb.cursor.Line < len(tb.lines) {
		tb.selection = &Selection{
			Start: Position{Line: tb.cursor.Line, Column: 0},
			End:   Position{Line: tb.cursor.Line, Column: len(tb.lines[tb.cursor.Line])},
		}
		tb.cursor = tb.selection.End
	}
}

func (tb *TextBuffer) GetSelectedText() string {
	if tb.selection == nil {
		return ""
	}

	start, end := tb.normalizeSelection()

	if start.Line == end.Line {
		line := tb.lines[start.Line]
		if start.Column >= len(line) {
			return ""
		}
		endCol := end.Column
		if endCol > len(line) {
			endCol = len(line)
		}
		return line[start.Column:endCol]
	}

	var result strings.Builder

	if start.Line < len(tb.lines) {
		line := tb.lines[start.Line]
		if start.Column < len(line) {
			result.WriteString(line[start.Column:])
		}
		result.WriteString("\n")
	}

	for i := start.Line + 1; i < end.Line && i < len(tb.lines); i++ {
		result.WriteString(tb.lines[i])
		result.WriteString("\n")
	}

	if end.Line < len(tb.lines) && end.Line > start.Line {
		line := tb.lines[end.Line]
		endCol := end.Column
		if endCol > len(line) {
			endCol = len(line)
		}
		result.WriteString(line[:endCol])
	}

	return result.String()
}

func (tb *TextBuffer) InsertText(text string) {
	tb.saveState()

	tb.selectAllOriginalCursor = nil

	if tb.selection != nil {
		tb.deleteSelection()
	}

	lines := strings.Split(text, "\n")
	currentLine := tb.lines[tb.cursor.Line]

	if len(lines) == 1 {
		before := currentLine[:tb.cursor.Column]
		after := currentLine[tb.cursor.Column:]
		tb.lines[tb.cursor.Line] = before + text + after
		tb.cursor.Column += len(text)
	} else {
		before := currentLine[:tb.cursor.Column]
		after := currentLine[tb.cursor.Column:]

		tb.lines[tb.cursor.Line] = before + lines[0]

		newLines := make([]string, len(tb.lines)+len(lines)-1)
		copy(newLines, tb.lines[:tb.cursor.Line+1])
		copy(newLines[tb.cursor.Line+1:], lines[1:])
		copy(newLines[tb.cursor.Line+len(lines):], tb.lines[tb.cursor.Line+1:])

		tb.lines = newLines

		lastLineIdx := tb.cursor.Line + len(lines) - 1
		tb.lines[lastLineIdx] += after

		tb.cursor.Line = lastLineIdx
		tb.cursor.Column = len(lines[len(lines)-1])
	}

	tb.selection = nil
}

func (tb *TextBuffer) DeleteSelection() bool {
	if tb.selection == nil {
		return false
	}

	tb.saveState()
	tb.selectAllOriginalCursor = nil
	tb.deleteSelection()
	return true
}

func (tb *TextBuffer) DeleteChar(backward bool) {
	if tb.selection != nil {
		tb.DeleteSelection()
		tb.selection = nil
		return
	}

	tb.saveState()

	tb.selectAllOriginalCursor = nil

	if backward {
		if tb.cursor.Column > 0 {

			line := tb.lines[tb.cursor.Line]
			tb.lines[tb.cursor.Line] = line[:tb.cursor.Column-1] + line[tb.cursor.Column:]
			tb.cursor.Column--

		} else if tb.cursor.Line > 0 {

			prevLine := tb.lines[tb.cursor.Line-1]
			currentLine := tb.lines[tb.cursor.Line]
			tb.lines[tb.cursor.Line-1] = prevLine + currentLine
			tb.lines = append(tb.lines[:tb.cursor.Line], tb.lines[tb.cursor.Line+1:]...)
			tb.cursor.Line--
			tb.cursor.Column = len(prevLine)

		}
	} else {
		if tb.cursor.Column < len(tb.lines[tb.cursor.Line]) {

			line := tb.lines[tb.cursor.Line]
			tb.lines[tb.cursor.Line] = line[:tb.cursor.Column] + line[tb.cursor.Column+1:]

		} else if tb.cursor.Line < len(tb.lines)-1 {

			currentLine := tb.lines[tb.cursor.Line]
			nextLine := tb.lines[tb.cursor.Line+1]
			tb.lines[tb.cursor.Line] = currentLine + nextLine
			tb.lines = append(tb.lines[:tb.cursor.Line+1], tb.lines[tb.cursor.Line+2:]...)

		}
	}
}

func (tb *TextBuffer) Undo() bool {
	if tb.historyIndex > 0 {
		tb.historyIndex--
		state := tb.history[tb.historyIndex]
		tb.lines = make([]string, len(state.Lines))
		copy(tb.lines, state.Lines)
		tb.cursor = state.Cursor
		tb.selection = nil
		tb.selectAllOriginalCursor = nil
		return true
	}
	return false
}

func (tb *TextBuffer) Redo() bool {
	if tb.historyIndex < len(tb.history)-1 {
		tb.historyIndex++
		state := tb.history[tb.historyIndex]
		tb.lines = make([]string, len(state.Lines))
		copy(tb.lines, state.Lines)
		tb.cursor = state.Cursor
		tb.selection = nil
		tb.selectAllOriginalCursor = nil
		return true
	}
	return false
}

func (tb *TextBuffer) GoToLine(line int) {
	tb.cursor = Position{
		Line:   tb.clampLine(line),
		Column: 0,
	}
	tb.selection = nil
	tb.selectAllOriginalCursor = nil
}

func (tb *TextBuffer) FindText(query string, caseSensitive bool) []Position {
	var positions []Position

	searchQuery := query
	if !caseSensitive {
		searchQuery = strings.ToLower(query)
	}

	for lineIdx, line := range tb.lines {
		searchLine := line
		if !caseSensitive {
			searchLine = strings.ToLower(line)
		}

		col := 0
		for {
			idx := strings.Index(searchLine[col:], searchQuery)
			if idx == -1 {
				break
			}
			positions = append(positions, Position{
				Line:   lineIdx,
				Column: col + idx,
			})
			col += idx + 1
		}
	}

	return positions
}

func (tb *TextBuffer) clampPosition(pos Position) Position {
	if pos.Line < 0 {
		pos.Line = 0
	} else if pos.Line >= len(tb.lines) {
		pos.Line = len(tb.lines) - 1
	}

	if line := tb.lines[pos.Line]; pos.Column > len(line) {
		pos.Column = len(line)
	} else if pos.Column < 0 {
		pos.Column = 0
	}

	return pos
}

func (tb *TextBuffer) clampLine(line int) int {
	if line < 0 {
		return 0
	}
	if line >= len(tb.lines) {
		return len(tb.lines) - 1
	}
	return line
}

func (tb *TextBuffer) normalizeSelection() (Position, Position) {
	if tb.selection == nil {
		return tb.cursor, tb.cursor
	}

	start, end := tb.selection.Start, tb.selection.End

	if start.Line > end.Line || (start.Line == end.Line && start.Column > end.Column) {
		start, end = end, start
	}

	return start, end
}

func (tb *TextBuffer) deleteSelection() {
	if tb.selection == nil {
		return
	}

	start, end := tb.normalizeSelection()

	if start.Line == end.Line {

		line := tb.lines[start.Line]
		tb.lines[start.Line] = line[:start.Column] + line[end.Column:]
		tb.cursor = start

	} else {

		startLine := tb.lines[start.Line][:start.Column]
		endLine := tb.lines[end.Line][end.Column:]

		newLines := make([]string, len(tb.lines)-(end.Line-start.Line))
		copy(newLines, tb.lines[:start.Line])
		newLines[start.Line] = startLine + endLine
		copy(newLines[start.Line+1:], tb.lines[end.Line+1:])

		tb.lines = newLines
		tb.cursor = start

	}

	tb.selection = nil
}

func (tb *TextBuffer) findNextWordBoundary(pos Position) Position {
	if pos.Line >= len(tb.lines) {
		return pos
	}

	line := tb.lines[pos.Line]
	col := pos.Column

	for col < len(line) && !unicode.IsSpace(rune(line[col])) {
		col++
	}

	for col < len(line) && unicode.IsSpace(rune(line[col])) {
		col++
	}

	if col >= len(line) && pos.Line < len(tb.lines)-1 {
		return Position{Line: pos.Line + 1, Column: 0}
	}

	return Position{Line: pos.Line, Column: col}
}

func (tb *TextBuffer) findPrevWordBoundary(pos Position) Position {
	if pos.Line < 0 {
		return pos
	}

	line := tb.lines[pos.Line]
	col := pos.Column

	if col > len(line) {
		col = len(line)
	}

	if col > 0 {
		col--

		for col > 0 && unicode.IsSpace(rune(line[col])) {
			col--
		}

		for col > 0 && !unicode.IsSpace(rune(line[col-1])) {
			col--
		}
	} else if pos.Line > 0 {
		return Position{Line: pos.Line - 1, Column: len(tb.lines[pos.Line-1])}
	}

	return Position{Line: pos.Line, Column: col}
}

func (tb *TextBuffer) GetWordBoundsAtCursor() (int, int) {
    if tb.cursor.Line >= len(tb.lines) {
        return -1, -1
    }
    plainLine := tb.lines[tb.cursor.Line]
    col := tb.cursor.Column
    if col >= len(plainLine) || unicode.IsSpace(rune(plainLine[col])) {
        return -1, -1
    }
    start := col
    for start > 0 && !unicode.IsSpace(rune(plainLine[start-1])) {
        start--
    }
    end := col + 1
    for end < len(plainLine) && !unicode.IsSpace(rune(plainLine[end])) {
        end++
    }
    return start, end
}

func (tb *TextBuffer) saveState() {
	if tb.historyIndex < len(tb.history)-1 {
		tb.history = tb.history[:tb.historyIndex+1]
	}

	state := TextState{
		Lines:  make([]string, len(tb.lines)),
		Cursor: tb.cursor,
	}
	copy(state.Lines, tb.lines)

	tb.history = append(tb.history, state)
	tb.historyIndex = len(tb.history) - 1

	if len(tb.history) > tb.maxHistory {
		tb.history = tb.history[1:]
		tb.historyIndex--
	}
}