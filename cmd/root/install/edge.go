// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

type Edge struct {
	cmdparser.Cmd
	MssqlBase
}

func (c *Edge) DefineCommand(subCommands ...cmdparser.Command) {
	const repo = "azure-sql-edge"

	c.Cmd.Options = cmdparser.Options{
		Use:   "mssql-edge",
		Short: "Install SQL Server Edge",
		Examples: []cmdparser.ExampleInfo{{
			Description: "Install SQL Server Edge in a container",
			Steps:       []string{"sqlcmd install mssql-edge"}}},
		Run: c.MssqlBase.Run,
	}

	c.Cmd.DefineCommand(subCommands...)
	c.AddFlags(c.AddFlag, repo, "edge")
}
