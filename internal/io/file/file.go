// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package file

import (
	"github.com/microsoft/go-sqlcmd/internal/io/folder"
	"os"
	"path/filepath"
)

// CreateEmptyIfNotExists creates an empty file with the given filename if it
// does not already exist. If the parent directory of the file does not exist, the
// function will create it. The function is useful for ensuring that a file is
// present before writing to it.
func CreateEmptyIfNotExists(filename string) {
	if filename == "" {
		panic("filename must not be empty")
	}

	d, _ := filepath.Split(filename)
	if d != "" && !Exists(d) {
		trace("Folder %v does not exist, creating", d)
		folder.MkdirAll(d)
	}
	if !Exists(filename) {
		trace("File %v does not exist, creating empty 0 byte file", filename)
		handle, err := os.Create(filename)
		checkErr(err)
		defer func() {
			err := handle.Close()
			checkErr(err)
		}()
	}
}

// Exists checks if a file with the given filename exists in the file system. It
// returns a boolean value indicating whether the file exists or not.
func Exists(filename string) (exists bool) {
	if filename == "" {
		panic("filename must not be empty")
	}

	if _, err := os.Stat(filename); err == nil {
		exists = true
	}

	return
}

// Remove is used to remove a file with the specified filename. The function
// takes in the name of the file as an argument and deletes it from the file system.
func Remove(filename string) {
	err := os.Remove(filename)
	checkErr(err)
}
