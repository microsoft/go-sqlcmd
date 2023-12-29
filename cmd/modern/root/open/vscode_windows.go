// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/credman"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// Type Vscode is used to implement the "open vscode" which launches VS Code
type Vscode struct {
	cmdparser.Cmd

	credential credman.Credential
}

// On Windows, the process blocks until the user exits ADS, let user know they can
// Ctrl+C here.
func (c *Vscode) displayPreLaunchInfo() {
	output := c.Output()

	output.Info(localizer.Sprintf("Press Ctrl+C to exit this process..."))
}

// PersistCredentialForAds stores a SQL password in the Windows Credential Manager
// for the given hostname and endpoint.
func (c *Vscode) PersistCredentialForAds(
	hostname string,
	endpoint sqlconfig.Endpoint,
	user *sqlconfig.User,
) {
}
