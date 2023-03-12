// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"os"
	"testing"
)

func TestDeleteUser(t *testing.T) {
	cmdparser.TestSetup(t)

	// SQLCMDPASSWORD is already set in the build pipelines, so don't
	// overwrite it here.
	if os.Getenv("SQLCMDPASSWORD") == "" {
		os.Setenv("SQLCMDPASSWORD", "it's-a-secret")
	}
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*AddUser]("--username user1 --password-encryption none")
	cmdparser.TestCmd[*DeleteUser]("--name user")
}
