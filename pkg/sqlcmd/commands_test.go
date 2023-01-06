// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/microsoft/go-mssqldb/azuread"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuitCommand(t *testing.T) {
	s := &Sqlcmd{}
	err := quitCommand(s, nil, 1)
	require.ErrorIs(t, err, ErrExitRequested)
	err = quitCommand(s, []string{"extra parameters"}, 2)
	require.Error(t, err, "Quit should error out with extra parameters")
	assert.NotErrorIs(t, err, ErrExitRequested, "Error with extra arguments")
}

func TestCommandParsing(t *testing.T) {
	type commandTest struct {
		line string
		cmd  string
		args []string
	}
	c := newCommands()
	commands := []commandTest{
		{"quite", "", nil},
		{"quit", "QUIT", []string{""}},
		{":QUIT\n", "QUIT", []string{""}},
		{" QUIT \n", "QUIT", []string{""}},
		{"quit extra\n", "QUIT", []string{"extra"}},
		{`:Out c:\folder\file`, "OUT", []string{`c:\folder\file`}},
		{` :Error c:\folder\file`, "ERROR", []string{`c:\folder\file`}},
		{`:Setvar A1 "some value" `, "SETVAR", []string{`A1 "some value" `}},
		{` :Listvar`, "LISTVAR", []string{""}},
		{`:EXIT (select 100 as count)`, "EXIT", []string{"(select 100 as count)"}},
		{`:EXIT ( )`, "EXIT", []string{"( )"}},
		{`EXIT `, "EXIT", []string{""}},
		{`:Connect someserver -U someuser`, "CONNECT", []string{"someserver -U someuser"}},
		{`:r c:\$(var)\file.sql`, "READFILE", []string{`c:\$(var)\file.sql`}},
		{`:!! notepad`, "EXEC", []string{" notepad"}},
		{`:!!notepad`, "EXEC", []string{"notepad"}},
		{` !! dir c:\`, "EXEC", []string{` dir c:\`}},
		{`!!dir c:\`, "EXEC", []string{`dir c:\`}},
	}

	for _, test := range commands {
		cmd, args := c.matchCommand(test.line)
		if test.cmd != "" {
			if assert.NotNil(t, cmd, "No command found for `%s`", test.line) {
				assert.Equal(t, test.cmd, cmd.name, "Incorrect command for `%s`", test.line)
				assert.Equal(t, test.args, args, "Incorrect arguments for `%s`", test.line)
			}
		} else {
			assert.Nil(t, cmd, "Unexpected match for %s", test.line)
		}
	}
}

func TestCustomBatchSeparator(t *testing.T) {
	c := newCommands()
	err := c.SetBatchTerminator("me!")
	if assert.NoError(t, err, "SetBatchTerminator should succeed") {
		cmd, args := c.matchCommand("  me! 5 \n")
		if assert.NotNil(t, cmd, "matchCommand didn't find GO for custom batch separator") {
			assert.Equal(t, "GO", cmd.name, "command name")
			assert.Equal(t, "5", strings.TrimSpace(args[0]), "go argument")
		}
	}
}

func TestVarCommands(t *testing.T) {
	vars := InitializeVariables(false)
	s := New(nil, "", vars)
	buf := &memoryBuffer{buf: new(bytes.Buffer)}
	s.SetOutput(buf)
	err := setVarCommand(s, []string{"ABC 100"}, 1)
	assert.NoError(t, err, "setVarCommand ABC 100")
	err = setVarCommand(s, []string{"XYZ 200"}, 2)
	assert.NoError(t, err, "setVarCommand XYZ 200")
	err = listVarCommand(s, []string{""}, 3)
	assert.NoError(t, err, "listVarCommand")
	s.SetOutput(nil)
	varmap := s.vars.All()
	o := buf.buf.String()
	t.Logf("Listvar output:\n'%s'", o)
	output := strings.Split(o, SqlcmdEol)
	for i, v := range builtinVariables {
		line := strings.Split(output[i], " = ")
		assert.Equalf(t, v, line[0], "unexpected variable printed at index %d", i)
		val := strings.Trim(line[1], `"`)
		assert.Equalf(t, varmap[v], val, "Unexpected value for variable %s", v)
	}
	assert.Equalf(t, `ABC = "100"`, output[len(output)-3], "Penultimate non-empty line should be ABC")
	assert.Equalf(t, `XYZ = "200"`, output[len(output)-2], "Last non-empty line should be XYZ")
	assert.Equalf(t, "", output[len(output)-1], "Last line should be empty")

}

// memoryBuffer has both Write and Close methods for use as io.WriteCloser
type memoryBuffer struct {
	buf *bytes.Buffer
}

func (b *memoryBuffer) Write(p []byte) (n int, err error) {
	return b.buf.Write(p)
}

func (b *memoryBuffer) Close() error {
	return nil
}

func TestResetCommand(t *testing.T) {
	var err error

	// setup a test sqlcmd
	vars := InitializeVariables(false)
	s := New(nil, "", vars)
	buf := &memoryBuffer{buf: new(bytes.Buffer)}
	s.SetOutput(buf)

	// insert a test batch
	s.batch.Reset([]rune("select 1"))
	_, _, err = s.batch.Next()
	assert.NoError(t, err, "Inserting test batch")
	assert.Equal(t, s.batch.batchline, int(2), "Batch line updated after test batch insert")

	// execute reset command and validate results
	err = resetCommand(s, nil, 1)
	assert.Equal(t, s.batch.batchline, int(1), "Batch line not reset properly")
	assert.NoError(t, err, "Executing :reset command")
}

func TestListCommand(t *testing.T) {
	var err error

	// setup a test sqlcmd
	vars := InitializeVariables(false)
	s := New(nil, "", vars)
	buf := &memoryBuffer{buf: new(bytes.Buffer)}
	s.SetOutput(buf)

	// insert test batch
	s.batch.Reset([]rune("select 1"))
	_, _, err = s.batch.Next()
	assert.NoError(t, err, "Inserting test batch")

	// execute list command and verify results
	err = listCommand(s, nil, 1)
	assert.NoError(t, err, "Executing :list command")
	s.SetOutput(nil)
	o := buf.buf.String()
	assert.Equal(t, o, "select 1"+SqlcmdEol, ":list output not equal to batch")
}

func TestConnectCommand(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	prompted := false
	s.lineIo = &testConsole{
		OnPasswordPrompt: func(prompt string) ([]byte, error) {
			prompted = true
			return []byte{}, nil
		},
	}
	err := connectCommand(s, []string{"someserver -U someuser"}, 1)
	assert.NoError(t, err, "connectCommand with valid arguments doesn't return an error on connect failure")
	assert.True(t, prompted, "connectCommand with user name and no password should prompt for password")
	assert.NotEqual(t, "someserver", s.Connect.ServerName, "On connection failure, sqlCmd.Connect does not copy inputs")

	err = connectCommand(s, []string{}, 2)
	assert.EqualError(t, err, InvalidCommandError("CONNECT", 2).Error(), ":Connect with no arguments should return an error")
	c := newConnect(t)

	authenticationMethod := ""
	password := ""
	username := ""
	if canTestAzureAuth() {
		authenticationMethod = "-G " + azuread.ActiveDirectoryDefault
	}
	if c.Password != "" {
		password = "-P " + c.Password
	}
	if c.UserName != "" {
		username = "-U " + c.UserName
	}
	s.vars.Set("servername", c.ServerName)
	s.vars.Set("to", "111")
	buf.buf.Reset()
	err = connectCommand(s, []string{fmt.Sprintf("$(servername) %s %s %s -l $(to)", username, password, authenticationMethod)}, 3)
	if assert.NoError(t, err, "connectCommand with valid parameters should not return an error") {
		// not using assert to avoid printing passwords in the log
		assert.NotContains(t, buf.buf.String(), "$(servername)", "ConnectDB should have succeeded")
		if s.Connect.UserName != c.UserName || c.Password != s.Connect.Password || s.Connect.LoginTimeoutSeconds != 111 {
			t.Fatalf("After connect, sqlCmd.Connect is not updated %+v", s.Connect)
		}
	}
}

func TestErrorCommand(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	file, err := os.CreateTemp("", "sqlcmderr")
	assert.NoError(t, err, "os.CreateTemp")
	defer os.Remove(file.Name())
	fileName := file.Name()
	_ = file.Close()
	err = errorCommand(s, []string{""}, 1)
	assert.EqualError(t, err, InvalidCommandError("OUT", 1).Error(), "errorCommand with empty file name")
	err = errorCommand(s, []string{fileName}, 1)
	assert.NoError(t, err, "errorCommand")
	// Only some error kinds go to the error output
	err = runSqlCmd(t, s, []string{"print N'message'", "RAISERROR(N'Error', 16, 1)", "SELECT 1", ":SETVAR 1", "GO"})
	assert.NoError(t, err, "runSqlCmd")
	s.SetError(nil)
	errText, err := os.ReadFile(file.Name())
	if assert.NoError(t, err, "ReadFile") {
		assert.Regexp(t, "Msg 50000, Level 16, State 1, Server .*, Line 2"+SqlcmdEol+"Error"+SqlcmdEol, string(errText), "Error file contents")
	}
}

func TestOnErrorCommand(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	s.SetOutput(buf)
	err := onerrorCommand(s, []string{""}, 1)
	assert.EqualError(t, err, InvalidCommandError("ON ERROR", 1).Error(), "onerrorCommand with empty content")
	err = runSqlCmd(t, s, []string{":ONERROR ignore", "printtgit N'message'", "SELECT @@versionn", "GO"})
	assert.NoError(t, err, "runSqlCmd")
	o := buf.buf.String()
	assert.Equal(t, 0, s.Exitcode, "ExitCode")
	assert.Contains(t, o, "Must declare the scalar variable \"@@versionn\"", "output not equal to expected")
	err = runSqlCmd(t, s, []string{":ONERROR exit", "printtgit N'message'", "SELECT @@versionn", "GO"})
	assert.NoError(t, err, "runSqlCmd")
	assert.Equal(t, 1, s.Exitcode, "ExitCode")
	// -b sets ExitOnError true
	s.Connect.ExitOnError = true
	err = runSqlCmd(t, s, []string{":ONERROR ignore", "printtgit N'message'", "SELECT @@versionn", "GO"})
	// when ignore is set along with -b command , ignore takes precedence and resets ExitOnError
	assert.Equal(t, false, s.Connect.ExitOnError, "ExitOnError")
	assert.NoError(t, err, "runSqlCmd")
	// checking ExitonError with  Exit option
	err = runSqlCmd(t, s, []string{":ONERROR exit", "printtgit N'message'", "SELECT @@versionn", "GO"})
	assert.Equal(t, true, s.Connect.ExitOnError, "ExitOnError")
	assert.NoError(t, err, "runSqlCmd")
}
func TestResolveArgumentVariables(t *testing.T) {
	type argTest struct {
		arg string
		val string
		err string
	}

	args := []argTest{
		{"$(var1)", "var1val", ""},
		{"$(var1", "$(var1", ""},
		{`C:\folder\$(var1)\$(var2)\$(var1)\file.sql`, `C:\folder\var1val\$(var2)\var1val\file.sql`, "Sqlcmd: Error: 'var2' scripting variable not defined."},
		{`C:\folder\$(var1\$(var2)\$(var1)\file.sql`, `C:\folder\$(var1\$(var2)\var1val\file.sql`, "Sqlcmd: Error: 'var2' scripting variable not defined."},
	}
	vars := InitializeVariables(false)
	s := New(nil, "", vars)
	s.vars.Set("var1", "var1val")
	buf := &memoryBuffer{buf: new(bytes.Buffer)}
	defer buf.Close()
	s.SetError(buf)
	for _, test := range args {
		actual, _ := resolveArgumentVariables(s, []rune(test.arg), false)
		assert.Equal(t, test.val, actual, "Incorrect argument parsing of "+test.arg)
		assert.Contains(t, buf.buf.String(), test.err, "Error output mismatch for "+test.arg)
		buf.buf.Reset()
	}
	actual, err := resolveArgumentVariables(s, []rune("$(var1)$(var2)"), true)
	if assert.ErrorContains(t, err, UndefinedVariable("var2").Error(), "fail on unresolved variable") {
		assert.Empty(t, actual, "fail on unresolved variable")
	}
}

func TestExecCommand(t *testing.T) {
	vars := InitializeVariables(false)
	s := New(nil, "", vars)
	s.vars.Set("var1", "hello")
	buf := &memoryBuffer{buf: new(bytes.Buffer)}
	defer buf.Close()
	s.SetOutput(buf)
	err := execCommand(s, []string{`echo $(var1)`}, 1)
	if assert.NoError(t, err, "execCommand with valid arguments") {
		assert.Equal(t, buf.buf.String(), "hello"+SqlcmdEol, "echo output should be in sqlcmd output")
	}
}

func TestDisableSysCommandBlocksExec(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	s.Cmd.DisableSysCommands(false)
	c := []string{"set nocount on", ":!! echo hello", "select 100", "go"}
	err := runSqlCmd(t, s, c)
	if assert.NoError(t, err, ":!! with warning should not raise error") {
		assert.Contains(t, buf.buf.String(), ErrCommandsDisabled.Error()+SqlcmdEol+"100"+SqlcmdEol)
		assert.Equal(t, 0, s.Exitcode, "ExitCode after warning")
	}
	buf.buf.Reset()
	s.Cmd.DisableSysCommands(true)
	err = runSqlCmd(t, s, c)
	if assert.NoError(t, err, ":!! with error should not return error") {
		assert.Contains(t, buf.buf.String(), ErrCommandsDisabled.Error()+SqlcmdEol)
		assert.NotContains(t, buf.buf.String(), "100", "query should not run when syscommand disabled")
		assert.Equal(t, 1, s.Exitcode, "ExitCode after error")
	}
}

func TestEditCommand(t *testing.T) {
	s, buf := setupSqlCmdWithMemoryOutput(t)
	defer buf.Close()
	s.vars.Set(SQLCMDEDITOR, "echo select 5000> ")
	c := []string{"set nocount on", "go", "select 100", ":ed", "go"}
	err := runSqlCmd(t, s, c)
	if assert.NoError(t, err, ":ed should not raise error") {
		assert.Equal(t, "1> select 5000"+SqlcmdEol+"5000"+SqlcmdEol+SqlcmdEol, buf.buf.String(), "Incorrect output from query after :ed command")
	}
}
