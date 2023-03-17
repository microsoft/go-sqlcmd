// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

var decryptCallback func(cipherText string, encryptionMethod string) (secret string)

func Initialize(
	errorHandler func(err error),
	traceHandler func(format string, a ...any),
	decryptHandler func(cipherText string, encryptionMethod string) (secret string)) {
	errorCallback = errorHandler
	traceCallback = traceHandler
	decryptCallback = decryptHandler
}
