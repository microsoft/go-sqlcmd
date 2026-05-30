// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build !windows

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

// DefineCommand sets up the ssms command for non-Windows platforms
func (c *Ssms) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "ssms",
		Short: localizer.Sprintf("Open SQL Server Management Studio and connect to current context"),
		Examples: []cmdparser.ExampleOptions{{
			Description: localizer.Sprintf("Open SSMS and connect using the current context"),
			Steps:       []string{"sqlcmd open ssms"}}},
		Run: c.run,
	}

	c.Cmd.DefineCommand(options)
}

// run fails immediately on non-Windows platforms
func (c *Ssms) run() {
	output := c.Output()
	output.Fatal(localizer.Sprintf("SSMS is only available on Windows. Use 'sqlcmd open vscode' instead."))
}
