// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// TestAddUser tests that the `sqlcmd config add-user` command
// works as expected
func TestAddUser(t *testing.T) {

	// SQLCMDPASSWORD is already set in the build pipelines, so don't
	// overwrite it here.
	if os.Getenv("SQLCMDPASSWORD") == "" {
		os.Setenv("SQLCMDPASSWORD", "it's-a-secret")
	}
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddUser]("--username user1 --password-encryption none")
}

// TestNegAddUser tests that the `sqlcmd config add-user` command
// fails when the auth-type is invalid
func TestNegAddUser(t *testing.T) {
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*AddUser]("--username user1 --auth-type bad-bad")
	})
}

// TestNegAddUser2 tests that the `sqlcmd config add-user` command
// fails when the auth-type is not basic and --password-encryption is set
func TestNegAddUser2(t *testing.T) {
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*AddUser]("--username user1 --auth-type other --password-encryption dpapi")
	})
}

// TestNegAddUser3 tests that the `sqlcmd config add-user` command
// fails when the SQLCMD_PASSWORD environment variable is not set
func TestNegAddUser3(t *testing.T) {
	if os.Getenv("SQLCMDPASSWORD") != "" {
		os.Setenv("SQLCMDPASSWORD", "")
	}

	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*AddUser]("--username user1")
	})
}

// TestNegAddUser4 tests that the `sqlcmd config add-user` command
// when username is not set
func TestNegAddUser4(t *testing.T) {
	os.Setenv("SQLCMD_PASSWORD", "whatever")

	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*AddUser]()
	})
}
