// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_configureViper(t *testing.T) {
	assert.Panics(t, func() {
		_ = configureViper("")
	})
}

func Test_configureViperValidExtensions(t *testing.T) {
	tests := []string{"config.yaml", "config.yml", "sqlconfig", "/path/to/config.YAML"}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			err := configureViper(name)
			assert.NoError(t, err)
		})
	}
}

func Test_configureViperInvalidExtension(t *testing.T) {
	err := configureViper("config.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "YAML format")
	assert.Contains(t, err.Error(), ".txt")
}

func Test_validateConfigFileExtension(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{"yaml extension", "config.yaml", false},
		{"yml extension", "config.yml", false},
		{"no extension", "sqlconfig", false},
		{"uppercase YAML", "config.YAML", false},
		{"txt extension", "config.txt", true},
		{"json extension", "config.json", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfigFileExtension(tt.file)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
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
