package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type Highlighter struct {
	lexer           chroma.Lexer
	formatter       chroma.Formatter
	style           *chroma.Style
	cachedContent   string
	cachedResult    []string
	contentHash     string
	maxCacheSize    int
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
		lexer:        lexer,
		formatter:    formatter,
		style:        style,
		maxCacheSize: 1024 * 1024, // 1MB cache limit
	}
}

func (h *Highlighter) Highlight(content string) (string, error) {
	// Check cache first
	contentHash := h.calculateHash(content)
	if h.contentHash == contentHash && h.cachedContent == content {
		return strings.Join(h.cachedResult, "\n"), nil
	}

	// Skip highlighting for very large content to maintain performance
	if len(content) > h.maxCacheSize {
		return content, nil
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

	result := buf.String()
	
	// Cache the result
	h.cachedContent = content
	h.cachedResult = strings.Split(result, "\n")
	h.contentHash = contentHash

	return result, nil
}

// calculateHash generates a hash for content to enable efficient cache checking
func (h *Highlighter) calculateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// HighlightLines highlights specific lines for incremental updates
func (h *Highlighter) HighlightLines(lines []string, startLine int) ([]string, error) {
	// Validate startLine parameter
	if startLine < 0 || startLine >= len(lines) {
		return []string{}, nil
	}
	
	// Extract lines from startLine to end
	linesToHighlight := lines[startLine:]
	content := strings.Join(linesToHighlight, "\n")
	highlighted, err := h.Highlight(content)
	if err != nil {
		return linesToHighlight, err
	}
	return strings.Split(highlighted, "\n"), nil
}

func (m *Model) applySyntaxHighlighting() {
	if !m.highlightingEnabled || m.highlighter == nil {
		m.highlighter = NewHighlighter(m.filename)
		if !m.highlightingEnabled {
			m.highlightedContent = m.textBuffer.GetLines()
			return
		}
	}

	content := m.textBuffer.GetContent()
	contentHash := m.highlighter.calculateHash(content)
	
	// Check if content has changed
	if m.lastHighlightedHash == contentHash && len(m.highlightedContent) > 0 {
		return // No need to re-highlight
	}

	// For large files, disable highlighting to maintain performance
	lines := m.textBuffer.GetLines()
	if len(content) > m.highlighter.maxCacheSize || len(lines) > 5000 {
		m.highlightingEnabled = false
		m.highlightedContent = lines
		return
	}

	highlighted, err := m.highlighter.Highlight(content)
	if err == nil {
		m.highlightedContent = strings.Split(highlighted, "\n")
		m.lastHighlightedHash = contentHash
		m.dirtyLines = make(map[int]bool) // Clear dirty lines after full highlight
	} else {
		m.highlightedContent = m.textBuffer.GetLines()
	}
}

// markLinesDirty marks specific lines as needing re-highlighting
func (m *Model) markLinesDirty(startLine, endLine int) {
	if !m.highlightingEnabled {
		return
	}
	for i := startLine; i <= endLine; i++ {
		m.dirtyLines[i] = true
	}
}

// applyIncrementalHighlighting re-highlights only dirty lines for better performance
func (m *Model) applyIncrementalHighlighting() {
	if !m.highlightingEnabled || len(m.dirtyLines) == 0 {
		return
	}

	// For small number of dirty lines, do incremental highlighting
	if len(m.dirtyLines) < 50 {
		lines := m.textBuffer.GetLines()
		for lineNum := range m.dirtyLines {
			if lineNum >= 0 && lineNum < len(lines) {
				// Highlight a small context around the dirty line
				startCtx := max(0, lineNum-2)
				endCtx := min(len(lines)-1, lineNum+2)
				contextLines := lines[startCtx:endCtx+1]
				
				highlighted, err := m.highlighter.HighlightLines(contextLines, startCtx)
				if err == nil {
					// Update only the highlighted portion
					for i, highlightedLine := range highlighted {
						if startCtx+i < len(m.highlightedContent) {
							m.highlightedContent[startCtx+i] = highlightedLine
						}
					}
				}
			}
		}
		m.dirtyLines = make(map[int]bool) // Clear dirty lines
	} else {
		// Too many dirty lines, do full re-highlighting
		m.applySyntaxHighlighting()
	}
}
