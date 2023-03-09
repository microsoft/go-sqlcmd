// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

import (
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
)

type Sql interface {
	Connect(endpoint Endpoint, user *User, options ConnectOptions)
	Query(text string)
	ExecuteString(text string) string
}

type ConnectOptions struct {
	Interactive bool
}
