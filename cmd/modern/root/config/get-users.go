// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
)

// GetUsers implements the `sqlcmd config get-users` command
type GetUsers struct {
	cmdparser.Cmd

	name     string
	detailed bool
}

func (c *GetUsers) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "get-users",
		Short: "Display one or many users from the sqlconfig file",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "List all the users in your sqlconfig file",
				Steps:       []string{"sqlcmd config get-users"},
			},
			{
				Description: "List all the users in your sqlconfig file",
				Steps:       []string{"sqlcmd config get-users --detailed"},
			},
			{
				Description: "Describe one user in your sqlconfig file",
				Steps:       []string{"sqlcmd config get-users user1"},
			},
		},
		Run: c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  "User name to view details of"})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.detailed,
		Name:  "detailed",
		Usage: "Include user details"})
}

func (c *GetUsers) run() {
	output := c.Output()

	if c.name != "" {
		if config.UserNameExists(c.name) {
			user := config.GetUser(c.name)
			output.Struct(user)
		} else {
			output.FatalfWithHints(
				[]string{"To view available users run `sqlcmd config get-users"},
				"error: no user exists with the name: \"%v\"",
				c.name)
		}
	} else {
		config.OutputUsers(output.Struct, c.detailed)
	}
}
