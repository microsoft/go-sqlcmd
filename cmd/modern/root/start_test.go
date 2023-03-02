package root

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/config"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestNegStart tests that the `sqlcmd start` command fails when
// no context is defined
func TestNegStart(t *testing.T) {
	cmdparser.TestSetup(t)
	assert.Panics(t, func() {
		cmdparser.TestCmd[*Start]()
	})
}

// TestNegStart2 tests that the `sqlcmd start` command fails when
// no container is included in endpoint
func TestNegStart2(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*config.AddEndpoint]()
	cmdparser.TestCmd[*config.AddContext]("--endpoint endpoint")
	assert.Panics(t, func() {
		cmdparser.TestCmd[*Start]()
	})
}
