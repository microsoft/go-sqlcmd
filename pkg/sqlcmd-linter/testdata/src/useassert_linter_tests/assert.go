package sqlcmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDontUseFatal(t *testing.T) {
	t.Fatal("this should fail")        // want "Use assert package methods instead of Fatal"
	t.Fatalf("this should %s", "fail") // want "Use assert package methods instead of Fatalf"
	t.Fail()                           // want "Use assert package methods instead of Fail"
	assert.NoError(t, fmt.Errorf("what"))
	t.FailNow() // want "Use assert package methods instead of FailNow"

}

func TestDontUseRecover(t *testing.T) {
	defer func() { assert.NotNil(t, recover(), "The code did not panic as expected") }() // want "Use assert.Panics instead of recover()"
}
