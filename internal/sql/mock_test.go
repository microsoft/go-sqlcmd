// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

import "testing"
import . "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"

func TestMockConnect(t *testing.T) {
	mockObj := mock{}
	mockObj.Connect(Endpoint{}, nil, "", ConnectOptions{})
}

func TestMockQuery(t *testing.T) {
	mockObj := mock{}
	mockObj.Query("")
}
