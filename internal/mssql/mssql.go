// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package mssql

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/sqlconfig"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"os"
)

func Connect(
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

func Query(s *sqlcmd.Sqlcmd, text string) {
	s.Query = text
	s.SetOutput(os.Stdout)
	s.SetError(os.Stderr)
	err := s.Run(true, false)
	checkErr(err)
}
