// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

// Type Ads is used to implement the "open ads" which launches Azure
// Data Studio and establishes a connection to the SQL Server for the current
// context
type Ssms struct {
	cmdparser.Cmd
}

func (c *Ssms) PersistCredentialForAds(hostname string, endpoint sqlconfig.Endpoint, user *sqlconfig.User) {
}

func (c *Ssms) displayPreLaunchInfo() {

}
