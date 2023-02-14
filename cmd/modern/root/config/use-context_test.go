// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUseContext(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddEndpoint]()
	cmdparser.TestCmd[*AddContext]("--endpoint endpoint")
	cmdparser.TestCmd[*UseContext]("--name context")
}

func TestNegUseContext(t *testing.T) {
	assert.Panics(t, func() {

		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*UseContext]("does-not-exist")
	})
}
