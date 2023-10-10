// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetUsers(t *testing.T) {
	if os.Getenv("SQLCMD_PASSWORD") == "" {
		os.Setenv("SQLCMD_PASSWORD", "whatever")
		defer os.Setenv("SQLCMD_PASSWORD", "")
	}

	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddUser]("--name user --username user  --password-encryption none")
	cmdparser.TestCmd[*GetUsers]()
	cmdparser.TestCmd[*GetUsers]("user")
}

func TestNegGetUsers(t *testing.T) {
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*GetUsers]("does-not-exist")
	})
}
