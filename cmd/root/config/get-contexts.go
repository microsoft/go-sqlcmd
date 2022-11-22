// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type GetContexts struct {
	cmdparser.Cmd

	name     string
	detailed bool
}

func (c *GetContexts) DefineCommand(...cmdparser.Command) {
	c.Cmd.Options = cmdparser.Options{
		Use:   "get-contexts",
		Short: "Display one or many contexts from the sqlconfig file",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "List all the context names in your sqlconfig file",
				Steps:       []string{"sqlcmd config get-contexts"},
			},
			{
				Description: "List all the contexts in your sqlconfig file",
				Steps:       []string{"sqlcmd config get-contexts --detailed"},
			},
			{
				Description: "Describe one context in your sqlconfig file",
				Steps:       []string{"sqlcmd config get-contexts my-context"},
			},
		},
		Run: c.run,

		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagInfo{Flag: "name", Value: &c.name},
	}

	c.Cmd.DefineCommand()

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.name,
		Name:   "name",
		Usage:  "Context name to view details of"})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.detailed,
		Name:  "detailed",
		Usage: "Include context details"})
}

func (c *GetContexts) run() {
	if c.name != "" {
		if config.ContextExists(c.name) {
			context := config.GetContext(c.name)
			output.Struct(context)
		} else {
			output.FatalfWithHints(
				[]string{"To view available contexts run `sqlcmd config get-contexts`"},
				"error: no context exists with the name: \"%v\"",
				c.name)
		}
	} else {
		config.OutputContexts(output.Struct, c.detailed)
	}
}
