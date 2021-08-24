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
	}
	for _, test := range tests {
		r := []rune(test.s)
		c, end := rune(strings.TrimSpace(test.s)[0]), len(r)
		if c != '\'' && c != '"' {
			t.Fatalf("test %+v incorrect!", test)
		}
		pos, ok := readString(r, test.i+1, end, c)
		assert.Equal(t, test.ok, ok, "test %+v ok", test)
		if !ok {
			continue
		}
		assert.Equal(t, c, r[pos], "test %+v last character")
		v := string(r[test.i : pos+1])
		assert.Equal(t, test.exp, v, "test %+v returned string", test)
	}
}
