// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
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
		Short: "Delete a context",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Delete a context",
				Steps: []string{
					"sqlcmd config delete-context --name my-context --cascade",
					"sqlcmd config delete-context my-context --cascade"},
			},
		},
		Run: c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  "Name of context to delete"})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:        &c.cascade,
		Name:        "cascade",
		DefaultBool: true,
		Usage:       "Delete the context's endpoint and user as well"})
}

// run is responsible for deleting a context in a configuration. It first checks if
// a name is provided and if the context exists. If the cascade flag is set, it will
// also delete the associated endpoint and user. It then deletes the context
// and prints a message to the output indicating the context has been deleted.
func (c *DeleteContext) run() {
	output := c.Output()

	if c.name == "" {
		output.FatalWithHints([]string{"Use the --name flag to pass in a context name to delete"},
			"A 'name' is required")
	}

	if config.ContextExists(c.name) {
		context := config.GetContext(c.name)

		if c.cascade {
			config.DeleteEndpoint(context.Endpoint)
			if context.ContextDetails.User != nil && *context.ContextDetails.User != "" {
				config.DeleteUser(*context.ContextDetails.User)
			}
		}

		config.DeleteContext(c.name)

		output.Infof("Context '%v' deleted", c.name)
	} else {
		output.FatalfWithHintExamples([][]string{
			{"View available contexts", "sqlcmd config get-contexts"},
		},
			"Context '%v' does not exist", c.name)
	}
}
