// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
)

// Root type implements the very top-level command for sqlcmd (which contains
// all the sub-commands, like install, query, config etc.
type Root struct {
	cmdparser.Cmd

	configFilename string
	loggingLevel   int
	outputType     string
}

// DefineCommand defines the top-level sqlcmd sub-commands.
// It sets the cli name, description, and subcommands, and adds global flags.
// It also provides usage examples for sqlcmd.
func (c *Root) DefineCommand(...cmdparser.CommandOptions) {
	examples := []cmdparser.ExampleOptions{
		{
			Description: "Install, Query, Uninstall SQL Server",
			Steps: []string{
				"sqlcmd install mssql",
				`sqlcmd query "SELECT @@version"`,
				"sqlcmd uninstall"}}}

	commandOptions := cmdparser.CommandOptions{
		Use:         "sqlcmd",
		Short:       "sqlcmd: Install/Create/Query SQL Server, Azure SQL, and Tools",
		SubCommands: c.SubCommands(),
		Examples:    examples,
	}

	c.Cmd.DefineCommand(commandOptions)
	c.addGlobalFlags()
}

// SubCommands returns a slice of subcommands for the Root command.
// The returned subcommands are Config, Install, query, and Uninstall.
func (c *Root) SubCommands() []cmdparser.Command {
	dependencies := c.Dependencies()

	return []cmdparser.Command{
		cmdparser.New[*root.Config](dependencies),
		cmdparser.New[*root.Install](dependencies),
		cmdparser.New[*root.Query](dependencies),
		cmdparser.New[*root.Uninstall](dependencies),
	}
}

// Execute runs the application based on the command-line
// parameters the user has passed in.
func (c *Root) Execute() {
	c.Cmd.Execute()
}

// IsValidSubCommand is TEMPORARY code, that will be removed when
// we enable the new cobra based CLI by default.  It returns true if the
// command-line provided by the user indicates they want the new cobra
// based CLI, e.g. sqlcmd install, or sqlcmd query, or sqlcmd --help etc.
func (c *Root) IsValidSubCommand(command string) bool {
	return c.IsSubCommand(command)
}

func (c *Root) addGlobalFlags() {
	c.AddFlag(cmdparser.FlagOptions{
		Bool:      &globalOptions.TrustServerCertificate,
		Name:      "trust-server-certificate",
		Shorthand: "C",
		Usage:     "Whether to trust the certificate presented by the endpoint for encryption",
	})

	c.AddFlag(cmdparser.FlagOptions{
		String:    &globalOptions.DatabaseName,
		Name:      "database-name",
		Shorthand: "d",
		Usage:     "The initial database for the connection",
	})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:      &globalOptions.UseTrustedConnection,
		Name:      "use-trusted-connection",
		Shorthand: "E",
		Usage:     "Whether to use integrated security",
	})

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.configFilename,
		DefaultString: config.DefaultFileName(),
		Name:          "sqlconfig",
		Usage:         "Configuration file",
	})

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.outputType,
		DefaultString: "yaml",
		Name:          "output",
		Shorthand:     "o",
		Usage:         "output type (yaml, json or xml)",
	})

	c.AddFlag(cmdparser.FlagOptions{
		Int:        (*int)(&c.loggingLevel),
		DefaultInt: 2,
		Name:       "verbosity",
		Shorthand:  "v",
		Usage:      "Log level, error=0, warn=1, info=2, debug=3, trace=4",
	})
}
