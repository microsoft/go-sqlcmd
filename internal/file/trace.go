// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package file

var traceCallback func(format string, a ...any)

func trace(format string, a ...any) {
	traceCallback(format, a...)
}
