// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package edge

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"testing"
)

func TestEdgeGetTags(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*GetTags]()
}
