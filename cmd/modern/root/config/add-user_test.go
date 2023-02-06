// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestAddUser(t *testing.T) {
	os.Setenv("SQLCMD_PASSWORD", "it's-a-secret")
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddUser]("--username user1")
}

func TestNegAddUser(t *testing.T) {
	assert.Panics(t, func() {
		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*AddUser]("--username user1 --auth-type bad-bad")
	})
}

func TestNegAddUser2(t *testing.T) {
	assert.Panics(t, func() {
		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*AddUser]("--username user1 --auth-type other --encrypt-password")
	})
}

func TestNegAddUser3(t *testing.T) {
	assert.Panics(t, func() {

		os.Setenv("SQLCMD_PASSWORD", "")

		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*AddUser]("--username user1")
	})
}

func TestNegAddUser4(t *testing.T) {
	assert.Panics(t, func() {

		os.Setenv("SQLCMD_PASSWORD", "whatever")

		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*AddUser]()
	})
}
