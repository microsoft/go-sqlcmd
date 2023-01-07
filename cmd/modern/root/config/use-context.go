// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
)

// UseContext implements the `sqlcmd config use-context` command
type UseContext struct {
	cmdparser.Cmd

	name string
}

func (c *UseContext) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "use-context",
		Short: "Set the current context",
		Examples: []cmdparser.ExampleOptions{{
			Description: "Set the mssql context (endpoint/user) to be the current context",
			Steps:       []string{"sqlcmd config use-context mssql"}}},
		Aliases: []string{"use", "change-context", "set-context"},
		Run:     c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  "Name of context to set as current context"})
}

func (c *UseContext) run() {
	output := c.Output()

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
