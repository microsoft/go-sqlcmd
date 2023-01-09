// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/test"
	"testing"
)

func TestGetEndpoints(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddEndpoint]("--name endpoint")
	cmdparser.TestCmd[*GetEndpoints]()
	cmdparser.TestCmd[*GetEndpoints]("endpoint")

}

func TestNegGetEndpoints(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*GetEndpoints]("does-not-exist")
}
