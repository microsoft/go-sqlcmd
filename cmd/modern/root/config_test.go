// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"testing"
)

// TestConfig runs a sanity test of `sqlcmd config`
func TestConfig(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*Config]()
}
