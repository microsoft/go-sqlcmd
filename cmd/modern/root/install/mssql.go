// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/install/mssql"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/pal"
)

// Mssql implements the `sqlcmd install mssql command and sub-commands
type Mssql struct {
	cmdparser.Cmd
	MssqlBase
}

func (c *Mssql) DefineCommand(...cmdparser.CommandOptions) {
	const repo = "mssql/server"

	options := cmdparser.CommandOptions{
		Use:   "mssql",
		Short: "Install SQL Server",
		Examples: []cmdparser.ExampleOptions{{
			Description: "Install SQL Server in a container",
			Steps:       []string{"sqlcmd install mssql"}}},
		Run:         c.MssqlBase.Run,
		SubCommands: c.SubCommands(),
	}

	c.MssqlBase.SetCrossCuttingConcerns(dependency.Options{
		EndOfLine: pal.LineBreak(),
		Output:    c.Output(),
	})

	c.Cmd.DefineCommand(options)
	c.AddFlags(c.AddFlag, repo, "mssql")
}

func (c *Mssql) SubCommands() []cmdparser.Command {
	return []cmdparser.Command{
		cmdparser.New[*mssql.GetTags](c.Dependencies()),
	}
}
