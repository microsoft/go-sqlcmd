package databaseurl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractUrl(t *testing.T) {
	type test struct {
		inputURL    string
		expectedURL string
	}

	tests := []test{
		{"https://example.com/testdb.bak,myDbName", "https://example.com/testdb.bak"},
		{"https://example.com/testdb.bak", "https://example.com/testdb.bak"},
		{"https://example.com,", "https://example.com,"},
	}

	for _, testcase := range tests {
		u := NewDatabaseUrl(testcase.inputURL)
		assert.Equal(t, testcase.expectedURL, u.String(),
			"Extracted URL does not match expected URL")
	}
}

func TestGetDbNameIfExists(t *testing.T) {
	t.Skip("stuartpa: Fix before code-review")
	
	type test struct {
		input                   string
		expectedIdentifierOp    string
		expectedNonIdentifierOp string
	}

	tests := []test{
		// Positive Testcases
		// Database name specified
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,myDbName", "myDbName", "myDbName"},
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,myDb Name", "myDb Name", "myDb Name"},
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,myDb Na,me", "myDb Na,me", "myDb Na,me"},
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,[myDb Na,me]", "[myDb Na,me]]", "[myDb Na,me]]"},
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,[myDb Na'me]", "[myDb Na''me]]", "[myDb Na'me]]"},
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,[myDb ,Nam,e]", "[myDb ,Nam,e]]", "[myDb ,Nam,e]]"},

		// Delimiter between filename and databaseName is part of the filename
		// Decoded filename: my random .bak bac'kp]ack.bak
		{"https://example.com/my%20random%20.bak%20bac%27kp%5Dack.bak,[myDb ,Nam,e]", "[myDb ,Nam,e]]", "[myDb ,Nam,e]]"},

		// Database name not specified
		{"https://example.com/my%20random%20.bak%20bac%27kp%5Dack.bak", "my random .bak bac''kp]]ack", "my random .bak bac'kp]]ack"},
		{"https://example.com/my%20random%20.bak%20bac%27kp%5Dack.bak,", "my random .bak bac''kp]]ack", "my random .bak bac'kp]]ack"},

		//Negative Testcases
		{"https://example.com,myDbName", "", ""},
	}

	for _, testcase := range tests {
		u := NewDatabaseUrl(testcase.input)

		assert.Equal(t, testcase.expectedIdentifierOp, u.DatabaseNameAsTsqlIdentifier,
			"Unexpected database name as identifier")
		assert.Equal(t, testcase.expectedNonIdentifierOp, u.DatabaseNameAsNonTsqlIdentifier,
			"Unexpected database name as non-identifier")
	}
}
