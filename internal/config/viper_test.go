// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_configureViper(t *testing.T) {
	assert.Panics(t, func() {
		configureViper("")
	})
}

func Test_validateConfigFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "valid yaml extension",
			filename: "config.yaml",
			wantErr:  false,
		},
		{
			name:     "valid yml extension",
			filename: "config.yml",
			wantErr:  false,
		},
		{
			name:     "no extension (default sqlconfig)",
			filename: "sqlconfig",
			wantErr:  false,
		},
		{
			name:     "no extension with path",
			filename: "/home/user/.sqlcmd/sqlconfig",
			wantErr:  false,
		},
		{
			name:     "invalid txt extension",
			filename: "config.txt",
			wantErr:  true,
		},
		{
			name:     "invalid json extension",
			filename: "config.json",
			wantErr:  true,
		},
		{
			name:     "invalid xml extension",
			filename: "config.xml",
			wantErr:  true,
		},
		{
			name:     "uppercase YAML extension",
			filename: "config.YAML",
			wantErr:  false,
		},
		{
			name:     "uppercase YML extension",
			filename: "config.YML",
			wantErr:  false,
		},
		{
			name:     "mixed case yaml extension",
			filename: "config.Yaml",
			wantErr:  false,
		},
		{
			name:     "multiple dots with valid extension",
			filename: "my.config.yaml",
			wantErr:  false,
		},
		{
			name:     "multiple dots with invalid extension",
			filename: "my.config.txt",
			wantErr:  true,
		},
		{
			name:     "backup file with valid extension",
			filename: "config.backup.yaml",
			wantErr:  false,
		},
		{
			name:     "backup file with invalid extension",
			filename: "config.backup.txt",
			wantErr:  true,
		},
		{
			name:     "hidden file with yaml extension",
			filename: ".config.yaml",
			wantErr:  false,
		},
		{
			name:     "hidden file with yml extension",
			filename: ".config.yml",
			wantErr:  false,
		},
		{
			name:     "hidden file with invalid extension",
			filename: ".config.txt",
			wantErr:  true,
		},
		{
			name:     "file with only dot and yaml",
			filename: ".yaml",
			wantErr:  false,
		},
		{
			name:     "file with only dot and yml",
			filename: ".yml",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfigFileExtension(tt.filename)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for filename: %s", tt.filename)
				assert.Contains(t, err.Error(), "Configuration files must use YAML format")
			} else {
				assert.NoError(t, err, "Expected no error for filename: %s", tt.filename)
			}
		})
	}
}

func Test_configureViper_withInvalidExtension(t *testing.T) {
	err := configureViper("myconfig.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Configuration files must use YAML format")
	assert.Contains(t, err.Error(), ".txt")
}

func Test_configureViper_withValidExtensions(t *testing.T) {
	testCases := []string{
		"config.yaml",
		"config.yml",
		"sqlconfig",
		"/path/to/config.yaml",
	}

	for _, filename := range testCases {
		t.Run(filename, func(t *testing.T) {
			err := configureViper(filename)
			assert.NoError(t, err, "Expected no error for filename: %s", filename)
		})
	}
}

func Test_Load(t *testing.T) {
	SetFileNameForTest(t)
	Clean()
	Load()
}

func TestNeg_Load(t *testing.T) {
	filename = ""
	assert.Panics(t, func() {
		Load()
	})
}

func TestNeg_Save(t *testing.T) {
	filename = ""
	assert.Panics(t, func() {
		Save()
	})
}
