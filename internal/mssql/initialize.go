// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package mssql

var decryptCallback func(cipherText string, decrypt bool) (secret string)

func init() {
	Initialize(
		func(err error) {
			if err != nil {
				panic(err)
			}
		},
		func(format string, a ...any) {},
		func(cipherText string, decrypt bool) (secret string) { return })
}

func Initialize(
	errorHandler func(err error),
	traceHandler func(format string, a ...any),
	decryptHandler func(cipherText string, decrypt bool) (secret string)) {
	errorCallback = errorHandler
	traceCallback = traceHandler
	decryptCallback = decryptHandler
}
