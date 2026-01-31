// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

func (t *SSMS) searchLocations() []string {
	// SSMS is not available on Linux
	return []string{}
}

func (t *SSMS) installText() string {
	return `SQL Server Management Studio (SSMS) is only available on Windows.

For Linux, please use:
- Visual Studio Code with the MSSQL extension: sqlcmd open vscode
- Azure Data Studio: sqlcmd open ads`
}
