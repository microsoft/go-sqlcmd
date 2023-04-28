package databaseurl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDatabaseUrl(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/testdb.bak,myDbName", "myDbName"},
		{"https://example.com/testdb.bak", "testdb"},
		{"https://example.com/test.foo", "test"},
		{"https://example.com/test.foo,test", "test"},
		{"https://example.com/test.7z,tsql_name", "tsql_name"},
		{"https://example.com/test.mdf,tsql_name?foo=bar", "tsql_name"},
		{"https://example.com/test.mdf,tsql_name#link?foo=bar", "tsql_name"},
		{"https://example.com/test.mdf?foo=bar", "test"},
		{"https://example.com/test.mdf#link?foo=bar", "test"},
		{"https://example.com/test,test", "test"},
		{"https://example.com,", ""},
		{"https://example.com", ""},
		{"test.7z,tsql_name", "tsql_name"},
		{"test.mdf,tsql_name", "tsql_name"},
		{"test.mdf", "test"},
		{"c:\\test.mdf", "test"},
		{"c:\\test.mdf,tsql_name", "tsql_name"},
		{"file://test.mdf,tsql_name", "tsql_name"},
		{"file://test.mdf", "test"},
		{"file://c:\\test.mdf", "test"},
		{"file://c:\\folder\\test.mdf", "test"},
		{"file://c:/test.mdf", "test"},
		{"file://c:/folder/test.mdf", "test"},
		{"file:\\test.mdf,tsql_name", "tsql_name"},
		{"file:\\test.mdf", "test"},
		{"file:\\c:\\test.mdf", "test"},
		{"file:\\c:\\folder\\test.mdf", "test"},
		{"file:\\c:/test.mdf", "test"},
		{"file:\\c:/folder/test.mdf", "test"},
		{"\\\\server\\share\\test.mdf", "test"},
		{"\\\\server\\share\\folder\\test.mdf", "test"},
		{"\\\\server\\share\\folder\\test.mdf,db_name", "db_name"},
	}
	for _, tt := range tests {
		t.Run("DatabaseURLTest-"+tt.url, func(t *testing.T) {
			url := NewDatabaseUrl(tt.url)
			assert.Equalf(t, tt.want, url.DatabaseName, "NewDatabaseUrl(%v)", url.DatabaseName)
		})
	}
}
