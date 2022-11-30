// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package mssql

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type GetTags struct {
	cmdparser.Cmd
}

func (c *GetTags) DefineCommand(output.Output, ...cmdparser.Command) {
	c.Cmd.SetOptions(cmdparser.Options{
		Use:   "get-tags",
		Short: "Get tags available for mssql install",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "List tags",
				Steps:       []string{"sqlcmd install mssql get-tags"},
			},
		},
		Aliases: []string{"gt", "lt"},
		Run:     c.run,
	})

	c.Cmd.DefineCommand()

}

func (c *GetTags) run() {
	output := c.Output()

	tags := container.ListTags(
		"mssql/server",
		"https://mcr.microsoft.com",
	)
	output.Struct(tags)
}
