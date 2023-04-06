package uri

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
		u := NewUri(testcase.inputURL)
		assert.Equal(t, testcase.expectedURL, u.ActualUrl(),
			"Extracted URL does not match expected URL")
	}
}

func TestParseDbName(t *testing.T) {
	type test struct {
		inputURL    string
		expectedURL string
	}

	tests := []test{
		{"https://example.com/testdb.bak,myDbName", "myDbName"},
		{"https://example.com/testdb.bak", "testdb"},
		{"https://example.com/test.foo", "test"},
		{"https://example.com/test.foo,test", "test"},
		{"https://example.com/test.7z,tql_name", "tsql_name"},
		{"https://example.com/test.mdf,tsql_name?foo=bar", "tsql_name"},
		{"https://example.com/test.mdf,tsql_name#link?foo=bar", "tsql_name"},
		{"https://example.com/test.mdf?foo=bar", "test"},
		{"https://example.com/test.mdf#link?foo=bar", "test"},
		{"https://example.com/test,test", "test"},
		{"https://example.com,", ""},
		{"https://example.com", ""},
		{"test.7z,tql_name", "tsql_name"},
		{"test.mdf,tql_name", "tsql_name"},
		{"test.mdf", "test"},
		{"c:\test.mdf", "test"},
		{"c:\test.mdf,tsql_name", "tsql_name"},
		{"file://test.mdf,tsql_name", "tsql_name"},
		{"file://test.mdf", "test"},
	}

	for _, testcase := range tests {
		u := NewUri(testcase.inputURL)
		assert.Equal(t, testcase.expectedURL, u.ParseDbName(),
			"Extracted DB Name does not match expected DB Name")
	}
}

func TestGetDbNameIfExists(t *testing.T) {

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
		u := NewUri(testcase.input)

		assert.Equal(t, testcase.expectedIdentifierOp, u.GetDbNameAsIdentifier(),
			"Unexpected database name as identifier")
		assert.Equal(t, testcase.expectedNonIdentifierOp, u.GetDbNameAsNonIdentifier(),
			"Unexpected database name as non-identifier")
	}
}
