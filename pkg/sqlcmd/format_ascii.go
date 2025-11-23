package sqlcmd

import (
	"database/sql"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/go-sqlcmd/internal/color"
	"golang.org/x/term"
)

type asciiFormatter struct {
	*sqlCmdFormatterType
	rows [][]string
}

func NewSQLCmdAsciiFormatter(vars *Variables, removeTrailingSpaces bool, ccb ControlCharacterBehavior) Formatter {
	return &asciiFormatter{
		sqlCmdFormatterType: &sqlCmdFormatterType{
			removeTrailingSpaces: removeTrailingSpaces,
			format:               "ascii",
			colorizer:            color.New(false),
			ccb:                  ccb,
			vars:                 vars,
		},
	}
}

func (f *asciiFormatter) BeginResultSet(cols []*sql.ColumnType) {
	f.sqlCmdFormatterType.BeginResultSet(cols)
	f.rows = make([][]string, 0)
}

func (f *asciiFormatter) AddRow(row *sql.Rows) string {
	values, err := f.scanRow(row)
	if err != nil {
		f.mustWriteErr(err.Error())
		return ""
	}
	f.rows = append(f.rows, values)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func (f *asciiFormatter) EndResultSet() {
	if len(f.rows) > 0 || len(f.columnDetails) > 0 {
		f.printAsciiTable()
	}
	f.rows = nil
	f.writeOut(SqlcmdEol, color.TextTypeNormal)
}

func (f *asciiFormatter) printAsciiTable() {
	colWidths := make([]int, len(f.columnDetails))

	for i, c := range f.columnDetails {
		colWidths[i] = utf8.RuneCountInString(c.col.Name())
	}

	for _, row := range f.rows {
		for i, val := range row {
			if i < len(colWidths) {
				l := utf8.RuneCountInString(val)
				if l > colWidths[i] {
					colWidths[i] = l
				}
			}
		}
	}

	maxWidth := int(f.vars.ScreenWidth())
	if maxWidth <= 0 {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			maxWidth = w - 1
		} else {
			maxWidth = 1000000
		}
	}

	totalWidth := 1
	for _, w := range colWidths {
		totalWidth += w + 3
	}

	if totalWidth <= maxWidth {
		f.printTableSegment(colWidths, 0, len(colWidths)-1)
	} else {
		startCol := 0
		for startCol < len(colWidths) {
			currentWidth := 1
			endCol := startCol
			for endCol < len(colWidths) {
				w := colWidths[endCol] + 3
				if currentWidth+w > maxWidth {
					break
				}
				currentWidth += w
				endCol++
			}

			if endCol == startCol {
				endCol++
			}

			f.printTableSegment(colWidths, startCol, endCol-1)
			startCol = endCol
			if startCol < len(colWidths) {
				f.writeOut(SqlcmdEol, color.TextTypeNormal)
			}
		}
	}
}

func (f *asciiFormatter) printTableSegment(colWidths []int, startCol, endCol int) {
	if startCol > endCol {
		return
	}

	divider := "+"
	for i := startCol; i <= endCol; i++ {
		divider += strings.Repeat("-", colWidths[i]+2) + "+"
	}
	f.writeOut(divider+SqlcmdEol, color.TextTypeNormal)

	header := "|"
	for i := startCol; i <= endCol; i++ {
		name := f.columnDetails[i].col.Name()
		header += " " + padRightString(name, colWidths[i]) + " |"
	}
	f.writeOut(header+SqlcmdEol, color.TextTypeNormal)
	f.writeOut(divider+SqlcmdEol, color.TextTypeNormal)

	for _, row := range f.rows {
		line := "|"
		for i := startCol; i <= endCol; i++ {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			isNumeric := isNumericType(f.columnDetails[i].col.DatabaseTypeName())

			if isNumeric {
				line += " " + padLeftString(val, colWidths[i]) + " |"
			} else {
				line += " " + padRightString(val, colWidths[i]) + " |"
			}
		}
		f.writeOut(line+SqlcmdEol, color.TextTypeNormal)
	}
	f.writeOut(divider+SqlcmdEol, color.TextTypeNormal)
}

func padRightString(s string, width int) string {
	l := utf8.RuneCountInString(s)
	if l >= width {
		return s
	}
	return s + strings.Repeat(" ", width-l)
}

func padLeftString(s string, width int) string {
	l := utf8.RuneCountInString(s)
	if l >= width {
		return s
	}
	return strings.Repeat(" ", width-l) + s
}

func isNumericType(typeName string) bool {
	switch typeName {
	case "TINYINT", "SMALLINT", "INT", "BIGINT", "REAL", "FLOAT", "DECIMAL", "NUMERIC", "MONEY", "SMALLMONEY":
		return true
	}
	return false
}
