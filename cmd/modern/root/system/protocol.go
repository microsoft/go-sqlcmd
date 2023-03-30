// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package system

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/system/protocol"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/pal"
)

// Protocol defines the `sqlcmd system protocol` sub-commands
type Protocol struct {
	cmdparser.Cmd
}

func (c *Protocol) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:         "protocol",
		Short:       `Register/Unregsiter the sqlcmd:// protocol handler`,
		SubCommands: c.SubCommands(),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Add context for existing endpoint and user (use SQLCMD_PASSWORD or SQLCMDPASSWORD)",
				Steps: []string{
					fmt.Sprintf("%s SQLCMD_PASSWORD=<placeholderpassword>", pal.CreateEnvVarKeyword()),
					"sqlcmd config add-user --name sa1434 --username sa",
					fmt.Sprintf("%s SQLCMD_PASSWORD=", pal.CreateEnvVarKeyword()),
					"sqlcmd config add-endpoint --name ep1434 --address localhost --port 1434",
					"sqlcmd config add-context --name mssql1434 --user sa1434 --endpoint ep1434"},
			},
		},
	}

	c.Cmd.DefineCommand(options)
}

// SubCommands sets up all the sub-commands for `sqlcmd config`
func (c *Protocol) SubCommands() []cmdparser.Command {
	dependencies := c.Dependencies()

	return []cmdparser.Command{
		cmdparser.New[*protocol.AddHandler](dependencies),
		cmdparser.New[*protocol.DeleteHandler](dependencies),
	}
}
