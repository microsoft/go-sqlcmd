// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// Type VSCode is used to implement the "open vscode" which launches Visual
// Studio Code and establishes a connection to the SQL Server for the current
// context
type VSCode struct {
	cmdparser.Cmd
}

func (c *VSCode) displayPreLaunchInfo() {
	output := c.Output()

	output.Info(localizer.Sprintf("Opening VS Code..."))
	output.Info(localizer.Sprintf("After VS Code opens:"))
	output.Info(localizer.Sprintf("1. Install the MSSQL extension if not already installed (Cmd+Shift+X, search 'SQL Server (mssql)')"))
	output.Info(localizer.Sprintf("2. Open the Command Palette (F1 or Cmd+Shift+P)"))
	output.Info(localizer.Sprintf("3. Type 'MS SQL: Connect' and select the 'sqlcmd-%s' profile", config.CurrentContextName()))
}
