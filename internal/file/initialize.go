// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package file

import (
	"github.com/microsoft/go-sqlcmd/internal/folder"
)

func init() {
	Initialize(
		func(err error) {
			if err != nil {
				panic(err)
			}
		},
		func(format string, a ...any) {})
}

func Initialize(
	errorHandler func(err error),
	traceHandler func(format string, a ...any)) {
	errorCallback = errorHandler
	traceCallback = traceHandler

	// this file helper depends on the folder helper (for example, to create folder paths
	// in passed in file names if the folders don't exist
	folder.Initialize(errorHandler, traceHandler)
}
