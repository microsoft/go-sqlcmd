// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
)

type Start struct {
	cmdparser.Cmd
}

func (c *Start) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "start",
		Short: "Start current context",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Start the current context",
				Steps:       []string{`sqlcmd start`}},
		},
		Run: c.run,
	}

	c.Cmd.DefineCommand(options)
}

func (c *Start) run() {
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
			"Starting %q for context %q",
			endpoint.ContainerDetails.Image,
			config.CurrentContextName(),
		)
		err := controller.ContainerStart(id)
		c.CheckErr(err)
	} else {
		output.FatalfWithHintExamples([][]string{
			{"Create new context with a sql container ", "sqlcmd create mssql"},
		}, "Current context does not have a container")
	}
}
