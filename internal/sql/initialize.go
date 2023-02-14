// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

var decryptCallback func(cipherText string, decrypt bool) (secret string)

func Initialize(
	errorHandler func(err error),
	traceHandler func(format string, a ...any),
	decryptHandler func(cipherText string, decrypt bool) (secret string)) {
	errorCallback = errorHandler
	traceCallback = traceHandler
	decryptCallback = decryptHandler
}
