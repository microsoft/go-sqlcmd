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

	// build pins which VS Code build to configure and launch: "stable",
	// "insiders", or "" to prefer stable then insiders.
	build string
}

func (c *VSCode) displayPreLaunchInfo() {
	output := c.Output()

	output.Info(localizer.Sprintf("Opening VS Code..."))
	output.Info(localizer.Sprintf("Use the '%s' connection profile to connect", config.CurrentContextName()))
}
