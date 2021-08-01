// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"database/sql"
	"io"
	"strings"

	"github.com/microsoft/go-sqlcmd/util"
	"github.com/microsoft/go-sqlcmd/variables"
)

const (
	defaultMaxDisplayWidth = 1024 * 1024
	maxPadWidth            = 8000
)

type Formatter interface {
	BeginBatch(query string, vars *variables.Variables, out io.Writer, err io.Writer)
	EndBatch()
	BeginResultSet([]*sql.ColumnType)
	EndResultSet()
	AddRow(*sql.Rows)
	AddMessage(string)
	AddError(err error)
}

type columnDetail struct {
	displayWidth       int64
	leftJustify        bool
	zeroesAfterDecimal bool
	col                sql.ColumnType
}

// The default formatter based on the native sqlcmd style
type sqlCmdFormatterType struct {
	out                  io.Writer
	err                  io.Writer
	vars                 *variables.Variables
	colsep               string
	removeTrailingSpaces bool
	columnDetails        []columnDetail
}

func NewSqlCmdDefaultFormatter(removeTrailingSpaces bool) Formatter {
	return &sqlCmdFormatterType{
		removeTrailingSpaces: removeTrailingSpaces,
	}
}

func (f *sqlCmdFormatterType) BeginBatch(_ string, vars *variables.Variables, out io.Writer, err io.Writer) {
	f.out = out
	f.err = err
	f.vars = vars
	f.colsep = vars.ColumnSeparator()
}

func (f *sqlCmdFormatterType) EndBatch() {
}

// Calculate the widths for each column and print the column names
// Since sql.ColumnType only provides sizes for variable length types we will
// base our numbers for most types on https://docs.microsoft.com/sql/odbc/reference/appendixes/column-size
func (f *sqlCmdFormatterType) BeginResultSet(cols []*sql.ColumnType) {
	f.columnDetails = calcColumnDetails(cols, f.vars.MaxFixedColumnWidth(), f.vars.MaxVarColumnWidth())
	if f.vars.MaxVarColumnWidth() > 0 {
		f.printColumnHeadings()
	}
}

func (f *sqlCmdFormatterType) EndResultSet()      {}
func (f *sqlCmdFormatterType) AddRow(*sql.Rows)   {}
func (f *sqlCmdFormatterType) AddMessage(string)  {}
func (f *sqlCmdFormatterType) AddError(err error) {}

func (f *sqlCmdFormatterType) printColumnHeadings() {
	names := new(strings.Builder)
	sep := new(strings.Builder)

	var leftPad, rightPad int64
	for i, c := range f.columnDetails {
		nameLen := int64(len([]rune(c.col.Name())))
		if f.removeTrailingSpaces {
			if nameLen == 0 {

				// special case for unnamed columns when using -W
				// print a single -
				rightPad = 1
				sep = util.PadRight(sep, 1, "-")
			} else {
				sep = util.PadRight(sep, nameLen, "-")
			}
		} else {
			length := min64(c.displayWidth, maxPadWidth)
			if nameLen < length {
				rightPad = length - nameLen
			}
			sep = util.PadRight(sep, length, "-")
		}
		names = util.PadRight(names, leftPad, " ")
		names.WriteString(c.col.Name()[:min64(nameLen, c.displayWidth)])
		names = util.PadRight(names, rightPad, " ")
		if i != len(f.columnDetails)-1 {
			names.WriteString(f.colsep)
			sep.WriteString(f.colsep)
		}
	}
	names.WriteString(SqlcmdEol)
	sep.WriteString(SqlcmdEol)
	names = fitToScreen(names, f.vars.ScreenWidth())
	sep = fitToScreen(sep, f.vars.ScreenWidth())
	f.out.Write([]byte(names.String()))
	f.out.Write([]byte(sep.String()))
}

func fitToScreen(s *strings.Builder, width int64) *strings.Builder {
	str := s.String()
	runes := []rune(str)
	if width == 0 || int64(len(runes)) < width {
		return s
	}

	line := new(strings.Builder)
	line.Grow(len(str))
	var c int64
	for i, r := range runes {
		if c == width {
			// We have printed a line's worth
			// if the next character is not part of a carriage return write our Eol
			if (SqlcmdEol == "\r\n" && (i == len(runes)-1 || (i < len(runes)-1 && string(runes[i:i+2]) != SqlcmdEol))) || (SqlcmdEol == "\n" && r != '\n') {
				line.WriteString(SqlcmdEol)
				c = 0
			}
		}
		line.WriteRune(r)
		if r == '\n' {
			c = 0
			// we are assuming \r is a non-printed character
			// The likelihood of a \r not being followed by \n is low
		} else if r == '\r' && SqlcmdEol == "\r\n" {
			c = 0
		} else {
			c++
		}
	}
	return line
}

func calcColumnDetails(cols []*sql.ColumnType, fixed int64, variable int64) (columnDetails []columnDetail) {
	columnDetails = make([]columnDetail, len(cols))
	for i, c := range cols {
		length, _ := c.Length()
		nameLen := int64(len([]rune(c.Name())))
		columnDetails[i].col = *c
		columnDetails[i].leftJustify = true
		columnDetails[i].zeroesAfterDecimal = false
		if length == 0 {
			columnDetails[i].displayWidth = defaultMaxDisplayWidth
		} else {
			columnDetails[i].displayWidth = length
		}
		switch c.DatabaseTypeName() {
		// Types with 0 size from sql.ColumnType
		case "BIT":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(1, nameLen)
		case "TINYINT":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(3, nameLen)
		case "SMALLINT":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(6, nameLen)
		case "INT":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(11, nameLen)
		case "BIGINT":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(21, nameLen)
		case "REAL":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(14, nameLen)
			columnDetails[i].zeroesAfterDecimal = true
		case "FLOAT":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(24, nameLen)
			columnDetails[i].zeroesAfterDecimal = true
		case "DECIMAL":
			columnDetails[i].leftJustify = false
			d, _, ok := c.DecimalSize()
			// maybe panic on !ok?
			if !ok {
				d = 24
			}
			columnDetails[i].displayWidth = max64(d+2, nameLen)
			columnDetails[i].zeroesAfterDecimal = true
		case "DATE":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(16, nameLen)
		case "DATETIME":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(23, nameLen)
		case "SMALLDATETIME":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(19, nameLen)
			columnDetails[i].zeroesAfterDecimal = true
		case "DATETIME2":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(38, nameLen)
			columnDetails[i].zeroesAfterDecimal = true
		case "DATETIMEOFFSET":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(45, nameLen)
		case "UNIQUEIDENTIFIER":
			columnDetails[i].displayWidth = max64(36, nameLen)
		// Types that can be fixed or variable
		case "VARCHAR":
			if length > 8000 {
				columnDetails[i].displayWidth = variable
			} else {
				columnDetails[i].displayWidth = min64(fixed, max64(length, nameLen))
			}
		case "NVARCHAR":
			if length > 4000 {
				columnDetails[i].displayWidth = variable
			} else {
				columnDetails[i].displayWidth = min64(fixed, max64(length, nameLen))
			}
		case "VARBINARY":
			if length <= 8000 {
				columnDetails[i].displayWidth = min64(fixed, max64(length, nameLen))
				columnDetails[i].displayWidth = variable
			}
		// Fixed length types
		case "CHAR", "NCHAR", "VARIANT":
			columnDetails[i].displayWidth = min64(fixed, max64(length, nameLen))
		// Variable length types
		// TODO: Fix BINARY once we have a driver with fix for https://github.com/denisenkom/go-mssqldb/issues/685
		case "XML", "TEXT", "NTEXT", "IMAGE", "BINARY":
			columnDetails[i].displayWidth = variable
		default:
			columnDetails[i].displayWidth = length
		}
		// When max var length is 0 we don't print column headers and print every value with unlimited width
		if variable == 0 {
			columnDetails[i].displayWidth = 0
		}
	}
	return columnDetails
}
