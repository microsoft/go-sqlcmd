// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/pkg/console"
	"os"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
)

// Connect is used to connect to a SQL Server using the specified endpoint
// and user details. The console parameter is used to output messages during
// the connection process. The function returns a Sqlcmd instance that can
// be used to run SQL commands on the server.
func (m *mssql) Connect(
	endpoint sqlconfig.Endpoint,
	user *sqlconfig.User,
	database string,
	options ConnectOptions,
) {
	v := sqlcmd.InitializeVariables(true)
	if options.Interactive {
		m.console = console.NewConsole("")
		defer m.console.Close()
	} else {
		m.console = nil
	}
	m.sqlcmd = sqlcmd.New(m.console, "", v)
	m.sqlcmd.Format = sqlcmd.NewSQLCmdDefaultFormatter(false)
	connect := sqlcmd.ConnectSettings{
		ServerName: fmt.Sprintf(
			"%s,%d",
			endpoint.EndpointDetails.Address,
			endpoint.EndpointDetails.Port),
		ApplicationName: "sqlcmd",
	}

	if database != "" {
		connect.Database = database
	}

	if user == nil {
		connect.UseTrustedConnection = true
	} else {
		if user.AuthenticationType == "basic" {
			connect.UseTrustedConnection = false
			connect.UserName = user.BasicAuth.Username
			connect.Password = decryptCallback(
				user.BasicAuth.Password,
				user.BasicAuth.PasswordEncrypted,
			)
		} else {
			panic("Authentication not supported")
		}
	}

	trace("Connecting to server %v", connect.ServerName)
	err := m.sqlcmd.ConnectDb(&connect, true)
	checkErr(err)
}

// Query is helper function that allows running a given SQL query on a
// provided sqlcmd.Sqlcmd object. It takes the sqlcmd.Sqlcmd object and the
// query text as inputs, and runs the query using the Run method of
// the sqlcmd.Sqlcmd object. It sets the standard output and standard error
// to be the same as the current process, and returns the error if any occurred
// during the execution of the query.
func (m *mssql) Query(text string) {
	if m.console == nil {
		m.sqlcmd.Query = text
		m.sqlcmd.SetOutput(os.Stdout)
		m.sqlcmd.SetError(os.Stderr)
		trace("Running query: %v", text)
		err := m.sqlcmd.Run(true, false)
		checkErr(err)
	} else {
		err := m.sqlcmd.Run(false, true)
		checkErr(err)
	}
}
