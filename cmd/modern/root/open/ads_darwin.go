package open

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

type Ads struct {
	cmdparser.Cmd
}

func (c *Ads) persistCredentialForAds(hostname string, endpoint sqlconfig.Endpoint, user *sqlconfig.User) {
	panic("not implemented")
}
