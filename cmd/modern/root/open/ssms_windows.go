// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

func (c *SSMS) displayPreLaunchInfo() {
	output := c.Output()
	output.Info(localizer.Sprintf("Launching SQL Server Management Studio..."))
}
