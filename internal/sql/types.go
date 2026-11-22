// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

import "github.com/microsoft/go-sqlcmd/pkg/sqlcmd"

// mssql implements for SQL Server
type mssql struct {
	sqlcmd  *sqlcmd.Sqlcmd
	console sqlcmd.Console
}

// mock impoements for unit testing which uses a Hello World container (no
// SQL)
type mock struct{}
