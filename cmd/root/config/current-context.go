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

func (c *CurrentContext) DefineCommand(output.Output, ...cmdparser.Command) {
	c.Cmd.SetOptions(cmdparser.Options{
		Use:   "current-context",
		Short: "Display the current-context",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "Display the current-context",
				Steps: []string{
					"sqlcmd config current-context"},
			},
		},
		Run: c.run})

	c.Cmd.DefineCommand()
}

func (c *CurrentContext) run() {
	output := c.Output()

	output.Infof("%v\n", config.GetCurrentContextName())
}
