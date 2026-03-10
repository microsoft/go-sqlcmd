// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build !windows

package tool

func (t *SSMS) searchLocations() []string {
	return []string{}
}

func (t *SSMS) installText() string {
	return `SQL Server Management Studio (SSMS) is only available on Windows.

Please use:
- Visual Studio Code with the MSSQL extension: sqlcmd open vscode
- Azure Data Studio: sqlcmd open ads`
}
