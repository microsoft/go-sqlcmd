// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddContext(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddEndpoint]()
	cmdparser.TestCmd[*AddContext]("--endpoint endpoint")
}

func TestNegAddContext(t *testing.T) {
	assert.Panics(t, func() {
		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*AddContext]("--endpoint does-not-exist")
	})
}

func TestNegAddContext2(t *testing.T) {
	assert.Panics(t, func() {
		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*AddEndpoint]()
		cmdparser.TestCmd[*AddContext]("--endpoint endpoint --user does-not-exist")
	})
}
