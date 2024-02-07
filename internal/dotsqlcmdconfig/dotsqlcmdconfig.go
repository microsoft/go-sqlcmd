// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package dotsqlcmdconfig

import (
	"fmt"
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlcmdconfig"
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"path/filepath"
	"testing"
)

var config Sqlcmdconfig
var filename string

// SetFileName sets the filename for the file that the application reads from and
// writes to. The file is created if it does not already exist, and Viper is configured
// to use the given filename.
func SetFileName(name string) {
	if name == "" {
		panic("name is empty")
	}

	filename = name

	file.CreateEmptyIfNotExists(filename)
}

func DatabaseNames() (dbs []string) {
	for _, db := range config.Databases {
		fmt.Println("db.Name: " + db.Name)
		dbs = append(dbs, db.Name)
	}

	return
}

func DatabaseFiles(ordinal int) (files []string) {
	if ordinal < 0 || ordinal >= len(config.Databases) {
		return
	}
	db := config.Databases[ordinal]

	for _, file := range db.DatabaseDetails.Use {
		files = append(files, file.Uri)
	}

	return
}

func AddonTypes() (addons []string) {
	for _, addon := range config.AddOns {
		addons = append(addons, addon.Type)
	}

	return
}

func AddonFiles(ordinal int) (files []string) {
	addon := config.AddOns[ordinal]

	for _, file := range addon.AddOnDetails.Use {
		files = append(files, file.Uri)
	}

	return
}

func SetFileNameForTest(t *testing.T) {
	SetFileName(filepath.Join(".sqlcmd", "sqlcmd.yaml"))
}

func DefaultFileName() (filename string) {
	filename = filepath.Join(".sqlcmd", "sqlcmd.yaml")

	return
}
