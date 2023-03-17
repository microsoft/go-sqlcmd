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
