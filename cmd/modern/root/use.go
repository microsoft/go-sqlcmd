// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"github.com/microsoft/go-sqlcmd/internal/sql"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest"
)

type Use struct {
	cmdparser.Cmd

	url          string
	useMechanism string

	sql sql.Sql
}

func (c *Use) DefineCommand(...cmdparser.CommandOptions) {
	examples := []cmdparser.ExampleOptions{
		{
			Description: "Download AdventureWorksLT into container for current context, set as default database",
			Steps:       []string{`sqlcmd use https://aka.ms/AdventureWorksLT.bak`}},
	}

	options := cmdparser.CommandOptions{
		Use:                        "use",
		Short:                      fmt.Sprintf("Download database (into container) (%s)", ingest.ValidFileExtensions()),
		Examples:                   examples,
		Run:                        c.run,
		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{Flag: "url", Value: &c.url},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.url,
		Name:   "url",
		Usage:  "Name of context to set as current context"})

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.useMechanism,
		DefaultString: "",
		Name:          "use-mechanism",
		Usage:         "Mechanism to use to bring database online (attach, restore, dacfx)",
	})
}

func (c *Use) run() {
	output := useOutput{output: c.Output()}

	controller := container.NewController()
	id := config.ContainerId()

	if !config.CurrentContextEndpointHasContainer() {
		output.FatalNoContainerInCurrentContext()
	}

	if !controller.ContainerRunning(id) {
		output.FatalContainerNotRunning()
	}

	endpoint, user := config.CurrentContext()

	c.sql = sql.NewSql(sql.SqlOptions{})
	c.sql.Connect(endpoint, user, sql.ConnectOptions{Database: "master"})

	useDatabase := ingest.NewIngest(c.url, controller, ingest.IngestOptions{
		Mechanism: c.useMechanism,
	})

	if !useDatabase.SourceFileExists() {
		output.FatalDatabaseSourceFileNotExist(c.url)
	}

	// Copy source file (e.g. .bak/.bacpac etc.) for database to be made available to container
	useDatabase.CopyToContainer(id)

	if useDatabase.IsExtractionNeeded() {
		output.output.Infof("Extracting files from %q", useDatabase.UrlFilename())
		useDatabase.Extract()
	}

	useDatabase.BringOnline(
		c.sql.Query,
		user.BasicAuth.Username,
		secret.Decode(user.BasicAuth.Password, user.BasicAuth.PasswordEncryption),
	)
}

func (c *Use) query(commandText string) {
	c.sql.Query(commandText)
}
