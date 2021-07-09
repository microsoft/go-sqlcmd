package sqlcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuitCommand(t *testing.T) {
	s := &Sqlcmd{}
	err := Quit(s, nil, 1)
	require.ErrorIs(t, err, ErrExitRequested)
	err = Quit(s, []string{"extra parameters"}, 2)
	require.Error(t, err, "Quit should error out with extra parameters")
	assert.NotErrorIs(t, err, ErrExitRequested, "Error with extra arguments")
}

func TestCommandParsing(t *testing.T) {
	type commandTest struct {
		line string
		cmd  string
		args []string
	}

	commands := []commandTest{
		{"quite", "", nil},
		{"quit", "QUIT", []string{""}},
		{":QUIT\n", "QUIT", []string{""}},
		{" QUIT \n", "QUIT", []string{" "}},
		{"quit extra\n", "QUIT", []string{" extra"}},
	}

	for _, test := range commands {
		cmd, args := matchCommand(test.line)
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
