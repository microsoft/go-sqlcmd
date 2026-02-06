// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"path/filepath"
)

func (t *SSMS) searchLocations() []string {
	programFiles := os.Getenv("ProgramFiles")
	programFilesX86 := os.Getenv("ProgramFiles(x86)")

	return []string{
		filepath.Join(programFiles, "Microsoft SQL Server Management Studio 20\\Common7\\IDE\\Ssms.exe"),
		filepath.Join(programFilesX86, "Microsoft SQL Server Management Studio 20\\Common7\\IDE\\Ssms.exe"),
		filepath.Join(programFiles, "Microsoft SQL Server Management Studio 19\\Common7\\IDE\\Ssms.exe"),
		filepath.Join(programFilesX86, "Microsoft SQL Server Management Studio 19\\Common7\\IDE\\Ssms.exe"),
		filepath.Join(programFiles, "Microsoft SQL Server Management Studio 18\\Common7\\IDE\\Ssms.exe"),
		filepath.Join(programFilesX86, "Microsoft SQL Server Management Studio 18\\Common7\\IDE\\Ssms.exe"),
	}
}

func (t *SSMS) installText() string {
	return `Install using a package manager:

    winget install Microsoft.SQLServerManagementStudio
    # or
    choco install sql-server-management-studio

Or download the latest version from:

    https://aka.ms/ssmsfullsetup

Note: SSMS is only available on Windows.`
}
