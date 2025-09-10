package main

import (
	"bytes"
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// HighlightCache represents a cache entry for highlighted content
type HighlightCache struct {
	content   string
	timestamp time.Time
}

type Highlighter struct {
	lexer     chroma.Lexer
	formatter chroma.Formatter
	style     *chroma.Style
	mu        sync.RWMutex
	cache     map[string]HighlightCache
	maxCache  int
}

func NewHighlighter(filename string) *Highlighter {
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("doom-one")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	return &Highlighter{
		lexer:     lexer,
		formatter: formatter,
		style:     style,
		cache:     make(map[string]HighlightCache),
		maxCache:  100, // Cache up to 100 line ranges
	}
}

// Highlight highlights the entire content (legacy method for compatibility)
func (h *Highlighter) Highlight(content string) (string, error) {
	iterator, err := h.lexer.Tokenise(nil, content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = h.formatter.Format(&buf, h.style, iterator)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// HighlightLines highlights specific lines with caching for performance
func (h *Highlighter) HighlightLines(ctx context.Context, lines []string, startLine, endLine int) ([]string, error) {
	// Validate and clamp bounds to prevent slice out of range panics
	if len(lines) == 0 {
		return []string{}, nil
	}
	
	// Clamp startLine to valid range
	if startLine < 0 {
		startLine = 0
	}
	if startLine >= len(lines) {
		startLine = len(lines) - 1
	}
	
	// Clamp endLine to valid range
	if endLine < 0 {
		endLine = 0
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}
	
	// Ensure startLine <= endLine
	if startLine > endLine {
		startLine, endLine = endLine, startLine
	}

	// Create cache key based on line range and content hash
	cacheKey := h.createCacheKey(lines, startLine, endLine)
	
	h.mu.RLock()
	if cached, exists := h.cache[cacheKey]; exists {
		// Check if cache is still valid (within 5 seconds)
		if time.Since(cached.timestamp) < 5*time.Second {
			h.mu.RUnlock()
			return strings.Split(cached.content, "\n"), nil
		}
	}
	h.mu.RUnlock()

	// Extract the range of lines to highlight
	lineRange := lines[startLine:endLine+1]
	content := strings.Join(lineRange, "\n")

	// Highlight the content
	highlighted, err := h.highlightContent(ctx, content)
	if err != nil {
		slog.Warn("Failed to highlight content", "error", err)
		return lineRange, nil // Return original lines on error
	}

	// Cache the result
	h.cacheResult(cacheKey, highlighted)

	return strings.Split(highlighted, "\n"), nil
}

// highlightContent performs the actual highlighting with context support
func (h *Highlighter) highlightContent(ctx context.Context, content string) (string, error) {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	iterator, err := h.lexer.Tokenise(nil, content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = h.formatter.Format(&buf, h.style, iterator)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// createCacheKey creates a cache key based on line content and range
func (h *Highlighter) createCacheKey(lines []string, startLine, endLine int) string {
	var keyBuilder strings.Builder
	keyBuilder.WriteString("range:")
	keyBuilder.WriteString(strconv.Itoa(startLine))
	keyBuilder.WriteString("-")
	keyBuilder.WriteString(strconv.Itoa(endLine))
	keyBuilder.WriteString(":")
	
	// Add a hash of the content for cache invalidation
	for i := startLine; i <= endLine && i < len(lines); i++ {
		if i > startLine {
			keyBuilder.WriteString("\n")
		}
		keyBuilder.WriteString(lines[i])
	}
	
	return keyBuilder.String()
}

// cacheResult stores the highlighted result in cache
func (h *Highlighter) cacheResult(key, content string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Clean old cache entries if we're at capacity
	if len(h.cache) >= h.maxCache {
		h.cleanOldEntries()
	}

	h.cache[key] = HighlightCache{
		content:   content,
		timestamp: time.Now(),
	}
}

// cleanOldEntries removes old cache entries to make room for new ones
func (h *Highlighter) cleanOldEntries() {
	// Remove entries older than 30 seconds
	cutoff := time.Now().Add(-30 * time.Second)
	for key, entry := range h.cache {
		if entry.timestamp.Before(cutoff) {
			delete(h.cache, key)
		}
	}

	// If still at capacity, remove oldest entries
	if len(h.cache) >= h.maxCache {
		oldestKey := ""
		oldestTime := time.Now()
		for key, entry := range h.cache {
			if entry.timestamp.Before(oldestTime) {
				oldestTime = entry.timestamp
				oldestKey = key
			}
		}
		if oldestKey != "" {
			delete(h.cache, oldestKey)
		}
	}
}

// ClearCache clears the highlighting cache
func (h *Highlighter) ClearCache() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cache = make(map[string]HighlightCache)
}

// applySyntaxHighlighting applies lazy syntax highlighting only to visible lines
func (m *Model) applySyntaxHighlighting() {
	if m.highlighter == nil {
		return
	}

	// Get a thread-safe copy of the lines to prevent race conditions
	lines := m.textBuffer.GetLines()
	if len(lines) == 0 {
		return
	}

	// Calculate visible line range with buffer
	visibleStart, visibleEnd := m.calculateVisibleRange()
	
	// Additional safety check to ensure bounds are valid
	if visibleStart < 0 || visibleEnd >= len(lines) || visibleStart > visibleEnd {
		slog.Warn("Invalid visible range calculated", "start", visibleStart, "end", visibleEnd, "totalLines", len(lines))
		return
	}
	
	// Use context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Highlight only the visible range
	highlightedRange, err := m.highlighter.HighlightLines(ctx, lines, visibleStart, visibleEnd)
	if err != nil {
		slog.Warn("Failed to apply syntax highlighting", "error", err)
		return
	}

	// Update only the visible portion of highlighted lines
	m.updateHighlightedRange(visibleStart, visibleEnd, highlightedRange)
}

// calculateVisibleRange determines which lines need highlighting based on viewport
func (m *Model) calculateVisibleRange() (start, end int) {
	totalLines := len(m.textBuffer.lines)
	if totalLines == 0 {
		return 0, 0
	}

	// Calculate visible area with buffer for smooth scrolling
	bufferSize := 10 // Lines to highlight beyond visible area
	viewportHeight := m.height - 2 // Account for status bar
	
	start = max(0, m.viewportY-bufferSize)
	end = min(totalLines-1, m.viewportY+viewportHeight+bufferSize)
	
	// Ensure start <= end and both are within valid bounds
	if start >= totalLines {
		start = totalLines - 1
	}
	if end < 0 {
		end = 0
	}
	if start > end {
		start = end
	}
	
	return start, end
}

// updateHighlightedRange updates the highlighted lines for a specific range
func (m *Model) updateHighlightedRange(start, end int, highlightedRange []string) {
	// Ensure highlightedLines slice is properly sized
	if len(m.highlightedLines) != len(m.textBuffer.lines) {
		m.highlightedLines = make([]string, len(m.textBuffer.lines))
		// Copy original lines as fallback
		copy(m.highlightedLines, m.textBuffer.lines)
	}

	// Update only the highlighted range
	for i, highlightedLine := range highlightedRange {
		lineIndex := start + i
		if lineIndex >= 0 && lineIndex < len(m.highlightedLines) {
			m.highlightedLines[lineIndex] = highlightedLine
		}
	}
}

// invalidateHighlightCache clears syntax highlighting cache when content changes
func (m *Model) invalidateHighlightCache() {
	if m.highlighter != nil {
		m.highlighter.ClearCache()
	}
}