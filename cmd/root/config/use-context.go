// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type UseContext struct {
	cmdparser.Cmd

	name string
}

func (c *UseContext) DefineCommand(...cmdparser.Command) {
	c.Cmd.Options = cmdparser.Options{
		Use:   "use-context",
		Short: "Display one or many users from the sqlconfig file",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "Use the context for the user@mssql sql instance",
				Steps:       []string{"sqlcmd config use-context user@mssql"},
			},
		},
		Aliases: []string{"use", "change-context", "set-context"},
		Run:     c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagInfo{Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand()

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  "Name of context to set as current context"})
}

func (c *UseContext) run() {
	if config.ContextExists(c.name) {
		config.SetCurrentContextName(c.name)
		output.InfofWithHints([]string{
			"To run a query:    sqlcmd query \"SELECT @@SERVERNAME\"",
			"To remove:         sqlcmd uninstall"},
			"Switched to context \"%v\".", c.name)
	} else {
		output.FatalfWithHints([]string{"To view available contexts run `sqlcmd config get-contexts`"},
			"No context exists with the name: \"%v\"", c.name)
	}
}
