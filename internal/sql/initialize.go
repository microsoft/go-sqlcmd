// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

var decryptCallback func(cipherText string, encryptionMethod string) (secret string)

func Initialize(
	enableTraceLogging bool,
	errorHandler func(err error),
	traceHandler func(format string, a ...any),
	decryptHandler func(cipherText string, encryptionMethod string) (secret string)) {
	traceLogging = enableTraceLogging
	errorCallback = errorHandler
	traceCallback = traceHandler
	decryptCallback = decryptHandler
}
