// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"os"
	"testing"
)

// TestConnectionStrings tests that the `sqlcmd config connection-strings` command
// works as expected
func TestConnectionStrings(t *testing.T) {
	cmdparser.TestSetup(t)

	output := output.New(output.Options{HintHandler: func(hints []string) {}, ErrorHandler: func(err error) {}})
	options := internal.InitializeOptions{
		ErrorHandler: func(err error) {
			if err != nil {
				panic(err)
			}
		},
		HintHandler:  func(strings []string) {},
		TraceHandler: output.Tracef,
		LineBreak:    "\n",
	}
	internal.Initialize(options)

	if os.Getenv("SQLCMDPASSWORD") == "" &&
		os.Getenv("SQLCMD_PASSWORD") == "" {
		os.Setenv("SQLCMD_PASSWORD", "it's-a-secret")
	}

	cmdparser.TestCmd[*AddEndpoint]()
	cmdparser.TestCmd[*AddUser]("--username user")
	cmdparser.TestCmd[*AddContext]("--endpoint endpoint --user user")
	cmdparser.TestCmd[*ConnectionStrings]()

	// Add endpoint with no user
	cmdparser.TestCmd[*AddContext]("--endpoint endpoint")
	cmdparser.TestCmd[*ConnectionStrings]()

	// Add endpoint to Azure SQL (connection strings won't Trust server cert)
	cmdparser.TestCmd[*AddEndpoint]("--address server.database.windows.net")
	cmdparser.TestCmd[*AddUser]("--username user")
	cmdparser.TestCmd[*AddContext]("--endpoint endpoint2 --user user")

	cmdparser.TestCmd[*ConnectionStrings]()
}
