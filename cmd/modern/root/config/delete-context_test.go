// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteContext(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddUser]("--username user --auth-type other")
	cmdparser.TestCmd[*AddEndpoint]()
	cmdparser.TestCmd[*AddContext]("--endpoint endpoint --user user")
	cmdparser.TestCmd[*DeleteContext]("--name context")
}

func TestNegDeleteContext(t *testing.T) {
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*DeleteContext]()
	})
}

func TestNegDeleteContext2(t *testing.T) {
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*DeleteContext]("--name does-not-exist")
	})
}
