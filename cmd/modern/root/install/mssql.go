// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/install/mssql"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/pal"
)

// Mssql implements the `sqlcmd install mssql` command and sub-commands
type Mssql struct {
	cmdparser.Cmd
	MssqlBase
}

func (c *Mssql) DefineCommand(...cmdparser.CommandOptions) {
	const repo = "mssql/server"

	options := cmdparser.CommandOptions{
		Use:   "mssql",
		Short: "Install/Create SQL Server in a container",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Install/Create SQL Server in a container",
				Steps:       []string{"sqlcmd create mssql"}},
			{
				Description: "See all release tags for SQL Server, install previous version",
				Steps: []string{
					"sqlcmd create mssql get-tags",
					"sqlcmd create mssql --tag 2019-latest",
				},
			},
			{
				Description: "Create SQL Server, download and attach AdventureWorks sample database",
				Steps:       []string{"sqlcmd create mssql --use https://aka.ms/AdventureWorksLT.bak"}},
			{
				Description: "Create SQL Server, download and attach AdventureWorks sample database with different database name",
				Steps:       []string{"sqlcmd create mssql --use https://aka.ms/AdventureWorksLT.bak,adventureworks"}},
			{
				Description: "Create SQL Server with an empty user database",
				Steps:       []string{"sqlcmd create mssql --database db1"}},
			{
				Description: "Install/Create SQL Server with full logging",
				Steps:       []string{"sqlcmd create mssql --verbosity 4"}},
		},
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
