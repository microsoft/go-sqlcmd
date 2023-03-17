package root

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/config"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestNegStop tests that the `sqlcmd stop` command fails when
// no context is defined
func TestNegStop(t *testing.T) {
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*Stop]()
	})
}

// TestNegStop2 tests that the `sqlcmd stop` command fails when
// no container is included in endpoint
func TestNegStop2(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*config.AddEndpoint]()
	cmdparser.TestCmd[*config.AddContext]("--endpoint endpoint")
	assert.Panics(t, func() {
		cmdparser.TestCmd[*Stop]()
	})
}
