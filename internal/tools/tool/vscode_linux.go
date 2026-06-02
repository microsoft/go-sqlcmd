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

	var locations []string
	for _, build := range t.buildsToSearch() {
		cli := "code"
		if build == "insiders" {
			cli = "code-insiders"
		}
		if p, err := exec.LookPath(cli); err == nil {
			locations = append(locations, p)
		}
		locations = append(locations,
			filepath.Join("/", "usr", "bin", cli),
		)
		// Skip the per-user path when $HOME is empty -- filepath.Join would
		// produce a relative ".local/bin/<cli>" that could match an unintended
		// binary in the working directory.
		if userProfile != "" {
			locations = append(locations, filepath.Join(userProfile, ".local", "bin", cli))
		}
		locations = append(locations,
			filepath.Join("/", "snap", "bin", cli),
		)
	}
	return locations
}

// launch executes VS Code with the given args. On Linux exeName is already the
// `code` CLI, which handles --open-url correctly whether or not VS Code is
// already running, so we go through the standard tool.Run path.
func (t *VSCode) launch(args []string) (int, error) {
	return t.tool.Run(args)
}

func (t *VSCode) installText() string {
	return `Install using a package manager:

    # Debian/Ubuntu
    sudo apt install code

    # Fedora/RHEL
    sudo dnf install code

    # Snap
    sudo snap install code --classic

Or download the latest version from:

    https://code.visualstudio.com/download

After installation, the MSSQL extension is required. Running "sqlcmd open vscode"
opens a vscode:// URL that prompts VS Code to install the extension on first use.
You can also install it directly via Extensions (Ctrl+Shift+X) and search for
"SQL Server (mssql)".`
}
