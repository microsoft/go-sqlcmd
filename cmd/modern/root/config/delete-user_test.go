// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDeleteUser(t *testing.T) {
	cmdparser.TestSetup(t)

	// SQLCMDPASSWORD is already set in the build pipelines, so don't
	// overwrite it here.
	if os.Getenv("SQLCMDPASSWORD") == "" && os.Getenv("SQLCMD_PASSWORD") == "" {
		os.Setenv("SQLCMDPASSWORD", "it's-a-secret")
		defer os.Setenv("SQLCMDPASSWORD", "")
	}
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddUser]("--username user1 --password-encryption none")
	cmdparser.TestCmd[*DeleteUser]("--name user")
}

func TestNegDeleteUserNoName(t *testing.T) {
	cmdparser.TestSetup(t)

	assert.Panics(t, func() {
		cmdparser.TestCmd[*DeleteUser]()
	})
}

func TestNegDeleteUserInvalidName(t *testing.T) {
	cmdparser.TestSetup(t)

	assert.Panics(t, func() {
		cmdparser.TestCmd[*DeleteUser]("--name bad-bad")
	})
}
