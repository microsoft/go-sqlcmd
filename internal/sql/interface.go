package sql

import (
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
)

type Sql interface {
	Connect(endpoint Endpoint, user *User, interactive bool)
	Query(text string)
}
