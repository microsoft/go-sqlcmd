// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// DeleteUser implements the `sqlcmd config delete-user` command
type DeleteUser struct {
	cmdparser.Cmd

	name string
}

func (c *DeleteUser) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "delete-user",
		Short: localizer.Sprintf("Delete a user"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("Delete a user"),
				Steps: []string{
					"sqlcmd config delete-user --name user1",
					"sqlcmd config delete-user user1"}},
		},
		Run: c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{
			Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  localizer.Sprintf("Name of user to delete")})
}

func (c *DeleteUser) run() {
	output := c.Output()

	if c.name == "" {
		output.Fatal(localizer.Sprintf("User name must be provided.  Provide user name with %s flag", localizer.NameFlag))
	}

	if config.UserNameExists(c.name) {
		config.DeleteUser(c.name)
	} else {
		output.FatalfWithHintExamples([][]string{
			{localizer.Sprintf("View users"), "sqlcmd config get-users"},
		},
			localizer.Sprintf("User %q does not exist", c.name))
	}

	output.Info(localizer.Sprintf("User %q deleted", c.name))
}
