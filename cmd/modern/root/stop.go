// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

type Stop struct {
	cmdparser.Cmd
}

func (c *Stop) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "stop",
		Short: localizer.Sprintf("Stop current context"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("Stop the current context"),
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
			{localizer.Sprintf("To view available contexts", "sqlcmd config get-contexts")},
		}, localizer.Sprintf("No current context"))
	}
	if config.CurrentContextEndpointHasContainer() {
		controller := container.NewController()
		id := config.ContainerId()
		endpoint, _ := config.CurrentContext()

		output.Infof(
			localizer.Sprintf("Stopping %q for context %q"),
			endpoint.ContainerDetails.Image,
			config.CurrentContextName(),
		)
		err := controller.ContainerStop(id)
		c.CheckErr(err)
	} else {
		output.FatalfWithHintExamples([][]string{
			{localizer.Sprintf("Create a new context with a SQL Server container ", "sqlcmd create mssql")},
		}, localizer.Sprintf("Current context does not have a container"))
	}
}
