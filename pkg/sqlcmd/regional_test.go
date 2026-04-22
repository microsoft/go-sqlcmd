// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestRegionalSettings_Disabled(t *testing.T) {
	r := NewRegionalSettings(false)
	assert.False(t, r.IsEnabled())

	// When disabled, all values should pass through unchanged
	assert.Equal(t, "1234.56", r.FormatNumber("1234.56"))
	assert.Equal(t, "1234.5600", r.FormatMoney("1234.5600"))

	testTime := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	assert.Equal(t, "2024-01-15", r.FormatDate(testTime))
	assert.Equal(t, "14:30:45", r.FormatTime(testTime, 0))
}

func TestRegionalSettings_Enabled(t *testing.T) {
	r := NewRegionalSettings(true)
	assert.True(t, r.IsEnabled())

	// When enabled, values should be formatted according to locale
	// The specific format depends on the system locale, so we just verify it works
	number := r.FormatNumber("1234.56")
	assert.NotEmpty(t, number)

	money := r.FormatMoney("1234.5600")
	assert.NotEmpty(t, money)
}

func TestRegionalSettings_NullHandling(t *testing.T) {
	r := NewRegionalSettings(true)

	// NULL values should pass through unchanged
	assert.Equal(t, "NULL", r.FormatNumber("NULL"))
	assert.Equal(t, "NULL", r.FormatMoney("NULL"))
	assert.Equal(t, "", r.FormatNumber(""))
	assert.Equal(t, "", r.FormatMoney(""))
}

func TestGetDecimalSeparator(t *testing.T) {
	tests := []struct {
		locale   string
		expected string
	}{
		{"en-US", "."},
		{"en-GB", "."},
		{"de-DE", ","},
		{"de-CH", "."},  // Swiss German uses . for decimal
		{"fr-FR", ","},
		{"fr-CH", ","},  // Swiss French keeps comma
		{"it-CH", "."},  // Swiss Italian uses . for decimal
		{"es-ES", ","},
		{"ja-JP", "."},
		{"zh-CN", "."},
	}

	for _, tc := range tests {
		t.Run(tc.locale, func(t *testing.T) {
			tag := language.MustParse(tc.locale)
			sep := getDecimalSeparator(tag)
			assert.Equal(t, tc.expected, sep, "Decimal separator for %s", tc.locale)
		})
	}
}

func TestGetThousandSeparator(t *testing.T) {
	tests := []struct {
		locale   string
		expected string
	}{
		{"en-US", ","},
		{"en-GB", ","},
		{"de-DE", "."},
		{"de-CH", "\u2019"}, // Swiss German uses typographic apostrophe
		{"fr-FR", "\u00a0"},
		{"fr-CH", "\u2019"}, // Swiss French uses typographic apostrophe
		{"sv-SE", "\u00a0"},
		{"ru-RU", "\u00a0"},
		{"ja-JP", ","},
	}

	for _, tc := range tests {
		t.Run(tc.locale, func(t *testing.T) {
			tag := language.MustParse(tc.locale)
			sep := getThousandSeparator(tag)
			assert.Equal(t, tc.expected, sep, "Thousand separator for %s", tc.locale)
		})
	}
}

func TestAddThousandSeparators(t *testing.T) {
	enUS := language.MustParse("en-US")
	deDE := language.MustParse("de-DE")

	tests := []struct {
		input    string
		locale   language.Tag
		expected string
	}{
		{"1", enUS, "1"},
		{"12", enUS, "12"},
		{"123", enUS, "123"},
		{"1234", enUS, "1,234"},
		{"12345", enUS, "12,345"},
		{"123456", enUS, "123,456"},
		{"1234567", enUS, "1,234,567"},
		{"1234567890", enUS, "1,234,567,890"},
		{"1234", deDE, "1.234"},
		{"1234567", deDE, "1.234.567"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := addThousandSeparators(tc.input, tc.locale)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetLocaleDateTimeFormats(t *testing.T) {
	tests := []struct {
		locale      string
		wantDate    string
		wantTime    string
		description string
	}{
		{"en-US", "01/02/2006", "03:04:05 PM", "US English uses M/D/Y and 12-hour time"},
		{"en-GB", "02/01/2006", "15:04:05", "UK English uses D/M/Y and 24-hour time"},
		{"de-DE", "02.01.2006", "15:04:05", "German uses D.M.Y format"},
		{"ja-JP", "2006/01/02", "15:04:05", "Japanese uses Y/M/D format"},
		{"fi-FI", "2006-01-02", "15:04:05", "Finnish uses ISO format"},
	}

	for _, tc := range tests {
		t.Run(tc.locale, func(t *testing.T) {
			tag := language.MustParse(tc.locale)
			dateFmt, timeFmt := getLocaleDateTimeFormats(tag)
			assert.Equal(t, tc.wantDate, dateFmt, tc.description)
			assert.Equal(t, tc.wantTime, timeFmt, tc.description)
		})
	}
}

func TestFormatOffset(t *testing.T) {
	tests := []struct {
		hours    int
		minutes  int
		expected string
	}{
		{0, 0, "+00:00"},
		{5, 30, "+05:30"},
		{-5, 0, "-05:00"},
		{-8, 0, "-08:00"},
		{12, 45, "+12:45"},
		{-12, 0, "-12:00"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatOffset(tc.hours, tc.minutes)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPow10(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 1},
		{1, 10},
		{2, 100},
		{3, 1000},
		{6, 1000000},
		{-1, 1},  // negative clamped to 1
		{-5, 1},  // negative clamped to 1
	}

	for _, tc := range tests {
		result := pow10(tc.n)
		assert.Equal(t, tc.expected, result)
	}
}

func TestPadLeftStr(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		pad      rune
		expected string
	}{
		{"5", 2, '0', "05"},
		{"12", 2, '0', "12"},
		{"1", 4, '0', "0001"},
		{"abc", 5, ' ', "  abc"},
	}

	for _, tc := range tests {
		result := padLeftStr(tc.input, tc.length, tc.pad)
		assert.Equal(t, tc.expected, result)
	}
}

func TestNewSQLCmdDefaultFormatterWithRegional(t *testing.T) {
	// Test that the formatter is created correctly with regional settings
	f := NewSQLCmdDefaultFormatterWithRegional(false, ControlIgnore, true)
	assert.NotNil(t, f)

	// Test without regional settings
	f2 := NewSQLCmdDefaultFormatterWithRegional(false, ControlIgnore, false)
	assert.NotNil(t, f2)

	// Test backward compatibility - NewSQLCmdDefaultFormatter should work
	f3 := NewSQLCmdDefaultFormatter(false, ControlIgnore)
	assert.NotNil(t, f3)
}

func TestFormatMoneyRounding(t *testing.T) {
	r := &RegionalSettings{enabled: true, tag: language.MustParse("en-US")}

	// Exact 4 digits - no rounding needed
	assert.Equal(t, "1.2345", r.FormatMoney("1.2345"))

	// More than 4 digits - round
	assert.Equal(t, "1.2346", r.FormatMoney("1.23456"))  // 5th digit >= 5, round up
	assert.Equal(t, "1.2345", r.FormatMoney("1.23454"))  // 5th digit < 5, truncate
	assert.Equal(t, "2.0000", r.FormatMoney("1.99999"))  // carry propagates to integer
	assert.Equal(t, "10.0000", r.FormatMoney("9.99999")) // carry propagates, integer grows
}

func TestIncrementIntString(t *testing.T) {
	assert.Equal(t, "2", incrementIntString("1"))
	assert.Equal(t, "10", incrementIntString("9"))
	assert.Equal(t, "100", incrementIntString("99"))
	assert.Equal(t, "1000", incrementIntString("999"))
	assert.Equal(t, "124", incrementIntString("123"))
}
