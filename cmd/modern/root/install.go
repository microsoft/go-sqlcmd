// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/install"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

// Install defines the `sqlcmd install` sub-commands
type Install struct {
	cmdparser.Cmd
}

func (c *Install) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:         "install",
		Short:       "Install/Create #SQLFamily and Tools",
		Aliases:     []string{"create"},
		SubCommands: c.SubCommands(),
	}

	c.Cmd.DefineCommand(options)
}

// SubCommands sets up the sub-commands for `sqlcmd install` such as
// `sqlcmd install mssql` and `sqlcmd install azsql-edge`
func (c *Install) SubCommands() []cmdparser.Command {
	dependencies := c.Dependencies()

	return []cmdparser.Command{
		cmdparser.New[*install.Mssql](dependencies),
		cmdparser.New[*install.Edge](dependencies),
	}
}
