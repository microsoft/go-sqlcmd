// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package folder

import (
	"os"
)

func MkdirAll(folder string) {
	if folder == "" {
		panic("folder must not be empty")
	}
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		trace("Folder %v does not exist, creating", folder)
		err := os.MkdirAll(folder, os.ModePerm)
		checkErr(err)
	}
}

func RemoveAll(folder string) {
	err := os.RemoveAll(folder)
	checkErr(err)
}
