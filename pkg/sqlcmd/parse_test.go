package sqlcmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	for _, test := range tests {
		r := []rune(test.s)
		c, end := rune(strings.TrimSpace(test.s)[0]), len(r)
		if c != '\'' && c != '"' {
			t.Fatalf("test %+v incorrect!", test)
		}
		pos, ok, err := readString(r, test.i+1, end, c, uint(0))
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
	for _, test := range tests {
		r := []rune(test)
		_, ok, err := readString(r, 1, len(test), '\'', 10)
		assert.Falsef(t, ok, "ok for %s", test)
		assert.EqualErrorf(t, err, "Sqlcmd: Error: Syntax error at line 10.", "expected err for %s", test)
	}
}
