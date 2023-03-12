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
		defer os.Setenv("SQLCMDPASSWORD", "")
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
	if os.Getenv("SQLCMD_PASSWORD") == "" {
		os.Setenv("SQLCMD_PASSWORD", "it's-a-secret")
		defer os.Setenv("SQLCMD_PASSWORD", "")
	}
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*AddUser]()
	})
}

// TestNegAddUserNoUserName tests that the `sqlcmd config add-user` command
// when username is not set
func TestNegAddUserNoUserName(t *testing.T) {
	if os.Getenv("SQLCMD_PASSWORD") == "" {
		os.Setenv("SQLCMD_PASSWORD", "it's-a-secret")
		defer os.Setenv("SQLCMD_PASSWORD", "")
	}
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*AddUser]("--password-encryption none")
	})
}

// TestNegAddUserBadEncryptionMethod tests that the `sqlcmd config add-user` command
// when encryption method is not valid
func TestNegAddUserBadEncryptionMethod(t *testing.T) {
	if os.Getenv("SQLCMD_PASSWORD") == "" {
		os.Setenv("SQLCMD_PASSWORD", "it's-a-secret")
		defer os.Setenv("SQLCMD_PASSWORD", "")
	}
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*AddUser]("--username sa --password-encryption bad-bad")
	})
}

func TestNegBothPasswordEnvVarsSet(t *testing.T) {
	if os.Getenv("SQLCMD_PASSWORD") == "" {
		os.Setenv("SQLCMD_PASSWORD", "whatever")
		defer os.Setenv("SQLCMD_PASSWORD", "")
	}

	if os.Getenv("SQLCMDPASSWORD") == "" {
		os.Setenv("SQLCMDPASSWORD", "whatever")
		defer os.Setenv("SQLCMDPASSWORD", "")
	}

	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*AddUser]("--username sa --password-encryption none")
	})
}

func TestNegNoPasswordEnvVarsSet(t *testing.T) {
	if os.Getenv("SQLCMD_PASSWORD") != "" ||
		os.Getenv("SQLCMDPASSWORD") != "" {
		os.Setenv("SQLCMD_PASSWORD", "whatever")
		defer os.Setenv("SQLCMD_PASSWORD", "")
	}

	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*AddUser]("--username sa --password-encryption none")
	})
}
