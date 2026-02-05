// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"context"
	"strings"
	"testing"

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

func TestFormatterFloatFormatting(t *testing.T) {
	// Test that float formatting matches ODBC sqlcmd behavior
	// This addresses the issue where go-sqlcmd was using scientific notation
	// while ODBC sqlcmd uses decimal notation
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer func() { _ = buf.Close() }()

	// Set SQLCMDMAXVARTYPEWIDTH to a non-zero value so FLOAT columns use the 24-char display width
	// This enables the width-based fallback logic to be tested properly
	s.vars.Set(SQLCMDMAXVARTYPEWIDTH, "256")

	// Test query with float values from the issue
	query := `SELECT 
		CAST(788991.19988463481 AS FLOAT) as Longitude1,
		CAST(4713347.3103808956 AS FLOAT) as Latitude1,
		CAST(789288.40771771886 AS FLOAT) as Longitude2,
		CAST(4712632.075629076 AS FLOAT) as Latitude2,
		CAST(788569.36558582436 AS FLOAT) as Longitude3,
		CAST(4714608.0418091472 AS FLOAT) as Latitude3`

	err := runSqlCmd(t, s, []string{query, "GO"})
	assert.NoError(t, err, "runSqlCmd returned error")

	output := buf.buf.String()

	// Verify that the output contains decimal notation, not scientific notation
	// Scientific notation would look like "4.713347310380896e+06"
	// Decimal notation should look like "4713347.3103808956"
	assert.NotContains(t, output, "e+", "Output should not contain scientific notation (e+)")
	assert.NotContains(t, output, "E+", "Output should not contain scientific notation (E+)")

	// Verify that specific expected values are present (allowing for precision differences)
	assert.Contains(t, output, "788991.1998846", "Output should contain decimal representation of Longitude1")
	assert.Contains(t, output, "4713347.310380", "Output should contain decimal representation of Latitude1")
	assert.Contains(t, output, "789288.4077177", "Output should contain decimal representation of Longitude2")
	assert.Contains(t, output, "4712632.075629", "Output should contain decimal representation of Latitude2")
	assert.Contains(t, output, "788569.3655858", "Output should contain decimal representation of Longitude3")
	assert.Contains(t, output, "4714608.041809", "Output should contain decimal representation of Latitude3")
}

func TestFormatterFloatFormattingExtremeValues(t *testing.T) {
	// Test that extreme float values fall back to scientific notation
	// to avoid truncation issues with very large or very small numbers
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer func() { _ = buf.Close() }()

	// Set SQLCMDMAXVARTYPEWIDTH to a non-zero value so FLOAT columns use the 24-char display width
	// This allows the fallback behavior to be tested
	s.vars.Set(SQLCMDMAXVARTYPEWIDTH, "256")

	// Test query with extreme float values that would exceed the 24-char display width
	query := `SELECT 
		CAST(1e100 AS FLOAT) as VeryLarge,
		CAST(1e-100 AS FLOAT) as VerySmall`

	err := runSqlCmd(t, s, []string{query, "GO"})
	assert.NoError(t, err, "runSqlCmd returned error")

	output := buf.buf.String()

	// Verify that extremely large values use scientific notation with positive exponent
	// (because decimal format would exceed the 24-char column width)
	assert.Contains(t, output, "e+", "Output should contain scientific notation (e+) for very large values")

	// Verify that extremely small values use scientific notation with negative exponent
	assert.Contains(t, output, "e-", "Output should contain scientific notation (e-) for very small values")
}

func TestFormatterFloatFormattingExtremeValuesUnlimitedWidth(t *testing.T) {
	// Test that extreme float values fall back to scientific notation even when
	// displayWidth is 0 (unlimited), using type default widths as threshold
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer func() { _ = buf.Close() }()

	// Leave SQLCMDMAXVARTYPEWIDTH at 0 (setupSqlCmdWithMemoryOutput default)
	// This sets displayWidth to 0 for all columns, testing the fallback logic
	// that uses type default widths (24 for FLOAT, 14 for REAL)

	// Test query with extreme float values
	query := `SELECT 
		CAST(1e100 AS FLOAT) as VeryLarge,
		CAST(1e-100 AS FLOAT) as VerySmall`

	err := runSqlCmd(t, s, []string{query, "GO"})
	assert.NoError(t, err, "runSqlCmd returned error")

	output := buf.buf.String()

	// Verify that extreme values still use scientific notation even with unlimited width
	// (fallback should use type default widths to prevent unbounded output)
	assert.Contains(t, output, "e+", "Output should contain scientific notation (e+) for very large values even with unlimited width")
	assert.Contains(t, output, "e-", "Output should contain scientific notation (e-) for very small values even with unlimited width")
}

func TestFormatterRealFormatting(t *testing.T) {
	// Test that REAL (float32) values use decimal notation for typical values
	// and fall back to scientific notation for extreme values
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer func() { _ = buf.Close() }()

	// Set SQLCMDMAXVARTYPEWIDTH to a non-zero value so REAL columns use the 14-char display width
	s.vars.Set(SQLCMDMAXVARTYPEWIDTH, "256")

	// Test query with REAL values (both typical and extreme)
	query := `SELECT 
		CAST(123.456789 AS REAL) as TypicalValue,
		CAST(1e30 AS REAL) as ExtremeValue`

	err := runSqlCmd(t, s, []string{query, "GO"})
	assert.NoError(t, err, "runSqlCmd returned error")

	output := buf.buf.String()

	// Split output into lines to examine the data row separately from headers
	lines := strings.Split(output, SqlcmdEol)
	var dataLine string
	for _, line := range lines {
		// Find the data line (contains actual values, not headers or separators)
		if strings.Contains(line, "123.") {
			dataLine = line
			break
		}
	}

	// Verify that typical REAL values use decimal notation (not scientific)
	assert.Contains(t, dataLine, "123.456", "Output should contain decimal representation of typical REAL value")

	// Check that the typical value portion doesn't use scientific notation
	// Parse columns using whitespace (the default column separator)
	fields := strings.Fields(dataLine)
	if len(fields) > 0 {
		// Find the field containing the typical value
		for _, field := range fields {
			if strings.Contains(field, "123.") {
				assert.NotContains(t, field, "e", "Typical REAL value should not use scientific notation")
				break
			}
		}
	}

	// Verify that extreme REAL values use scientific notation
	assert.Contains(t, output, "e+", "Output should contain scientific notation for extreme REAL value")
}
