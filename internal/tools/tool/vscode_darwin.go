// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"path/filepath"
)

func (t *VSCode) searchLocations() []string {
	userProfile := os.Getenv("HOME")

	var locations []string
	for _, build := range t.buildsToSearch() {
		app := "Visual Studio Code.app"
		if build == "insiders" {
			app = "Visual Studio Code - Insiders.app"
		}
		locations = append(locations,
			filepath.Join("/", "Applications", app),
			filepath.Join(userProfile, "Downloads", app),
		)
	}
	return locations
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
