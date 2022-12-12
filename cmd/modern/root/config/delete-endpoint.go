// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"

	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
)

// DeleteEndpoint implements the `sqlcmd config delete-endpoint` command
type DeleteEndpoint struct {
	cmdparser.Cmd

	name string
}

func (c *DeleteEndpoint) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "delete-endpoint",
		Short: "Delete an endpoint",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Delete an endpoint",
				Steps: []string{
					"sqlcmd config delete-endpoint --name my-endpoint",
					"sqlcmd config delete-context endpoint"},
			},
		},
		Run: c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  "Name of endpoint to delete"})
}

// run is used to delete an endpoint with the given name. If the specified endpoint
// does not exist, the function will print an error message and return. If the
// endpoint exists, it will be deleted and a success message will be printed.
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
