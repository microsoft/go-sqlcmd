package color

import (
	"io"
	"os"

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
	// Write prints s to w using the current color scheme. If w is not a terminal or if it is redirected, no color codes are printed
	Write(w io.Writer, s string, scheme string, t TextType)
}

type chromaColorizer struct {
	forceColor bool
}

func New(forceColor bool) Colorizer {
	return &chromaColorizer{forceColor: forceColor}
}

func (c *chromaColorizer) Write(w io.Writer, s string, scheme string, t TextType) {
	colorize := scheme != ""
	// only colorize if w is a terminal and it's not redirected, or if forceColor is set
	if colorize && !c.forceColor {
		if f, ok := w.(*os.File); ok {
			if f == os.Stdout || f == os.Stderr {
				i, _ := f.Stat()
				colorize = (i.Mode() & os.ModeCharDevice) == os.ModeCharDevice
			} else {
				colorize = false
			}
		} else {
			colorize = false
		}
	}
	if !colorize {
		t = TextTypeNormal
	}
	switch t {
	case TextTypeNormal:
		_, _ = w.Write([]byte(s))
	case TextTypeTSql:
		if err := quick.Highlight(w, s, "transact-sql", "terminal256", scheme); err != nil {
			_, _ = w.Write([]byte(s))
		}
	}
}
