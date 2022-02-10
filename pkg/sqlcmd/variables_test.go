// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicVariableOperations(t *testing.T) {
	variables := Variables{
		"var1": "val1",
	}
	variables.Set("var2", "val2")
	assert.Contains(t, variables, "VAR2", "Set should add a capitalized key")
	all := variables.All()
	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}
	assert.ElementsMatch(t, []string{"var1", "VAR2"}, keys, "All returns every key")
	assert.Equal(t, "val2", all["VAR2"], "VAR2 set value")

}

func TestSetvarFailsForReadOnlyVariables(t *testing.T) {
	variables := Variables{}
	variables.Set("SQLCMDDBNAME", "somedatabase")
	err := variables.Setvar("SQLCMDDBNAME", "newdatabase")
	assert.Error(t, err, "setting a readonly variable fails")
	assert.Equal(t, "somedatabase", variables.SQLCmdDatabase(), "readonly variable shouldn't be changed by Setvar")
}

func TestEnvironmentVariablesAsInput(t *testing.T) {
	os.Setenv("SQLCMDSERVER", "someserver")
	defer os.Unsetenv("SQLCMDSERVER")
	os.Setenv("x", "somevalue")
	defer os.Unsetenv("x")
	vars := InitializeVariables(true).All()
	assert.Equal(t, "someserver", vars["SQLCMDSERVER"], "InitializeVariables should read a valid environment variable from the known list")
	_, ok := vars["x"]
	assert.False(t, ok, "InitializeVariables should skip variables not in the known list")
}

func TestSqlServerSplitsName(t *testing.T) {
	vars := Variables{
		SQLCMDSERVER: `tcp:someserver/someinstance`,
	}
	serverName, instance, port, err := vars.SQLCmdServer()
	if assert.NoError(t, err, "tcp:server/someinstance") {
		assert.Equal(t, "someserver", serverName, "server name for instance")
		assert.Equal(t, uint64(0), port, "port for instance")
		assert.Equal(t, "someinstance", instance, "instance for instance")
	}
	vars = Variables{
		SQLCMDSERVER: `tcp:someserver,1111`,
	}
	serverName, instance, port, err = vars.SQLCmdServer()
	if assert.NoError(t, err, "tcp:server,1111") {
		assert.Equal(t, "someserver", serverName, "server name for port number")
		assert.Equal(t, uint64(1111), port, "port for port number")
		assert.Equal(t, "", instance, "instance for port number")
	}
}

func TestParseValue(t *testing.T) {
	type test struct {
		raw   string
		val   string
		valid bool
	}
	tests := []test{
		{`""`, "", true},
		{`"`, "", false},
		{`"""`, "", false},
		{`no quotes`, "", false},
		{`"is quoted"`, "is quoted", true},
		{`" " single quote "`, "", false},
		{`" "" escaped quotes "" "`, ` " escaped quotes " `, true},
	}

	for _, tst := range tests {
		v, err := ParseValue(tst.raw)
		if tst.valid {
			if assert.NoErrorf(t, err, "Unexpected error for value %s", tst.raw) {
				assert.Equalf(t, tst.val, v, "Incorrect parsed value for %s", tst.raw)
			}
		} else {
			assert.Errorf(t, err, "Expected error for %s", tst.raw)
		}
	}
}

func TestValidIdentifier(t *testing.T) {
	type test struct {
		raw   string
		valid bool
	}
	tests := []test{
		{"1A", false},
		{"A1", true},
		{"A+", false},
		{"A-_b", true},
	}
	for _, tst := range tests {
		err := ValidIdentifier(tst.raw)
		if tst.valid {
			assert.NoErrorf(t, err, "%s is valid", tst.raw)
		} else {
			assert.Errorf(t, err, "%s is invalid", tst.raw)
		}
	}
}
