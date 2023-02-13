// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/config"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

// Config defines the `sqlcmd config` sub-commands
type Config struct {
	cmdparser.Cmd
}

// DefineCommand defines the `sqlcmd config` command, which is only
// more sub-commands (`sqlcmd config` does not `run` anything itself)
func (c *Config) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:         "config",
		Short:       `Modify sqlconfig files using subcommands like "sqlcmd config use-context mssql"`,
		SubCommands: c.SubCommands(),
	}

	c.Cmd.DefineCommand(options)
}

// SubCommands sets up all the sub-commands for `sqlcmd config`
func (c *Config) SubCommands() []cmdparser.Command {
	dependencies := c.Dependencies()

	return []cmdparser.Command{
		cmdparser.New[*config.AddContext](dependencies),
		cmdparser.New[*config.AddEndpoint](dependencies),
		cmdparser.New[*config.AddUser](dependencies),
		cmdparser.New[*config.ConnectionStrings](dependencies),
		cmdparser.New[*config.CurrentContext](dependencies),
		cmdparser.New[*config.DeleteContext](dependencies),
		cmdparser.New[*config.DeleteEndpoint](dependencies),
		cmdparser.New[*config.DeleteUser](dependencies),
		cmdparser.New[*config.GetContexts](dependencies),
		cmdparser.New[*config.GetEndpoints](dependencies),
		cmdparser.New[*config.GetUsers](dependencies),
		cmdparser.New[*config.UseContext](dependencies),
		cmdparser.New[*config.View](dependencies),
	}
}
