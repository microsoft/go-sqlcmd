// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/transform"
)

func TestParseCodePage(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		wantInput   int
		wantOutput  int
		wantErr     bool
		errContains string
	}{
		{
			name:       "empty string",
			arg:        "",
			wantInput:  0,
			wantOutput: 0,
			wantErr:    false,
		},
		{
			name:       "single codepage sets both",
			arg:        "65001",
			wantInput:  65001,
			wantOutput: 65001,
			wantErr:    false,
		},
		{
			name:       "input only",
			arg:        "i:1252",
			wantInput:  1252,
			wantOutput: 0,
			wantErr:    false,
		},
		{
			name:       "output only",
			arg:        "o:65001",
			wantInput:  0,
			wantOutput: 65001,
			wantErr:    false,
		},
		{
			name:       "input and output",
			arg:        "i:1252,o:65001",
			wantInput:  1252,
			wantOutput: 65001,
			wantErr:    false,
		},
		{
			name:       "output and input reversed",
			arg:        "o:65001,i:1252",
			wantInput:  1252,
			wantOutput: 65001,
			wantErr:    false,
		},
		{
			name:       "uppercase prefix",
			arg:        "I:1252,O:65001",
			wantInput:  1252,
			wantOutput: 65001,
			wantErr:    false,
		},
		{
			name:        "invalid codepage number",
			arg:         "abc",
			wantErr:     true,
			errContains: "invalid codepage",
		},
		{
			name:        "invalid input codepage",
			arg:         "i:abc",
			wantErr:     true,
			errContains: "invalid input codepage",
		},
		{
			name:        "invalid output codepage",
			arg:         "o:xyz",
			wantErr:     true,
			errContains: "invalid output codepage",
		},
		{
			name:        "unsupported codepage",
			arg:         "99999",
			wantErr:     true,
			errContains: "codepage", // Error message varies by platform
		},
		{
			name:        "comma only produces no codepage",
			arg:         ",",
			wantErr:     true,
			errContains: "invalid codepage",
		},
		{
			name:        "whitespace only produces no codepage",
			arg:         "   ",
			wantErr:     true,
			errContains: "invalid codepage",
		},
		{
			name:        "multiple commas produce no codepage",
			arg:         ",,,",
			wantErr:     true,
			errContains: "invalid codepage",
		},
		{
			name:       "Japanese Shift JIS",
			arg:        "932",
			wantInput:  932,
			wantOutput: 932,
			wantErr:    false,
		},
		{
			name:       "Chinese GBK",
			arg:        "936",
			wantInput:  936,
			wantOutput: 936,
			wantErr:    false,
		},
		{
			name:       "Korean",
			arg:        "949",
			wantInput:  949,
			wantOutput: 949,
			wantErr:    false,
		},
		{
			name:       "Traditional Chinese Big5",
			arg:        "950",
			wantInput:  950,
			wantOutput: 950,
			wantErr:    false,
		},
		{
			name:       "EBCDIC",
			arg:        "37",
			wantInput:  37,
			wantOutput: 37,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := ParseCodePage(tt.arg)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			assert.NoError(t, err)
			if tt.arg == "" {
				assert.Nil(t, settings)
				return
			}
			assert.NotNil(t, settings)
			assert.Equal(t, tt.wantInput, settings.InputCodePage)
			assert.Equal(t, tt.wantOutput, settings.OutputCodePage)
		})
	}
}

func TestGetEncoding(t *testing.T) {
	tests := []struct {
		codepage int
		wantNil  bool // UTF-8 returns nil encoding
		wantErr  bool
	}{
		// Unicode
		{65001, true, false}, // UTF-8
		{1200, false, false}, // UTF-16LE
		{1201, false, false}, // UTF-16BE

		// OEM/DOS
		{437, false, false},
		{850, false, false},
		{866, false, false},

		// Windows
		{874, false, false},
		{1250, false, false},
		{1251, false, false},
		{1252, false, false},
		{1253, false, false},
		{1254, false, false},
		{1255, false, false},
		{1256, false, false},
		{1257, false, false},
		{1258, false, false},

		// ISO-8859
		{28591, false, false},
		{28592, false, false},
		{28605, false, false},

		// Cyrillic
		{20866, false, false},
		{21866, false, false},

		// Macintosh
		{10000, false, false},
		{10007, false, false},

		// EBCDIC
		{37, false, false},
		{1047, false, false},
		{1140, false, false},

		// CJK
		{932, false, false},   // Japanese Shift JIS
		{20932, false, false}, // EUC-JP
		{50220, false, false}, // ISO-2022-JP
		{949, false, false},   // Korean EUC-KR
		{936, false, false},   // Chinese GBK
		{54936, false, false}, // GB18030
		{950, false, false},   // Big5

		// Invalid
		{99999, false, true},
		{12345, false, true},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.codepage), func(t *testing.T) {
			enc, err := GetEncoding(tt.codepage)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, enc, "UTF-8 should return nil encoding")
			} else {
				assert.NotNil(t, enc, "non-UTF-8 codepage should return encoding")
			}
		})
	}
}

func TestSupportedCodePages(t *testing.T) {
	cps := SupportedCodePages()

	// Should have entries
	assert.Greater(t, len(cps), 0, "should return codepages")

	// Each returned codepage should be valid in GetEncoding
	for _, cp := range cps {
		_, err := GetEncoding(cp.CodePage)
		assert.NoError(t, err, "SupportedCodePages entry %d should be valid in GetEncoding", cp.CodePage)
		assert.NotEmpty(t, cp.Name, "codepage %d should have a name", cp.CodePage)
		assert.NotEmpty(t, cp.Description, "codepage %d should have a description", cp.CodePage)
	}

	// Result should be sorted by codepage number
	for i := 1; i < len(cps); i++ {
		assert.Less(t, cps[i-1].CodePage, cps[i].CodePage, "codepages should be sorted")
	}

	// Check some well-known codepages are present
	known := map[int]bool{
		65001: false, // UTF-8
		1252:  false, // Windows Western
		437:   false, // DOS US
		932:   false, // Japanese
	}
	for _, cp := range cps {
		if _, ok := known[cp.CodePage]; ok {
			known[cp.CodePage] = true
		}
	}
	for cp, found := range known {
		assert.True(t, found, "well-known codepage %d should be in list", cp)
	}
}

func TestGetEncodingWindowsFallback(t *testing.T) {
	// Japanese EBCDIC (20290) is not in our built-in registry but is available on Windows
	// This test verifies that the Windows API fallback works for codepages not in our registry
	cp := 20290 // IBM EBCDIC Japanese Katakana Extended

	enc, err := GetEncoding(cp)

	// On Windows, this should succeed because the Windows API can handle this codepage
	// On other platforms, this should fail with a helpful error message
	if err != nil {
		// Expected on non-Windows platforms
		assert.Contains(t, err.Error(), "codepage")
	} else {
		// Expected on Windows - verify the encoding works
		assert.NotNil(t, enc)

		// Test round-trip encoding/decoding
		// EBCDIC 'A' is 0xC1
		decoder := enc.NewDecoder()
		decoded, err := decoder.String(string([]byte{0xC1}))
		assert.NoError(t, err, "decoder should work")
		assert.Equal(t, "A", decoded, "EBCDIC 0xC1 should decode to 'A'")

		encoder := enc.NewEncoder()
		encoded, err := encoder.String("A")
		assert.NoError(t, err, "encoder should work")
		assert.Equal(t, []byte{0xC1}, []byte(encoded), "'A' should encode to EBCDIC 0xC1")
	}

	// Also test that a completely made-up codepage fails on all platforms
	_, err = GetEncoding(99999)
	assert.Error(t, err, "invalid codepage should fail on all platforms")
	assert.Contains(t, err.Error(), "codepage")
}

func TestWindowsEncodingStreaming(t *testing.T) {
	// This test exercises that the Windows API fallback encoding can be used in
	// streaming-like scenarios and that it handles single-byte data and
	// incomplete UTF-8 input correctly.

	// Japanese EBCDIC (20290) is a good test case as it's only available via Windows API
	cp := 20290 // IBM EBCDIC Japanese Katakana Extended

	enc, err := GetEncoding(cp)
	if err != nil {
		t.Skip("Codepage 20290 not available on this platform")
	}

	// Test decoder streaming with transform.Reader
	t.Run("decoder streaming", func(t *testing.T) {
		// Create a simple EBCDIC encoded string: "ABC" = 0xC1 0xC2 0xC3
		ebcdicData := []byte{0xC1, 0xC2, 0xC3}

		decoder := enc.NewDecoder()

		// Simulate streaming by processing one byte at a time
		var result []byte
		for i := 0; i < len(ebcdicData); i++ {
			decoder.Reset() // Reset between chunks for clean state
			dst := make([]byte, 32)
			nDst, _, err := decoder.Transform(dst, ebcdicData[i:i+1], i == len(ebcdicData)-1)
			if err != nil && err != transform.ErrShortSrc {
				t.Fatalf("Transform failed at byte %d: %v", i, err)
			}
			result = append(result, dst[:nDst]...)
		}
		assert.Equal(t, "ABC", string(result), "streaming decode should produce 'ABC'")
	})

	// Test encoder streaming
	t.Run("encoder streaming", func(t *testing.T) {
		// Test encoding "ABC" one character at a time
		input := "ABC"
		encoder := enc.NewEncoder()

		var result []byte
		for i := 0; i < len(input); i++ {
			encoder.Reset() // Reset between chunks for clean state
			dst := make([]byte, 32)
			nDst, _, err := encoder.Transform(dst, []byte(input[i:i+1]), i == len(input)-1)
			if err != nil && err != transform.ErrShortSrc {
				t.Fatalf("Transform failed at char %d: %v", i, err)
			}
			result = append(result, dst[:nDst]...)
		}
		expected := []byte{0xC1, 0xC2, 0xC3} // "ABC" in EBCDIC
		assert.Equal(t, expected, result, "streaming encode should produce EBCDIC ABC")
	})

	// Test encoder handles incomplete UTF-8 correctly
	t.Run("encoder incomplete UTF-8", func(t *testing.T) {
		encoder := enc.NewEncoder()
		dst := make([]byte, 32)

		// Send first byte of a 2-byte UTF-8 sequence (é = 0xC3 0xA9)
		incompleteUTF8 := []byte{0xC3} // First byte of é
		_, _, err := encoder.Transform(dst, incompleteUTF8, false)
		// Should return ErrShortSrc because the sequence is incomplete
		assert.Equal(t, transform.ErrShortSrc, err, "incomplete UTF-8 should return ErrShortSrc when not at EOF")

		// At EOF, incomplete sequence should be an error
		encoder.Reset()
		_, _, err = encoder.Transform(dst, incompleteUTF8, true)
		assert.Error(t, err, "incomplete UTF-8 at EOF should return error")
	})
}
