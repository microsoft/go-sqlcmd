// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/test"
	"testing"
)

func TestGetUsers(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddUser]("--name user --username user")
	cmdparser.TestCmd[*GetUsers]()
	cmdparser.TestCmd[*GetUsers]("user")
}

func TestNegGetUsers(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*GetUsers]("does-not-exist")
}
