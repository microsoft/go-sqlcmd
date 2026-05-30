// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build !windows

package open

import (
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// run fails immediately on non-Windows platforms.
func (c *Ssms) run() {
	output := c.Output()
	output.Fatal(localizer.Sprintf("SSMS is only available on Windows. Use 'sqlcmd open vscode' instead."))
}
