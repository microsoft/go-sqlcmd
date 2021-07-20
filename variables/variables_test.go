package variables

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicVariableOperations(t *testing.T) {
	variables = Variables{
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
	variables = Variables{}
	err := Setvar("SQLCMDDBNAME", "somedatabase")
	assert.NoError(t, err, "Setvar should succeed when SQLCMDDBNAME is not set")
	err = Setvar("SQLCMDDBNAME", "newdatabase")
	assert.EqualError(t, err, "Sqlcmd: Error: The scripting variable: 'SQLCMDDBNAME' is read-only")
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
	serverName, instance, port, err := vars.SqlCmdServer()
	if assert.NoError(t, err, "tcp:server/someinstance") {
		assert.Equal(t, "someserver", serverName, "server name for instance")
		assert.Equal(t, uint64(0), port, "port for instance")
		assert.Equal(t, "someinstance", instance, "instance for instance")
	}
	vars = Variables{
		SQLCMDSERVER: `tcp:someserver,1111`,
	}
	serverName, instance, port, err = vars.SqlCmdServer()
	if assert.NoError(t, err, "tcp:server,1111") {
		assert.Equal(t, "someserver", serverName, "server name for port number")
		assert.Equal(t, uint64(1111), port, "port for port number")
		assert.Equal(t, "", instance, "instance for port number")
	}
}
