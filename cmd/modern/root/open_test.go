package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"testing"
)

// TestOpen runs a sanity test of `sqlcmd open`
func TestOpen(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*Open]()
}
