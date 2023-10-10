// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// GetContexts implements the `sqlcmd config get-contexts` command
type GetContexts struct {
	cmdparser.Cmd

	name     string
	detailed bool
}

func (c *GetContexts) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "get-contexts",
		Short: localizer.Sprintf("Display one or many contexts from the sqlconfig file"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("List all the context names in your sqlconfig file"),
				Steps:       []string{"sqlcmd config get-contexts"},
			},
			{
				Description: localizer.Sprintf("List all the contexts in your sqlconfig file"),
				Steps:       []string{"sqlcmd config get-contexts --detailed"},
			},
			{
				Description: localizer.Sprintf("Describe one context in your sqlconfig file"),
				Steps:       []string{"sqlcmd config get-contexts my-context"},
			},
		},
		Run: c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  localizer.Sprintf("Context name to view details of")})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.detailed,
		Name:  "detailed",
		Usage: localizer.Sprintf("Include context details")})
}

func (c *GetContexts) run() {
	output := c.Output()

	if c.name != "" {
		if config.ContextExists(c.name) {
			context := config.GetContext(c.name)
			output.Struct(context)
		} else {
			output.FatalWithHints(
				[]string{localizer.Sprintf("To view available contexts run `%s`", localizer.GetContextCommand)},
				localizer.Sprintf("error: no context exists with the name: \"%v\"", c.name))
		}
	} else {
		config.OutputContexts(output.Struct, c.detailed)
	}
}
