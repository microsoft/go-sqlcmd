// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
)

type Root struct {
	cmdparser.Cmd
}

func (c *Root) DefineCommand(subCommands ...cmdparser.Command) {
	c.Cmd.Options = cmdparser.Options{
		Use:   "sqlcmd",
		Short: "sqlcmd: a command-line interface for the #SQLFamily",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "Run a query",
				Steps:       []string{`sqlcmd query "SELECT @@SERVERNAME"`}}},
	}

	c.Cmd.DefineCommand(subCommands...)
	c.addGlobalFlags()
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
		String:        &configFilename,
		DefaultString: config.DefaultFileName(),
		Name:          "sqlconfig",
		Usage:         "Configuration file",
	})

	c.AddFlag(cmdparser.FlagOptions{
		String:        &outputType,
		DefaultString: "yaml",
		Name:          "output",
		Shorthand:     "o",
		Usage:         "output type (yaml, json or xml)",
	})

	c.AddFlag(cmdparser.FlagOptions{
		Int:        &loggingLevel,
		DefaultInt: 2,
		Name:       "verbosity",
		Shorthand:  "v",
		Usage:      "Log level, error=0, warn=1, info=2, debug=3, trace=4",
	})
}
