// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type AddEndpoint struct {
	cmdparser.Cmd

	name    string
	address string
	port    int
}

func (c *AddEndpoint) DefineCommand(...cmdparser.Command) {
	c.Cmd.Options = cmdparser.Options{
		Use:   "add-endpoint",
		Short: "Add an endpoint",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "Add a default endpoint",
				Steps:       []string{"sqlcmd config add-endpoint --name my-endpoint --address localhost --port 1433"},
			},
		},
		Run: c.run,
	}

	c.Cmd.DefineCommand()

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.name,
		Name:          "name",
		DefaultString: "endpoint",
		Usage:         "Display name for the endpoint",
	})

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.address,
		Name:          "address",
		DefaultString: "localhost",
		Usage:         "The network address to connect to, e.g. 127.0.0.1 etc.",
	})

	c.AddFlag(cmdparser.FlagOptions{
		Int:        &c.port,
		Name:       "port",
		DefaultInt: 1433,
		Usage:      "The network port to connect to, e.g. 1433 etc.",
	})
}

func (c *AddEndpoint) run() {
	if c.name == "containerId" {
		panic("containerId")
	}

	endpoint := sqlconfig.Endpoint{
		EndpointDetails: sqlconfig.EndpointDetails{
			Address: c.address,
			Port:    c.port,
		},
		Name: c.name,
	}

	uniqueEndpointName := config.AddEndpoint(endpoint)
	output.InfofWithHintExamples([][]string{
		{"Add a context for this endpoint", fmt.Sprintf("sqlcmd config add-context --endpoint %v", uniqueEndpointName)},
		{"View endpoint names", "sqlcmd config get-endpoints"},
		{"View endpoint details", fmt.Sprintf("sqlcmd config get-endpoints %v", uniqueEndpointName)},
		{"View all endpoints details", "sqlcmd config get-endpoints --detailed"},
		{"Delete this endpoint", fmt.Sprintf("sqlcmd config delete-endpoint %v", uniqueEndpointName)},
	},
		"Endpoint '%v' added (address: '%v', port: '%v')", uniqueEndpointName, c.address, c.port)
}
