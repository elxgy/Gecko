package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode"
)

// Custom error types for better error handling
var (
	ErrInvalidPosition = errors.New("invalid position")
	ErrInvalidRange    = errors.New("invalid range")
	ErrEmptyBuffer     = errors.New("buffer is empty")
	ErrHistoryEmpty    = errors.New("history is empty")
)

type Position struct {
	Line   int
	Column int
}

// IsValid checks if the position is valid for the given buffer
func (p Position) IsValid(lineCount int, getLineLength func(int) int) bool {
	if p.Line < 0 || p.Line >= lineCount {
		return false
	}
	if p.Column < 0 || p.Column > getLineLength(p.Line) {
		return false
	}
	return true
}

type Selection struct {
	Start Position
	End   Position
}

// IsValid checks if the selection is valid
func (s *Selection) IsValid(lineCount int, getLineLength func(int) int) bool {
	if s == nil {
		return false
	}
	return s.Start.IsValid(lineCount, getLineLength) && s.End.IsValid(lineCount, getLineLength)
}

// Normalize ensures start comes before end
func (s *Selection) Normalize() (Position, Position) {
	start, end := s.Start, s.End
	if start.Line > end.Line || (start.Line == end.Line && start.Column > end.Column) {
		start, end = end, start
	}
	return start, end
}

// GapBuffer represents a gap buffer for efficient text editing
type GapBuffer struct {
	buffer   []rune
	gapStart int
	gapEnd   int
}

// NewGapBuffer creates a new gap buffer with initial content
func NewGapBuffer(content string) *GapBuffer {
	runes := []rune(content)
	gapSize := max(len(runes), 256) // Initial gap size
	buffer := make([]rune, len(runes)+gapSize)
	copy(buffer, runes)

	return &GapBuffer{
		buffer:   buffer,
		gapStart: len(runes),
		gapEnd:   len(buffer),
	}
}

// moveGapTo moves the gap to the specified position
func (gb *GapBuffer) moveGapTo(pos int) {
	if pos < gb.gapStart {
		// Move gap left: copy data from [pos:gapStart] to the end of the gap
		dist := gb.gapStart - pos
		copy(gb.buffer[gb.gapEnd-dist:], gb.buffer[pos:gb.gapStart])
		gb.gapStart = pos
		gb.gapEnd -= dist
	} else if pos > gb.gapStart {
		// Move gap right: copy data from [gapEnd:gapEnd+dist] to [gapStart:gapStart+dist]
		dist := pos - gb.gapStart
		copy(gb.buffer[gb.gapStart:], gb.buffer[gb.gapEnd:gb.gapEnd+dist])
		gb.gapStart += dist
		gb.gapEnd += dist
	}
}

// Insert inserts text at the specified position
func (gb *GapBuffer) Insert(pos int, text string) {
	runes := []rune(text)
	gb.moveGapTo(pos)

	// Expand gap if necessary
	if len(runes) > gb.gapEnd-gb.gapStart {
		gb.expandGap(len(runes))
	}

	copy(gb.buffer[gb.gapStart:], runes)
	gb.gapStart += len(runes)
}

// Delete deletes text from start to end position (end is exclusive)
func (gb *GapBuffer) Delete(start, end int) {
	if start > end {
		start, end = end, start
	}

	// Bounds checking
	if start < 0 {
		start = 0
	}
	if end > gb.Len() {
		end = gb.Len()
	}
	if start >= end {
		return // Nothing to delete
	}

	// Move gap to start position
	gb.moveGapTo(start)

	// Expand the gap to include the deleted range
	gb.gapEnd += end - start
}

// expandGap expands the gap to accommodate more text
func (gb *GapBuffer) expandGap(minSize int) {
	newGapSize := max(minSize*2, 256)
	newBuffer := make([]rune, len(gb.buffer)+newGapSize)

	// Copy text before gap
	copy(newBuffer, gb.buffer[:gb.gapStart])

	// Copy text after gap
	copy(newBuffer[gb.gapStart+newGapSize:], gb.buffer[gb.gapEnd:])

	gb.buffer = newBuffer
	gb.gapEnd = gb.gapStart + newGapSize
}

// String returns the content as a string
func (gb *GapBuffer) String() string {
	result := make([]rune, 0, len(gb.buffer)-(gb.gapEnd-gb.gapStart))
	result = append(result, gb.buffer[:gb.gapStart]...)
	result = append(result, gb.buffer[gb.gapEnd:]...)
	return string(result)
}

// Len returns the length of the content (excluding gap)
func (gb *GapBuffer) Len() int {
	return len(gb.buffer) - (gb.gapEnd - gb.gapStart)
}

type TextBuffer struct {
	mu                      sync.RWMutex
	lines                   []string
	cursor                  Position
	selection               *Selection
	history                 []TextState
	historyIndex            int
	maxHistory              int
	selectAllOriginalCursor *Position
	// Performance optimization: cache frequently accessed data
	lastLineCount   int
	lastContentHash uint64
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
		lines:         lines,
		cursor:        Position{Line: 0, Column: 0},
		maxHistory:    100,
		lastLineCount: len(lines),
	}

	// Calculate content hash after initialization
	tb.lastContentHash = tb.calculateContentHash(lines)
	tb.saveState()
	return tb
}

// calculateContentHash calculates a simple hash of the content for change detection
func (tb *TextBuffer) calculateContentHash(lines []string) uint64 {
	// For large files, only hash a subset to avoid performance issues
	if len(lines) > 1000 {
		// Hash first 100, middle 100, and last 100 lines for large files
		return tb.calculatePartialHash(lines)
	}

	var hash uint64 = 5381
	for _, line := range lines {
		for _, char := range line {
			hash = ((hash << 5) + hash) + uint64(char)
		}
		hash = ((hash << 5) + hash) + uint64('\n') // Add newline to hash
	}
	return hash
}

// calculatePartialHash computes hash for large files using sampling
func (tb *TextBuffer) calculatePartialHash(lines []string) uint64 {
	var hash uint64 = 5381
	totalLines := len(lines)

	// Hash first 100 lines
	for i := 0; i < min(100, totalLines); i++ {
		for _, char := range lines[i] {
			hash = ((hash << 5) + hash) + uint64(char)
		}
		hash = ((hash << 5) + hash) + uint64('\n')
	}

	// Hash middle 100 lines
	midStart := max(100, totalLines/2-50)
	midEnd := min(totalLines, midStart+100)
	for i := midStart; i < midEnd; i++ {
		for _, char := range lines[i] {
			hash = ((hash << 5) + hash) + uint64(char)
		}
		hash = ((hash << 5) + hash) + uint64('\n')
	}

	// Hash last 100 lines
	lastStart := max(midEnd, totalLines-100)
	for i := lastStart; i < totalLines; i++ {
		for _, char := range lines[i] {
			hash = ((hash << 5) + hash) + uint64(char)
		}
		hash = ((hash << 5) + hash) + uint64('\n')
	}

	// Include total line count in hash to detect structural changes
	hash = ((hash << 5) + hash) + uint64(totalLines)
	return hash
}

// validatePosition ensures a position is within valid bounds
func (tb *TextBuffer) validatePosition(pos Position) error {
	if len(tb.lines) == 0 {
		// Initialize buffer if it's empty
		tb.lines = []string{""}
		tb.cursor = Position{Line: 0, Column: 0}
	}
	if pos.Line < 0 || pos.Line >= len(tb.lines) {
		return fmt.Errorf("%w: line %d out of range [0, %d)", ErrInvalidPosition, pos.Line, len(tb.lines))
	}
	if pos.Column < 0 || pos.Column > len(tb.lines[pos.Line]) {
		return fmt.Errorf("%w: column %d out of range [0, %d]", ErrInvalidPosition, pos.Column, len(tb.lines[pos.Line]))
	}
	return nil
}

// getLineLength returns the length of a line safely
func (tb *TextBuffer) getLineLength(lineIdx int) int {
	if lineIdx < 0 || lineIdx >= len(tb.lines) {
		return 0
	}
	return len(tb.lines[lineIdx])
}

func (tb *TextBuffer) GetContent() string {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return strings.Join(tb.lines, "\n")
}

func (tb *TextBuffer) GetLines() []string {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	lines := make([]string, len(tb.lines))
	copy(lines, tb.lines)
	return lines
}

// GetLinesRange returns a subset of lines for viewport rendering
func (tb *TextBuffer) GetLinesRange(start, end int) []string {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	if start < 0 {
		start = 0
	}
	if end > len(tb.lines) {
		end = len(tb.lines)
	}
	if start >= end {
		return []string{}
	}

	lines := make([]string, end-start)
	copy(lines, tb.lines[start:end])
	return lines
}

// GetLineCount returns the total number of lines
func (tb *TextBuffer) GetLineCount() int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return len(tb.lines)
}

// GetLine returns a single line safely
func (tb *TextBuffer) GetLine(lineIdx int) string {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	if lineIdx < 0 || lineIdx >= len(tb.lines) {
		return ""
	}
	return tb.lines[lineIdx]
}

func (tb *TextBuffer) GetCursor() Position {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.cursor
}

func (tb *TextBuffer) GetSelection() *Selection {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	if tb.selection == nil {
		return nil
	}
	selection := *tb.selection
	return &selection
}

func (tb *TextBuffer) HasSelection() bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.selection != nil
}

func (tb *TextBuffer) SetSelection(selection *Selection) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.selection = selection
}

func (tb *TextBuffer) ClearSelection() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.selection = nil
}

func (tb *TextBuffer) SetCursor(pos Position) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.cursor = tb.clampPosition(pos)
}

func (tb *TextBuffer) restoreSelectAllCursor() {
	if tb.selectAllOriginalCursor != nil {
		tb.cursor = *tb.selectAllOriginalCursor
		tb.selectAllOriginalCursor = nil
	}
}

func (tb *TextBuffer) MoveCursor(line, column int) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Initialize buffer if it's empty
	if len(tb.lines) == 0 {
		tb.lines = []string{""}
		tb.cursor = Position{Line: 0, Column: 0}
	}

	// Clamp values to valid ranges
	if line < 0 {
		line = 0
	}
	if line >= len(tb.lines) {
		line = len(tb.lines) - 1
	}

	if column < 0 {
		column = 0
	}
	if column > len(tb.lines[line]) {
		column = len(tb.lines[line])
	}

	tb.cursor = Position{Line: line, Column: column}
	return nil
}

func (tb *TextBuffer) MoveCursorDelta(deltaLine, deltaColumn int, extend bool) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if err := tb.ensureBufferInitialized(); err != nil {
		return err
	}

	if tb.handleSelectAllRestore(extend) {
		return nil
	}

	tb.handleSelectionStart(extend)
	newPos := tb.calculateNewPosition(deltaLine, deltaColumn)
	tb.cursor = tb.clampPosition(newPos)
	tb.updateSelectionAfterMove(extend)

	return nil
}

// ensureBufferInitialized initializes the buffer if it's empty
func (tb *TextBuffer) ensureBufferInitialized() error {
	if len(tb.lines) == 0 {
		tb.lines = []string{""}
		tb.cursor = Position{Line: 0, Column: 0}
	}
	return nil
}

// handleSelectAllRestore handles restoration of cursor after select all
func (tb *TextBuffer) handleSelectAllRestore(extend bool) bool {
	if !extend && tb.selectAllOriginalCursor != nil {
		tb.restoreSelectAllCursor()
		tb.selection = nil
		return true
	}
	if tb.selectAllOriginalCursor != nil {
		tb.selectAllOriginalCursor = nil
	}
	return false
}

// handleSelectionStart initializes selection if extending
func (tb *TextBuffer) handleSelectionStart(extend bool) {
	if extend && tb.selection == nil {
		tb.selection = &Selection{
			Start: tb.cursor,
			End:   tb.cursor,
		}
	}
}

// calculateNewPosition calculates the new cursor position with line wrapping
func (tb *TextBuffer) calculateNewPosition(deltaLine, deltaColumn int) Position {
	newLine := tb.cursor.Line + deltaLine
	newCol := tb.cursor.Column + deltaColumn

	// Handle line wrapping when moving left
	if newCol < 0 && newLine > 0 {
		newLine--
		newCol = len(tb.lines[newLine])
	}

	// Handle line wrapping when moving right
	if newLine >= 0 && newLine < len(tb.lines) &&
		newCol > len(tb.lines[newLine]) && newLine < len(tb.lines)-1 {
		newLine++
		newCol = 0
	}

	return Position{Line: newLine, Column: newCol}
}

// updateSelectionAfterMove updates selection state after cursor movement
func (tb *TextBuffer) updateSelectionAfterMove(extend bool) {
	if extend && tb.selection != nil {
		tb.selection.End = tb.cursor
	} else if !extend {
		tb.selection = nil
	}
}

func (tb *TextBuffer) MoveToWordBoundary(forward bool, extend bool) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
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
	tb.mu.Lock()
	defer tb.mu.Unlock()
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
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if tb.cursor.Line < len(tb.lines) {
		tb.selection = &Selection{
			Start: Position{Line: tb.cursor.Line, Column: 0},
			End:   Position{Line: tb.cursor.Line, Column: len(tb.lines[tb.cursor.Line])},
		}
		tb.cursor = tb.selection.End
	}
}

func (tb *TextBuffer) SelectWord() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if tb.cursor.Line >= len(tb.lines) {
		return
	}
	start, end := tb.GetWordBoundsAtCursor()
	if start == end {
		return
	}
	tb.selection = &Selection{
		Start: Position{Line: tb.cursor.Line, Column: start},
		End:   Position{Line: tb.cursor.Line, Column: end},
	}
	tb.cursor.Column = end
}

func (tb *TextBuffer) GetSelectedText() string {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	if tb.selection == nil {
		return ""
	}

	start, end := tb.normalizeSelection()

	if start.Line < 0 || start.Line >= len(tb.lines) || end.Line < 0 || end.Line >= len(tb.lines) {
		return ""
	}

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

func (tb *TextBuffer) InsertText(text string) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if err := tb.prepareForInsertion(); err != nil {
		return err
	}

	lines := strings.Split(text, "\n")
	if len(lines) == 1 {
		tb.insertSingleLine(text)
	} else {
		tb.insertMultipleLines(lines)
	}

	tb.finalizeInsertion(lines)
	return nil
}

// prepareForInsertion handles pre-insertion setup and validation
func (tb *TextBuffer) prepareForInsertion() error {
	if err := tb.ensureBufferInitialized(); err != nil {
		return err
	}

	if err := tb.validatePosition(tb.cursor); err != nil {
		return fmt.Errorf("invalid cursor position: %w", err)
	}

	tb.saveState()
	tb.selectAllOriginalCursor = nil

	if tb.selection != nil {
		if err := tb.deleteSelection(); err != nil {
			return fmt.Errorf("failed to delete selection: %w", err)
		}
	}

	// Ensure cursor is valid after potential selection deletion
	tb.cursor = tb.clampPosition(tb.cursor)
	if err := tb.ensureBufferInitialized(); err != nil {
		return err
	}

	return nil
}

// insertSingleLine handles single-line text insertion
func (tb *TextBuffer) insertSingleLine(text string) {
	currentLine := tb.lines[tb.cursor.Line]
	before := currentLine[:tb.cursor.Column]
	after := currentLine[tb.cursor.Column:]
	tb.lines[tb.cursor.Line] = before + text + after
	tb.cursor.Column += len(text)
}

// insertMultipleLines handles multi-line text insertion
func (tb *TextBuffer) insertMultipleLines(lines []string) {
	currentLine := tb.lines[tb.cursor.Line]
	before := currentLine[:tb.cursor.Column]
	after := currentLine[tb.cursor.Column:]

	// Update first line
	tb.lines[tb.cursor.Line] = before + lines[0]

	// Choose insertion strategy based on file size
	if len(tb.lines) > 1000 {
		tb.insertLinesEfficient(tb.cursor.Line+1, lines[1:], after)
	} else {
		tb.insertLinesStandard(lines, after)
	}

	// Update cursor position
	lastLineIdx := tb.cursor.Line + len(lines) - 1
	tb.lines[lastLineIdx] += after
	tb.cursor.Line = lastLineIdx
	tb.cursor.Column = len(lines[len(lines)-1])
}

// insertLinesStandard handles line insertion for smaller files
func (tb *TextBuffer) insertLinesStandard(lines []string, suffix string) {
	newLines := make([]string, len(tb.lines)+len(lines)-1)
	copy(newLines, tb.lines[:tb.cursor.Line+1])
	copy(newLines[tb.cursor.Line+1:], lines[1:])
	copy(newLines[tb.cursor.Line+len(lines):], tb.lines[tb.cursor.Line+1:])
	tb.lines = newLines
}

// finalizeInsertion handles post-insertion cleanup and optimization
func (tb *TextBuffer) finalizeInsertion(lines []string) {
	tb.selection = nil
	tb.lastLineCount = len(tb.lines)

	// Only recalculate hash for smaller files or single-line changes
	if len(tb.lines) < 500 || len(lines) == 1 {
		tb.lastContentHash = tb.calculateContentHash(tb.lines)
	}
}

// insertLinesEfficient inserts lines efficiently for large files
func (tb *TextBuffer) insertLinesEfficient(insertAt int, newLines []string, suffix string) {
	if len(newLines) == 0 {
		return
	}

	// Create new slice with enough capacity
	totalLines := len(tb.lines) + len(newLines)
	newSlice := make([]string, 0, totalLines)

	// Copy lines before insertion point
	newSlice = append(newSlice, tb.lines[:insertAt]...)

	// Add new lines
	newSlice = append(newSlice, newLines...)

	// Add suffix to the last inserted line
	if len(newLines) > 0 {
		newSlice[insertAt+len(newLines)-1] += suffix
	}

	// Copy lines after insertion point
	newSlice = append(newSlice, tb.lines[insertAt:]...)

	tb.lines = newSlice
}

func (tb *TextBuffer) DeleteSelection() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if tb.selection == nil {
		return false
	}

	tb.saveState()
	tb.selectAllOriginalCursor = nil
	if err := tb.deleteSelection(); err != nil {
		// Log error but don't fail the operation
		return false
	}
	return true
}

func (tb *TextBuffer) DeleteChar(backward bool) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if err := tb.ensureBufferInitialized(); err != nil {
		return err
	}

	if err := tb.validatePosition(tb.cursor); err != nil {
		return fmt.Errorf("invalid cursor position: %w", err)
	}

	// Handle selection deletion first
	if tb.selection != nil {
		return tb.handleSelectionDeletion()
	}

	tb.saveState()
	tb.selectAllOriginalCursor = nil

	if backward {
		tb.deleteBackward()
	} else {
		tb.deleteForward()
	}

	tb.updatePerformanceTracking()
	return nil
}

// handleSelectionDeletion deletes the current selection
func (tb *TextBuffer) handleSelectionDeletion() error {
	tb.saveState()
	tb.selectAllOriginalCursor = nil
	if err := tb.deleteSelection(); err != nil {
		return fmt.Errorf("failed to delete selection: %w", err)
	}
	return nil
}

// deleteBackward handles backward character deletion
func (tb *TextBuffer) deleteBackward() {
	if tb.cursor.Column > 0 {
		// Delete character in current line
		line := tb.lines[tb.cursor.Line]
		tb.lines[tb.cursor.Line] = line[:tb.cursor.Column-1] + line[tb.cursor.Column:]
		tb.cursor.Column--
		return
	}

	if tb.cursor.Line > 0 {
		// Join with previous line
		prevLine := tb.lines[tb.cursor.Line-1]
		currentLine := tb.lines[tb.cursor.Line]
		tb.lines[tb.cursor.Line-1] = prevLine + currentLine
		tb.lines = append(tb.lines[:tb.cursor.Line], tb.lines[tb.cursor.Line+1:]...)
		tb.cursor.Line--
		tb.cursor.Column = len(prevLine)
	}
}

// deleteForward handles forward character deletion
func (tb *TextBuffer) deleteForward() {
	if tb.cursor.Column < len(tb.lines[tb.cursor.Line]) {
		// Delete character in current line
		line := tb.lines[tb.cursor.Line]
		tb.lines[tb.cursor.Line] = line[:tb.cursor.Column] + line[tb.cursor.Column+1:]
		return
	}

	if tb.cursor.Line < len(tb.lines)-1 {
		// Join with next line
		currentLine := tb.lines[tb.cursor.Line]
		nextLine := tb.lines[tb.cursor.Line+1]
		tb.lines[tb.cursor.Line] = currentLine + nextLine
		tb.lines = append(tb.lines[:tb.cursor.Line+1], tb.lines[tb.cursor.Line+2:]...)
	}
}

// updatePerformanceTracking updates performance tracking fields
func (tb *TextBuffer) updatePerformanceTracking() {
	tb.lastLineCount = len(tb.lines)
	tb.lastContentHash = tb.calculateContentHash(tb.lines)
}

func (tb *TextBuffer) Undo() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
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
	tb.mu.Lock()
	defer tb.mu.Unlock()
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
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.cursor = Position{
		Line:   tb.clampLine(line),
		Column: 0,
	}
	tb.selection = nil
	tb.selectAllOriginalCursor = nil
}

func (tb *TextBuffer) FindText(query string, caseSensitive bool) []Position {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.findTextWithContext(context.Background(), query, caseSensitive)
}

func (tb *TextBuffer) FindTextWithContext(ctx context.Context, query string, caseSensitive bool) []Position {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.findTextWithContext(ctx, query, caseSensitive)
}

func (tb *TextBuffer) findTextWithContext(ctx context.Context, query string, caseSensitive bool) []Position {
	var positions []Position

	searchQuery := query
	if !caseSensitive {
		searchQuery = strings.ToLower(query)
	}

	for lineIdx, line := range tb.lines {
		select {
		case <-ctx.Done():
			return positions
		default:
		}

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
	if len(tb.lines) == 0 {
		return Position{Line: 0, Column: 0}
	}

	if pos.Line < 0 {
		pos.Line = 0
	} else if pos.Line >= len(tb.lines) {
		pos.Line = len(tb.lines) - 1
	}

	if pos.Line < len(tb.lines) {
		if line := tb.lines[pos.Line]; pos.Column > len(line) {
			pos.Column = len(line)
		} else if pos.Column < 0 {
			pos.Column = 0
		}
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

func (tb *TextBuffer) deleteSelection() error {
	if tb.selection == nil {
		return nil
	}

	start, end := tb.normalizeSelection()

	if err := tb.validatePosition(start); err != nil {
		return fmt.Errorf("invalid selection start: %w", err)
	}
	if err := tb.validatePosition(end); err != nil {
		return fmt.Errorf("invalid selection end: %w", err)
	}

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

	// Update performance tracking fields
	tb.lastLineCount = len(tb.lines)
	tb.lastContentHash = tb.calculateContentHash(tb.lines)

	return nil
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
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	if len(tb.lines) == 0 {
		return -1, -1
	}

	line := tb.lines[tb.cursor.Line]
	if len(line) == 0 {
		// If line is empty, no word to highlight
		return -1, -1
	}

	// Clamp cursor position to valid range
	cursorCol := tb.cursor.Column
	if cursorCol >= len(line) {
		cursorCol = len(line) - 1
	}
	if cursorCol < 0 {
		return -1, -1
	}

	// If cursor is on whitespace, no word to highlight
	if isWhitespace(line[cursorCol]) {
		return -1, -1
	}

	// Find word start - go backwards until we hit a word boundary
	wordStart := cursorCol
	for wordStart > 0 && !isWordBoundary(line[wordStart-1]) && !isWhitespace(line[wordStart-1]) {
		wordStart--
	}

	// Find word end - go forwards until we hit a word boundary
	wordEnd := cursorCol + 1
	for wordEnd < len(line) && !isWordBoundary(line[wordEnd]) && !isWhitespace(line[wordEnd]) {
		wordEnd++
	}

	// Ensure we have a valid word (at least one character)
	if wordStart >= wordEnd {
		return -1, -1
	}

	// Check if the word contains at least one alphanumeric character
	hasAlphaNum := false
	for i := wordStart; i < wordEnd; i++ {
		if isAlphaNumeric(line[i]) {
			hasAlphaNum = true
			break
		}
	}

	if !hasAlphaNum {
		return -1, -1
	}

	return wordStart, wordEnd
}

// isWordBoundary returns true if the character is a word boundary
func isWordBoundary(ch byte) bool {
	return ch == '.' || ch == ',' || ch == ';' || ch == ':' ||
		ch == '!' || ch == '?' || ch == '(' || ch == ')' ||
		ch == '[' || ch == ']' || ch == '{' || ch == '}' ||
		ch == '<' || ch == '>' || ch == '"' || ch == '\'' ||
		ch == '/' || ch == '\\' || ch == '|' || ch == '&' ||
		ch == '*' || ch == '+' || ch == '-' || ch == '=' ||
		ch == '@' || ch == '#' || ch == '$' || ch == '%' ||
		ch == '^' || ch == '~' || ch == '`'
}

// isWhitespace returns true if the character is whitespace
func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// isAlphaNumeric returns true if the character is alphanumeric or underscore
func isAlphaNumeric(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '_'
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
