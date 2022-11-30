// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type GetEndpoints struct {
	cmdparser.Cmd

	name     string
	detailed bool
}

func (c *GetEndpoints) DefineCommand(output.Output, ...cmdparser.Command) {
	c.Cmd.SetOptions(cmdparser.Options{
		Use:   "get-endpoints",
		Short: "Display one or many endpoints from the sqlconfig file",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "List all the endpoints in your sqlconfig file",
				Steps:       []string{"sqlcmd config get-endpoints"}},
			{
				Description: "List all the endpoints in your sqlconfig file",
				Steps:       []string{"sqlcmd config get-endpoints --detailed"}},
			{
				Description: "Describe one endpoint in your sqlconfig file",
				Steps:       []string{"sqlcmd config get-endpoints my-endpoint"}},
		},
		Run:                        c.run,
		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagInfo{Flag: "name", Value: &c.name},
	})

	c.Cmd.DefineCommand()

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  "Endpoint name to view details of"})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.detailed,
		Name:  "detailed",
		Usage: "Include endpoint details"})
}

func (c *GetEndpoints) run() {
	output := c.Output()

	if c.name != "" {
		if config.EndpointExists(c.name) {
			context := config.GetEndpoint(c.name)
			output.Struct(context)
		} else {
			output.FatalfWithHints(
				[]string{"To view available endpoints run `sqlcmd config get-endpoints"},
				"error: no endpoint exists with the name: \"%v\"",
				c.name)
		}
	} else {
		config.OutputEndpoints(output.Struct, c.detailed)
	}
}
