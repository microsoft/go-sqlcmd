package color

import (
	"io"
	"os"
	"sort"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/chroma/v2/styles"
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

var typeMap map[TextType]chroma.TokenType = map[TextType]chroma.TokenType{
	TextTypeCell:      chroma.StringOther,
	TextTypeHeader:    chroma.GenericHeading,
	TextTypeSeparator: chroma.StringDelimiter,
	TextTypeError:     chroma.GenericError,
	TextTypeWarning:   chroma.GenericEmph,
}

// Colorizer has methods to write colorized text to a stream
type Colorizer interface {
	// Write prints s to w using the current color scheme. If w is not a terminal or if it is redirected, no color codes are printed
	Write(w io.Writer, s string, scheme string, t TextType) error
	// Styles returns the array of available style names
	Styles() []string
}

type chromaColorizer struct {
	forceColor bool
}

func New(forceColor bool) Colorizer {
	return &chromaColorizer{forceColor: forceColor}
}

func (c *chromaColorizer) Write(w io.Writer, s string, scheme string, t TextType) (err error) {
	style := styles.Get(scheme)
	colorize := scheme != "" && style != nil
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
	// We use this lexer simply for token iteration
	lexer := lexers.Get("plaintext")
	formatter := formatters.Get("terminal16m")
	if !colorize || lexer == nil || formatter == nil {
		t = TextTypeNormal
	}
	switch t {
	case TextTypeNormal:
		_, err = w.Write([]byte(s))
	case TextTypeTSql:
		if err = quick.Highlight(w, s, "transact-sql", "terminal16m", scheme); err != nil {
			_, err = w.Write([]byte(s))
		}
	default:
		tokens := chroma.Literator(chroma.Token{
			Type: typeMap[t], Value: s})
		if err = formatter.Format(w, style, tokens); err != nil {
			_, err = w.Write([]byte(s))
		}
	}
	return
}

func (c *chromaColorizer) Styles() []string {
	s := make([]string, len(styles.Registry))
	i := 0
	for key := range styles.Registry {
		s[i] = key
		i++
	}
	sort.Strings(s)
	return s
}
