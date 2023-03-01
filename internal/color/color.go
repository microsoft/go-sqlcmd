package color

import (
	"io"

	"github.com/alecthomas/chroma/v2/quick"
)

// TextType defines a category of text that can be colorized
type TextType int

const (
	// TextTypeNormal is non-categorized text that prints with the default color
	TextTypeNormal TextType = iota
	// TextTypeTSql is for transact-sql syntax
	TextTypeTSql
	// TextTypeHeader is for a table column header or cell label
	TextTypeHeader
	// TextTypeCell is for a cell value
	TextTypeCell
	// TextTypeSeparator is for characters that delimit columns and rows
	TextTypeSeparator
	// TextTypeError is for error messages
	TextTypeError
	// TextTypeWarning is for warning messages
	TextTypeWarning
)

// Colorizer has methods to write colorized text to a stream and to add color to HTML content
type Colorizer interface {
	// Write prints s to w using the current color scheme. It assumes w is compatible with ANSI color codes.
	Write(w io.Writer, s string, t TextType)
}

type chromaColorizer struct {
	scheme string
}

func New(scheme string) Colorizer {
	return &chromaColorizer{scheme: scheme}
}

func (c *chromaColorizer) Write(w io.Writer, s string, t TextType) {
	if c.scheme == "" {
		t = TextTypeNormal
	}
	switch t {
	case TextTypeNormal:
		_, _ = w.Write([]byte(s))
	case TextTypeTSql:
		if err := quick.Highlight(w, s, "transact-sql", "terminal256", c.scheme); err != nil {
			_, _ = w.Write([]byte(s))
		}
	}
}
