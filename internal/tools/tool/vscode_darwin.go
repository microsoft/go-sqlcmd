// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"path/filepath"
)

// searchLocations returns the .app bundle paths to probe. tool_darwin.go's
// generateCommandLine launches via "open -a <path> --args ...", which expects
// either a registered app name or a .app bundle path; pointing it at the
// in-bundle code binary would fail with "Unable to find application".
func (t *VSCode) searchLocations() []string {
	userProfile := os.Getenv("HOME")

	var locations []string
	for _, build := range t.buildsToSearch() {
		app := "Visual Studio Code.app"
		if build == "insiders" {
			app = "Visual Studio Code - Insiders.app"
		}
		locations = append(locations, filepath.Join("/", "Applications", app))
		// Skip the per-user paths when $HOME is empty -- filepath.Join would
		// produce a relative "Applications/..." that could match a directory
		// in the working directory.
		if userProfile != "" {
			locations = append(locations,
				filepath.Join(userProfile, "Applications", app),
				filepath.Join(userProfile, "Downloads", app),
			)
		}
	}
	return locations
}

func (t *VSCode) installText() string {
	return `Install using Homebrew:

    brew install --cask visual-studio-code

Or download the latest version from:

    https://code.visualstudio.com/download

After installation, the MSSQL extension is required. Running "sqlcmd open vscode"
opens a vscode:// URL that prompts VS Code to install the extension on first use.
You can also install it directly via Extensions (Cmd+Shift+X) and search for
"SQL Server (mssql)".`
}
