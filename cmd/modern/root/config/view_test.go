// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"testing"
)

func TestView(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*View]()
}
