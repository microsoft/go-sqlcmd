package open

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

// Type Ads is used to implement the "open ads" which launches Azure
// Data Studio and establishes a connection to the SQL Server for the current
// context
type Ads struct {
	cmdparser.Cmd
}

func (c *Ads) persistCredentialForAds(hostname string, endpoint sqlconfig.Endpoint, user *sqlconfig.User) {
	panic("not implemented")
}
