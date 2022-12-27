package cmdparser

import (
	"github.com/microsoft/go-sqlcmd/internal"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"strings"
	"testing"
)

// Test.go contains functions useful for creating compact unit tests for the
// CLI application using this package, e.g. a unit test can be two lines of code:
//
// 	  cmdparser.TestSetup(t)
//	  cmdparser.TestCmd[*GetEndpoints]()
//
// This is a complete unit test that runs the sqlcmd config get-endpoints command
// line.

// Setup internal packages for testing
func TestSetup(t *testing.T) {
	o := output.New(output.Options{})
	internal.Initialize(
		internal.InitializeOptions{
			ErrorHandler: func(err error) {
				if err != nil {
					panic(err)
				}
			},
			TraceHandler: o.Tracef,
			HintHandler: func(strings []string) {
				o.Infof("HINTS: %v"+sqlcmd.SqlcmdEol, strings)
			},
			LineBreak: sqlcmd.SqlcmdEol,
		})
	config.SetFileNameForTest(t)
	t.Log("Initialized internal packages for testing")
}

// Run a command expecing it to pass, passing in any supplied args (args are split on " " (space))
func TestCmd[T PtrAsReceiverWrapper[pointerType], pointerType any](args ...string) {
	err := testCmd[T](args...)

	// DEVNOTE: I don't think the code will ever get here (c.Command().Execute() will
	// always panic first. This is here to silence code checkers, that require the err return
	// variable be checked.
	if err != nil {
		panic(err)
	}
}

func testCmd[T PtrAsReceiverWrapper[pointerType], pointerType any](args ...string) error {
	c := New[T](dependency.Options{
		Output: output.New(output.Options{LoggingLevel: verbosity.Trace}),
	})
	c.DefineCommand()
	if len(args) > 1 {
		panic("Only provide one string of args, they will be split on space")
	} else if len(args) == 1 {
		c.SetArgsForUnitTesting(strings.Split(args[0], " "))
	} else {
		c.SetArgsForUnitTesting([]string{})
	}
	err := c.Command().Execute()
	return err
}
