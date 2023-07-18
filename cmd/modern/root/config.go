// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"fmt"

	"github.com/microsoft/go-sqlcmd/cmd/modern/root/config"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/internal/telemetry"
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
		Short:       localizer.Sprintf(`Modify sqlconfig files using subcommands like "%s"`, localizer.UseContextCommand),
		SubCommands: c.SubCommands(),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("Add context for existing endpoint and user (use %s or %s)", localizer.PasswordEnvVar, localizer.PasswordEnvVar2),
				Steps: []string{
					fmt.Sprintf("%s SQLCMD_PASSWORD=<placeholderpassword>", pal.CreateEnvVarKeyword()),
					"sqlcmd config add-user --name sa1434 --username sa",
					fmt.Sprintf("%s SQLCMD_PASSWORD=", pal.CreateEnvVarKeyword()),
					"sqlcmd config add-endpoint --name ep1434 --address localhost --port 1434",
					"sqlcmd config add-context --name mssql1434 --user sa1434 --endpoint ep1434"},
			},
		},
	}

	c.Cmd.DefineCommand(options)
	telemetry.TrackEvent("config")
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
