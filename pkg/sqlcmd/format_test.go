// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"context"
	"strings"
	"testing"

	mssql "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-sqlcmd/internal/color"
	"github.com/stretchr/testify/assert"
)

func TestFitToScreen(t *testing.T) {
	type fitTest struct {
		width int64
		raw   string
		fit   string
	}

	tests := []fitTest{
		{0, "this is a string", "this is a string"},
		{9, "12345678", "12345678"},
		{9, "123456789", "123456789"},
		{9, "123456789A", "123456789" + SqlcmdEol + "A"},
		{9, "123456789" + SqlcmdEol, "123456789" + SqlcmdEol},
		{9, "12345678" + SqlcmdEol + "9A", "12345678" + SqlcmdEol + "9A"},
		{9, "123456789\rA", "123456789" + SqlcmdEol + "\rA"},
	}

	for _, test := range tests {

		line := new(strings.Builder)
		line.WriteString(test.raw)
		t.Log(test.raw)
		f := fitToScreen(line, test.width).String()
		assert.Equal(t, test.fit, f, "Mismatched fit for raw string: '%s'", test.raw)
	}
}

func TestCalcColumnDetails(t *testing.T) {
	type colTest struct {
		fixed    int64
		variable int64
		query    string
		details  []columnDetail
		max      int
	}

	tests := []colTest{
		{8, 8,
			"select 100 as '123456789ABC', getdate() as '987654321', 'string' as col1",
			[]columnDetail{
				{leftJustify: false, displayWidth: 12},
				{leftJustify: false, displayWidth: 23},
				{leftJustify: true, displayWidth: 6},
			},
			12,
		},
	}

	db, err := ConnectDb(t)
	if assert.NoError(t, err, "ConnectDB failed") {
		defer db.Close()
		for x, test := range tests {
			rows, err := db.QueryContext(context.Background(), test.query)
			if assert.NoError(t, err, "Query failed: %s", test.query) {
				defer rows.Close()
				cols, err := rows.ColumnTypes()
				if assert.NoError(t, err, "ColumnTypes failed:%s", test.query) {
					actual, max := calcColumnDetails(cols, test.fixed, test.variable)
					for i, a := range actual {
						if test.details[i].displayWidth != a.displayWidth ||
							test.details[i].leftJustify != a.leftJustify ||
							test.details[i].zeroesAfterDecimal != a.zeroesAfterDecimal {
							assert.Failf(t, "", "[%d] Incorrect test details for column [%s] in query '%s':%+v", x, cols[i].Name(), test.query, a)
						}
						assert.Equal(t, test.max, max, "[%d] Max column name length incorrect", x)
					}
				}
			}
		}
	}
}

func TestFormatFloat64UsesDatabaseTypePrecision(t *testing.T) {
	tests := []struct {
		name         string
		value        float64
		databaseType string
		expected     string
	}{
		{"REAL scientific notation", float64(float32(1e-20)), "REAL", "1e-20"},
		{"FLOAT keeps double precision", 1.2345678901234567, "FLOAT", "1.2345678901234567"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, formatFloat64(test.value, test.databaseType))
		})
	}
}

func TestScanRowFormatsFloatingPointTypesAtTheirPrecision(t *testing.T) {
	db, err := ConnectDb(t)
	if !assert.NoError(t, err, "ConnectDB failed") {
		return
	}
	defer db.Close()

	rows, err := db.QueryContext(context.Background(),
		"SELECT CAST(1E-20 AS REAL) AS r, CAST(1.2345678901234567 AS FLOAT) AS f")
	if !assert.NoError(t, err, "Query failed") {
		return
	}
	defer rows.Close()

	cols, err := rows.ColumnTypes()
	if !assert.NoError(t, err, "ColumnTypes failed") {
		return
	}
	details, _ := calcColumnDetails(cols, 0, 0)
	f := &sqlCmdFormatterType{columnDetails: details}
	if assert.True(t, rows.Next(), "Expected one row") {
		values, err := f.scanRow(rows)
		assert.NoError(t, err, "scanRow failed")
		assert.Equal(t, []string{"1e-20", "1.2345678901234567"}, values)
	}
}

func TestControlCharacterBehavior(t *testing.T) {
	type ccbTest struct {
		raw                 string
		replaced            string
		removed             string
		consecutivereplaced string
	}

	tests := []ccbTest{
		{"no control", "no control", "no control", "no control"},
		{string(rune(1)) + "tabs\t\treturns\r\n\r\n", " tabs  returns    ", "tabsreturns", " tabs returns "},
	}

	for _, test := range tests {
		s := applyControlCharacterBehavior(test.raw, ControlReplace)
		assert.Equalf(t, test.replaced, s, "Incorrect Replaced for '%s'", test.raw)
		s = applyControlCharacterBehavior(test.raw, ControlRemove)
		assert.Equalf(t, test.removed, s, "Incorrect Remove for '%s'", test.raw)
		s = applyControlCharacterBehavior(test.raw, ControlReplaceConsecutive)
		assert.Equalf(t, test.consecutivereplaced, s, "Incorrect ReplaceConsecutive for '%s'", test.raw)
	}
}

func TestDecodeBinary(t *testing.T) {
	type decodeTest struct {
		b []byte
		s string
	}

	tests := []decodeTest{
		{[]byte("123456ABCDEF"), "313233343536414243444546"},
		{[]byte{0x12, 0x34, 0x56}, "123456"},
	}
	for _, test := range tests {
		a := decodeBinary(test.b)
		assert.Equalf(t, test.s, a, "Incorrect decoded binary string for %v", test.b)
	}
}

func BenchmarkDecodeBinary(b *testing.B) {
	b.ReportAllocs()
	bytes := make([]byte, 10000)
	for i := 0; i < 10000; i++ {
		bytes[i] = byte(i % 0xff)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := decodeBinary(bytes)
		if len(s) != 20000 {
			b.Fatalf("Incorrect length of returned string. Should be 20k, was %d", len(s))
		}
	}
}

func TestFormatterColorizer(t *testing.T) {

	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	s.vars.Set(SQLCMDCOLORSCHEME, "emacs")
	s.Format.(*sqlCmdFormatterType).colorizer = color.New(true)
	err := runSqlCmd(t, s, []string{"select 'name' as name", "GO"})
	assert.NoError(t, err, "runSqlCmd returned error")
	output := buf.buf.String()
	// Verify the colorized output contains ANSI escape codes and expected content
	assert.Contains(t, output, "\x1b[", "output should contain ANSI escape codes")
	assert.Contains(t, output, "name", "output should contain column value")
	assert.Contains(t, output, "(1 row affected)", "output should contain row count")
}

func TestFormatterXmlMode(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	s.Format.XmlMode(true)
	err := runSqlCmd(t, s, []string{"select name from sys.databases where name='master' for xml auto ", "GO"})
	assert.NoError(t, err, "runSqlCmd returned error")
	assert.Equal(t, `<sys.databases name="master"/>`+SqlcmdEol, buf.buf.String())
}

func TestAddErrorStripsMssqlPrefixByDefault(t *testing.T) {
	out, errOut := new(strings.Builder), new(strings.Builder)
	vars := InitializeVariables(false)
	f := NewSQLCmdDefaultFormatter(vars, false, ControlIgnore)
	f.BeginBatch("", vars, out, errOut)

	f.AddError(mssql.Error{Number: 50000, State: 1, Class: 16, Message: "Something failed", ServerName: "server", LineNo: 7})

	got := errOut.String()
	assert.Contains(t, got, "Msg 50000, Level 16, State 1, Server server, Line 7")
	assert.Contains(t, got, "Something failed")
	assert.NotContains(t, got, "mssql:")
	assert.NotContains(t, got, "(50000)")
}

func TestAddErrorWithRawErrorsKeepsMssqlPrefix(t *testing.T) {
	out, errOut := new(strings.Builder), new(strings.Builder)
	vars := InitializeVariables(false)
	f := NewSQLCmdDefaultFormatter(vars, false, ControlIgnore, WithRawErrors(true))
	f.BeginBatch("", vars, out, errOut)

	f.AddError(mssql.Error{Number: 50000, State: 1, Class: 16, Message: "Something failed", ServerName: "server", LineNo: 7})

	got := errOut.String()
	assert.Contains(t, got, "Msg 50000, Level 16, State 1, Server server, Line 7")
	assert.Contains(t, got, "mssql: Something failed")
}

func TestAddErrorWithRawErrorsAppliesToAsciiFormatter(t *testing.T) {
	out, errOut := new(strings.Builder), new(strings.Builder)
	vars := InitializeVariables(false)
	vars.Set(SQLCMDFORMAT, "ascii")
	f := NewSQLCmdDefaultFormatter(vars, false, ControlIgnore, WithRawErrors(true))
	f.BeginBatch("", vars, out, errOut)

	f.AddError(mssql.Error{Number: 50000, State: 1, Class: 16, Message: "Something failed", ServerName: "server", LineNo: 7})

	assert.Contains(t, errOut.String(), "mssql: Something failed", "ascii formatter must honor WithRawErrors")
}
