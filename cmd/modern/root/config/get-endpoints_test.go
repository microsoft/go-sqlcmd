// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetEndpoints(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddEndpoint]("--name endpoint")
	cmdparser.TestCmd[*GetEndpoints]()
	cmdparser.TestCmd[*GetEndpoints]("endpoint")

}

func TestNegGetEndpoints(t *testing.T) {
	assert.Panics(t, func() {

		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*GetEndpoints]("does-not-exist")
	})
}
