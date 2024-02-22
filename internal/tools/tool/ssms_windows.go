// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"path/filepath"
)

// Search in this order
//
//	User Insiders Install
//	System Insiders Install
//	User non-Insiders install
//	System non-Insiders install
func (t *SqlServerManagementStudio) searchLocations() []string {
	programFiles := os.Getenv("ProgramFiles(x86)")

	// BUGBUG: Go looking in the registry for where SSMS is

	// C:\Program Files (x86)\Microsoft SQL Server Management Studio 19\Common7\IDE
	return []string{
		filepath.Join(programFiles, "Microsoft SQL Server Management Studio 19\\Common7\\IDE\\ssms.exe"),
	}
}

func (t *SqlServerManagementStudio) installText() string {
	return `Download the latest 'User Installer' .msi from:

    https://go.microsoft.com/fwlink/?linkid=2150927

More information can be found here:

    https://docs.microsoft.com/sql/azure-data-studio/download-azure-data-studio#get-azure-data-studio-for-windows`
}
