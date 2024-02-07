// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package dotsqlcmdconfig

var traceCallback func(format string, a ...any)

func trace(format string, a ...any) {
	traceCallback(format, a...)
}
