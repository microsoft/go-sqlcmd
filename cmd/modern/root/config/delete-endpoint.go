// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// DeleteEndpoint implements the `sqlcmd config delete-endpoint` command
type DeleteEndpoint struct {
	cmdparser.Cmd

	name string
}

func (c *DeleteEndpoint) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "delete-endpoint",
		Short: localizer.Sprintf("Delete an endpoint"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("Delete an endpoint"),
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
		Usage:  localizer.Sprintf("Name of endpoint to delete")})
}

// run is used to delete an endpoint with the given name. If the specified endpoint
// does not exist, the function will print an error message and return. If the
// endpoint exists, it will be deleted and a success message will be printed.
func (c *DeleteEndpoint) run() {
	output := c.Output()

	if c.name == "" {
		output.Fatal(localizer.Sprintf("Endpoint name must be provided.  Provide endpoint name with %s flag", localizer.NameFlag))
	}

	if config.EndpointExists(c.name) {
		config.DeleteEndpoint(c.name)
	} else {
		output.FatalfWithHintExamples([][]string{
			{localizer.Sprintf("View endpoints"), "sqlcmd config get-endpoints"},
		},
			localizer.Sprintf("Endpoint '%v' does not exist", c.name))
	}

	output.Infof(localizer.Sprintf("Endpoint '%v' deleted", c.name))
}
