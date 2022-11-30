// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package output

import (
	"github.com/microsoft/go-sqlcmd/internal/output/formatter"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"io"
	"os"
)

// init initializes the package for unit testing.  For production, use
// the Initialize method to inject fully functional dependencies
func init() {
	errorHandler := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	Initialize(
		errorHandler,
		func(hints []string) {},
		os.Stdout,
		"yaml",
		verbosity.Info,
	)
}

func Initialize(
	errorHandler func(err error),
	hintHandler func(hints []string),
	standardOutput io.WriteCloser,
	serializationFormat string,
	verbosity verbosity.Enum,
) Output {
	errorCallback = errorHandler
	hintCallback = hintHandler
	f := formatter.NewFormatter(serializationFormat, standardOutput, errorHandler)
	o := NewOutput(f, verbosity, standardOutput)
	o.Tracef("Initializing output as '%v\n'", serializationFormat)

	return o
}
