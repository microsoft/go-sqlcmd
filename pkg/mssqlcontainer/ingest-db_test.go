package mssqlcontainer

import (
	"fmt"
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
		assert.Equal(t, testcase.expectedURL, extractUrl(testcase.inputURL), "Extracted URL does not match expected URL")
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
		dbname := parseDbName(testcase.input)
		dbnameAsIdentifier := getDbNameAsIdentifier(dbname)
		dbnameAsNonIdentifier := getDbNameAsNonIdentifier(dbname)
		assert.Equal(t, testcase.expectedIdentifierOp, dbnameAsIdentifier, "Unexpected database name as identifier")
		assert.Equal(t, testcase.expectedNonIdentifierOp, dbnameAsNonIdentifier, "Unexpected database name as non-identifier")
	}
}

func TestTzOutput(t *testing.T) {
	stdout := `Path = Readme_2010.txt
Size = 1157
Packed Size = 680
Modified = 2018-09-11 11:45:55.2543593
Attributes = A
CRC = B243D895
Encrypted = -
Method = LZMA2:27
Block = 0

Path = StackOverflow2010.mdf
Size = 8980398080
Packed Size = 1130813973
Modified = 2018-09-11 11:30:55.3142494
Attributes = A
CRC = 8D688B2A
Encrypted = -
Method = LZMA2:27
Block = 1

Path = StackOverflow2010_log.ldf
Size = 268312576
Packed Size = 37193161
Modified = 2018-09-11 11:30:55.3152489
Attributes = A
CRC = BCA9F91F
Encrypted = -
Method = LZMA2:27
Block = 2`

	paths := extractPaths(stdout)

	fmt.Println(paths)
}
