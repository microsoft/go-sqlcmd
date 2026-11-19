// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/test"
)

type SSMS struct {
	tool

	// version is the requested SSMS major version (for example "21"). Empty
	// means "latest installed". It feeds the vswhere lookup in searchLocations.
	version string
}

func (t *SSMS) Init() {
	t.SetToolDescription(Description{
		Name:        "ssms",
		Purpose:     "SQL Server Management Studio (SSMS) is an integrated environment for managing SQL Server infrastructure.",
		InstallText: t.installText()})

	t.resolveExePath()
}

// SetVersion pins the SSMS major version to discover and re-resolves the exe
// path. Call after NewTool and before IsInstalled.
func (t *SSMS) SetVersion(version string) {
	t.version = version
	t.resolveExePath()
}

func (t *SSMS) resolveExePath() {
	t.exeName = ""
	t.installed = nil
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
