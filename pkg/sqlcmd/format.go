// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"

	mssql "github.com/denisenkom/go-mssqldb"
)

const (
	defaultMaxDisplayWidth = 1024 * 1024
	maxPadWidth            = 8000
)

// Formatter defines methods to process query output
type Formatter interface {
	// BeginBatch is called before the query runs
	BeginBatch(query string, vars *Variables, out io.Writer, err io.Writer)
	// EndBatch is the last function called during batch execution and signals the end of the batch
	EndBatch()
	// BeginResultSet is called when a new result set is encountered
	BeginResultSet([]*sql.ColumnType)
	// EndResultSet is called after all rows in a result set have been processed
	EndResultSet()
	// AddRow is called for each row in a result set
	AddRow(*sql.Rows)
	// AddMessage is called for every information message returned by the server during the batch
	AddMessage(string)
	// AddError is called for each error encountered during batch execution
	AddError(err error)
}

// ControlCharacterBehavior specifies the text handling required for control characters in the output
type ControlCharacterBehavior int

const (
	// ControlIgnore preserves control characters in the output
	ControlIgnore ControlCharacterBehavior = iota
	// ControlReplace replaces control characters with spaces, 1 space per character
	ControlReplace
	// ControlRemove removes control characters from the output
	ControlRemove
	// ControlReplaceConsecutive replaces multiple consecutive control characters with a single space
	ControlReplaceConsecutive
)

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
	vars                 *Variables
	colsep               string
	removeTrailingSpaces bool
	ccb                  ControlCharacterBehavior
	columnDetails        []columnDetail
	rowcount             int
	writepos             int64
}

// NewSQLCmdDefaultFormatter returns a Formatter that mimics the original ODBC-based sqlcmd formatter
func NewSQLCmdDefaultFormatter(removeTrailingSpaces bool) Formatter {
	return &sqlCmdFormatterType{
		removeTrailingSpaces: removeTrailingSpaces,
	}
}

// Adds the given string to the current line, wrapping it based on the screen width setting
func (f *sqlCmdFormatterType) WriteOut(s string) {
	w := f.vars.ScreenWidth()
	if w == 0 {
		f.out.Write([]byte(s))
		return
	}

	r := []rune(s)
	for i := 0; true; {
		if i == len(r) {
			f.out.Write([]byte(string(r)))
			return
		} else if f.writepos == w {
			f.out.Write([]byte(string(r[:i])))
			f.out.Write([]byte(SqlcmdEol))
			r = []rune(string(r[i:]))
			f.writepos = 0
			i = 0
		} else {
			c := r[i]
			if c != '\r' && c != '\n' {
				f.writepos++
			} else {
				f.writepos = 0
			}
			i++
		}
	}
}

// Stores the settings to use for processing the current batch
// TODO: add a third io.Writer for messages when we add -r support
func (f *sqlCmdFormatterType) BeginBatch(_ string, vars *Variables, out io.Writer, err io.Writer) {
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
	f.rowcount = 0
	f.columnDetails = calcColumnDetails(cols, f.vars.MaxFixedColumnWidth(), f.vars.MaxVarColumnWidth())
	if f.vars.RowsBetweenHeaders() > -1 {
		f.printColumnHeadings()
	}
}

// Writes a blank line to the designated output writer
func (f *sqlCmdFormatterType) EndResultSet() {
	f.WriteOut(SqlcmdEol)
}

// Writes the current row to the designated output writer
func (f *sqlCmdFormatterType) AddRow(row *sql.Rows) {

	f.writepos = 0
	values, err := f.scanRow(row)
	if err != nil {
		f.err.Write([]byte(err.Error()))
		return
	}

	// values are the full values, look at the displaywidth of each column and truncate accordingly
	for i, v := range values {
		if i > 0 {
			f.WriteOut(f.vars.ColumnSeparator())
		}
		f.printColumnValue(v, i)
	}
	f.rowcount++
	gap := f.vars.RowsBetweenHeaders()
	if gap > 0 && (int64(f.rowcount)%gap == 0) {
		f.WriteOut(SqlcmdEol)
		f.printColumnHeadings()
	}
	f.WriteOut(SqlcmdEol)
}

// Writes a non-error message to the designated message writer
func (f *sqlCmdFormatterType) AddMessage(string) {}

// Writes an error to the designated err Writer
func (f *sqlCmdFormatterType) AddError(err error) {
	b := new(strings.Builder)
	msg := err.Error()
	switch e := (err).(type) {
	case mssql.Error:
		b.WriteString(fmt.Sprintf("Msg %d, Level %d, State %d, Server %s, Line %d%s", e.Number, e.Class, e.State, e.ServerName, e.LineNo, SqlcmdEol))
		msg = strings.TrimPrefix(msg, "mssql: ")
	}
	b.WriteString(msg)
	b.WriteString(SqlcmdEol)
	f.err.Write([]byte(fitToScreen(b, f.vars.ScreenWidth()).String()))
}

// Prints column headings based on columnDetail, variables, and command line arguments
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
				sep = PadRight(sep, 1, "-")
			} else {
				sep = PadRight(sep, nameLen, "-")
			}
		} else {
			length := min64(c.displayWidth, maxPadWidth)
			if nameLen < length {
				rightPad = length - nameLen
			}
			sep = PadRight(sep, length, "-")
		}
		names = PadRight(names, leftPad, " ")
		names.WriteString(c.col.Name()[:min64(nameLen, c.displayWidth)])
		names = PadRight(names, rightPad, " ")
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

// Wraps the input string every width characters when width > 0
// When width == 0 returns the input Builder
// When width > 0 returns a new Builder containing the wrapped string
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

// Given the array of driver-provided columnType values and the sqlcmd size limits,
// return an array of columnDetail objects describing the output format for each column
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
				if fixed > 0 {
					length = min64(fixed, length)
				}
				columnDetails[i].displayWidth = max64(length, nameLen)
			}
		case "NVARCHAR":
			if length > 4000 {
				columnDetails[i].displayWidth = variable
			} else {
				if fixed > 0 {
					length = min64(fixed, length)
				}
				columnDetails[i].displayWidth = max64(length, nameLen)
			}
		case "VARBINARY":
			if length <= 8000 {
				if fixed > 0 {
					length = min64(fixed, length)
				}
				columnDetails[i].displayWidth = max64(length, nameLen)
			} else {
				columnDetails[i].displayWidth = variable
			}
		// Fixed length types
		case "CHAR", "NCHAR", "VARIANT":
			if fixed > 0 {
				length = min64(fixed, length)
			}
			columnDetails[i].displayWidth = max64(length, nameLen)
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

// scanRow fetches the next row and converts each value to the appropriate string representation
func (f *sqlCmdFormatterType) scanRow(rows *sql.Rows) ([]string, error) {
	r := make([]interface{}, len(f.columnDetails))
	for i := range r {
		r[i] = new(interface{})
	}
	if err := rows.Scan(r...); err != nil {
		return nil, err
	}
	row := make([]string, len(f.columnDetails))
	for n, z := range r {
		j := z.(*interface{})
		if *j == nil {
			row[n] = "NULL"
		} else {
			switch x := (*j).(type) {
			case []byte:
				if isBinaryDataType(&f.columnDetails[n].col) {
					row[n] = decodeBinary(x)
				} else {
					row[n] = string(x)
				}
			case string:
				row[n] = x
			case time.Time:
				// Go lacks any way to get the user's preferred time format or even the system default
				row[n] = x.String()
			case fmt.Stringer:
				row[n] = x.String()
			// not sure why go-mssql reports bit as bool
			case bool:
				if x {
					row[n] = "1"
				} else {
					row[n] = "0"
				}
			default:
				var err error
				if row[n], err = fmt.Sprintf("%v", x), nil; err != nil {
					return nil, err
				}
			}
		}
	}
	return row, nil
}

// Prints the final version of a cell based on formatting variables and command line parameters
func (f *sqlCmdFormatterType) printColumnValue(val string, col int) {
	c := f.columnDetails[col]
	s := new(strings.Builder)
	if isNeedingControlCharacterTreatment(&c.col) {
		val = applyControlCharacterBehavior(val, f.ccb)
	}

	if isNeedingHexPrefix(&c.col) {
		val = "0x" + val
	}

	s.WriteString(val)
	r := []rune(val)
	if !f.removeTrailingSpaces {
		if f.vars.MaxVarColumnWidth() != 0 || !isLargeVariableType(&c.col) {
			padding := c.displayWidth - min64(c.displayWidth, int64(len(r)))
			if padding > 0 {
				if c.leftJustify {
					s = PadRight(s, padding, " ")
				} else {
					s = PadLeft(s, padding, " ")
				}
			}
		}
	}

	r = []rune(s.String())
	if c.displayWidth > 0 && int64(len(r)) > c.displayWidth {
		s.Reset()
		s.WriteString(string(r[:c.displayWidth]))
	}
	f.WriteOut(s.String())
}

func isLargeVariableType(col *sql.ColumnType) bool {
	l, _ := col.Length()
	switch col.DatabaseTypeName() {

	case "VARCHAR", "VARBINARY":
		return l > 8000
	case "NVARCHAR":
		return l > 4000
	case "XML", "TEXT", "NTEXT", "IMAGE":
		return true
	}
	return false
}

func isNeedingControlCharacterTreatment(col *sql.ColumnType) bool {
	switch col.DatabaseTypeName() {
	case "CHAR", "VARCHAR", "TEXT", "NTEXT", "NCHAR", "NVARCHAR", "XML":
		return true
	}
	return false
}
func isBinaryDataType(col *sql.ColumnType) bool {
	switch col.DatabaseTypeName() {
	case "BINARY", "VARBINARY":
		return true
	}
	return false
}

func isNeedingHexPrefix(col *sql.ColumnType) bool {
	return isBinaryDataType(col) // || col.DatabaseTypeName() == "UDT"
}

func isControlChar(r rune) bool {
	c := int(r)
	return c == 0x7f || (c >= 0 && c <= 0x1f)
}

func applyControlCharacterBehavior(val string, ccb ControlCharacterBehavior) string {
	if ccb == ControlIgnore {
		return val
	}
	b := new(strings.Builder)
	r := []rune(val)
	if ccb == ControlReplace {
		for _, l := range r {
			if isControlChar(l) {
				b.WriteRune(' ')
			} else {
				b.WriteRune(l)
			}
		}
	} else {
		for i := 0; i < len(r); {
			if !isControlChar(r[i]) {
				b.WriteRune(r[i])
				i++
			} else {
				for ; i < len(r) && isControlChar(r[i]); i++ {
				}
				if ccb == ControlReplaceConsecutive {
					b.WriteRune(' ')
				}
			}
		}
	}
	return b.String()
}

// Per https://docs.microsoft.com/sql/odbc/reference/appendixes/sql-to-c-binary
var hexDigits = []rune{'A', 'B', 'C', 'D', 'E', 'F'}

func decodeBinary(b []byte) string {

	s := new(strings.Builder)
	s.Grow(len(b) * 2)
	for _, ch := range b {
		b1 := ch >> 4
		b2 := ch & 0x0f
		if b1 >= 10 {
			s.WriteRune(hexDigits[b1-10])
		} else {
			s.WriteRune(rune('0' + b1))
		}
		if b2 >= 10 {
			s.WriteRune(hexDigits[b2-10])
		} else {
			s.WriteRune(rune('0' + b2))
		}
	}
	return s.String()
}
