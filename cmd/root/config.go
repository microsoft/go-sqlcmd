// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type Config struct {
	cmdparser.Cmd
}

func (c *Config) DefineCommand(output output.Output, subCommands ...cmdparser.Command) {
	c.Cmd.SetOptions(cmdparser.Options{
		Use:   "config",
		Short: `Modify sqlconfig files using subcommands like "sqlcmd config use-context mssql"`,
	})
	c.Cmd.DefineCommand(
}
