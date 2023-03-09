// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

import . "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"

// Connect is a mock implementation used to speed up unit testing of other units
func (m *mock) Connect(
	endpoint Endpoint,
	user *User,
	options ConnectOptions,
) {
}

// Query is a mock implementation used to speed up unit testing of other units
func (m *mock) Query(text string) {
}

func (m *mock) ExecuteString(text string) string {
	return ""
}
