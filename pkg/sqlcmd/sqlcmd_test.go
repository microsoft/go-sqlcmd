// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/user"
	"runtime"
	"strings"
	"testing"

	"github.com/microsoft/go-mssqldb/azuread"
	"github.com/microsoft/go-mssqldb/msdsn"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const oneRowAffected = "(1 row affected)"

func TestConnectionStringFromSqlCmd(t *testing.T) {
	type connectionStringTest struct {
		settings         *ConnectSettings
		connectionString string
	}

	pwd := uuid.New().String()

	commands := []connectionStringTest{

		{&ConnectSettings{}, "sqlserver://."},
		{
			&ConnectSettings{TrustServerCertificate: true, WorkstationName: "mystation", Database: "somedatabase"},
			"sqlserver://.?database=somedatabase&trustservercertificate=true&workstation+id=mystation",
		},
		{
			&ConnectSettings{WorkstationName: "mystation", Encrypt: "false", Database: "somedatabase"},
			"sqlserver://.?database=somedatabase&encrypt=false&workstation+id=mystation",
		},
		{
			&ConnectSettings{TrustServerCertificate: true, Password: pwd, ServerName: `someserver\instance`, Database: "somedatabase", UserName: "someuser"},
			fmt.Sprintf("sqlserver://someuser:%s@someserver/instance?database=somedatabase&trustservercertificate=true", pwd),
		},
		{
			&ConnectSettings{TrustServerCertificate: true, UseTrustedConnection: true, Password: pwd, ServerName: `tcp:someserver,1045`, UserName: "someuser"},
			"sqlserver://someserver:1045?protocol=tcp&trustservercertificate=true",
		},
		{
			&ConnectSettings{ServerName: `tcp:someserver,1045`},
			"sqlserver://someserver:1045?protocol=tcp",
		},
		{
			&ConnectSettings{ServerName: "someserver", AuthenticationMethod: azuread.ActiveDirectoryServicePrincipal, UserName: "myapp@mytenant", Password: pwd},
			fmt.Sprintf("sqlserver://myapp%%40mytenant:%s@someserver", pwd),
		},
		{
			&ConnectSettings{ServerName: `\\someserver\pipe\sql\query`},
			"sqlserver://someserver?pipe=sql%5Cquery&protocol=np",
		},
	}

	for i, test := range commands {

		connectionString, err := test.settings.ConnectionString()
		if assert.NoError(t, err, "Unexpected error from [%d] %+v", i, test.settings) {
			assert.Equal(t, test.connectionString, connectionString, "Wrong connection string from [%d]: %+v", i, test.settings)
		}
	}
}

/*
	The following tests require a working SQL instance and rely on SqlCmd environment variables

to manage the initial connection string. The default connection when no environment variables are
set will be to localhost using Windows auth.
*/
func TestSqlCmdConnectDb(t *testing.T) {
	v := InitializeVariables(true)
	s := &Sqlcmd{vars: v}
	s.Connect = newConnect(t)
	err := s.ConnectDb(nil, false)
	if assert.NoError(t, err, "ConnectDb should succeed") {
		sqlcmduser := os.Getenv(SQLCMDUSER)
		if sqlcmduser == "" {
			u, _ := user.Current()
			sqlcmduser = u.Username
		}
		assert.Equal(t, sqlcmduser, s.vars.SQLCmdUser(), "SQLCMDUSER variable should match connected user")
	}
}

func ConnectDb(t testing.TB) (*sql.Conn, error) {
	v := InitializeVariables(true)
	s := &Sqlcmd{vars: v}
	s.Connect = newConnect(t)
	err := s.ConnectDb(nil, false)
	return s.db, err
}

func TestSqlCmdQueryAndExit(t *testing.T) {
	s, file := setupSqlcmdWithFileOutput(t)
	defer os.Remove(file.Name())
	s.Query = "select $(X"
	err := s.Run(true, false)
	if assert.NoError(t, err, "s.Run(once = true)") {
		s.SetOutput(nil)
		bytes, err := os.ReadFile(file.Name())
		if assert.NoError(t, err, "os.ReadFile") {
			assert.Equal(t, "Sqlcmd: Error: Syntax error at line 1"+SqlcmdEol, string(bytes), "Incorrect output from Run")
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
		assert.Equal(t, "select 100 as num"+SqlcmdEol+"select 'string' as title", s.batch.String(), "s.batch.String() after IncludeFile singlebatchnogo.sql false")
		bytes, err := os.ReadFile(file.Name())
		if assert.NoError(t, err, "os.ReadFile") {
			assert.Equal(t, "", string(bytes), "Incorrect output from Run")
		}
		file, err = os.CreateTemp("", "sqlcmdout")
		assert.NoError(t, err, "os.CreateTemp")
		defer os.Remove(file.Name())
		s.SetOutput(file)
		// The second file has a go so it will execute all statements before it
		err = s.IncludeFile(dataPath+"twobatchnoendinggo.sql", false)
		if assert.NoError(t, err, "IncludeFile twobatchnoendinggo.sql false") {
			assert.Equal(t, "-", s.batch.State(), "s.batch.State() after IncludeFile twobatchnoendinggo.sql false")
			assert.Equal(t, "select 'string' as title", s.batch.String(), "s.batch.String() after IncludeFile twobatchnoendinggo.sql false")
			s.SetOutput(nil)
			bytes, err := os.ReadFile(file.Name())
			if assert.NoError(t, err, "os.ReadFile") {
				assert.Equal(t, "100"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol+"string"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol+"100"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol, string(bytes), "Incorrect output from Run")
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
			assert.Equal(t, "100"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol+"string"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol, string(bytes), "Incorrect output from Run")
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
				assert.Equal(t, "100"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol+"string"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol, string(bytes), "Incorrect output from Run")
			}
		}
	}
}

func TestIncludeFileWithVariables(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	dataPath := "testdata" + string(os.PathSeparator)
	err := s.IncludeFile(dataPath+"variablesnogo.sql", true)
	if assert.NoError(t, err, "IncludeFile variablesnogo.sql true") {
		assert.Equal(t, "=", s.batch.State(), "s.batch.State() after IncludeFile variablesnogo.sql true")
		assert.Equal(t, "", s.batch.String(), "s.batch.String() after IncludeFile variablesnogo.sql true")
		s.SetOutput(nil)
		o := buf.buf.String()
		assert.Equal(t, "100"+SqlcmdEol+SqlcmdEol, o)
	}
}

func TestIncludeFileMultilineString(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	dataPath := "testdata" + string(os.PathSeparator)
	err := s.IncludeFile(dataPath+"blanks.sql", true)
	if assert.NoError(t, err, "IncludeFile blanks.sql true") {
		assert.Equal(t, "=", s.batch.State(), "s.batch.State() after IncludeFile blanks.sql true")
		assert.Equal(t, "", s.batch.String(), "s.batch.String() after IncludeFile blanks.sql true")
		s.SetOutput(nil)
		o := buf.buf.String()
		assert.Equal(t, "line 1"+SqlcmdEol+SqlcmdEol+SqlcmdEol+SqlcmdEol+"line2"+SqlcmdEol+SqlcmdEol, o)
	}
}

func TestIncludeFileQuotedIdentifiers(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	dataPath := "testdata" + string(os.PathSeparator)
	err := s.IncludeFile(dataPath+"quotedidentifiers.sql", true)
	if assert.NoError(t, err, "IncludeFile quotedidentifiers.sql true") {
		assert.Equal(t, "=", s.batch.State(), "s.batch.State() after IncludeFile quotedidentifiers.sql true")
		assert.Equal(t, "", s.batch.String(), "s.batch.String() after IncludeFile quotedidentifiers.sql true")
		s.SetOutput(nil)
		o := buf.buf.String()
		assert.Equal(t, `ab 1 a"b`+SqlcmdEol+SqlcmdEol, o)
	}
}

func TestGetRunnableQuery(t *testing.T) {
	v := InitializeVariables(false)
	v.Set("var1", "v1")
	v.Set("var2", "variable2")

	type test struct {
		raw string
		q   string
	}
	tests := []test{
		{"$(var1)", "v1"},
		{"$ (var2)", "$ (var2)"},
		{"select '$(VAR1) $(VAR2)' as  c", "select 'v1 variable2' as  c"},
		{" $(VAR1) ' $(VAR2) ' as  $(VAR1)", " v1 ' variable2 ' as  v1"},
	}
	s := New(nil, "", v)
	for _, test := range tests {
		s.batch.Reset([]rune(test.raw))
		_, _, _ = s.batch.Next()
		s.Connect.DisableVariableSubstitution = false
		t.Log(test.raw)
		r := s.getRunnableQuery(test.raw)
		assert.Equalf(t, test.q, r, `runnableQuery for "%s"`, test.raw)
		s.Connect.DisableVariableSubstitution = true
		r = s.getRunnableQuery(test.raw)
		assert.Equalf(t, test.raw, r, `runnableQuery without variable subs for "%s"`, test.raw)
	}
}

func TestExitInitialQuery(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	_ = s.vars.Setvar("var1", "1200")
	s.Query = "EXIT(SELECT '$(var1)', 2100)"
	err := s.Run(true, false)
	if assert.NoError(t, err, "s.Run(once = true)") {
		s.SetOutput(nil)
		o := buf.buf.String()
		assert.Equal(t, "1200 2100"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol, o, "Output")
		assert.Equal(t, 1200, s.Exitcode, "ExitCode")
	}

}

func TestExitCodeSetOnError(t *testing.T) {
	s, _ := setupSqlCmdWithMemoryOutput(t)
	s.Connect.ErrorSeverityLevel = 12
	retcode, err := s.runQuery("RAISERROR (N'Testing!' , 11, 1)")
	assert.NoError(t, err, "!ExitOnError 11")
	assert.Equal(t, -101, retcode, "Raiserror below ErrorSeverityLevel")
	retcode, err = s.runQuery("RAISERROR (N'Testing!' , 14, 1)")
	assert.NoError(t, err, "!ExitOnError 14")
	assert.Equal(t, 14, retcode, "Raiserror above ErrorSeverityLevel")
	s.Connect.ExitOnError = true
	retcode, err = s.runQuery("RAISERROR (N'Testing!' , 11, 1)")
	assert.NoError(t, err, "ExitOnError and Raiserror below ErrorSeverityLevel")
	assert.Equal(t, -101, retcode, "Raiserror below ErrorSeverityLevel")
	retcode, err = s.runQuery("RAISERROR (N'Testing!' , 14, 1)")
	assert.ErrorIs(t, err, ErrExitRequested, "ExitOnError and Raiserror above ErrorSeverityLevel")
	assert.Equal(t, 14, retcode, "ExitOnError and Raiserror above ErrorSeverityLevel")
	s.Connect.ErrorSeverityLevel = 0
	retcode, err = s.runQuery("RAISERROR (N'Testing!' , 11, 1)")
	assert.ErrorIs(t, err, ErrExitRequested, "ExitOnError and ErrorSeverityLevel = 0, Raiserror above 10")
	assert.Equal(t, 1, retcode, "ExitOnError and ErrorSeverityLevel = 0, Raiserror above 10")
	retcode, err = s.runQuery("RAISERROR (N'Testing!' , 5, 1)")
	assert.NoError(t, err, "ExitOnError and ErrorSeverityLevel = 0, Raiserror below 10")
	assert.Equal(t, -101, retcode, "ExitOnError and ErrorSeverityLevel = 0, Raiserror below 10")
	retcode, err = s.runQuery("RAISERROR (15001, 10, 127)")
	assert.ErrorIs(t, err, ErrExitRequested, "RAISERROR with state 127")
	assert.Equal(t, 15001, retcode, "RAISERROR (15001, 10, 127)")
}

func TestSqlCmdExitOnError(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	s.Connect.ExitOnError = true
	err := runSqlCmd(t, s, []string{"select 1", "GO", ":setvar", "select 2", "GO"})
	o := buf.buf.String()
	assert.EqualError(t, err, "Sqlcmd: Error: Syntax error at line 3 near command ':SETVAR'.", "Run should return an error")
	assert.Equal(t, "1"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol+"Sqlcmd: Error: Syntax error at line 3 near command ':SETVAR'."+SqlcmdEol, o, "Only first select should run")
	assert.Equal(t, 1, s.Exitcode, "s.ExitCode for a syntax error")

	s, buf = setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	s.Connect.ExitOnError = true
	s.Connect.ErrorSeverityLevel = 15
	s.vars.Set(SQLCMDERRORLEVEL, "14")
	err = runSqlCmd(t, s, []string{"raiserror(N'13', 13, 1)", "GO", "raiserror(N'14', 14, 1)", "GO", "raiserror(N'15', 15, 1)", "GO", "SELECT 'nope'", "GO"})
	o = buf.buf.String()
	assert.NotContains(t, o, "Level 13", "Level 13 should be filtered from the output")
	assert.NotContains(t, o, "nope", "Last select should not be run")
	assert.Contains(t, o, "Level 14", "Level 14 should be in the output")
	assert.Contains(t, o, "Level 15", "Level 15 should be in the output")
	assert.Equal(t, 15, s.Exitcode, "s.ExitCode for a syntax error")
	assert.NoError(t, err, "Run should not return an error for a SQL error")
}

func TestSqlCmdSetErrorLevel(t *testing.T) {
	s, _ := setupSqlCmdWithMemoryOutput(t)
	s.Connect.ErrorSeverityLevel = 15
	err := runSqlCmd(t, s, []string{"select bad as bad", "GO", "select 1", "GO"})
	assert.NoError(t, err, "runSqlCmd should have no error")
	assert.Equal(t, 16, s.Exitcode, "Select error should be the exit code")
}

type testConsole struct {
	PromptText       string
	OnPasswordPrompt func(prompt string) ([]byte, error)
	OnReadLine       func() (string, error)
}

func (tc *testConsole) Readline() (string, error) {
	return tc.OnReadLine()
}

func (tc *testConsole) ReadPassword(prompt string) ([]byte, error) {
	return tc.OnPasswordPrompt(prompt)
}

func (tc *testConsole) SetPrompt(s string) {
	tc.PromptText = s
}

func (tc *testConsole) Close() {

}

func TestPromptForPasswordNegative(t *testing.T) {
	prompted := false
	console := &testConsole{
		OnPasswordPrompt: func(prompt string) ([]byte, error) {
			assert.Equal(t, "Password:", prompt, "Incorrect password prompt")
			prompted = true
			return []byte{}, nil
		},
		OnReadLine: func() (string, error) {
			assert.Fail(t, "ReadLine should not be called")
			return "", nil
		},
	}
	v := InitializeVariables(true)
	s := New(console, "", v)
	c := newConnect(t)
	s.Connect.UserName = "someuser"
	s.Connect.ServerName = c.ServerName
	err := s.ConnectDb(nil, false)
	assert.True(t, prompted, "Password prompt not shown for SQL auth")
	assert.Error(t, err, "ConnectDb")
	prompted = false
	if canTestAzureAuth() {
		s.Connect.AuthenticationMethod = azuread.ActiveDirectoryPassword
		err = s.ConnectDb(nil, false)
		assert.True(t, prompted, "Password prompt not shown for AD Password auth")
		assert.Error(t, err, "ConnectDb")
		prompted = false
	}
}

func TestPromptForPasswordPositive(t *testing.T) {
	prompted := false
	c := newConnect(t)
	if c.Password == "" {
		// See if azure variables are set for activedirectoryserviceprincipal
		c.UserName = os.Getenv("AZURE_CLIENT_ID") + "@" + os.Getenv("AZURE_TENANT_ID")
		c.Password = os.Getenv("AZURE_CLIENT_SECRET")
		c.AuthenticationMethod = azuread.ActiveDirectoryServicePrincipal
		if c.Password == "" {
			t.Skip("No password available")
		}
	}
	password := c.Password
	c.Password = ""
	console := &testConsole{
		OnPasswordPrompt: func(prompt string) ([]byte, error) {
			assert.Equal(t, "Password:", prompt, "Incorrect password prompt")
			prompted = true
			return []byte(password), nil
		},
		OnReadLine: func() (string, error) {
			assert.Fail(t, "ReadLine should not be called")
			return "", nil
		},
	}
	v := InitializeVariables(true)
	s := New(console, "", v)
	// attempt without password prompt
	err := s.ConnectDb(c, true)
	assert.False(t, prompted, "ConnectDb with nopw=true should not prompt for password")
	assert.Error(t, err, "ConnectDb with nopw==true and no password provided")
	err = s.ConnectDb(c, false)
	assert.True(t, prompted, "ConnectDb with !nopw should prompt for password")
	assert.NoError(t, err, "ConnectDb with !nopw and valid password returned from prompt")
	assert.Equal(t, password, s.Connect.Password, "Password not stored in the connection")
}

func TestVerticalLayoutNoColumns(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	s.vars.Set(SQLCMDFORMAT, "vert")
	_, err := s.runQuery("SELECT 100 as 'column1', 2000 as 'col2', 300")
	assert.NoError(t, err, "runQuery failed")
	assert.Equal(t,
		"100"+SqlcmdEol+"2000"+SqlcmdEol+"300"+SqlcmdEol+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol,
		buf.buf.String(), "Query without column headers")
}

func TestSelectGuidColumn(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	_, err := s.runQuery("select convert(uniqueidentifier, N'3ddba21e-ff0f-4d24-90b4-f355864d7865')")
	assert.NoError(t, err, "runQuery failed")
	assert.Equal(t, "3ddba21e-ff0f-4d24-90b4-f355864d7865"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol, buf.buf.String(), "select a uniqueidentifier should work")
}

func TestSelectNullGuidColumn(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	_, err := s.runQuery("select convert(uniqueidentifier,null)")
	assert.NoError(t, err, "runQuery failed")
	assert.Equal(t, "NULL"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol, buf.buf.String(), "select a null uniqueidentifier should work")
}

func TestVerticalLayoutWithColumns(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	s.vars.Set(SQLCMDFORMAT, "vert")
	s.vars.Set(SQLCMDMAXVARTYPEWIDTH, "256")
	_, err := s.runQuery("SELECT 100 as 'column1', 2000 as 'col2', 300")
	assert.NoError(t, err, "runQuery failed")
	assert.Equal(t,
		"column1 100"+SqlcmdEol+"col2    2000"+SqlcmdEol+"        300"+SqlcmdEol+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol,
		buf.buf.String(), "Query without column headers")

}

func TestSqlCmdDefersToPrintError(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	s.PrintError = func(msg string, severity uint8) bool {
		return severity > 10
	}
	err := runSqlCmd(t, s, []string{"PRINT 'this has severity 10'", "RAISERROR (N'Testing!' , 11, 1)", "GO"})
	if assert.NoError(t, err, "runSqlCmd failed") {
		assert.Equal(t, "this has severity 10"+SqlcmdEol, buf.buf.String(), "Errors should be filtered by s.PrintError")
	}
}

func TestSqlCmdMaintainsConnectionBetweenBatches(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	err := runSqlCmd(t, s, []string{"CREATE TABLE #tmp1 (col1 int)", "insert into #tmp1 values (1)", "GO", "select * from #tmp1", "drop table #tmp1", "GO"})
	if assert.NoError(t, err, "runSqlCmd failed") {
		assert.Equal(t, oneRowAffected+SqlcmdEol+"1"+SqlcmdEol+SqlcmdEol+oneRowAffected+SqlcmdEol, buf.buf.String(), "Sqlcmd uses the same connection for all queries")
	}
}

func TestDateTimeFormats(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	err := s.IncludeFile(`testdata/selectdates.sql`, true)
	if assert.NoError(t, err, "selectdates.sql") {
		assert.Equal(t,
			`2022-03-05 14:01:02.000 2021-01-02 11:06:02.2000 2021-05-05 00:00:00.000000 +00:00 2019-01-11 13:00:00 14:01:02.0000000 2011-02-03`+SqlcmdEol+SqlcmdEol,
			buf.buf.String(),
			"Unexpected date format output")

	}
}

func TestQueryServerPropertyReturnsColumnName(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	s.vars.Set(SQLCMDMAXVARTYPEWIDTH, "100")
	defer buf.Close()
	err := runSqlCmd(t, s, []string{"select SERVERPROPERTY('EngineEdition') AS DatabaseEngineEdition", "GO"})
	if assert.NoError(t, err, "select should succeed") {
		assert.Contains(t, buf.buf.String(), "DatabaseEngineEdition", "Column name missing from output")
	}
}

func TestSqlCmdOutputAndError(t *testing.T) {
	s, outfile, errfile := setupSqlcmdWithFileErrorOutput(t)
	defer os.Remove(outfile.Name())
	defer os.Remove(errfile.Name())
	s.Query = "select $(X"
	err := s.Run(true, false)
	if assert.NoError(t, err, "s.Run(once = true)") {
		bytes, err := os.ReadFile(errfile.Name())
		if assert.NoError(t, err, "os.ReadFile") {
			assert.Equal(t, "Sqlcmd: Error: Syntax error at line 1"+SqlcmdEol, string(bytes), "Expected syntax error not received for query execution")
		}
	}
	s.Query = "select '1'"
	err = s.Run(true, false)
	if assert.NoError(t, err, "s.Run(once = true)") {
		bytes, err := os.ReadFile(outfile.Name())
		if assert.NoError(t, err, "os.ReadFile") {
			assert.Equal(t, "1"+SqlcmdEol+SqlcmdEol+"(1 row affected)"+SqlcmdEol, string(bytes), "Unexpected output for query execution")
		}
	}

	s, outfile, errfile = setupSqlcmdWithFileErrorOutput(t)
	defer os.Remove(outfile.Name())
	defer os.Remove(errfile.Name())
	dataPath := "testdata" + string(os.PathSeparator)
	err = s.IncludeFile(dataPath+"testerrorredirection.sql", false)
	if assert.NoError(t, err, "IncludeFile testerrorredirection.sql false") {
		bytes, err := os.ReadFile(outfile.Name())
		if assert.NoError(t, err, "os.ReadFile outfile") {
			assert.Equal(t, "1"+SqlcmdEol+SqlcmdEol+"(1 row affected)"+SqlcmdEol, string(bytes), "Unexpected output for sql file execution in outfile")
		}
		bytes, err = os.ReadFile(errfile.Name())
		if assert.NoError(t, err, "os.ReadFile errfile") {
			assert.Equal(t, "Sqlcmd: Error: Syntax error at line 3"+SqlcmdEol, string(bytes), "Expected syntax error not found in errfile")
		}
	}
}

// runSqlCmd uses lines as input for sqlcmd instead of relying on file or console input
func runSqlCmd(t testing.TB, s *Sqlcmd, lines []string) error {
	t.Helper()
	i := 0
	s.batch.read = func() (string, error) {
		if i < len(lines) {
			index := i
			i++
			return lines[index], nil
		}
		return "", io.EOF
	}
	return s.Run(false, false)
}

func setupSqlCmdWithMemoryOutput(t testing.TB) (*Sqlcmd, *memoryBuffer) {
	t.Helper()
	v := InitializeVariables(true)
	v.Set(SQLCMDMAXVARTYPEWIDTH, "0")
	s := New(nil, "", v)
	s.Connect = newConnect(t)
	s.Format = NewSQLCmdDefaultFormatter(true)
	buf := &memoryBuffer{buf: new(bytes.Buffer)}
	s.SetOutput(buf)
	err := s.ConnectDb(nil, true)
	assert.NoError(t, err, "s.ConnectDB")
	return s, buf
}

func setupSqlcmdWithFileOutput(t testing.TB) (*Sqlcmd, *os.File) {
	t.Helper()
	v := InitializeVariables(true)
	v.Set(SQLCMDMAXVARTYPEWIDTH, "0")
	s := New(nil, "", v)
	s.Connect = newConnect(t)
	s.Format = NewSQLCmdDefaultFormatter(true)
	file, err := os.CreateTemp("", "sqlcmdout")
	assert.NoError(t, err, "os.CreateTemp")
	s.SetOutput(file)
	err = s.ConnectDb(nil, true)
	if err != nil {
		os.Remove(file.Name())
	}
	assert.NoError(t, err, "s.ConnectDB")
	return s, file
}

func setupSqlcmdWithFileErrorOutput(t testing.TB) (*Sqlcmd, *os.File, *os.File) {
	t.Helper()
	v := InitializeVariables(true)
	v.Set(SQLCMDMAXVARTYPEWIDTH, "0")
	s := New(nil, "", v)
	s.Connect = newConnect(t)
	s.Format = NewSQLCmdDefaultFormatter(true)
	outfile, err := os.CreateTemp("", "sqlcmdout")
	assert.NoError(t, err, "os.CreateTemp")
	errfile, err := os.CreateTemp("", "sqlcmderr")
	assert.NoError(t, err, "os.CreateTemp")
	s.SetOutput(outfile)
	s.SetError(errfile)
	err = s.ConnectDb(nil, true)
	if err != nil {
		os.Remove(outfile.Name())
		os.Remove(errfile.Name())
	}
	assert.NoError(t, err, "s.ConnectDB")
	return s, outfile, errfile
}

// Assuming public Azure, use AAD when SQLCMDUSER environment variable is not set
func canTestAzureAuth() bool {
	server := os.Getenv(SQLCMDSERVER)
	userName := os.Getenv(SQLCMDUSER)
	return strings.Contains(server, ".database.windows.net") && userName == ""
}

func newConnect(t testing.TB) *ConnectSettings {
	t.Helper()
	connect := ConnectSettings{
		UserName:   os.Getenv(SQLCMDUSER),
		Database:   os.Getenv(SQLCMDDBNAME),
		ServerName: os.Getenv(SQLCMDSERVER),
		Password:   os.Getenv(SQLCMDPASSWORD),
	}
	if canTestAzureAuth() {
		t.Log("Using ActiveDirectoryDefault")
		connect.AuthenticationMethod = azuread.ActiveDirectoryDefault
	}
	return &connect
}

func TestSqlcmdPrefersSharedMemoryProtocol(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip()
	}
	assert.EqualValuesf(t, "lpc", msdsn.ProtocolParsers[0].Protocol(), "lpc should be first protocol")
	assert.EqualValuesf(t, "np", msdsn.ProtocolParsers[1].Protocol(), "np should be second protocol")
	assert.EqualValuesf(t, "tcp", msdsn.ProtocolParsers[2].Protocol(), "tcp should be third protocol")
}
