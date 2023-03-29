package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
