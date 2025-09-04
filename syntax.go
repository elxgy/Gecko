package main

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type Highlighter struct {
	lexer     chroma.Lexer
	formatter chroma.Formatter
	style     *chroma.Style
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
	}
}

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

func (m *Model) applySyntaxHighlighting() {
	if m.highlighter == nil {
		m.highlighter = NewHighlighter(m.filename)
	}

	content := m.textBuffer.GetContent()
	highlighted, err := m.highlighter.Highlight(content)
	if err == nil {
		m.highlightedContent = strings.Split(highlighted, "\n")
	} else {
		m.highlightedContent = m.textBuffer.GetLines()
	}
}