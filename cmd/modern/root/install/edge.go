// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/install/edge"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/pal"
)

// Edge implements the `sqlcmd install azsql-edge command and sub-commands
type Edge struct {
	cmdparser.Cmd
	MssqlBase
}

func (c *Edge) DefineCommand(...cmdparser.CommandOptions) {
	const repo = "azure-sql-edge"

	options := cmdparser.CommandOptions{
		Use:   "azsql-edge",
		Short: "Install Azure Sql Edge",
		Examples: []cmdparser.ExampleOptions{{
			Description: "Install/Create Azure SQL Edge in a container",
			Steps:       []string{"sqlcmd create azsql-edge"}}},
		Run:         c.MssqlBase.Run,
		SubCommands: c.SubCommands(),
	}

	c.MssqlBase.SetCrossCuttingConcerns(dependency.Options{
		EndOfLine: pal.LineBreak(),
		Output:    c.Output(),
	})

	c.Cmd.DefineCommand(options)
	c.AddFlags(c.AddFlag, repo, "edge")
}

func (c *Edge) SubCommands() []cmdparser.Command {
	return []cmdparser.Command{
		cmdparser.New[*edge.GetTags](c.Dependencies()),
	}
}
