// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package output

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/output/formatter"
	"os"
	"strings"
)

// New initializes a new Output instance with the specified options. If options
// are not provided, default values are used. The function sets the error callback
// and the hint callback based on the value of the unitTesting field in the
// provided options. If unitTesting is true, the error callback is set to
// panic on error, otherwise it is set to use cobra.CheckErr to handle errors.
func New(options Options) *Output {
	if options.LoggingLevel == 0 {
		options.LoggingLevel = 2
	}
	if options.StandardWriter == nil {
		options.StandardWriter = os.Stdout
	}
	if options.ErrorHandler == nil {
		if isRunningInTestExecutor(options) {
			options.ErrorHandler = func(err error) {
				if err != nil {
					panic(err)
				}
			}
		} else {
			panic("Must provide Error Handler (the process (" + os.Args[0] + ") host is not a test executor)")
		}

	}
	if options.HintHandler == nil {
		if isRunningInTestExecutor(options) {
			options.HintHandler = func(hints []string) {
				fmt.Println(hints)
			}
		} else {
			panic("Must provide hint handler (the process " + os.Args[0] + " host is not a test executor)")
		}
	}

	f := formatter.New(formatter.Options{
		SerializationFormat: options.OutputType,
		StandardOutput:      options.StandardWriter,
		ErrorHandler:        options.ErrorHandler,
	})

	return &Output{
		formatter:           f,
		loggingLevel:        options.LoggingLevel,
		standardWriteCloser: options.StandardWriter,
		errorCallback:       options.ErrorHandler,
		hintCallback:        options.HintHandler,
	}
}

func isRunningInTestExecutor(options Options) bool {
	if (strings.HasSuffix(os.Args[0], ".test") || // are we in go test on *nix?
		strings.HasSuffix(os.Args[0], ".test.exe") || // are we in go test on windows?
		(len(os.Args) > 1 && os.Args[1] == "-test.v")) && // are we in goland unittest?
		!options.unitTesting {
		return true
	} else {
		return false
	}
}
