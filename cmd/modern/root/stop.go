// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
)

type Stop struct {
	cmdparser.Cmd
}

func (c *Stop) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "stop",
		Short: "Stop current context",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Stop the current context",
				Steps:       []string{`sqlcmd stop`}},
		},
		Run: c.run,
	}

	c.Cmd.DefineCommand(options)
}

func (c *Stop) run() {
	output := c.Output()

	if config.CurrentContextName() == "" {
		output.FatalfWithHintExamples([][]string{
			{"To view available contexts", "sqlcmd config get-contexts"},
		}, "No current context")
	}
	if config.CurrentContextEndpointHasContainer() {
		controller := container.NewController()
		id := config.ContainerId()
		endpoint, _ := config.CurrentContext()

		output.Infof(
			"Stopping %q for context %q",
			endpoint.ContainerDetails.Image,
			config.CurrentContextName(),
		)
		err := controller.ContainerStop(id)
		c.CheckErr(err)
	} else {
		output.FatalfWithHintExamples([][]string{
			{"Create a new context with a SQL Server container ", "sqlcmd create mssql"},
		}, "Current context does not have a container")
	}
}
