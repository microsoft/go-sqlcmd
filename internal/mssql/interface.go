package mssql

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
)

type MssqlInterface interface {
	Connect(endpoint sqlconfig.Endpoint, user *sqlconfig.User, console sqlcmd.Console) *sqlcmd.Sqlcmd
	Query(s *sqlcmd.Sqlcmd, text string)
}
