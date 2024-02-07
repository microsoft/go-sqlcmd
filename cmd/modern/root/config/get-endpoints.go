// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// GetEndpoints implements the `sqlcmd config get-endpoints` command
type GetEndpoints struct {
	cmdparser.Cmd

	name     string
	detailed bool
	value    string
}

func (c *GetEndpoints) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "get-endpoints",
		Short: localizer.Sprintf("Display one or many endpoints from the sqlconfig file"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("List all the endpoints in your sqlconfig file"),
				Steps:       []string{"sqlcmd config get-endpoints"}},
			{
				Description: localizer.Sprintf("List all the endpoints in your sqlconfig file"),
				Steps:       []string{"sqlcmd config get-endpoints --detailed"}},
			{
				Description: localizer.Sprintf("Describe one endpoint in your sqlconfig file"),
				Steps:       []string{"sqlcmd config get-endpoints my-endpoint"}},
		},
		Run:                        c.run,
		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  localizer.Sprintf("Endpoint name to view details of")})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.detailed,
		Name:  "detailed",
		Usage: localizer.Sprintf("Include endpoint details")})

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.value,
		Name:   "value",
		Usage:  localizer.Sprintf("Value to get, (endpoint, port")})
}

func (c *GetEndpoints) run() {
	output := c.Output()

	if c.name != "" {
		if config.EndpointExists(c.name) {
			context := config.GetEndpoint(c.name)

			if c.value == "" {
				output.Struct(context)
			} else {
				if c.value == "address" {
					output.Struct(context.Address)
				} else if c.value == "port" {
					output.Struct(context.Port)
				} else {
					panic("Invalid value")
				}
			}

		} else {
			output.FatalWithHints(
				[]string{localizer.Sprintf("To view available endpoints run `%s`", localizer.GetEndpointsCommand)},
				localizer.Sprintf("error: no endpoint exists with the name: \"%v\"", c.name))
		}
	} else {
		config.OutputEndpoints(output.Struct, c.detailed)
	}
}
