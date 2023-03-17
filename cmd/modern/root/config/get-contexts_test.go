// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetContexts(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddEndpoint]("--name endpoint")
	cmdparser.TestCmd[*AddContext]("--endpoint endpoint")
	cmdparser.TestCmd[*GetContexts]()
	cmdparser.TestCmd[*GetContexts]("context")
}

func TestNegGetContexts(t *testing.T) {
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*GetContexts]("does-not-exist")
	})
}
