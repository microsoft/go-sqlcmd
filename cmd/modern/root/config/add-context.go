// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
)

// AddContext implements the `sqlcmd config add-context` command
type AddContext struct {
	cmdparser.Cmd

	name         string
	endpointName string
	userName     string
}

func (c *AddContext) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "add-context",
		Short: "Add a context",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Add a default context",
				Steps: []string{
					"sqlcmd config add-endpoint --name localhost-1433",
					"sqlcmd config add-context --name my-context --endpoint localhost-1433"}},
		},
		Run: c.run}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.name,
		Name:          "name",
		DefaultString: "context",
		Usage:         "Display name for the context"})

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.endpointName,
		Name:   "endpoint",
		Usage:  "Name of endpoint this context will use, use `sqlcmd config get-endpoints` to see list"})

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.userName,
		Name:   "user",
		Usage:  "Name of user this context will use, use `sqlcmd config get-users` to see list"})
}

// run adds a context to the configuration and sets it as the current context. The
// context consists of an endpoint and an optional user. The function checks
// if the specified endpoint and user exist and if not, it returns an error with
// suggestions on how to create them. If the context is successfully added, it
// outputs a message indicating the current context.
func (c *AddContext) run() {
	output := c.Output()
	context := sqlconfig.Context{
		ContextDetails: sqlconfig.ContextDetails{
			Endpoint: c.endpointName,
			User:     &c.userName,
		},
		Name: c.name,
	}

	if c.endpointName == "" || !config.EndpointExists(c.endpointName) {
		output.FatalfWithHintExamples([][]string{
			{"View existing endpoints to choose from", "sqlcmd config get-endpoints"},
			{"Add a new local endpoint", "sqlcmd create"},
			{"Add an already existing endpoint", "sqlcmd config add-endpoint --address localhost --port 1433"}},
			"Endpoint required to add context.  Endpoint '%v' does not exist.  Use --endpoint flag", c.endpointName)
	}

	if c.userName != "" {
		if !config.UserNameExists(c.userName) {
			output.FatalfWithHintExamples([][]string{
				{"View list of users", "sqlcmd config get-users"},
				{"Add the user", fmt.Sprintf("sqlcmd config add-user --name %v", c.userName)},
				{"Add an endpoint", "sqlcmd create"}},
				"User '%v' does not exist", c.userName)
		}
	}

	context.Name = config.AddContext(context)
	config.SetCurrentContextName(context.Name)
	output.InfofWithHintExamples([][]string{
		{"Open in Azure Data Studio", "sqlcmd open ads"},
		{"To start interactive query session", "sqlcmd query"},
		{"To run a query", "sqlcmd query \"SELECT @@version\""},
	}, "Current Context '%v'", context.Name)
}
