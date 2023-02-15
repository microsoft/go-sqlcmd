// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteEndpoint(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddEndpoint]()
	cmdparser.TestCmd[*DeleteEndpoint]("--name endpoint")
}

func TestNegDeleteEndpoint(t *testing.T) {
	assert.Panics(t, func() {

		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*DeleteEndpoint]()
	})
}

func TestNegDeleteEndpoint2(t *testing.T) {
	assert.Panics(t, func() {

		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*DeleteEndpoint]("--name does-not-exist")
	})
}
