// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

import (
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
)

type Sql interface {
	Connect(endpoint Endpoint, user *User, options ConnectOptions)
	Query(text string)
	ExecuteSqlFile(filename string)
	ScalarString(query string) string
}

type ConnectOptions struct {
	Database    string
	LogLevel    int
	Interactive bool
}
