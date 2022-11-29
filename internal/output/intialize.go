// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package output

import (
	"fmt"
	. "github.com/microsoft/go-sqlcmd/internal/output/formatter"
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
	formatter = &Yaml{Base: Base{
		StandardOutput:       standardWriteCloser,
		ErrorHandlerCallback: errorHandler,
	}}

	Initialize(
		errorHandler,
		func(format string, a ...any) {},
		func(hints []string) {},
		os.Stdout,
		"yaml",
		verbosity.Info,
	)
}

func Initialize(
	errorHandler func(err error),
	traceHandler func(format string, a ...any),
	hintHandler func(hints []string),
	standardOutput io.WriteCloser,
	serializationFormat string,
	verbosity verbosity.Enum,
) {
	errorCallback = errorHandler
	traceCallback = traceHandler
	hintCallback = hintHandler
	standardWriteCloser = standardOutput
	loggingLevel = verbosity

	trace("Initializing output as '%v'", serializationFormat)

	switch serializationFormat {
	case "json":
		formatter = &Json{Base: Base{
			StandardOutput:       standardWriteCloser,
			ErrorHandlerCallback: errorHandler}}
	case "yaml":
		formatter = &Yaml{Base: Base{
			StandardOutput:       standardWriteCloser,
			ErrorHandlerCallback: errorHandler}}
	case "xml":
		formatter = &Xml{Base: Base{
			StandardOutput:       standardWriteCloser,
			ErrorHandlerCallback: errorHandler}}
	default:
		panic(fmt.Sprintf("Format '%v' not supported", serializationFormat))
	}
}
