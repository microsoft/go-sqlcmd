// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"

	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// AddEndpoint implements the `sqlcmd config add-endpoint` command
type AddEndpoint struct {
	cmdparser.Cmd

	name    string
	address string
	port    int
}

func (c *AddEndpoint) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "add-endpoint",
		Short: localizer.Sprintf("Add an endpoint"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("Add a default endpoint"),
				Steps:       []string{"sqlcmd config add-endpoint --name my-endpoint --address localhost --port 1433"},
			},
		},
		Run: c.run}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.name,
		Name:          "name",
		DefaultString: "endpoint",
		Usage:         localizer.Sprintf("Display name for the endpoint"),
	})

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.address,
		Name:          "address",
		DefaultString: "localhost",
		Usage:         localizer.Sprintf("The network address to connect to, e.g. 127.0.0.1 etc."),
	})

	c.AddFlag(cmdparser.FlagOptions{
		Int:        &c.port,
		Name:       "port",
		DefaultInt: 1433,
		Usage:      localizer.Sprintf("The network port to connect to, e.g. 1433 etc."),
	})
}

// run adds an endpoint to the system configuration. It creates a sqlconfig.Endpoint
// struct with the given parameters, and then adds the struct to the system configuration
// using the config.AddEndpoint method. If the endpoint is successfully added, it prints
// a message with information about the endpoint and hints on how to use the endpoint.
func (c *AddEndpoint) run() {
	output := c.Output()

	endpoint := sqlconfig.Endpoint{
		EndpointDetails: sqlconfig.EndpointDetails{
			Address: c.address,
			Port:    c.port,
		},
		Name: c.name,
	}

	uniqueEndpointName := config.AddEndpoint(endpoint)
	output.InfofWithHintExamples([][]string{
		{localizer.Sprintf("Add a context for this endpoint"), fmt.Sprintf("sqlcmd config add-context --endpoint %v", uniqueEndpointName)},
		{localizer.Sprintf("View endpoint names"), "sqlcmd config get-endpoints"},
		{localizer.Sprintf("View endpoint details"), fmt.Sprintf("sqlcmd config get-endpoints %v", uniqueEndpointName)},
		{localizer.Sprintf("View all endpoints details"), "sqlcmd config get-endpoints --detailed"},
		{localizer.Sprintf("Delete this endpoint"), fmt.Sprintf("sqlcmd config delete-endpoint %v", uniqueEndpointName)},
	},
		localizer.Sprintf("Endpoint '%v' added (address: '%v', port: '%v')", uniqueEndpointName, c.address, c.port))
}
