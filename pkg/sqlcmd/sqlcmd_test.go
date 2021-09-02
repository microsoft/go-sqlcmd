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
	"github.com/stretchr/testify/assert"
)

func TestConnectionStringFromSqlCmd(t *testing.T) {
	type connectionStringTest struct {
		settings         *ConnectSettings
		setup            func(*Variables)
		connectionString string
	}

	pwd := uuid.New().String()

	commands := []connectionStringTest{

		{nil, nil, "sqlserver://."},
		{
			&ConnectSettings{TrustServerCertificate: true},
			func(vars *Variables) {
				_ = Setvar(SQLCMDDBNAME, "somedatabase")
			},
			"sqlserver://.?database=somedatabase&trustservercertificate=true",
		},
		{
			&ConnectSettings{TrustServerCertificate: true},
			func(vars *Variables) {
				vars.Set(SQLCMDSERVER, `someserver/instance`)
				vars.Set(SQLCMDDBNAME, "somedatabase")
				vars.Set(SQLCMDUSER, "someuser")
				vars.Set(SQLCMDPASSWORD, pwd)
			},
			fmt.Sprintf("sqlserver://someuser:%s@someserver/instance?database=somedatabase&trustservercertificate=true", pwd),
		},
		{
			&ConnectSettings{TrustServerCertificate: true, UseTrustedConnection: true},
			func(vars *Variables) {
				vars.Set(SQLCMDSERVER, `tcp:someserver,1045`)
				vars.Set(SQLCMDUSER, "someuser")
				vars.Set(SQLCMDPASSWORD, pwd)
			},
			"sqlserver://someserver:1045?trustservercertificate=true",
		},
		{
			nil,
			func(vars *Variables) {
				vars.Set(SQLCMDSERVER, `tcp:someserver,1045`)
			},
			"sqlserver://someserver:1045",
		},
	}

	for _, test := range commands {
		v := InitializeVariables(false)
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
	v := InitializeVariables(true)
	s := &Sqlcmd{vars: v}
	err := s.ConnectDb("", "", "", false)
	if assert.NoError(t, err, "ConnectDb should succeed") {
		sqlcmduser := os.Getenv(SQLCMDUSER)
		if sqlcmduser == "" {
			u, _ := user.Current()
			sqlcmduser = u.Username
		}
		assert.Equal(t, sqlcmduser, s.vars.SQLCmdUser(), "SQLCMDUSER variable should match connected user")
	}
}

func ConnectDb() (*sql.DB, error) {
	v := InitializeVariables(true)
	s := &Sqlcmd{vars: v}
	err := s.ConnectDb("", "", "", false)
	return s.db, err
}

func TestSqlCmdQueryAndExit(t *testing.T) {
	s, file := setupSqlcmdWithFileOutput(t)
	defer os.Remove(file.Name())
	s.Query = "select 100"
	err := s.Run(true, false)
	if assert.NoError(t, err, "s.Run(once = true)") {
		s.SetOutput(nil)
		bytes, err := os.ReadFile(file.Name())
		if assert.NoError(t, err, "os.ReadFile") {
			assert.Equal(t, "100"+SqlcmdEol+SqlcmdEol, string(bytes), "Incorrect output from Run")
		}
	}
}

// Simulate :r command
func TestIncludeFileNoExecutions(t *testing.T) {
	s, file := setupSqlcmdWithFileOutput(t)
	defer os.Remove(file.Name())
	dataPath := "testdata" + string(os.PathSeparator)
	err := s.IncludeFile(dataPath+"singlebatchnogo.sql", false)
	s.SetOutput(nil)
	if assert.NoError(t, err, "IncludeFile singlebatchnogo.sql false") {
		assert.Equal(t, "-", s.batch.State(), "s.batch.State() after IncludeFile singlebatchnogo.sql false")
		assert.Equal(t, "select 100 as num\nselect 'string' as title", s.batch.String(), "s.batch.String() after IncludeFile singlebatchnogo.sql false")
		bytes, err := os.ReadFile(file.Name())
		if assert.NoError(t, err, "os.ReadFile") {
			assert.Equal(t, "", string(bytes), "Incorrect output from Run")
		}
		file, err = os.CreateTemp("", "sqlcmdout")
		assert.NoError(t, err, "os.CreateTemp")
		s.SetOutput(file)
		// The second file has a go so it will execute all statements before it
		err = s.IncludeFile(dataPath+"twobatchnoendinggo.sql", false)
		if assert.NoError(t, err, "IncludeFile twobatchnoendinggo.sql false") {
			assert.Equal(t, "-", s.batch.State(), "s.batch.State() after IncludeFile twobatchnoendinggo.sql false")
			assert.Equal(t, "select 'string' as title", s.batch.String(), "s.batch.String() after IncludeFile twobatchnoendinggo.sql false")
			bytes, err := os.ReadFile(file.Name())
			if assert.NoError(t, err, "os.ReadFile") {
				assert.Equal(t, "100"+SqlcmdEol+SqlcmdEol+"string"+SqlcmdEol+SqlcmdEol+"100"+SqlcmdEol+SqlcmdEol, string(bytes), "Incorrect output from Run")
			}
		}
	}
}

// Simulate -i command line usage
func TestIncludeFileProcessAll(t *testing.T) {
	s, file := setupSqlcmdWithFileOutput(t)
	defer os.Remove(file.Name())
	dataPath := "testdata" + string(os.PathSeparator)
	err := s.IncludeFile(dataPath+"twobatchwithgo.sql", true)
	s.SetOutput(nil)
	if assert.NoError(t, err, "IncludeFile twobatchwithgo.sql true") {
		assert.Equal(t, "=", s.batch.State(), "s.batch.State() after IncludeFile twobatchwithgo.sql true")
		assert.Equal(t, "", s.batch.String(), "s.batch.String() after IncludeFile twobatchwithgo.sql true")
		bytes, err := os.ReadFile(file.Name())
		if assert.NoError(t, err, "os.ReadFile") {
			assert.Equal(t, "100"+SqlcmdEol+SqlcmdEol+"string"+SqlcmdEol+SqlcmdEol, string(bytes), "Incorrect output from Run")
		}
		file, err = os.CreateTemp("", "sqlcmdout")
		defer os.Remove(file.Name())
		assert.NoError(t, err, "os.CreateTemp")
		s.SetOutput(file)
		err = s.IncludeFile(dataPath+"twobatchnoendinggo.sql", true)
		if assert.NoError(t, err, "IncludeFile twobatchnoendinggo.sql true") {
			assert.Equal(t, "=", s.batch.State(), "s.batch.State() after IncludeFile twobatchnoendinggo.sql true")
			assert.Equal(t, "", s.batch.String(), "s.batch.String() after IncludeFile twobatchnoendinggo.sql true")
			bytes, err := os.ReadFile(file.Name())
			if assert.NoError(t, err, "os.ReadFile") {
				assert.Equal(t, "100"+SqlcmdEol+SqlcmdEol+"string"+SqlcmdEol+SqlcmdEol, string(bytes), "Incorrect output from Run")
			}
		}
	}
}
func setupSqlcmdWithFileOutput(t testing.TB) (*Sqlcmd, *os.File) {
	v := InitializeVariables(true)
	v.Set(SQLCMDMAXVARTYPEWIDTH, "0")
	s := New(nil, "", v)
	s.Format = NewSQLCmdDefaultFormatter(true)
	file, err := os.CreateTemp("", "sqlcmdout")
	assert.NoError(t, err, "os.CreateTemp")
	s.SetOutput(file)
	err = s.ConnectDb("", "", "", true)
	assert.NoError(t, err, "s.ConnectDB")
	return s, file
}
