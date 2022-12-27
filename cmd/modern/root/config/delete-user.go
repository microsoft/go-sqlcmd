// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
)

// DeleteUser implements the `sqlcmd config delete-user` command
type DeleteUser struct {
	cmdparser.Cmd

	name string
}

func (c *DeleteUser) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "delete-user",
		Short: "Delete a user",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Delete a user",
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
		Usage:  "Name of user to delete"})
}

func (c *DeleteUser) run() {
	output := c.Output()

	config.DeleteUser(c.name)
	output.Infof("User '%v' deleted", c.name)
}
