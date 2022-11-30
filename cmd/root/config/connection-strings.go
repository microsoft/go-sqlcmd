// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/internal/secret"
)

type ConnectionStrings struct {
	cmdparser.Cmd
}

func (c *ConnectionStrings) DefineCommand(output.Output, ...cmdparser.Command) {
	c.Cmd.SetOptions(cmdparser.Options{
		Use:   "connection-strings",
		Short: "Display connections strings for the current context",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "List connection strings for all client drivers",
				Steps: []string{
					"sqlcmd config connection-strings",
					"sqlcmd config cs"},
			},
		},
		Run:     c.run,
		Aliases: []string{"cs"},
	})

	c.Cmd.DefineCommand()
}

func (c *ConnectionStrings) run() {
	output := c.Output()

	// connectionStringFormats borrowed from "portal.azure.com" "connection strings" pane
	var connectionStringFormats = map[string]string{
		"ADO.NET": "Server=tcp:%s,%d;Initial Catalog=%s;Persist Security options=False;User ID=%s;Password=%s;MultipleActiveResultSets=False;Encode=True;TrustServerCertificate=False;Connection Timeout=30;",
		"JDBC":    "jdbc:sqlserver://%s:%d;database=%s;user=%s;password=%s;encrypt=true;trustServerCertificate=false;loginTimeout=30;",
		"ODBC":    "Driver={ODBC Driver 13 for SQL Server};Server=tcp:%s,%d;Database=%s;Uid=%s;Pwd=%s;Encode=yes;TrustServerCertificate=no;Connection Timeout=30;",
	}

	endpoint, user := config.GetCurrentContext()
	for k, v := range connectionStringFormats {
		connectionStringFormats[k] = fmt.Sprintf(v,
			endpoint.EndpointDetails.Address,
			endpoint.EndpointDetails.Port,
			"master",
			user.BasicAuth.Username,
			secret.Decode(user.BasicAuth.Password, user.BasicAuth.PasswordEncrypted))
	}

	format := pal.CmdLineWithEnvVars(
		[]string{"SQLCMDPASSWORD=%s"},
		"sqlcmd -S %s,%d -U %s",
	)

	connectionStringFormats["SQLCMD"] = fmt.Sprintf(format,
		secret.Decode(user.BasicAuth.Password, user.BasicAuth.PasswordEncrypted),
		endpoint.EndpointDetails.Address,
		endpoint.EndpointDetails.Port,
		user.BasicAuth.Username)

	output.Struct(connectionStringFormats)
}
