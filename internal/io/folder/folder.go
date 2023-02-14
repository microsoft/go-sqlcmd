// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package folder

import (
	"os"
)

// Getwd returns the current working directory
func Getwd() string {
	path, err := os.Getwd()
	checkErr(err)
	return path
}

// MkdirAll creates a directory with the given name if it does not already exist.
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

// RemoveAll removes a folder and all of its contents at the given path.
func RemoveAll(folder string) {
	err := os.RemoveAll(folder)
	checkErr(err)
}
