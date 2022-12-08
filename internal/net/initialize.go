// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package net

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
}
