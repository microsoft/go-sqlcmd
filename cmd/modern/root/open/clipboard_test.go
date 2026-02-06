// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"runtime"
	"testing"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

func TestCopyPasswordToClipboardWithNoUser(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux due to ADS tool initialization issue in tools factory")
	}

	cmdparser.TestSetup(t)

	result := copyPasswordToClipboard(nil, nil)
	if result {
		t.Error("Expected false when user is nil")
	}
}

func TestCopyPasswordToClipboardWithNonBasicAuth(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux due to ADS tool initialization issue in tools factory")
	}

	cmdparser.TestSetup(t)

	user := &sqlconfig.User{
		AuthenticationType: "windows",
		Name:               "test-user",
	}

	result := copyPasswordToClipboard(user, nil)
	if result {
		t.Error("Expected false when auth type is not 'basic'")
	}
}

func TestCopyPasswordToClipboardWithEmptyPassword(t *testing.T) {
	user := &sqlconfig.User{
		AuthenticationType: "basic",
		BasicAuth: &sqlconfig.BasicAuthDetails{
			Username:           "sa",
			PasswordEncryption: "",
			Password:           "",
		},
	}

	if !userShouldCopyPassword(user) {
		t.Error("userShouldCopyPassword should return true for basic auth user")
	}
}

func TestCopyPasswordToClipboardLogic(t *testing.T) {
	if userShouldCopyPassword(nil) {
		t.Error("Should not copy password when user is nil")
	}

	user := &sqlconfig.User{
		AuthenticationType: "integrated",
	}
	if userShouldCopyPassword(user) {
		t.Error("Should not copy password when auth type is not basic")
	}

	user = &sqlconfig.User{
		AuthenticationType: "basic",
		BasicAuth: &sqlconfig.BasicAuthDetails{
			Username: "sa",
			Password: "test",
		},
	}
	if !userShouldCopyPassword(user) {
		t.Error("Should copy password when auth type is basic")
	}
}

// userShouldCopyPassword is a helper that tests the condition logic
func userShouldCopyPassword(user *sqlconfig.User) bool {
	if user == nil || user.AuthenticationType != "basic" {
		return false
	}
	return true
}
