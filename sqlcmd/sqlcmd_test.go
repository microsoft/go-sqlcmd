// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"database/sql"
	"fmt"
	"os"
	"os/user"
	"testing"

	"github.com/google/uuid"
	"github.com/microsoft/go-sqlcmd/variables"
	"github.com/stretchr/testify/assert"
)

func TestConnectionStringFromSqlCmd(t *testing.T) {
	type connectionStringTest struct {
		settings         *ConnectSettings
		setup            func(*variables.Variables)
		connectionString string
	}

	pwd := uuid.New().String()

	commands := []connectionStringTest{

		{nil, nil, "sqlserver://."},
		{
			&ConnectSettings{TrustServerCertificate: true},
			func(vars *variables.Variables) {
				variables.Setvar(variables.SQLCMDDBNAME, "somedatabase")
			},
			"sqlserver://.?database=somedatabase&trustservercertificate=true",
		},
		{
			&ConnectSettings{TrustServerCertificate: true},
			func(vars *variables.Variables) {
				vars.Set(variables.SQLCMDSERVER, `someserver/instance`)
				vars.Set(variables.SQLCMDDBNAME, "somedatabase")
				vars.Set(variables.SQLCMDUSER, "someuser")
				vars.Set(variables.SQLCMDPASSWORD, pwd)
			},
			fmt.Sprintf("sqlserver://someuser:%s@someserver/instance?database=somedatabase&trustservercertificate=true", pwd),
		},
		{
			&ConnectSettings{TrustServerCertificate: true, UseTrustedConnection: true},
			func(vars *variables.Variables) {
				vars.Set(variables.SQLCMDSERVER, `tcp:someserver,1045`)
				vars.Set(variables.SQLCMDUSER, "someuser")
				vars.Set(variables.SQLCMDPASSWORD, pwd)
			},
			"sqlserver://someserver:1045?trustservercertificate=true",
		},
		{
			nil,
			func(vars *variables.Variables) {
				vars.Set(variables.SQLCMDSERVER, `tcp:someserver,1045`)
			},
			"sqlserver://someserver:1045",
		},
	}

	for _, test := range commands {
		v := variables.InitializeVariables(false)
		if test.setup != nil {
			test.setup(v)
		}
		s := &Sqlcmd{vars: v}
		if test.settings != nil {
			s.Connect = *test.settings
		}
		connectionString, err := s.ConnectionString()
		if assert.NoError(t, err, "Unexpected error from %+v", s) {
			assert.Equal(t, test.connectionString, connectionString, "Wrong connection string from: %+v", *s)
		}
	}
}

/* The following tests require a working SQL instance and rely on SqlCmd environment variables
to manage the initial connection string. The default connection when no environment variables are
set will be to localhost using Windows auth.

*/
func TestSqlCmdConnectDb(t *testing.T) {
	v := variables.InitializeVariables(true)
	s := &Sqlcmd{vars: v}
	err := s.ConnectDb("", "", "", false)
	if assert.NoError(t, err, "ConnectDb should succeed") {
		sqlcmduser := os.Getenv(variables.SQLCMDUSER)
		if sqlcmduser == "" {
			u, _ := user.Current()
			sqlcmduser = u.Username
		}
		assert.Equal(t, sqlcmduser, s.vars.SqlCmdUser(), "SQLCMDUSER variable should match connected user")
	}
}

func ConnectDb() (*sql.DB, error) {
	v := variables.InitializeVariables(true)
	s := &Sqlcmd{vars: v}
	err := s.ConnectDb("", "", "", false)
	return s.db, err
}
