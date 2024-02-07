// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// CurrentContext implements the `sqlcmd config current-context` command
type CurrentContext struct {
	cmdparser.Cmd
}

func (c *CurrentContext) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "current-context",
		Short: localizer.Sprintf("Display the name of the current-context"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("Display the current-context name"),
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
