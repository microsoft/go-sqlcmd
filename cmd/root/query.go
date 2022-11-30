// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/mssql"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/pkg/console"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
)

type Query struct {
	cmdparser.Cmd

	text string
}

func (c *Query) DefineCommand(output.Output, ...cmdparser.Command) {
	c.Cmd.SetOptions(cmdparser.Options{
		Use:   "query",
		Short: "Run a query against the current context",
		Examples: []cmdparser.ExampleInfo{
			{Description: "Run a query", Steps: []string{
				`sqlcmd query "SELECT @@SERVERNAME"`,
				`sqlcmd query --text "SELECT @@SERVERNAME"`,
				`sqlcmd query --query "SELECT @@SERVERNAME"`,
			}}},
		Run: c.run,
		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagInfo{
			Flag:  "text",
			Value: &c.text,
		},
	})

	c.Cmd.DefineCommand()

	c.AddFlag(cmdparser.FlagOptions{
		String:    &c.text,
		Name:      "text",
		Shorthand: "t",
		Usage:     "Command text to run"})

	// BUG(stuartpa): Decide on if --text or --query is best
	c.AddFlag(cmdparser.FlagOptions{
		String:    &c.text,
		Name:      "query",
		Shorthand: "q",
		Usage:     "Command text to run"})
}

func (c *Query) run() {
	endpoint, user := config.GetCurrentContext()

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
