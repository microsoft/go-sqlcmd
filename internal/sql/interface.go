// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

import (
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
)

type Sql interface {
	Connect(endpoint Endpoint, user *User, options ConnectOptions)
	Query(text string)
	ScalarString(query string) string
}

type ConnectOptions struct {
	Database string

	Interactive bool

	// ExitOnError makes Query and ScalarString fail when the server reports
	// an error, instead of only printing it.
	ExitOnError bool
}
