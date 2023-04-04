// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"github.com/microsoft/go-sqlcmd/internal/sql"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer"
)

type Use struct {
	cmdparser.Cmd

	url string

	sql sql.Sql
}

func (c *Use) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "use",
		Short: "Download (into container) and use database",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Download AdventureWorksLT into container for current context, set as default database",
				Steps:       []string{`sqlcmd use https://aka.ms/AdventureWorksLT.bak`}},
		},
		Run:                        c.run,
		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{Flag: "url", Value: &c.url},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.url,
		Name:   "url",
		Usage:  "Name of context to set as current context"})
}

func (c *Use) run() {
	output := c.Output()

	if config.CurrentContextName() == "" {
		output.FatalfWithHintExamples([][]string{
			{"To view available contexts", "sqlcmd config get-contexts"},
		}, "No current context")
	}
	if config.CurrentContextEndpointHasContainer() {
		controller := container.NewController()
		id := config.ContainerId()

		if !controller.ContainerRunning(id) {
			output.FatalfWithHintExamples([][]string{
				{"Start container for current context", "sqlcmd start"},
			}, "Container for current context is not running")
		}

		endpoint, user := config.CurrentContext()

		c.sql = sql.New(sql.SqlOptions{UnitTesting: false})
		c.sql.Connect(endpoint, user, sql.ConnectOptions{Database: "master", Interactive: false})

		mssqlcontainer.DownloadAndRestoreDb(
			controller,
			id,
			c.url,
			user.BasicAuth.Username,
			secret.Decode(user.BasicAuth.Password, user.BasicAuth.PasswordEncryption),
			c.query,
			c.Cmd.Output(),
		)

	} else {
		output.FatalfWithHintExamples([][]string{
			{"Create new context with a sql container ", "sqlcmd create mssql"},
		}, "Current context does not have a container")
	}
}

func (c *Use) query(commandText string) {
	c.sql.Query(commandText)
}
