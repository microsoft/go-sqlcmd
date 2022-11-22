// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type DeleteUser struct {
	cmdparser.Cmd

	name string
}

func (c *DeleteUser) DefineCommand(...cmdparser.Command) {
	c.Cmd.Options = cmdparser.Options{
		Use:   "delete-user",
		Short: "Delete a user",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "Delete a user",
				Steps: []string{
					"sqlcmd config delete-user --name user1",
					"sqlcmd config delete-user user1"}},
		},
		Run: c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagInfo{
			Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand()

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  "Name of user to delete"})
}

func (c *DeleteUser) run() {
	config.DeleteUser(c.name)
	output.Infof("User '%v' deleted", c.name)
}
