// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/config"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"os"
	"runtime"
	"testing"
)

// TestQuery runs a sanity test of `sqlcmd query` using the local instance on 1433
func TestQuery(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("stuartpa: This is failing in the pipeline (Login failed for user 'sa'.)")
	}
	cmdparser.TestSetup(t)

	// if SQLCMDSERVER != "" add an endpoint using the --address
	if os.Getenv("SQLCMDSERVER") == "" {
		cmdparser.TestCmd[*config.AddEndpoint]()
	} else {
		t.Logf("SQLCMDSERVER: %v", os.Getenv("SQLCMDSERVER"))
		cmdparser.TestCmd[*config.AddEndpoint](fmt.Sprintf("--address %v", os.Getenv("SQLCMDSERVER")))
	}

	// If the SQLCMDPASSWORD envvar is set, then add a basic user, otherwise
	// we'll use trusted auth
	if os.Getenv("SQLCMDPASSWORD") != "" &&
		os.Getenv("SQLCMDUSER") != "" {

		// sqlcmd uses the SQLCMD_PASSWORD env var, but the tests use the
		// SQLCMDPASSWORD env var
		err := os.Setenv("SQLCMD_PASSWORD", os.Getenv("SQLCMDPASSWORD"))
		assert.Nil(t, err)
		cmdparser.TestCmd[*config.AddUser](
			fmt.Sprintf("--name user1 --username %s",
				os.Getenv("SQLCMDUSER")))
		cmdparser.TestCmd[*config.AddContext]("--endpoint endpoint --user user1")
	} else {
		cmdparser.TestCmd[*config.AddContext]("--endpoint endpoint")
	}
	cmdparser.TestCmd[*config.View]() // displaying the config (info in-case test fails)
	cmdparser.TestCmd[*Query]("PRINT")
}
