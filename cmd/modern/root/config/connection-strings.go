// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/sql"
	"strings"

	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/internal/secret"
)

// ConnectionStrings implements the `sqlcmd config connection-strings` command
type ConnectionStrings struct {
	cmdparser.Cmd

	database string
}

func (c *ConnectionStrings) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "connection-strings",
		Short: "Display connections strings for the current context",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "List connection strings for all client drivers",
				Steps: []string{
					"sqlcmd config connection-strings",
					"sqlcmd config cs"},
			},
		},
		Run:     c.run,
		Aliases: []string{"cs"},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.database,
		Name:          "database",
		DefaultString: "",
		Shorthand:     "d",
		Usage:         "Database for the connection string (default is taken from the T/SQL login)"})
}

// run generates connection strings for the current context in multiple formats.
// The generated connection strings will include the current endpoint and user information.
func (c *ConnectionStrings) run() {
	output := c.Output()

	// connectionStringFormats borrowed from "portal.azure.com" "connection strings" pane
	var connectionStringFormats = map[string]string{
		"ADO.NET": "Server=tcp:%s,%d;Initial Catalog=%s;Persist Security Info=False;User ID=%s;Password=%s;MultipleActiveResultSets=False;Encrypt=True;TrustServerCertificate=%s;Connection Timeout=30;",
		"JDBC":    "jdbc:sqlserver://%s:%d;database=%s;user=%s;password=%s;encrypt=true;trustServerCertificate=%s;loginTimeout=30;",
		"ODBC":    "Driver={ODBC Driver 18 for SQL Server};Server=tcp:%s,%d;Database=%s;Uid=%s;Pwd=%s;Encrypt=yes;TrustServerCertificate=%s;Connection Timeout=30;",
		"GO":      "sqlserver://%s:%s@%s,%d?database=%s;encrypt=true;trustServerCertificate=%s;dial+timeout=30",
		"SQLCMD":  "sqlcmd -S %s,%d -U %s",
	}

	endpoint, user := config.CurrentContext()

	if c.database == "" {
		if endpoint.AssetDetails != nil && endpoint.AssetDetails.ContainerDetails != nil {
			controller := container.NewController()
			if controller.ContainerRunning(endpoint.AssetDetails.ContainerDetails.Id) {
				s := sql.New(sql.SqlOptions{})
				s.Connect(endpoint, user, sql.ConnectOptions{Interactive: false})
				c.database = s.ScalarString("PRINT DB_NAME()")
			} else {
				c.database = "master"
			}
		} else {
			c.database = "master"
		}
	}

	if user != nil {
		for k, v := range connectionStringFormats {
			if k == "GO" {
				connectionStringFormats[k] = fmt.Sprintf(
					v,
					user.BasicAuth.Username,
					secret.Decode(user.BasicAuth.Password, user.BasicAuth.PasswordEncrypted),
					endpoint.EndpointDetails.Address,
					endpoint.EndpointDetails.Port,
					c.database,
					c.stringForBoolean(c.trustServerCertificate(endpoint), k))
			} else if k == "SQLCMD" {
				format := pal.CmdLineWithEnvVars(
					[]string{"SQLCMDPASSWORD=%s"},
					"sqlcmd -S %s,%d -U %s -d %s",
				)

				connectionStringFormats[k] = fmt.Sprintf(format,
					secret.Decode(user.BasicAuth.Password, user.BasicAuth.PasswordEncrypted),
					endpoint.EndpointDetails.Address,
					endpoint.EndpointDetails.Port,
					user.BasicAuth.Username,
					c.database)
			} else {
				connectionStringFormats[k] = fmt.Sprintf(v,
					endpoint.EndpointDetails.Address,
					endpoint.EndpointDetails.Port,
					c.database,
					user.BasicAuth.Username,
					secret.Decode(user.BasicAuth.Password, user.BasicAuth.PasswordEncrypted),
					c.stringForBoolean(c.trustServerCertificate(endpoint), k))
			}
		}

		for k, v := range connectionStringFormats {
			output.Infof("%-8s %s", k+":", v)
		}
	} else {
		output.Infof("Connection Strings only supported for Basic Auth type")
	}
}

func (c *ConnectionStrings) trustServerCertificate(endpoint sqlconfig.Endpoint) (trustServerCertificate bool) {
	// Per issue:
	//	https://github.com/microsoft/go-sqlcmd/issues/249
	// set trustServerCertificate to "False" if connecting to Azure SQL
	trustServerCertificate = true
	if strings.HasSuffix(
		strings.ToLower(endpoint.EndpointDetails.Address),
		".database.windows.net") {
		trustServerCertificate = false
	}
	return
}

func (c *ConnectionStrings) stringForBoolean(value bool, protocol string) (s string) {
	// Each connection string has a different way of setting booleans
	if protocol == "JDBC" || protocol == "GO" {
		if value {
			s = "true"
		} else {
			s = "false"
		}
	} else if protocol == "ODBC" {
		if value {
			s = "yes"
		} else {
			s = "no"
		}
	} else {
		if value {
			s = "True"
		} else {
			s = "False"
		}
	}
	return
}
