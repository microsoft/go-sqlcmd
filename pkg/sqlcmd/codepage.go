// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

// codepageEntry defines a codepage with its encoding and metadata
type codepageEntry struct {
	encoding    encoding.Encoding // nil for UTF-8 (Go's native encoding)
	name        string
	description string
}

// codepageRegistry is the single source of truth for all supported codepages
// that work cross-platform. Both GetEncoding and SupportedCodePages use this
// registry. On Windows, additional codepages installed on the system are also
// available via the Windows API fallback in GetEncoding.
var codepageRegistry = map[int]codepageEntry{
	// Unicode
	65001: {nil, "UTF-8", "Unicode (UTF-8)"},
	1200:  {unicode.UTF16(unicode.LittleEndian, unicode.UseBOM), "UTF-16LE", "Unicode (UTF-16 Little-Endian)"},
	1201:  {unicode.UTF16(unicode.BigEndian, unicode.UseBOM), "UTF-16BE", "Unicode (UTF-16 Big-Endian)"},

	// OEM/DOS codepages
	437: {charmap.CodePage437, "CP437", "OEM United States"},
	850: {charmap.CodePage850, "CP850", "OEM Multilingual Latin 1"},
	852: {charmap.CodePage852, "CP852", "OEM Latin 2"},
	855: {charmap.CodePage855, "CP855", "OEM Cyrillic"},
	858: {charmap.CodePage858, "CP858", "OEM Multilingual Latin 1 + Euro"},
	860: {charmap.CodePage860, "CP860", "OEM Portuguese"},
	862: {charmap.CodePage862, "CP862", "OEM Hebrew"},
	863: {charmap.CodePage863, "CP863", "OEM Canadian French"},
	865: {charmap.CodePage865, "CP865", "OEM Nordic"},
	866: {charmap.CodePage866, "CP866", "OEM Russian"},

	// Windows codepages
	874:  {charmap.Windows874, "Windows-874", "Thai"},
	1250: {charmap.Windows1250, "Windows-1250", "Central European"},
	1251: {charmap.Windows1251, "Windows-1251", "Cyrillic"},
	1252: {charmap.Windows1252, "Windows-1252", "Western European"},
	1253: {charmap.Windows1253, "Windows-1253", "Greek"},
	1254: {charmap.Windows1254, "Windows-1254", "Turkish"},
	1255: {charmap.Windows1255, "Windows-1255", "Hebrew"},
	1256: {charmap.Windows1256, "Windows-1256", "Arabic"},
	1257: {charmap.Windows1257, "Windows-1257", "Baltic"},
	1258: {charmap.Windows1258, "Windows-1258", "Vietnamese"},

	// ISO-8859 codepages
	28591: {charmap.ISO8859_1, "ISO-8859-1", "Latin 1 (Western European)"},
	28592: {charmap.ISO8859_2, "ISO-8859-2", "Latin 2 (Central European)"},
	28593: {charmap.ISO8859_3, "ISO-8859-3", "Latin 3 (South European)"},
	28594: {charmap.ISO8859_4, "ISO-8859-4", "Latin 4 (North European)"},
	28595: {charmap.ISO8859_5, "ISO-8859-5", "Cyrillic"},
	28596: {charmap.ISO8859_6, "ISO-8859-6", "Arabic"},
	28597: {charmap.ISO8859_7, "ISO-8859-7", "Greek"},
	28598: {charmap.ISO8859_8, "ISO-8859-8", "Hebrew"},
	28599: {charmap.ISO8859_9, "ISO-8859-9", "Turkish"},
	28600: {charmap.ISO8859_10, "ISO-8859-10", "Nordic"},
	28603: {charmap.ISO8859_13, "ISO-8859-13", "Baltic"},
	28604: {charmap.ISO8859_14, "ISO-8859-14", "Celtic"},
	28605: {charmap.ISO8859_15, "ISO-8859-15", "Latin 9 (Western European with Euro)"},
	28606: {charmap.ISO8859_16, "ISO-8859-16", "Latin 10 (South-Eastern European)"},

	// Cyrillic
	20866: {charmap.KOI8R, "KOI8-R", "Russian"},
	21866: {charmap.KOI8U, "KOI8-U", "Ukrainian"},

	// Macintosh
	10000: {charmap.Macintosh, "Macintosh", "Mac Roman"},
	10007: {charmap.MacintoshCyrillic, "x-mac-cyrillic", "Mac Cyrillic"},

	// EBCDIC
	37:   {charmap.CodePage037, "IBM037", "EBCDIC US-Canada"},
	1047: {charmap.CodePage1047, "IBM1047", "EBCDIC Latin 1/Open System"},
	1140: {charmap.CodePage1140, "IBM01140", "EBCDIC US-Canada with Euro"},

	// Japanese
	932:   {japanese.ShiftJIS, "Shift_JIS", "Japanese (Shift-JIS)"},
	20932: {japanese.EUCJP, "EUC-JP", "Japanese (EUC)"},
	50220: {japanese.ISO2022JP, "ISO-2022-JP", "Japanese (JIS)"},
	50221: {japanese.ISO2022JP, "csISO2022JP", "Japanese (JIS-Allow 1 byte Kana)"},
	50222: {japanese.ISO2022JP, "ISO-2022-JP", "Japanese (JIS-Allow 1 byte Kana SO/SI)"},

	// Korean
	949:   {korean.EUCKR, "EUC-KR", "Korean"},
	51949: {korean.EUCKR, "EUC-KR", "Korean (EUC)"},

	// Simplified Chinese
	936:   {simplifiedchinese.GBK, "GBK", "Chinese Simplified (GBK)"},
	54936: {simplifiedchinese.GB18030, "GB18030", "Chinese Simplified (GB18030)"},
	52936: {simplifiedchinese.HZGB2312, "HZ-GB-2312", "Chinese Simplified (HZ)"},

	// Traditional Chinese
	950: {traditionalchinese.Big5, "Big5", "Chinese Traditional (Big5)"},
}

// CodePageSettings holds the input and output codepage settings
type CodePageSettings struct {
	InputCodePage  int
	OutputCodePage int
}

// ParseCodePage parses the -f codepage argument
// Format: codepage | i:codepage[,o:codepage] | o:codepage[,i:codepage]
func ParseCodePage(arg string) (*CodePageSettings, error) {
	if arg == "" {
		return nil, nil
	}

	settings := &CodePageSettings{}
	parts := strings.Split(arg, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.HasPrefix(strings.ToLower(part), "i:") {
			// Input codepage
			cp, err := strconv.Atoi(strings.TrimPrefix(strings.ToLower(part), "i:"))
			if err != nil {
				return nil, localizer.Errorf("invalid input codepage: %s", part)
			}
			settings.InputCodePage = cp
		} else if strings.HasPrefix(strings.ToLower(part), "o:") {
			// Output codepage
			cp, err := strconv.Atoi(strings.TrimPrefix(strings.ToLower(part), "o:"))
			if err != nil {
				return nil, localizer.Errorf("invalid output codepage: %s", part)
			}
			settings.OutputCodePage = cp
		} else {
			// Both input and output
			cp, err := strconv.Atoi(part)
			if err != nil {
				return nil, localizer.Errorf("invalid codepage: %s", part)
			}
			settings.InputCodePage = cp
			settings.OutputCodePage = cp
		}
	}

	// If a non-empty argument was provided but no codepage was parsed,
	// treat this as an error rather than silently disabling codepage handling.
	if settings.InputCodePage == 0 && settings.OutputCodePage == 0 {
		return nil, localizer.Errorf("invalid codepage: %s", arg)
	}

	// Validate codepages
	if settings.InputCodePage != 0 {
		if _, err := GetEncoding(settings.InputCodePage); err != nil {
			return nil, err
		}
	}
	if settings.OutputCodePage != 0 {
		if _, err := GetEncoding(settings.OutputCodePage); err != nil {
			return nil, err
		}
	}

	return settings, nil
}

// GetEncoding returns the encoding for a given Windows codepage number.
// Returns nil for UTF-8 (65001) since Go uses UTF-8 natively.
// If the codepage is not in the built-in registry, falls back to
// OS-specific support (Windows API on Windows, error on other platforms).
func GetEncoding(codepage int) (encoding.Encoding, error) {
	entry, ok := codepageRegistry[codepage]
	if !ok {
		// Fallback to system-provided codepage support
		return getSystemCodePageEncoding(codepage)
	}
	return entry.encoding, nil
}

// CodePageInfo describes a supported codepage
type CodePageInfo struct {
	CodePage    int
	Name        string
	Description string
}

// SupportedCodePages returns a list of all supported codepages with descriptions
func SupportedCodePages() []CodePageInfo {
	result := make([]CodePageInfo, 0, len(codepageRegistry))
	for cp, entry := range codepageRegistry {
		result = append(result, CodePageInfo{
			CodePage:    cp,
			Name:        entry.name,
			Description: entry.description,
		})
	}
	// Sort by codepage number for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].CodePage < result[j].CodePage
	})
	return result
}
