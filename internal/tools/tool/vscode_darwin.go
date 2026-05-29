// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"os/exec"
	"path/filepath"
)

func (t *VSCode) searchLocations() []string {
	userProfile := os.Getenv("HOME")

	// The .app bundle is a directory; the launchable binary lives at
	// Contents/Resources/app/bin/<cli>. Prefer the PATH shim the user
	// installs via "Shell Command: Install 'code' command in PATH", then
	// fall back to the in-bundle binary at the standard install locations.
	var locations []string
	for _, build := range t.buildsToSearch() {
		cli := "code"
		app := "Visual Studio Code.app"
		if build == "insiders" {
			cli = "code-insiders"
			app = "Visual Studio Code - Insiders.app"
		}
		if p, err := exec.LookPath(cli); err == nil {
			locations = append(locations, p)
		}
		binPath := filepath.Join("Contents", "Resources", "app", "bin", cli)
		locations = append(locations,
			filepath.Join("/", "Applications", app, binPath),
			filepath.Join(userProfile, "Applications", app, binPath),
			filepath.Join(userProfile, "Downloads", app, binPath),
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
