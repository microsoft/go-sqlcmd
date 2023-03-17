// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddUser(t *testing.T) {
	assert.Panics(t, func() {
		AddUser(User{
			Name:               "",
			AuthenticationType: "basic",
			BasicAuth:          nil,
		})
	})
}

func TestUserExists2(t *testing.T) {
	context := Context{
		ContextDetails: ContextDetails{},
		Name:           "context",
	}
	assert.False(t, UserExists(context))
}

func TestUserExists3(t *testing.T) {
	user := "user"
	context := Context{
		ContextDetails: ContextDetails{User: &user},
		Name:           "context",
	}
	assert.True(t, UserExists(context))
}

func TestNegAddUser(t *testing.T) {
	assert.Panics(t, func() {
		AddUser(User{
			Name:               "",
			AuthenticationType: "basic",
			BasicAuth: &BasicAuthDetails{
				Username:           "",
				PasswordEncryption: "none",
				Password:           "",
			},
		})
	})
}

func TestNegAddUser2(t *testing.T) {
	assert.Panics(t, func() {
		GetUser("doesnotexist")
	})
}
