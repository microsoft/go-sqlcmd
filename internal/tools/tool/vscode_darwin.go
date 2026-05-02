// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"path/filepath"
)

func (t *VSCode) searchLocations() []string {
	userProfile := os.Getenv("HOME")

	return []string{
		filepath.Join("/", "Applications", "Visual Studio Code - Insiders.app"),
		filepath.Join(userProfile, "Downloads", "Visual Studio Code - Insiders.app"),
		filepath.Join("/", "Applications", "Visual Studio Code.app"),
		filepath.Join(userProfile, "Downloads", "Visual Studio Code.app"),
	}
}

func (t *VSCode) installText() string {
	return `Install using Homebrew:

    brew install --cask visual-studio-code

Or download the latest version from:

    https://code.visualstudio.com/download

After installation, install the MSSQL extension:

    sqlcmd open vscode --install-extension

Or install it directly in VS Code via Extensions (Cmd+Shift+X) and search for "SQL Server (mssql)"`
}
