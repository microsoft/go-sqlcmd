// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type DeleteEndpoint struct {
	cmdparser.Cmd

	name string
}

func (c *DeleteEndpoint) DefineCommand(output.Output, ...cmdparser.Command) {
	c.Cmd.SetOptions(cmdparser.Options{
		Use:   "delete-endpoint",
		Short: "Delete an endpoint",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "Delete an endpoint",
				Steps: []string{
					"sqlcmd config delete-endpoint --name my-endpoint",
					"sqlcmd config delete-context endpoint"},
			},
		},
		Run: c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagInfo{Flag: "name", Value: &c.name},
	})

	c.Cmd.DefineCommand()

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  "Name of endpoint to delete"})
}

func (c *DeleteEndpoint) run() {
	output := c.Output()

	if c.name == "" {
		output.Fatal("Endpoint name must be provided.  Provide endpoint name with --name flag")
	}

	if config.EndpointExists(c.name) {
		config.DeleteEndpoint(c.name)
	} else {
		output.FatalfWithHintExamples([][]string{
			{"View endpoints", "sqlcmd config get-endpoints"},
		},
			fmt.Sprintf("Endpoint '%v' does not exist", c.name))
	}

	output.Infof("Endpoint '%v' deleted", c.name)
}
