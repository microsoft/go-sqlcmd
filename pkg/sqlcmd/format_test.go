// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"strings"
	"testing"

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
			rows, err := db.Query(test.query)
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
