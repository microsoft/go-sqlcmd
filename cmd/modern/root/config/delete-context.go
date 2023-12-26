// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// DeleteContext implements the `sqlcmd config delete-context` command
type DeleteContext struct {
	cmdparser.Cmd

	name    string
	cascade bool
}

func (c *DeleteContext) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "delete-context",
		Short: localizer.Sprintf("Delete a context"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("Delete a context (including its endpoint and user)"),
				Steps: []string{
					"sqlcmd config delete-context --name my-context --cascade",
					"sqlcmd config delete-context my-context --cascade"},
			},
			{
				Description: localizer.Sprintf("Delete a context (excluding its endpoint and user)"),
				Steps: []string{
					"sqlcmd config delete-context --name my-context --cascade=false",
					"sqlcmd config delete-context my-context --cascade=false"},
			},
		},
		Run: c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  localizer.Sprintf("Name of context to delete")})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:        &c.cascade,
		Name:        "cascade",
		DefaultBool: true,
		Usage:       localizer.Sprintf("Delete the context's endpoint and user as well")})
}

// run is responsible for deleting a context in a configuration. It first checks if
// a name is provided and if the context exists. If the cascade flag is set, it will
// also delete the associated endpoint and user. It then deletes the context
// and prints a message to the output indicating the context has been deleted.
func (c *DeleteContext) run() {
	output := c.Output()

	if c.name == "" {
		output.FatalWithHints([]string{localizer.Sprintf("Use the %s flag to pass in a context name to delete", localizer.NameFlag)},
			"A 'name' is required")
	}

	if config.ContextExists(c.name) {
		context := config.GetContext(c.name)

		if c.cascade {
			config.DeleteEndpoint(context.Endpoint)
			if config.UserExists(context) {
				config.DeleteUser(*context.ContextDetails.User)
			}
		}

		config.DeleteContext(c.name)

		output.Info(localizer.Sprintf("Context '%v' deleted", c.name))
	} else {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("View available contexts"), "sqlcmd config get-contexts"},
		},
			localizer.Sprintf("Context '%v' does not exist", c.name))
	}
}
