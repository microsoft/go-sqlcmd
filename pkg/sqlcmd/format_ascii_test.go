// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"bytes"
	"database/sql"
	"reflect"
	"testing"
	"unsafe"

	"github.com/microsoft/go-sqlcmd/internal/color"
	"github.com/stretchr/testify/assert"
)

func setColumnInfo(c *sql.ColumnType, name string, dbType string) {
	v := reflect.ValueOf(c).Elem()
	fName := v.FieldByName("name")
	if fName.IsValid() {
		reflect.NewAt(fName.Type(), unsafe.Pointer(fName.UnsafeAddr())).Elem().SetString(name)
	}
	fType := v.FieldByName("databaseType")
	if fType.IsValid() {
		reflect.NewAt(fType.Type(), unsafe.Pointer(fType.UnsafeAddr())).Elem().SetString(dbType)
	}
}

func TestAsciiFormatter(t *testing.T) {
	vars := InitializeVariables(false)
	vars.Set(SQLCMDFORMAT, "ascii")

	buf := new(bytes.Buffer)
	f := &asciiFormatter{
		sqlCmdFormatterType: &sqlCmdFormatterType{
			vars:      vars,
			out:       buf,
			colorizer: color.New(false),
			format:    "ascii",
		},
		rows:      [][]string{{"1", "test"}},
		colWidths: []int{2, 4},
	}

	// Mock column details
	f.columnDetails = make([]columnDetail, 2)
	setColumnInfo(&f.columnDetails[0].col, "id", "INT")
	setColumnInfo(&f.columnDetails[1].col, "name", "VARCHAR")

	f.printAsciiTable()

	expected := `+----+------+` + SqlcmdEol +
		`| id | name |` + SqlcmdEol +
		`+----+------+` + SqlcmdEol +
		`|  1 | test |` + SqlcmdEol +
		`+----+------+` + SqlcmdEol

	assert.Equal(t, expected, buf.String())
}

func TestAsciiFormatterWrapping(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	if s.db == nil {
		t.Skip("No database connection available")
	}
	defer buf.Close()

	s.vars.Set(SQLCMDFORMAT, "ascii")
	s.vars.Set(SQLCMDCOLWIDTH, "20") // Small width to force wrapping
	s.Format = NewSQLCmdDefaultFormatter(s.vars, false, ControlIgnore)

	// Select 3 columns that won't fit in 20 chars
	err := runSqlCmd(t, s, []string{"select 1 as id, 'test' as name, '0123456789' as descr", "GO"})
	assert.NoError(t, err, "runSqlCmd returned error")

	expectedPart1 := `+----+------+` + SqlcmdEol +
		`| id | name |` + SqlcmdEol +
		`+----+------+` + SqlcmdEol +
		`|  1 | test |` + SqlcmdEol +
		`+----+------+` + SqlcmdEol

	expectedPart2 := `+------------+` + SqlcmdEol +
		`| descr      |` + SqlcmdEol +
		`+------------+` + SqlcmdEol +
		`| 0123456789 |` + SqlcmdEol +
		`+------------+` + SqlcmdEol +
		`(1 row affected)` + SqlcmdEol

	assert.Contains(t, buf.buf.String(), expectedPart1)
	assert.Contains(t, buf.buf.String(), expectedPart2)
}

func TestAsciiFormatterTruncation(t *testing.T) {
	vars := InitializeVariables(false)
	vars.Set(SQLCMDCOLWIDTH, "20")

	buf := new(bytes.Buffer)
	f := &asciiFormatter{
		sqlCmdFormatterType: &sqlCmdFormatterType{
			vars:      vars,
			out:       buf,
			colorizer: color.New(false),
			format:    "ascii",
		},
		rows:      [][]string{{"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}}, // 50 chars
		colWidths: []int{50},
	}

	// Mock column details with empty column type (defaults to non-numeric, empty name)
	f.columnDetails = []columnDetail{{}}

	f.printAsciiTable()

	output := buf.String()

	// Expected behavior:
	// maxWidth = 20
	// maxColContentWidth = 20 - 4 = 16
	// colWidths[0] should be clamped to 16
	// The value should be truncated to 13 chars + "..." = 16 chars total.

	// Divider: + followed by 16 dashes + 2 dashes (padding) + +
	// Total width: 1 + 16 + 2 + 1 = 20
	// Divider line: +------------------+

	// Header: | <empty>            | (padded to 16)
	// Since name is empty.

	// Value: | AAAAAAAAAAAAA... | (13 A's followed by ...)

	expectedDivider := "+------------------+"
	expectedValue := "| AAAAAAAAAAAAA... |"

	assert.Contains(t, output, expectedDivider)
	assert.Contains(t, output, expectedValue)

	// Verify it does NOT contain the full string
	assert.NotContains(t, output, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
}
