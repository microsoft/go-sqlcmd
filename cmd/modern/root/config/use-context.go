// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// UseContext implements the `sqlcmd config use-context` command
type UseContext struct {
	cmdparser.Cmd

	name string
}

func (c *UseContext) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "use-context",
		Short: localizer.Sprintf("Set the current context"),
		Examples: []cmdparser.ExampleOptions{{
			Description: localizer.Sprintf("Set the mssql context (endpoint/user) to be the current context"),
			Steps:       []string{"sqlcmd config use-context mssql"}}},
		Aliases: []string{"use", "change-context", "set-context"},
		Run:     c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  localizer.Sprintf("Name of context to set as current context")})
}

func (c *UseContext) run() {
	output := c.Output()

	if config.ContextExists(c.name) {
		config.SetCurrentContextName(c.name)
		output.InfoWithHints([]string{
			localizer.Sprintf("To run a query:    %s", localizer.RunQueryExample),
			localizer.Sprintf("To remove:         %s", localizer.UninstallCommand)},
			localizer.Sprintf("Switched to context \"%v\".", c.name))
	} else {
		output.FatalWithHints([]string{localizer.Sprintf("To view available contexts run `%s`", localizer.GetContextCommand)},
			localizer.Sprintf("No context exists with the name: \"%v\"", c.name))
	}
}
