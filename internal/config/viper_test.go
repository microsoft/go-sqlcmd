// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/test"
	"testing"
)

func Test_configureViper(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	configureViper("")
}

func Test_Load(t *testing.T) {
	SetFileNameForTest(t)
	Clean()
	Load()
}

func TestNeg_Load(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()
	filename = ""
	Load()
}

func TestNeg_Save(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()
	filename = ""
	Save()
}
