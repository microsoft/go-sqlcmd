// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

type Mssql struct {
	cmdparser.Cmd
	MssqlBase
}

func (c *Mssql) DefineCommand(subCommands ...cmdparser.Command) {
	const repo = "mssql/server"

	c.Cmd.Options = cmdparser.Options{
		Use:   "mssql",
		Short: "Install SQL Server",
		Examples: []cmdparser.ExampleInfo{{
			Description: "Install SQL Server in a container",
			Steps:       []string{"sqlcmd install mssql"}}},
		Run: c.MssqlBase.Run,
	}

	c.Cmd.DefineCommand(subCommands...)
	c.AddFlags(c.AddFlag, repo, "mssql")

	return
}
