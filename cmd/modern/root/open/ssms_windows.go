// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// On Windows, display info before launching
func (c *SSMS) displayPreLaunchInfo() {
	output := c.Output()
	output.Info(localizer.Sprintf("Launching SQL Server Management Studio..."))
}
