// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/test"
)

type SqlServerManagementStudio struct {
	tool
}

func (t *SqlServerManagementStudio) Init() {
	t.tool.SetToolDescription(Description{
		Name:        "ssms",
		Purpose:     "Sql Server Management Studio is a tool for managing SQL Server instances",
		InstallText: t.installText()})

	for _, location := range t.searchLocations() {
		if file.Exists(location) {
			t.tool.SetExePathAndName(location)
			break
		}
	}
}

func (t *SqlServerManagementStudio) Run(args []string, options RunOptions) (int, error) {
	if !test.IsRunningInTestExecutor() {
		return t.tool.Run(args, options)
	} else {
		return 0, nil
	}
}
