// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package file

import (
	"github.com/microsoft/go-sqlcmd/internal/folder"
	"os"
	"path/filepath"
)

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

func Exists(filename string) (exists bool) {
	if filename == "" {
		panic("filename must not be empty")
	}

	if _, err := os.Stat(filename); err == nil {
		exists = true
	}

	return
}

func Remove(filename string) {
	err := os.Remove(filename)
	checkErr(err)
}
