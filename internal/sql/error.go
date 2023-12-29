// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

var errorCallback func(err error)
var traceLogging bool

func checkErr(err error) {
	errorCallback(err)
}
