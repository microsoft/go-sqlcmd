// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// Type Ads is used to implement the "open ads" which launches Azure
// Data Studio and establishes a connection to the SQL Server for the current
// context
type Ads struct {
	cmdparser.Cmd
}

func (c *Ads) persistCredentialForAds(hostname string, endpoint sqlconfig.Endpoint, user *sqlconfig.User) {
	// UNDONE: See - https://github.com/microsoft/go-sqlcmd/issues/257
}

// BUG(stuartpa): There is a bug in ADS that is naming credentials in Mac KeyChain
// using UTF16 encoding, when it should be UTF8.  This prevents us from creating
// an item in KeyChain that ADS can then re-use (because all the golang Keychain
// packages take a string for credential name, which is UTF8).  Rather than trying
// to clone the ADS bug here, we prompt the user without to get the credential which
// they'll have to enter into ADS (once, if they save password in the connection dialog)
func (c *Ads) displayPreLaunchInfo() {
	output := c.Output()

	output.Info(localizer.Sprintf("Temporary: To view connection information run:"))
	output.Info("")
	output.Info("\tsqlcmd config connection-strings")
	output.Info("")
	output.Info("(see issue for more information: https://github.com/microsoft/go-sqlcmd/issues/257)")
}
