// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/test"
	"testing"
)

func TestDeleteEndpoint(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddEndpoint]()
	cmdparser.TestCmd[*DeleteEndpoint]("--name endpoint")
}

func TestNegDeleteEndpoint(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*DeleteEndpoint]()
}

func TestNegDeleteEndpoint2(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()

	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*DeleteEndpoint]("--name does-not-exist")
}
