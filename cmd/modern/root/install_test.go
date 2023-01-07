// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"testing"
)

// TestInstall runs a sanity test of `sqlcmd install`
func TestInstall(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*Install]()
}
