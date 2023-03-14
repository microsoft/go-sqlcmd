package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDbNameIfExists(t *testing.T) {

	type test struct {
		input          string
		expectedOutput string
	}

	tests := []test{
		// Positive Testcases
		{"https://example.com,myDbName", "myDbName"},
		{"https://example.com,myDb Name", "myDb Name"},
		{"https://example.com,myDb Na,me", "myDb Na,me"},
		{"https://example.com,[myDb Na,me]", "[myDb Na,me]]"},
		{"https://example.com,[myDb Na'me]", "[myDb Na''me]]"},
		{"https://example.com,[myDb ,Nam,e]", "[myDb ,Nam,e]]"},

		//Negative Testcases
		{"https://example.commyDbName", ""},
	}

	for _, testcase := range tests {
		dbname := parseDbName(testcase.input)
		dbname = getEscapedDbName(dbname)
		assert.Equal(t, testcase.expectedOutput, dbname, "Unexpected value from getDbNameForScripts")
	}
}
