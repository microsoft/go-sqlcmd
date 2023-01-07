// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/mssql"
	"github.com/microsoft/go-sqlcmd/pkg/console"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
)

// Query defines the `sqlcmd query` command
type Query struct {
	cmdparser.Cmd

	text string
}

func (c *Query) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "query",
		Short: "Run a query against the current context",
		Examples: []cmdparser.ExampleOptions{
			{Description: "Run a query", Steps: []string{
				`sqlcmd query "SELECT @@SERVERNAME"`,
				`sqlcmd query --text "SELECT @@SERVERNAME"`,
				`sqlcmd query --query "SELECT @@SERVERNAME"`,
			}}},
		Run: c.run,
		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{
			Flag:  "text",
			Value: &c.text,
		},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String:    &c.text,
		Name:      "text",
		Shorthand: "t",
		Usage:     "Command text to run"})

	// BUG(stuartpa): Decide on if --text or --query is best (or leave both for convenience)
	c.AddFlag(cmdparser.FlagOptions{
		String:    &c.text,
		Name:      "query",
		Shorthand: "q",
		Usage:     "Command text to run"})
}

// run executes the Query command.
// It connects to a SQL Server endpoint using the current context from the config file,
// and either runs an interactive SQL console or executes the provided query.
// If an error occurs, it is handled by the CheckErr function.
func (c *Query) run() {
	endpoint, user := config.CurrentContext()

	var line sqlcmd.Console = nil
	if c.text == "" {
		line = console.NewConsole("")
		defer line.Close()
	}
	s := mssql.Connect(endpoint, user, line)
	if c.text == "" {
		err := s.Run(false, false)
		c.CheckErr(err)
	} else {
		mssql.Query(s, c.text)
	}
}
