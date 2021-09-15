package sqlcmd

import (
	"bytes"
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
