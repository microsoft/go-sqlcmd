// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"runtime"
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
	// Example usage steps
	steps := []string{"sqlcmd create mssql --accept-eula --using https://aka.ms/AdventureWorksLT.bak"}

	if runtime.GOOS == "windows" {
		steps = append(steps, "sqlcmd open ads")
	}

	steps = append(steps, `sqlcmd query "SELECT @version"`)
	steps = append(steps, "sqlcmd delete")

	examples := []cmdparser.ExampleOptions{{
		Description: "Install/Create, Query, Uninstall SQL Server",
		Steps:       steps}}

	commandOptions := cmdparser.CommandOptions{
		Use: "sqlcmd",
		Short: `sqlcmd: Install/Create/Query SQL Server, Azure SQL, and Tools

Feedback:
  https://github.com/microsoft/go-sqlcmd/issues/new`,
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

	subCommands := []cmdparser.Command{
		cmdparser.New[*root.Config](dependencies),
		cmdparser.New[*root.Install](dependencies),
		cmdparser.New[*root.Query](dependencies),
		cmdparser.New[*root.Start](dependencies),
		cmdparser.New[*root.Stop](dependencies),
		cmdparser.New[*root.Uninstall](dependencies),
	}

	// BUG:(stuartpa) - Add Mac / Linux support
	if runtime.GOOS == "windows" {
		subCommands = append(subCommands, cmdparser.New[*root.Open](dependencies))
	}

	return subCommands
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
	var unused bool
	c.AddFlag(cmdparser.FlagOptions{
		Bool:      &unused,
		Name:      "?",
		Shorthand: "?",
		Usage:     "help for backwards compatibility flags (-S, -U, -E etc.)",
	})

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.configFilename,
		DefaultString: config.DefaultFileName(),
		Name:          "sqlconfig",
		Usage:         "Configuration file",
	})

	/* BUG:(stuartpa) - At the moment this is a top level flag, but it doesn't
	work with all sub-commands (e.g. query), so removing for now.
	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.outputType,
		DefaultString: "json",
		Name:          "output",
		Shorthand:     "o",
		Usage:         "output type (yaml, json or xml)",
	})
	*/

	c.AddFlag(cmdparser.FlagOptions{
		Int:        (*int)(&c.loggingLevel),
		DefaultInt: 2,
		Name:       "verbosity",
		Usage:      "Log level, error=0, warn=1, info=2, debug=3, trace=4",
	})
}
