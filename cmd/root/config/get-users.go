// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type GetUsers struct {
	cmdparser.Cmd

	name     string
	detailed bool
}

func (c *GetUsers) DefineCommand(...cmdparser.Command) {
	c.Cmd.Options = cmdparser.Options{
		Use:   "get-users",
		Short: "Display one or many users from the sqlconfig file",
		Examples: []cmdparser.ExampleInfo{
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

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagInfo{Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand()

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
	if c.name != "" {
		if config.UserExists(c.name) {
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
