// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/test"
)

type VSCode struct {
	tool
}

func (t *VSCode) Init() {
	t.SetToolDescription(Description{
		Name:        "vscode",
		Purpose:     "Visual Studio Code is a code editor with support for database management through the MSSQL extension.",
		InstallText: t.installText()})

	for _, location := range t.searchLocations() {
		if file.Exists(location) {
			t.SetExePathAndName(location)
			break
		}
	}
}

func (t *VSCode) Run(args []string) (int, error) {
	if !test.IsRunningInTestExecutor() {
		return t.tool.Run(args)
	}
	return 0, nil
}

func (t *VSCode) RunWithOutput(args []string) (string, int, error) {
	if !test.IsRunningInTestExecutor() {
		return t.tool.RunWithOutput(args)
	}
	// In test mode, simulate extension list output
	for _, arg := range args {
		if arg == "--list-extensions" {
			return "ms-mssql.mssql\n", 0, nil
		}
	}
	return "", 0, nil
}
