// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package file

var errorCallback func(err error)

func checkErr(err error) {
	errorCallback(err)
}
