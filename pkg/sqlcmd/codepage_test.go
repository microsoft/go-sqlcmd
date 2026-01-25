// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			errContains: "unsupported codepage",
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
		t.Run(string(rune(tt.codepage)), func(t *testing.T) {
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
