// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUsers(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddUser]("--name user --username user")
	cmdparser.TestCmd[*GetUsers]()
	cmdparser.TestCmd[*GetUsers]("user")
}

func TestNegGetUsers(t *testing.T) {
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*GetUsers]("does-not-exist")
	})
}
