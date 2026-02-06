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
		filepath.Join("/", "usr", "bin", "code-insiders"),
		filepath.Join("/", "usr", "bin", "code"),
		filepath.Join(userProfile, ".local", "bin", "code-insiders"),
		filepath.Join(userProfile, ".local", "bin", "code"),
		filepath.Join("/", "snap", "bin", "code"),
	}
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

After installation, install the MSSQL extension:

    sqlcmd open vscode --install-extension

Or install it directly in VS Code via Extensions (Ctrl+Shift+X) and search for "SQL Server (mssql)"`
}
