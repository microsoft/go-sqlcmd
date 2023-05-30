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
		Short: localizer.Sprintf("Add a context"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("Add a context for a local instance of SQL Server on port 1433 using trusted authentication"),
				Steps: []string{
					"sqlcmd config add-endpoint --name localhost-1433",
					"sqlcmd config add-context --name mssql --endpoint localhost-1433"}},
		},
		Run: c.run}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.name,
		Name:          "name",
		DefaultString: "context",
		Usage:         localizer.Sprintf("Display name for the context")})

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.endpointName,
		Name:   "endpoint",
		Usage:  localizer.Sprintf("Name of endpoint this context will use")})

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.userName,
		Name:   "user",
		Usage:  localizer.Sprintf("Name of user this context will use")})
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
			{localizer.Sprintf("View existing endpoints to choose from"), "sqlcmd config get-endpoints"},
			{localizer.Sprintf("Add a new local endpoint"), "sqlcmd create"},
			{localizer.Sprintf("Add an already existing endpoint"), "sqlcmd config add-endpoint --address localhost --port 1433"}},
			localizer.Sprintf("Endpoint required to add context.  Endpoint '%v' does not exist.  Use %s flag", c.endpointName, localizer.EndpointFlag))
	}

	if c.userName != "" {
		if !config.UserNameExists(c.userName) {
			output.FatalfWithHintExamples([][]string{
				{localizer.Sprintf("View list of users"), "sqlcmd config get-users"},
				{localizer.Sprintf("Add the user"), fmt.Sprintf("sqlcmd config add-user --name %v", c.userName)},
				{localizer.Sprintf("Add an endpoint"), "sqlcmd create"}},
				localizer.Sprintf("User '%v' does not exist", c.userName))
		}
	}

	context.Name = config.AddContext(context)
	config.SetCurrentContextName(context.Name)
	output.InfofWithHintExamples([][]string{
		{localizer.Sprintf("Open in Azure Data Studio"), "sqlcmd open ads"},
		{localizer.Sprintf("To start interactive query session"), "sqlcmd query"},
		{localizer.Sprintf("To run a query"), "sqlcmd query \"SELECT @@version\""},
	}, localizer.Sprintf("Current Context '%v'", context.Name))
}
