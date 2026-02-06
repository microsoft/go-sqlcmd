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
func (t *VSCode) searchLocations() []string {
	userProfile := os.Getenv("USERPROFILE")
	programFiles := os.Getenv("ProgramFiles")

	return []string{
		filepath.Join(userProfile, "AppData\\Local\\Programs\\Microsoft VS Code Insiders\\Code - Insiders.exe"),
		filepath.Join(programFiles, "Microsoft VS Code Insiders\\Code - Insiders.exe"),
		filepath.Join(userProfile, "AppData\\Local\\Programs\\Microsoft VS Code\\Code.exe"),
		filepath.Join(programFiles, "Microsoft VS Code\\Code.exe"),
	}
}

func (t *VSCode) installText() string {
	return `Install using a package manager:

    winget install Microsoft.VisualStudioCode
    # or
    choco install vscode

Or download the latest version from:

    https://code.visualstudio.com/download

After installation, install the MSSQL extension:

    sqlcmd open vscode --install-extension

Or install it directly in VS Code via Extensions (Ctrl+Shift+X) and search for "SQL Server (mssql)"`
}
