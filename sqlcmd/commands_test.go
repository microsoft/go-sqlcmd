package sqlcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuitCommand(t *testing.T) {
	s := &Sqlcmd{}
	err := Quit(s, nil)
	require.ErrorIs(t, err, ErrExitRequested)
	err = Quit(s, []string{"extra parameters"})
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
		{":QUIT\n", "QUIT", []string{""}},
		{" QUIT \n", "QUIT", []string{" "}},
		{"quit extra\n", "QUIT", []string{" extra"}},
	}

	for _, test := range commands {
		cmd, args := matchCommand(test.line)
		if assert.NotNil(t, cmd, "No command found for "+test.line) {
			assert.Equal(t, test.cmd, cmd.name, "Incorrect command for "+test.line)
			assert.Equal(t, test.args, args, "Incorrect arguments for "+test.line)
		}
	}
}
