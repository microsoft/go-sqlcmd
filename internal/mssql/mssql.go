// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package mssql

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"os"
)

// Connect is used to connect to a SQL Server using the specified endpoint
// and user details. The console parameter is used to output messages during
// the connection process. The function returns a Sqlcmd instance that can
// be used to run SQL commands on the server.
func (m *MssqlType) Connect(
	endpoint sqlconfig.Endpoint,
	user *sqlconfig.User,
	console sqlcmd.Console,
) *sqlcmd.Sqlcmd {
	v := sqlcmd.InitializeVariables(true)
	s := sqlcmd.New(console, "", v)
	s.Format = sqlcmd.NewSQLCmdDefaultFormatter(false)
	connect := sqlcmd.ConnectSettings{
		ServerName: fmt.Sprintf(
			"%s,%d",
			endpoint.EndpointDetails.Address,
			endpoint.EndpointDetails.Port),
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
	err := s.ConnectDb(&connect, true)
	checkErr(err)
	return s
}

// Query is helper function that allows running a given SQL query on a
// provided sqlcmd.Sqlcmd object. It takes the sqlcmd.Sqlcmd object and the
// query text as inputs, and runs the query using the Run method of
// the sqlcmd.Sqlcmd object. It sets the standard output and standard error
// to be the same as the current process, and returns the error if any occurred
// during the execution of the query.
func (m *MssqlType) Query(s *sqlcmd.Sqlcmd, text string) {
	s.Query = text
	s.SetOutput(os.Stdout)
	s.SetError(os.Stderr)
	err := s.Run(true, false)
	checkErr(err)
}
