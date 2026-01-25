// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
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
func GetEncoding(codepage int) (encoding.Encoding, error) {
	switch codepage {
	// Unicode encodings
	case 65001:
		// UTF-8 - Go's native encoding, return nil to indicate no transformation needed
		return nil, nil
	case 1200:
		// UTF-16LE - Use ExpectBOM to strip BOM if present during input
		return unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM), nil
	case 1201:
		// UTF-16BE - Use ExpectBOM to strip BOM if present during input
		return unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM), nil

	// OEM/DOS codepages
	case 437:
		return charmap.CodePage437, nil
	case 850:
		return charmap.CodePage850, nil
	case 852:
		return charmap.CodePage852, nil
	case 855:
		return charmap.CodePage855, nil
	case 858:
		return charmap.CodePage858, nil
	case 860:
		return charmap.CodePage860, nil
	case 862:
		return charmap.CodePage862, nil
	case 863:
		return charmap.CodePage863, nil
	case 865:
		return charmap.CodePage865, nil
	case 866:
		return charmap.CodePage866, nil

	// Windows codepages
	case 874:
		return charmap.Windows874, nil
	case 1250:
		return charmap.Windows1250, nil
	case 1251:
		return charmap.Windows1251, nil
	case 1252:
		return charmap.Windows1252, nil
	case 1253:
		return charmap.Windows1253, nil
	case 1254:
		return charmap.Windows1254, nil
	case 1255:
		return charmap.Windows1255, nil
	case 1256:
		return charmap.Windows1256, nil
	case 1257:
		return charmap.Windows1257, nil
	case 1258:
		return charmap.Windows1258, nil

	// ISO-8859 codepages
	case 28591:
		return charmap.ISO8859_1, nil
	case 28592:
		return charmap.ISO8859_2, nil
	case 28593:
		return charmap.ISO8859_3, nil
	case 28594:
		return charmap.ISO8859_4, nil
	case 28595:
		return charmap.ISO8859_5, nil
	case 28596:
		return charmap.ISO8859_6, nil
	case 28597:
		return charmap.ISO8859_7, nil
	case 28598:
		return charmap.ISO8859_8, nil
	case 28599:
		return charmap.ISO8859_9, nil
	case 28600:
		return charmap.ISO8859_10, nil
	case 28603:
		return charmap.ISO8859_13, nil
	case 28604:
		return charmap.ISO8859_14, nil
	case 28605:
		return charmap.ISO8859_15, nil
	case 28606:
		return charmap.ISO8859_16, nil

	// Cyrillic
	case 20866:
		return charmap.KOI8R, nil
	case 21866:
		return charmap.KOI8U, nil

	// Macintosh
	case 10000:
		return charmap.Macintosh, nil
	case 10007:
		return charmap.MacintoshCyrillic, nil

	// EBCDIC codepages
	case 37:
		return charmap.CodePage037, nil
	case 1047:
		return charmap.CodePage1047, nil
	case 1140:
		return charmap.CodePage1140, nil

	// Japanese
	case 932:
		// Shift JIS (Windows-31J)
		return japanese.ShiftJIS, nil
	case 20932:
		// EUC-JP
		return japanese.EUCJP, nil
	case 50220, 50221, 50222:
		// ISO-2022-JP
		return japanese.ISO2022JP, nil

	// Korean
	case 949:
		// EUC-KR (Korean)
		return korean.EUCKR, nil
	case 51949:
		// EUC-KR alternate
		return korean.EUCKR, nil

	// Simplified Chinese
	case 936:
		// GBK (Simplified Chinese)
		return simplifiedchinese.GBK, nil
	case 54936:
		// GB18030
		return simplifiedchinese.GB18030, nil
	case 52936:
		// HZ-GB2312
		return simplifiedchinese.HZGB2312, nil

	// Traditional Chinese
	case 950:
		// Big5
		return traditionalchinese.Big5, nil

	default:
		return nil, localizer.Errorf("unsupported codepage %s", strconv.Itoa(codepage))
	}
}

// CodePageInfo describes a supported codepage
type CodePageInfo struct {
	CodePage    int
	Name        string
	Description string
}

// SupportedCodePages returns a list of all supported codepages with descriptions
func SupportedCodePages() []CodePageInfo {
	return []CodePageInfo{
		// Unicode
		{65001, "UTF-8", "Unicode (UTF-8)"},
		{1200, "UTF-16LE", "Unicode (UTF-16 Little-Endian)"},
		{1201, "UTF-16BE", "Unicode (UTF-16 Big-Endian)"},

		// OEM/DOS codepages
		{437, "CP437", "OEM United States"},
		{850, "CP850", "OEM Multilingual Latin 1"},
		{852, "CP852", "OEM Latin 2"},
		{855, "CP855", "OEM Cyrillic"},
		{858, "CP858", "OEM Multilingual Latin 1 + Euro"},
		{860, "CP860", "OEM Portuguese"},
		{862, "CP862", "OEM Hebrew"},
		{863, "CP863", "OEM Canadian French"},
		{865, "CP865", "OEM Nordic"},
		{866, "CP866", "OEM Russian"},

		// Windows codepages
		{874, "Windows-874", "Thai"},
		{1250, "Windows-1250", "Central European"},
		{1251, "Windows-1251", "Cyrillic"},
		{1252, "Windows-1252", "Western European"},
		{1253, "Windows-1253", "Greek"},
		{1254, "Windows-1254", "Turkish"},
		{1255, "Windows-1255", "Hebrew"},
		{1256, "Windows-1256", "Arabic"},
		{1257, "Windows-1257", "Baltic"},
		{1258, "Windows-1258", "Vietnamese"},

		// ISO-8859 codepages
		{28591, "ISO-8859-1", "Latin 1 (Western European)"},
		{28592, "ISO-8859-2", "Latin 2 (Central European)"},
		{28593, "ISO-8859-3", "Latin 3 (South European)"},
		{28594, "ISO-8859-4", "Latin 4 (North European)"},
		{28595, "ISO-8859-5", "Cyrillic"},
		{28596, "ISO-8859-6", "Arabic"},
		{28597, "ISO-8859-7", "Greek"},
		{28598, "ISO-8859-8", "Hebrew"},
		{28599, "ISO-8859-9", "Turkish"},
		{28600, "ISO-8859-10", "Nordic"},
		{28603, "ISO-8859-13", "Baltic"},
		{28604, "ISO-8859-14", "Celtic"},
		{28605, "ISO-8859-15", "Latin 9 (Western European with Euro)"},
		{28606, "ISO-8859-16", "Latin 10 (South-Eastern European)"},

		// Cyrillic
		{20866, "KOI8-R", "Russian"},
		{21866, "KOI8-U", "Ukrainian"},

		// Macintosh
		{10000, "Macintosh", "Mac Roman"},
		{10007, "x-mac-cyrillic", "Mac Cyrillic"},

		// EBCDIC
		{37, "IBM037", "EBCDIC US-Canada"},
		{1047, "IBM1047", "EBCDIC Latin 1/Open System"},
		{1140, "IBM01140", "EBCDIC US-Canada with Euro"},

		// Japanese
		{932, "Shift_JIS", "Japanese (Shift-JIS)"},
		{20932, "EUC-JP", "Japanese (EUC)"},
		{50220, "ISO-2022-JP", "Japanese (JIS)"},
		{50221, "csISO2022JP", "Japanese (JIS-Allow 1 byte Kana)"},
		{50222, "ISO-2022-JP", "Japanese (JIS-Allow 1 byte Kana SO/SI)"},

		// Korean
		{949, "EUC-KR", "Korean"},
		{51949, "EUC-KR", "Korean (EUC)"},

		// Simplified Chinese
		{936, "GBK", "Chinese Simplified (GBK)"},
		{54936, "GB18030", "Chinese Simplified (GB18030)"},
		{52936, "HZ-GB-2312", "Chinese Simplified (HZ)"},

		// Traditional Chinese
		{950, "Big5", "Chinese Traditional (Big5)"},
	}
}
