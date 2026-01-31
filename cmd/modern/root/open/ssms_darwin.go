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

func (c *Ssms) displayPreLaunchInfo() {
	output := c.Output()
	output.Info(localizer.Sprintf("SSMS is only available on Windows"))
	output.Info(localizer.Sprintf("Please use 'sqlcmd open vscode' or 'sqlcmd open ads' instead"))
}
