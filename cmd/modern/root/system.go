// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/system"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

// Config defines the `sqlcmd system` sub-commands
type System struct {
	cmdparser.Cmd
}

func (c *System) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:         "system",
		Short:       `Modify system wide options like the protocol handler or container runtime cache"`,
		SubCommands: c.SubCommands(),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Register the sqlcmd:// protocol handler",
				Steps: []string{
					"sqlcmd system protocol add-handler"},
			},
		},
	}

	c.Cmd.DefineCommand(options)
}

// SubCommands sets up all the sub-commands for `sqlcmd config`
func (c *System) SubCommands() []cmdparser.Command {
	dependencies := c.Dependencies()

	return []cmdparser.Command{
		cmdparser.New[*system.Protocol](dependencies),
	}
}
