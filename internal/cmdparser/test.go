// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"regexp"
	"testing"

	"github.com/microsoft/go-sqlcmd/internal"
	"github.com/microsoft/go-sqlcmd/internal/buffer"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"github.com/microsoft/go-sqlcmd/internal/telemetry"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
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
func TestCmd[T PtrAsReceiverWrapper[pointerType], pointerType any](args ...string) string {
	result, err := testCmd[T](args...)

	if err != nil {

		// DEVNOTE: I don't think the code will ever get here (c.Command().Execute() will
		// always panic first. This is here to silence code checkers, that require the err return
		// variable be used.
		panic(err)
	}
	return result
}

func testCmd[T PtrAsReceiverWrapper[pointerType], pointerType any](args ...string) (result string, err error) {
	telemetry.SetTelemetryClientFromInstrumentationKey("")

	buf := buffer.NewMemoryBuffer()
	defer func() { buf.Close() }()
	c := New[T](dependency.Options{
		Output: output.New(output.Options{
			StandardWriter: buf,
			LoggingLevel:   verbosity.Trace}),
	})
	c.DefineCommand()
	if len(args) > 1 {
		panic("Only provide one string of args, they will be split on space/quoted values (with spaces)")
	} else if len(args) == 1 {
		c.SetArgsForUnitTesting(splitStringIntoArgsSlice(args[0]))
	} else {
		c.SetArgsForUnitTesting([]string{})
	}
	err = c.Command().Execute()
	return buf.String(), err
}

// splitStringIntoArgsSlice uses a regular expression that matches either a
// quoted string or a non-whitespace sequence of characters. All the matches
// from the input string are extracted and returned a slice of strings
func splitStringIntoArgsSlice(argsAsString string) (args []string) {
	re := regexp.MustCompile(`"([^"]+)"|([^\s]+)`)
	matches := re.FindAllStringSubmatch(argsAsString, -1)
	for _, field := range matches {
		if field[1] != "" {
			args = append(args, field[1]) // quoted string
		} else {
			args = append(args, field[2]) // non-whitespace sequence
		}
	}
	return args
}
