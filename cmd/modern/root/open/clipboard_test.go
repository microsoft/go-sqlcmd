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
