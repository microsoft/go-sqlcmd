// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

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
	s, _ := setupSqlCmdWithMemoryOutput(t)
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
	assert.NotEqual(t, "someserver", s.Connect.ServerName, "On error, sqlCmd.Connect does not copy inputs")

	err = connectCommand(s, []string{}, 2)
	assert.EqualError(t, err, InvalidCommandError("CONNECT", 2).Error(), ":Connect with no arguments should return an error")
	c := newConnect(t)

	authenticationMethod := ""
	if c.Password == "" {
		c.UserName = os.Getenv("AZURE_CLIENT_ID") + "@" + os.Getenv("AZURE_TENANT_ID")
		c.Password = os.Getenv("AZURE_CLIENT_SECRET")
		authenticationMethod = "-G ActiveDirectoryServicePrincipal"
		if c.Password == "" {
			t.Log("Not trying :Connect with valid password due to no password being available")
			return
		}
		err = connectCommand(s, []string{fmt.Sprintf("%s -U %s -P %s %s", c.ServerName, c.UserName, c.Password, authenticationMethod)}, 3)
		assert.NoError(t, err, "connectCommand with valid parameters should not return an error")
		// not using assert to avoid printing passwords in the log
		if s.Connect.UserName != c.UserName || c.Password != s.Connect.Password {
			t.Fatal("After connect, sqlCmd.Connect is not updated")
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
