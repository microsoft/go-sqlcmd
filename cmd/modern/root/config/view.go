// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
)

// View implements the `sqlcmd config view` command
type View struct {
	cmdparser.Cmd

	raw bool
}

func (c *View) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "view",
		Short: "Display merged sqlconfig settings or a specified sqlconfig file",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Show sqlconfig settings, with REDACTED authentication data",
				Steps:       []string{"sqlcmd config view"},
			},
			{
				Description: "Show sqlconfig settings and raw authentication data",
				Steps:       []string{"sqlcmd config view --raw"},
			},
		},
		Aliases: []string{"show"},
		Run:     c.run,
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		Name:  "raw",
		Bool:  &c.raw,
		Usage: "Display raw byte data",
	})
}

func (c *View) run() {
	output := c.Output()

	contents := config.RedactedConfig(c.raw)
	output.Struct(contents)
}
