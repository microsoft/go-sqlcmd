// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"testing"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

func TestCopyPasswordToClipboardWithNoUser(t *testing.T) {
	cmdparser.TestSetup(t)

	if copyPasswordToClipboard(nil, nil) {
		t.Error("Expected false when user is nil")
	}
}

func TestCopyPasswordToClipboardWithNonBasicAuth(t *testing.T) {
	cmdparser.TestSetup(t)

	user := &sqlconfig.User{
		AuthenticationType: "windows",
		Name:               "test-user",
	}

	if copyPasswordToClipboard(user, nil) {
		t.Error("Expected false when auth type is not 'basic'")
	}
}
