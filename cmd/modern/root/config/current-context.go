// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
)

// CurrentContext implements the `sqlcmd config current-context` command
type CurrentContext struct {
	cmdparser.Cmd
}

func (c *CurrentContext) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "current-context",
		Short: "Display the current-context",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Display the current-context",
				Steps: []string{
					"sqlcmd config current-context"},
			},
		},
		Run: c.run}

	c.Cmd.DefineCommand(options)
}

func (c *CurrentContext) run() {
	output := c.Output()
	output.Infof("%v\n", config.CurrentContextName())
}
