package sql

import "github.com/microsoft/go-sqlcmd/pkg/sqlcmd"

type SqlType struct {
	sqlcmd  *sqlcmd.Sqlcmd
	console sqlcmd.Console
}

type SqlMock struct{}
