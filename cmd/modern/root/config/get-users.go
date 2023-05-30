// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
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
		Short: localizer.Sprintf("Display one or many users from the sqlconfig file"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("List all the users in your sqlconfig file"),
				Steps:       []string{"sqlcmd config get-users"},
			},
			{
				Description: localizer.Sprintf("List all the users in your sqlconfig file"),
				Steps:       []string{"sqlcmd config get-users --detailed"},
			},
			{
				Description: localizer.Sprintf("Describe one user in your sqlconfig file"),
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
		Usage:  localizer.Sprintf("User name to view details of")})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.detailed,
		Name:  "detailed",
		Usage: localizer.Sprintf("Include user details")})
}

func (c *GetUsers) run() {
	output := c.Output()

	if c.name != "" {
		if config.UserNameExists(c.name) {
			user := config.GetUser(c.name)
			output.Struct(user)
		} else {
			output.FatalfWithHints(
				[]string{localizer.Sprintf("To view available users run `%s`", localizer.GetUsersCommand)},
				localizer.Sprintf("error: no user exists with the name: \"%v\"", c.name))
		}
	} else {
		config.OutputUsers(output.Struct, c.detailed)
	}
}
