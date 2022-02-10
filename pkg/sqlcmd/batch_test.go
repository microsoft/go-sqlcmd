// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchNext(t *testing.T) {
	tests := []struct {
		s     string
		stmts []string
		cmds  []string
		state string
	}{
		{"", nil, nil, "="},
		{"select 1", []string{"select 1"}, nil, "-"},
		{"select $(x)\nquit", []string{"select $(x)"}, []string{"QUIT"}, "-"},
		{"select '$ (X' \nquite", []string{"select '$ (X' " + SqlcmdEol + "quite"}, nil, "-"},
		{":list\n:reset\n", nil, []string{"LIST", "RESET"}, "="},
		{"select 1\n:list\nselect 2", []string{"select 1" + SqlcmdEol + "select 2"}, []string{"LIST"}, "-"},
		{"select '1\n", []string{"select '1" + SqlcmdEol + ""}, nil, "'"},
		{"select 1 /* comment\nGO", []string{"select 1 /* comment" + SqlcmdEol + "GO"}, nil, "*"},
		{"select '1\n00' \n/* comm\nent*/\nGO 4", []string{"select '1" + SqlcmdEol + "00' " + SqlcmdEol + "/* comm" + SqlcmdEol + "ent*/"}, []string{"GO"}, "-"},
		{"$(x) $(y) 100\nquit", []string{"$(x) $(y) 100"}, []string{"QUIT"}, "-"},
		{"select 1\n:list", []string{"select 1"}, []string{"LIST"}, "-"},
		{"select 1\n:reset", []string{"select 1"}, []string{"RESET"}, "-"},
		{"select 1\n:exit()", []string{"select 1"}, []string{"EXIT"}, "-"},
		{"select 1\n:exit (select 10)", []string{"select 1"}, []string{"EXIT"}, "-"},
		{"select 1\n:exit", []string{"select 1"}, []string{"EXIT"}, "-"},
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
			if cmd != nil {
				cmds = append(cmds, cmd.name)
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

func TestBatchNextErrOnInvalidVariable(t *testing.T) {
	tests := []string{
		"select $(x",
		"$((x",
		"alter $( x)",
	}
	for _, test := range tests {
		b := NewBatch(sp(test, "\n"), newCommands())
		cmd, _, err := b.Next()
		assert.Nil(t, cmd, "cmd for "+test)
		assert.Equal(t, uint(1), b.linecount, "linecount should increment on a variable syntax error")
		assert.EqualErrorf(t, err, "Sqlcmd: Error: Syntax error at line 1.", "expected err for %s", test)
	}
}

func TestReadString(t *testing.T) {
	tests := []struct {
		// input string
		s string
		// index to start inside s
		i int
		// expected return string
		exp string
		// expected return bool
		ok bool
	}{
		{`'`, 0, ``, false},
		{` '`, 1, ``, false},
		{`'str' `, 0, `'str'`, true},
		{` 'str' `, 1, `'str'`, true},
		{`"str"`, 0, `"str"`, true},
		{`'str''str'`, 0, `'str''str'`, true},
		{` 'str''str' `, 1, `'str''str'`, true},
		{` "str''str" `, 1, `"str''str"`, true},
		// escaped \" aren't allowed in strings, so the second " would be next
		// double quoted string
		{`"str\""`, 0, `"str\"`, true},
		{` "str\"" `, 1, `"str\"`, true},
		{`''''`, 0, `''''`, true},
		{` '''' `, 1, `''''`, true},
		{`''''''`, 0, `''''''`, true},
		{` '''''' `, 1, `''''''`, true},
		{`'''`, 0, ``, false},
		{` ''' `, 1, ``, false},
		{`'''''`, 0, ``, false},
		{` ''''' `, 1, ``, false},
		{`"st'r"`, 0, `"st'r"`, true},
		{` "st'r" `, 1, `"st'r"`, true},
		{`"st''r"`, 0, `"st''r"`, true},
		{` "st''r" `, 1, `"st''r"`, true},
		{`'$(v)'`, 0, `'$(v)'`, true},
		{`'var $(var1) var2 $(var2)'`, 0, `'var $(var1) var2 $(var2)'`, true},
		{`'var $(var1) $`, 0, `'var $(var1) $`, false},
	}
	b := NewBatch(nil, newCommands())

	for _, test := range tests {
		r := []rune(test.s)
		c, end := rune(strings.TrimSpace(test.s)[0]), len(r)
		if c != '\'' && c != '"' {
			t.Fatalf("test %+v incorrect!", test)
		}
		pos, ok, err := b.readString(r, test.i+1, end, c, uint(0))
		assert.NoErrorf(t, err, "should be no error for %s", test)
		assert.Equal(t, test.ok, ok, "test %+v ok", test)
		if !ok {
			continue
		}
		assert.Equal(t, c, r[pos], "test %+v last character")
		v := string(r[test.i : pos+1])
		assert.Equal(t, test.exp, v, "test %+v returned string", test)
	}
}

func TestReadStringMalformVariable(t *testing.T) {
	tests := []string{
		"'select $(x'",
		"'  $((x'",
		"'alter $( x)",
	}
	b := NewBatch(nil, newCommands())
	for _, test := range tests {
		r := []rune(test)
		_, ok, err := b.readString(r, 1, len(test), '\'', 10)
		assert.Falsef(t, ok, "ok for %s", test)
		assert.EqualErrorf(t, err, "Sqlcmd: Error: Syntax error at line 10.", "expected err for %s", test)
	}
}

func TestReadStringVarmap(t *testing.T) {
	type mapTest struct {
		s string
		m map[int]string
	}
	tests := []mapTest{
		{`'var $(var1) var2 $(var2)'`, map[int]string{5: "var1", 18: "var2"}},
		{`'var $(va_1) var2 $(va-2)'`, map[int]string{5: "va_1", 18: "va-2"}},
	}
	for _, test := range tests {
		b := NewBatch(nil, newCommands())
		b.linevarmap = make(map[int]string)
		i, ok, err := b.readString([]rune(test.s), 1, len(test.s), '\'', 0)
		assert.Truef(t, ok, "ok returned by readString for %s", test.s)
		assert.NoErrorf(t, err, "readString for %s", test.s)
		assert.Equal(t, len(test.s)-1, i, "index returned by readString for %s", test.s)
		assert.Equalf(t, test.m, b.linevarmap, "linevarmap after readString %s", test.s)
	}
}

func TestBatchNextVarMap(t *testing.T) {
	type mapTest struct {
		s string
		m map[int]string
	}
	tests := []mapTest{
		{"'var $(var1)\nvar2 $(var2)\n'", map[int]string{5: "var1", 17 + len(SqlcmdEol): "var2"}},
		{"$(var1) select $(var2)\nselect 100\nselect '$(var3)'", map[int]string{
			0:                     "var1",
			15:                    "var2",
			40 + 2*len(SqlcmdEol): "var3"},
		},
	}
loop:
	for _, test := range tests {
		var err error
		b := NewBatch(sp(test.s, "\n"), newCommands())
		for {
			_, _, err = b.Next()
			if err == io.EOF {
				assert.Equalf(t, test.m, b.varmap, "varmap after Next %s. Batch:%s", test.s, escapeeol(b.String()))
				break loop
			} else {
				assert.NoErrorf(t, err, "Should have no error from Next")
			}
		}
	}
}

func escapeeol(s string) string {
	return strings.Replace(strings.Replace(s, "\n", `\n`, -1), "\r", `\r`, -1)
}
