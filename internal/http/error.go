// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package http

var errorCallback func(err error)

func checkErr(err error) {
	errorCallback(err)
}
