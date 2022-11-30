// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"github.com/microsoft/go-sqlcmd/internal"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type Root struct {
	cmdparser.Cmd

	loggingLevel   int
	outputType     string
	configFilename string
}

func (c *Root) ConfigFilename() string {
	return c.configFilename
}

func (c *Root) DefineCommand(output output.Output, subCommands ...cmdparser.Command) {
	c.Cmd.SetOptions(cmdparser.Options{
		Use:   "sqlcmd",
		Short: "sqlcmd: a command-line interface for the #SQLFamily",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "Run a query",
				Steps:       []string{`sqlcmd query "SELECT @@SERVERNAME"`}}},
	})

	c.Cmd.DefineCommand(
	c.addGlobalFlags()
}

// Execute runs the application based on the command-line
// parameters the user has passed in.
func (c *Root) Execute() {
	c.Cmd.Execute()
}

func (c *Root) LoggingLevel() int {
	return c.loggingLevel
}

func (c *Root) OutputType() string {
	return c.outputType
}

// Initialize initializes the command-line interface. The func passed into
// cmdparser.Initialize is called after the command-line from the user has been
// parsed, so the helpers are initialized with the values from the command-line
// like '-v 4' which sets the logging level to maximum etc.
func (c *Root) InitializeCallback() {
	config.SetFileName(c.ConfigFilename())
	config.Load()

	options := internal.InitializeOptions{
		ErrorHandler: c.checkErr,
		HintHandler:  c.displayHints,
		OutputType:   "yaml",
		LoggingLevel: 2,
	}
	o := internal.Initialize(options)
	c.SetOutput(o)
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
		Int:        &c.loggingLevel,
		DefaultInt: 2,
		Name:       "verbosity",
		Shorthand:  "v",
		Usage:      "Log level, error=0, warn=1, info=2, debug=3, trace=4",
	})
}

// checkErr uses Cobra to check err, and halts the application if err is not
// nil.  Pass (inject) checkErr into all dependencies (helpers etc.) as an
// errorHandler.
//
// To aid debugging issues, if the logging level is > 2 (e.g. -v 3 or -4), we
// panic which outputs a stacktrace.
func (c *Root) checkErr(err error) {
	if c.LoggingLevel() > 2 {
		if err != nil {
			panic(err)
		}
	}
	c.CheckErr(err)
}

// displayHints displays helpful information on what the user should do next
// to make progress.  displayHints is injected into dependencies (helpers etc.)
func (c *Root) displayHints(hints []string) {
	output := c.Output()

	if len(hints) > 0 {
		output.Infof("\nHINT:")
		for i, hint := range hints {
			output.Infof("  %d. %v", i+1, hint)
		}
	}
}
