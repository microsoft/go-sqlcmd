// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package edge

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type GetTags struct {
	cmdparser.Cmd
}

func (c *GetTags) DefineCommand(...cmdparser.Command) {
	c.Cmd.Options = cmdparser.Options{
		Use:   "get-tags",
		Short: "Get tags available for mssql edge install",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "List tags",
				Steps:       []string{"sqlcmd install mssql-edge get-tags"},
			},
		},
		Aliases: []string{"gt", "lt"},
		Run:     c.run,
	}

	c.Cmd.DefineCommand()
}

func (c *GetTags) run() {
	tags := container.ListTags(
		"azure-sql-edge",
		"https://mcr.microsoft.com",
	)
	output.Struct(tags)
}
