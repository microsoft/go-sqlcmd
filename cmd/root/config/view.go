// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type View struct {
	cmdparser.Cmd

	raw bool
}

func (c *View) DefineCommand(output.Output, ...cmdparser.Command) {
	c.Cmd.SetOptions(cmdparser.Options{
		Use:   "view",
		Short: "Display merged sqlconfig settings or a specified sqlconfig file",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "Show merged sqlconfig settings",
				Steps:       []string{"sqlcmd config view"},
			},
			{
				Description: "Show merged sqlconfig settings and raw authentication data",
				Steps:       []string{"sqlcmd config view --raw"},
			},
		},
		Aliases: []string{"use", "change-context", "set-context"},
		Run:     c.run,
	})

	c.Cmd.DefineCommand()

	c.AddFlag(cmdparser.FlagOptions{
		Name:  "raw",
		Bool:  &c.raw,
		Usage: "Display raw byte data",
	})
}

func (c *View) run() {
	output := c.Output()

	contents := config.GetRedactedConfig(c.raw)
	output.Struct(contents)
}
