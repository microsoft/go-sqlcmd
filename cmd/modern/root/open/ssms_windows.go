// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// Type Ssms is used to implement the "open ssms" which launches SQL Server
// Management Studio and establishes a connection to the SQL Server for the current
// context
type Ssms struct {
	cmdparser.Cmd
}

// On Windows, display info before launching
func (c *Ssms) displayPreLaunchInfo() {
	output := c.Output()
	output.Info(localizer.Sprintf("Launching SQL Server Management Studio..."))
}
