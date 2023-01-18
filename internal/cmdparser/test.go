package cmdparser

import (
	"github.com/microsoft/go-sqlcmd/internal"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"regexp"
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
		Output: output.New(output.Options{
			LoggingLevel: verbosity.Trace}),
	})
	c.DefineCommand()
	if len(args) > 1 {
		panic("Only provide one string of args, they will be split on space/quoted values (with spaces)")
	} else if len(args) == 1 {

		// This code uses a regular expression that matches either a quoted string
		// or a non-whitespace sequence of characters. The regexp.FindAllStringSubmatch
		// function then extracts all the matches from the input string and returns
		// them as a slice of string slices, where each inner slice contains the matched
		// string and any capture groups. In this case, the capture group is the
		// quoted string itself, which is what we want to extract.
		//
		// The code then iterates through the slice of string slices and appends the
		// quoted string or the non-quoted string to the fields slice, depending on
		// which type of match was found.
		re := regexp.MustCompile(`"([^"]+)"|([^\s]+)`)
		matches := re.FindAllStringSubmatch(args[0], -1)
		var fields []string
		for _, field := range matches {
			if field[1] != "" {
				fields = append(fields, field[1])
			} else {
				fields = append(fields, field[2])
			}
		}
		c.SetArgsForUnitTesting(fields)
	} else {
		c.SetArgsForUnitTesting([]string{})
	}
	err := c.Command().Execute()
	return err
}
