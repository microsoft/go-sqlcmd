// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/test"
)

type SSMS struct {
	tool
}

func (t *SSMS) Init() {
	t.SetToolDescription(Description{
		Name:        "ssms",
		Purpose:     "SQL Server Management Studio (SSMS) is an integrated environment for managing SQL Server infrastructure.",
		InstallText: t.installText()})

	for _, location := range t.searchLocations() {
		if file.Exists(location) {
			t.SetExePathAndName(location)
			break
		}
	}
}

func (t *SSMS) Run(args []string) (int, error) {
	if !test.IsRunningInTestExecutor() {
		return t.tool.Run(args)
	}
	return 0, nil
}
