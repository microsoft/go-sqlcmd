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

	errorLogEntryToWaitFor string
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
	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.errorLogEntryToWaitFor,
		DefaultString: "Recovery is complete",
		Name:          "errorlog-wait-line",
		Usage:         "Line in errorlog to wait for before exiting this process",
	})
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

		// BUG(stuartpa): Even if we block here for the final entry in the errorlog
		// "Recovery Complete", a sqlcmd query will still fail for a few seconds
		// with:
		//  sqlcmd query "SELECT @@version"
		//  EOF
		//  Error: EOF
		// I'm not sure what else to block here for before returning to ensure
		// a `sqlcmd query` will work
		controller.ContainerWaitForLogEntry(
			id, c.errorLogEntryToWaitFor)
	} else {
		output.FatalfWithHintExamples([][]string{
			{"Create new context with a sql container ", "sqlcmd create mssql"},
		}, "Current context does not have a container")
	}
}
