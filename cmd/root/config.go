// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

type Config struct {
	cmdparser.Cmd
}

func (c *Config) DefineCommand(subCommands ...cmdparser.Command) {
	c.Cmd.Options = cmdparser.Options{
		Use:   "config",
		Short: `Modify sqlconfig files using subcommands like "sqlcmd config use-context mssql"`,
	}
	c.Cmd.DefineCommand(subCommands...)
}
