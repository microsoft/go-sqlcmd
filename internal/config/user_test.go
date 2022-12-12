// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/test"
	"testing"
)

func TestAddUser(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	AddUser(User{
		Name:               "",
		AuthenticationType: "basic",
		BasicAuth:          nil,
	})
}

func TestNegAddUser(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()
	AddUser(User{
		Name:               "",
		AuthenticationType: "basic",
		BasicAuth: &BasicAuthDetails{
			Username:          "",
			PasswordEncrypted: false,
			Password:          "",
		},
	})
}

func TestNegAddUser2(t *testing.T) {
	defer func() { test.CatchExpectedError(recover(), t) }()
	GetUser("doesnotexist")
}
