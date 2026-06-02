// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// searchLocations returns the .app bundle paths to probe. We resolve at the
// bundle level (not the in-bundle `code` CLI) so IsInstalled keeps working
// when the bundle exists but the optional CLI shim is missing, and so
// tests/users can point --build at the bundle they care about. launch()
// derives the in-bundle CLI from the bundle path at execution time.
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

// launch executes VS Code with the given args. macOS's `open -a <App> --args`
// only forwards args on a cold launch; when VS Code is already running the
// args (including the vscode:// URL we use to drive the mssql extension) are
// silently dropped. Bypass `open` entirely and exec the in-bundle `code` CLI,
// which handles --open-url whether or not VS Code is already running.
func (t *VSCode) launch(args []string) (int, error) {
	cliPath := filepath.Join(t.exeName, "Contents", "Resources", "app", "bin", "code")
	if _, err := os.Stat(cliPath); err != nil {
		return 1, fmt.Errorf("VS Code CLI not found at %q: %w", cliPath, err)
	}
	cmd := exec.Command(cliPath, args...)
	return t.runCmd(cmd)
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
