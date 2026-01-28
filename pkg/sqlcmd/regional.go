// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/language"
)

// RegionalSettings provides locale-aware formatting for output when -R is used
type RegionalSettings struct {
	enabled bool
	tag     language.Tag
	dateFmt string
	timeFmt string
}

// NewRegionalSettings creates a new RegionalSettings instance
// If enabled is false, all format methods return values unchanged
func NewRegionalSettings(enabled bool) *RegionalSettings {
	r := &RegionalSettings{enabled: enabled}
	if enabled {
		r.tag = detectUserLocale()
		r.dateFmt, r.timeFmt = getLocaleDateTimeFormats(r.tag)
	}
	return r
}

// IsEnabled returns whether regional formatting is active
func (r *RegionalSettings) IsEnabled() bool {
	return r.enabled
}

// FormatNumber formats a numeric value with locale-specific thousand separators
// Used for DECIMAL and NUMERIC types. Formatting is done purely by string
// manipulation to preserve all digits of high-precision values.
func (r *RegionalSettings) FormatNumber(value string) string {
	if !r.enabled || value == "" || value == "NULL" {
		return value
	}

	// Handle leading sign
	negative := strings.HasPrefix(value, "-")
	if negative {
		value = value[1:]
	}

	// Split into integer and decimal parts using the SQL-style decimal point.
	// We do not change any digits; we only insert locale-specific separators.
	parts := strings.SplitN(value, ".", 2)
	intPart := parts[0]

	// Add thousand separators using locale convention (pure string manipulation)
	formatted := addThousandSeparators(intPart, r.tag)
	if len(parts) > 1 {
		formatted += getDecimalSeparator(r.tag) + parts[1]
	}
	if negative {
		formatted = "-" + formatted
	}
	return formatted
}

// FormatMoney formats a currency value with locale-specific formatting
// Used for MONEY and SMALLMONEY types
// Uses pure string manipulation to preserve precision for large values
func (r *RegionalSettings) FormatMoney(value string) string {
	if !r.enabled || value == "" || value == "NULL" {
		return value
	}

	// MONEY/SMALLMONEY are fixed-point with 4 decimal places.
	// Avoid float64 to prevent rounding of large values; format via string operations.
	negative := strings.HasPrefix(value, "-")
	cleanValue := value
	if negative {
		cleanValue = value[1:]
	}

	// Split into integer and fractional parts
	parts := strings.SplitN(cleanValue, ".", 2)
	intPart := parts[0]
	fracPart := ""
	if len(parts) > 1 {
		fracPart = parts[1]
	}

	// Normalize fractional part to exactly 4 digits, matching SQL Server MONEY display
	if len(fracPart) == 0 {
		fracPart = "0000"
	} else if len(fracPart) < 4 {
		fracPart = fracPart + strings.Repeat("0", 4-len(fracPart))
	} else if len(fracPart) > 4 {
		fracPart = fracPart[:4]
	}

	// Apply locale-specific thousand separators to the integer part
	formattedInt := addThousandSeparators(intPart, r.tag)

	// Combine with locale-specific decimal separator
	formatted := formattedInt + getDecimalSeparator(r.tag) + fracPart
	if negative {
		formatted = "-" + formatted
	}
	return formatted
}

// FormatDate formats a date value using locale-specific date format
// Used for DATE type
func (r *RegionalSettings) FormatDate(t time.Time) string {
	if !r.enabled {
		return t.Format("2006-01-02")
	}
	return t.Format(r.dateFmt)
}

// FormatDateTime formats a datetime value using locale-specific format
// Used for DATETIME, DATETIME2, SMALLDATETIME types
func (r *RegionalSettings) FormatDateTime(t time.Time, scale int, addOffset bool) string {
	if !r.enabled {
		return t.Format(dateTimeFormatString(scale, addOffset))
	}

	// Combine date and time in regional format
	datePart := t.Format(r.dateFmt)
	timePart := t.Format(r.timeFmt)

	result := datePart + " " + timePart
	if scale > 0 {
		// Add fractional seconds
		frac := t.Nanosecond() / (1000000000 / pow10(scale))
		result += getDecimalSeparator(r.tag) + padLeftStr(strconv.Itoa(frac), scale, '0')
	}
	if addOffset {
		_, offset := t.Zone()
		hours := offset / 3600
		minutes := (offset % 3600) / 60
		if minutes < 0 {
			minutes = -minutes
		}
		result += " " + formatOffset(hours, minutes)
	}
	return result
}

// FormatTime formats a time value using locale-specific time format
// Used for TIME type
func (r *RegionalSettings) FormatTime(t time.Time, scale int) string {
	if !r.enabled {
		format := "15:04:05"
		if scale > 0 {
			format = format + "." + strings.Repeat("0", scale)
		}
		return t.Format(format)
	}

	result := t.Format(r.timeFmt)
	if scale > 0 {
		frac := t.Nanosecond() / (1000000000 / pow10(scale))
		result += getDecimalSeparator(r.tag) + padLeftStr(strconv.Itoa(frac), scale, '0')
	}
	return result
}

// Helper functions

func pow10(n int) int {
	result := 1
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}

func padLeftStr(s string, length int, pad rune) string {
	for len(s) < length {
		s = string(pad) + s
	}
	return s
}

func formatOffset(hours, minutes int) string {
	sign := "+"
	if hours < 0 {
		sign = "-"
		hours = -hours
	}
	return sign + padLeftStr(strconv.Itoa(hours), 2, '0') + ":" + padLeftStr(strconv.Itoa(minutes), 2, '0')
}

// getDecimalSeparator returns the decimal separator for the given locale
func getDecimalSeparator(tag language.Tag) string {
	// Common decimal separators by language
	base, _ := tag.Base()
	switch base.String() {
	case "de", "fr", "es", "it", "pt", "nl", "pl", "cs", "sk", "hu", "ro", "bg", "hr", "sl", "sr", "tr", "el", "ru", "uk", "be", "fi", "sv", "no", "da", "is":
		return ","
	default:
		return "."
	}
}

// getThousandSeparator returns the thousand separator for the given locale
func getThousandSeparator(tag language.Tag) string {
	base, _ := tag.Base()
	switch base.String() {
	case "de", "fr", "es", "it", "pt", "nl", "pl", "cs", "sk", "hu", "ro", "bg", "hr", "sl", "sr", "tr", "el", "ru", "uk", "be", "fi", "sv", "no", "da", "is":
		// These locales use period or space as thousand separator
		return "."
	default:
		return ","
	}
}

// addThousandSeparators adds thousand separators to an integer string
func addThousandSeparators(s string, tag language.Tag) string {
	sep := getThousandSeparator(tag)
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder
	start := len(s) % 3
	if start == 0 {
		start = 3
	}
	result.WriteString(s[:start])
	for i := start; i < len(s); i += 3 {
		result.WriteString(sep)
		result.WriteString(s[i : i+3])
	}
	return result.String()
}

// getLocaleDateTimeFormats returns the date and time format strings for the locale
func getLocaleDateTimeFormats(tag language.Tag) (dateFmt, timeFmt string) {
	// Default to ISO format
	dateFmt = "2006-01-02"
	timeFmt = "15:04:05"

	base, _ := tag.Base()
	region, _ := tag.Region()

	// Set date format based on locale
	switch base.String() {
	case "en":
		if region.String() == "US" {
			dateFmt = "01/02/2006"
		} else {
			dateFmt = "02/01/2006"
		}
	case "de", "ru", "pl", "cs", "sk", "hu", "ro", "bg", "hr", "sl", "sr", "uk", "be":
		dateFmt = "02.01.2006"
	case "fr", "pt", "es", "it", "nl", "tr", "el":
		dateFmt = "02/01/2006"
	case "ja", "zh", "ko":
		dateFmt = "2006/01/02"
	case "fi", "sv", "no", "da", "is":
		dateFmt = "2006-01-02"
	}

	// Set time format based on locale (12hr vs 24hr)
	switch base.String() {
	case "en":
		if region.String() == "US" || region.String() == "CA" || region.String() == "AU" {
			timeFmt = "03:04:05 PM"
		}
	case "ja", "ko":
		// These can use 12hr with different AM/PM
		timeFmt = "15:04:05"
	}

	return dateFmt, timeFmt
}
