// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type CurrentContext struct {
	cmdparser.Cmd
}

func (c *CurrentContext) DefineCommand(...cmdparser.Command) {
	c.Cmd.Options = cmdparser.Options{
		Use:   "current-context",
		Short: "Display the current-context",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "Display the current-context",
				Steps: []string{
					"sqlcmd config current-context"},
			},
		},
		Run: c.run,
	}

	c.Cmd.DefineCommand()
}

func (c *CurrentContext) run() {
	output.Infof("%v\n", config.GetCurrentContextName())
}
