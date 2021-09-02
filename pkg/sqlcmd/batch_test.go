package sqlcmd

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchNextReset(t *testing.T) {
	tests := []struct {
		s     string
		stmts []string
		cmds  []string
		state string
	}{
		{"", nil, nil, "="},
		{"select 1", []string{"select 1"}, nil, "-"},
		{"select 1\nquit", []string{"select 1"}, []string{"QUIT"}, "="},
		{"select 1\nquite", []string{"select 1\nquite"}, nil, "-"},
		{"select 1\nquit\nselect 2", []string{"select 1", "select 2"}, []string{"QUIT"}, "-"},
		{"select '1\n", []string{"select '1\n"}, nil, "'"},
		{"select 1 /* comment\nGO", []string{"select 1 /* comment\nGO"}, nil, "*"},
		{"select '1\n00' \n/* comm\nent*/\nGO 4", []string{"select '1\n00' \n/* comm\nent*/"}, []string{"GO"}, "="},
	}
	for _, test := range tests {
		b := NewBatch(sp(test.s, "\n"), newCommands())
		var stmts, cmds []string
	loop:
		for {
			cmd, _, err := b.Next()
			switch {
			case err == io.EOF:
				// if we get EOF before a command we will try to run
				// whatever is in the buffer
				if s := b.String(); s != "" {
					stmts = append(stmts, s)
				}
				break loop
			case err != nil:
				t.Fatalf("test %s did not expect error, got: %v", test.s, err)
			}
			// resetting the buffer for every command purely for test purposes
			if cmd != nil {
				stmts = append(stmts, b.String())
				cmds = append(cmds, cmd.name)
				b.Reset(nil)
			}
		}
		assert.Equal(t, test.stmts, stmts, "Statements for %s", test.s)
		assert.Equal(t, test.state, b.State(), "State for %s", test.s)
		assert.Equal(t, test.cmds, cmds, "Commands for %s", test.s)
		b.Reset(nil)
		assert.Zero(t, b.Length, "Length after Reset")
		assert.Zero(t, len(b.Buffer), "len(Buffer) after Reset")
		assert.Zero(t, b.quote, "quote after Reset")
		assert.False(t, b.comment, "comment after Reset")
		assert.Equal(t, "=", b.State(), "State() after Reset")
	}
}

func sp(a, sep string) func() (string, error) {
	s := strings.Split(a, sep)
	return func() (string, error) {
		if len(s) > 0 {
			z := s[0]
			s = s[1:]
			return z, nil
		}
		return "", io.EOF
	}
}
