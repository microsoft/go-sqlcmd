// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"path/filepath"
)

// Search in this order - newer versions first
//
//	SSMS 20
//	SSMS 19
//	SSMS 18
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
	return `Download the latest version from:

    https://aka.ms/ssmsfullsetup

Or visit:

    https://docs.microsoft.com/sql/ssms/download-sql-server-management-studio-ssms

Note: SSMS is only available on Windows.`
}
