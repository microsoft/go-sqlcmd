// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-sqlcmd/internal/color"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
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
	// AddRow is called for each row in a result set. It returns the value of the first column
	AddRow(*sql.Rows) string
	// AddMessage is called for every information message returned by the server during the batch
	AddMessage(string)
	// AddError is called for each error encountered during batch execution
	AddError(err error)
	// XmlMode enables or disables XML rendering mode
	XmlMode(enable bool)
	// IsXmlMode returns whether XML mode is enabled
	IsXmlMode() bool
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

const (
	// Default display widths for float types
	realDefaultWidth  int64 = 14 // For REAL and SMALLMONEY
	floatDefaultWidth int64 = 24 // For FLOAT and MONEY
)

type columnDetail struct {
	displayWidth       int64
	leftJustify        bool
	zeroesAfterDecimal bool
	col                sql.ColumnType
	precision          int
	scale              int
}

// The default formatter based on the native sqlcmd style
// It supports both horizontal (default) and vertical layout for results.
// Both vertical and horizontal layouts respect column widths set by SQLCMD variables.
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
	format               string
	maxColNameLen        int
	colorizer            color.Colorizer
	xml                  bool
}

// NewSQLCmdDefaultFormatter returns a Formatter that mimics the original ODBC-based sqlcmd formatter
func NewSQLCmdDefaultFormatter(removeTrailingSpaces bool, ccb ControlCharacterBehavior) Formatter {
	return &sqlCmdFormatterType{
		removeTrailingSpaces: removeTrailingSpaces,
		format:               "horizontal",
		colorizer:            color.New(false),
		ccb:                  ccb,
	}
}

// Adds the given string to the current line, wrapping it based on the screen width setting
func (f *sqlCmdFormatterType) writeOut(s string, t color.TextType) {
	w := f.vars.ScreenWidth()
	if w == 0 {
		f.mustWriteOut(s, t)
		return
	}

	r := []rune(s)
	for i := 0; true; {
		if i == len(r) {
			f.mustWriteOut(string(r), t)
			return
		} else if f.writepos == w {
			f.mustWriteOut(string(r[:i]), t)
			f.mustWriteOut(SqlcmdEol, color.TextTypeNormal)
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

// BeginBatch stores the settings to use for processing the current batch
// TODO: add a third io.Writer for messages when we add -r support
func (f *sqlCmdFormatterType) BeginBatch(_ string, vars *Variables, out io.Writer, err io.Writer) {
	f.out = out
	f.err = err
	f.vars = vars
	f.colsep = vars.ColumnSeparator()
	f.format = vars.Format()
}

func (f *sqlCmdFormatterType) EndBatch() {
}

// Calculate the widths for each column and print the column names
// Since sql.ColumnType only provides sizes for variable length types we will
// base our numbers for most types on https://docs.microsoft.com/sql/odbc/reference/appendixes/column-size
func (f *sqlCmdFormatterType) BeginResultSet(cols []*sql.ColumnType) {
	f.rowcount = 0
	f.columnDetails, f.maxColNameLen = calcColumnDetails(cols, f.vars.MaxFixedColumnWidth(), f.vars.MaxVarColumnWidth())
	if f.vars.RowsBetweenHeaders() > -1 && f.format == "horizontal" && !f.xml {
		f.printColumnHeadings()
	}
}

// EndResultSet writes a blank line to the designated output writer
func (f *sqlCmdFormatterType) EndResultSet() {
	if !f.xml {
		f.writeOut(SqlcmdEol, color.TextTypeNormal)
	}
}

// AddRow writes the current row to the designated output writer
func (f *sqlCmdFormatterType) AddRow(row *sql.Rows) string {
	retval := ""
	values, err := f.scanRow(row)
	if err != nil {
		f.mustWriteErr(err.Error())
		return retval
	}
	retval = values[0]
	if f.xml {
		f.printColumnValue(retval, 0)
	} else if f.format == "horizontal" {
		// values are the full values, look at the displaywidth of each column and truncate accordingly
		for i, v := range values {
			if i > 0 {
				f.writeOut(f.vars.ColumnSeparator(), color.TextTypeSeparator)
			}
			f.printColumnValue(v, i)
		}
		f.rowcount++
		gap := f.vars.RowsBetweenHeaders()
		if gap > 0 && (int64(f.rowcount)%gap == 0) {
			f.writeOut(SqlcmdEol, color.TextTypeNormal)
			f.printColumnHeadings()
		}
	} else {
		f.addVerticalRow(values)
	}
	f.writeOut(SqlcmdEol, color.TextTypeNormal)
	return retval
}

func (f *sqlCmdFormatterType) addVerticalRow(values []string) {
	for i, v := range values {
		if f.vars.RowsBetweenHeaders() > -1 {
			builder := new(strings.Builder)
			name := f.columnDetails[i].col.Name()
			builder.WriteString(name)
			builder = padRight(builder, int64(f.maxColNameLen-len(name)+1), " ")
			f.writeOut(builder.String(), color.TextTypeHeader)
		}
		f.printColumnValue(v, i)
		f.writeOut(SqlcmdEol, color.TextTypeNormal)
	}
}

// AddMessage writes a non-error message to the designated message writer
func (f *sqlCmdFormatterType) AddMessage(msg string) {
	if !f.xml {
		f.mustWriteOut(msg+SqlcmdEol, color.TextTypeWarning)
	}
}

// AddError writes an error to the designated err Writer
func (f *sqlCmdFormatterType) AddError(err error) {
	print := true
	b := new(strings.Builder)
	if errors.Is(err, context.DeadlineExceeded) {
		err = localizer.Errorf("Timeout expired")
	}
	msg := err.Error()
	switch e := (err).(type) {
	case mssql.Error:
		if print = f.vars.ErrorLevel() <= 0 || e.Class >= uint8(f.vars.ErrorLevel()); print {
			if len(e.ProcName) > 0 {
				b.WriteString(localizer.Sprintf("Msg %#v, Level %d, State %d, Server %s, Procedure %s, Line %#v%s", e.Number, e.Class, e.State, e.ServerName, e.ProcName, e.LineNo, SqlcmdEol))
			} else {
				b.WriteString(localizer.Sprintf("Msg %#v, Level %d, State %d, Server %s, Line %#v%s", e.Number, e.Class, e.State, e.ServerName, e.LineNo, SqlcmdEol))
			}
			msg = strings.TrimPrefix(msg, "mssql: ")
		}
	}
	if print {
		b.WriteString(msg)
		b.WriteString(SqlcmdEol)
		f.mustWriteErr(fitToScreen(b, f.vars.ScreenWidth()).String())
	}
}

// XmlMode enables or disables XML mode
func (f *sqlCmdFormatterType) XmlMode(enable bool) {
	f.xml = enable
}

// IsXmlMode returns whether XML mode is enabled
func (f *sqlCmdFormatterType) IsXmlMode() bool {
	return f.xml
}

// Prints column headings based on columnDetail, variables, and command line arguments
func (f *sqlCmdFormatterType) printColumnHeadings() {
	names := new(strings.Builder)
	sep := new(strings.Builder)

	var leftPad, rightPad int64
	for i, c := range f.columnDetails {
		rightPad = 0
		nameLen := int64(len([]rune(c.col.Name())))
		if f.removeTrailingSpaces {
			if nameLen == 0 {
				// special case for unnamed columns when using -W
				// print a single -
				rightPad = 1
				sep = padRight(sep, 1, "-")
			} else {
				sep = padRight(sep, nameLen, "-")
			}
		} else {
			length := min64(c.displayWidth, maxPadWidth)
			if nameLen < length {
				rightPad = length - nameLen
			}
			sep = padRight(sep, length, "-")
		}
		names = padRight(names, leftPad, " ")
		names.WriteString(c.col.Name()[:min64(nameLen, c.displayWidth)])
		names = padRight(names, rightPad, " ")
		if i != len(f.columnDetails)-1 {
			names.WriteString(f.colsep)
			sep.WriteString(f.colsep)
		}
	}
	names.WriteString(SqlcmdEol)
	sep.WriteString(SqlcmdEol)
	names = fitToScreen(names, f.vars.ScreenWidth())
	sep = fitToScreen(sep, f.vars.ScreenWidth())
	f.mustWriteOut(names.String(), color.TextTypeHeader)
	f.mustWriteOut(sep.String(), color.TextTypeSeparator)
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
// Return an array of columnDetail objects describing the output format for each column.
// Return the length of the longest column name.
func calcColumnDetails(cols []*sql.ColumnType, fixed int64, variable int64) ([]columnDetail, int) {
	columnDetails := make([]columnDetail, len(cols))
	maxNameLen := 0
	for i, c := range cols {
		length, _ := c.Length()
		nameLen := int64(len([]rune(c.Name())))
		if nameLen > int64(maxNameLen) {
			maxNameLen = int(nameLen)
		}
		columnDetails[i].col = *c
		columnDetails[i].leftJustify = true
		columnDetails[i].zeroesAfterDecimal = false
		p, s, ok := c.DecimalSize()
		if ok {
			columnDetails[i].precision = int(p)
			columnDetails[i].scale = int(s)
		}
		if length == 0 {
			columnDetails[i].displayWidth = defaultMaxDisplayWidth
		} else {
			columnDetails[i].displayWidth = length
		}
		typeName := c.DatabaseTypeName()

		switch typeName {
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
		case "REAL", "SMALLMONEY":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(realDefaultWidth, nameLen)
			columnDetails[i].zeroesAfterDecimal = true
		case "FLOAT", "MONEY":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(floatDefaultWidth, nameLen)
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
		case "TIME":
			columnDetails[i].leftJustify = false
			columnDetails[i].displayWidth = max64(16, nameLen)
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
		case "SQL_VARIANT":
			if fixed > 0 {
				columnDetails[i].displayWidth = min64(fixed, 8000)
			} else {
				columnDetails[i].displayWidth = 8000
			}
		// Fixed length types
		case "CHAR", "NCHAR":
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
	return columnDetails, maxNameLen
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
				} else if f.columnDetails[n].col.DatabaseTypeName() == "UNIQUEIDENTIFIER" {
					// Unscramble the guid
					// see https://github.com/denisenkom/go-mssqldb/issues/56
					x[0], x[1], x[2], x[3] = x[3], x[2], x[1], x[0]
					x[4], x[5] = x[5], x[4]
					x[6], x[7] = x[7], x[6]
					if guid, err := uuid.FromBytes(x); err == nil {
						row[n] = guid.String()
					} else {
						// this should never happen
						row[n] = uuid.New().String()
					}
				} else {
					row[n] = string(x)
				}
			case string:
				row[n] = x
			case time.Time:
				// Go lacks any way to get the user's preferred time format or even the system default
				switch f.columnDetails[n].col.DatabaseTypeName() {
				case "DATE":
					row[n] = x.Format("2006-01-02")
				case "DATETIME":
					row[n] = x.Format(dateTimeFormatString(3, false))
				case "DATETIME2":
					row[n] = x.Format(dateTimeFormatString(f.columnDetails[n].scale, false))
				case "SMALLDATETIME":
					row[n] = x.Format(dateTimeFormatString(0, false))
				case "DATETIMEOFFSET":
					row[n] = x.Format(dateTimeFormatString(f.columnDetails[n].scale, true))
				case "TIME":
					format := "15:04:05"
					if f.columnDetails[n].scale > 0 {
						format = fmt.Sprintf("%s.%0*d", format, f.columnDetails[n].scale, 0)
					}
					row[n] = x.Format(format)
				default:
					row[n] = x.Format(time.RFC3339)
				}
			case fmt.Stringer:
				row[n] = x.String()
			// not sure why go-mssql reports bit as bool
			case bool:
				if x {
					row[n] = "1"
				} else {
					row[n] = "0"
				}
			case float64:
				// Format float64 to match ODBC sqlcmd behavior
				// Use 'f' format with -1 precision to avoid scientific notation for typical values
				// Fall back to 'g' format if the result would exceed the column display width

				// Use appropriate bitSize based on the SQL type (REAL=32, FLOAT=64)
				// REAL columns should use 32-bit precision even though the value is scanned as float64
				bitSize := 64
				typeName := f.columnDetails[n].col.DatabaseTypeName()
				if typeName == "REAL" || typeName == "SMALLMONEY" {
					bitSize = 32
				}

				formatted := strconv.FormatFloat(x, 'f', -1, bitSize)
				displayWidth := f.columnDetails[n].displayWidth

				// Use the type's default display width when displayWidth is 0 (unlimited)
				// to avoid extremely long strings for extreme values
				widthThreshold := displayWidth
				if widthThreshold == 0 {
					if typeName == "REAL" || typeName == "SMALLMONEY" {
						widthThreshold = realDefaultWidth
					} else {
						widthThreshold = floatDefaultWidth
					}
				}

				if int64(len(formatted)) > widthThreshold {
					// Use 'g' format for very large/small values to avoid truncation issues
					formatted = strconv.FormatFloat(x, 'g', -1, bitSize)
				}
				row[n] = formatted
			case float32:
				// Format float32 to match ODBC sqlcmd behavior
				// float32 values are rare (database/sql typically normalizes to float64)
				// Use bitSize 32 to maintain precision appropriate for the original float32 value
				formatted := strconv.FormatFloat(float64(x), 'f', -1, 32)
				displayWidth := f.columnDetails[n].displayWidth

				// Use default REAL display width when displayWidth is 0
				widthThreshold := displayWidth
				if widthThreshold == 0 {
					widthThreshold = realDefaultWidth
				}

				if int64(len(formatted)) > widthThreshold {
					// Use 'g' format for very large/small values to avoid truncation issues
					formatted = strconv.FormatFloat(float64(x), 'g', -1, 32)
				}
				row[n] = formatted
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

func dateTimeFormatString(scale int, addOffset bool) string {
	format := `2006-01-02 15:04:05`
	if scale > 0 {
		format = fmt.Sprintf("%s.%0*d", format, scale, 0)
	}
	if addOffset {
		format += " -07:00"
	}
	return format
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
	if !f.xml && f.format == "horizontal" {
		if !f.removeTrailingSpaces {
			if f.vars.MaxVarColumnWidth() != 0 || !isLargeVariableType(&c.col) {
				padding := c.displayWidth - min64(c.displayWidth, int64(len(r)))
				if padding > 0 {
					if c.leftJustify {
						s = padRight(s, padding, " ")
					} else {
						s = padLeft(s, padding, " ")
					}
				}
			}
		}

		r = []rune(s.String())
	}
	if !f.xml && (c.displayWidth > 0 && int64(len(r)) > c.displayWidth) {
		s.Reset()
		s.WriteString(string(r[:c.displayWidth]))
	}
	clr := color.TextTypeCell
	if f.xml {
		clr = color.TextTypeXml
	}
	f.writeOut(s.String(), clr)
}

func (f *sqlCmdFormatterType) mustWriteOut(s string, t color.TextType) {
	err := f.colorizer.Write(f.out, s, f.vars.ColorScheme(), t)
	if err != nil {
		panic(err)
	}
}

func (f *sqlCmdFormatterType) mustWriteErr(s string) {
	err := f.colorizer.Write(f.err, s, f.vars.ColorScheme(), color.TextTypeError)
	if err != nil {
		panic(err)
	}
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
