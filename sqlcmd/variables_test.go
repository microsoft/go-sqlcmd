package sqlcmd

import (
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
