// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"runtime"
	"testing"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/tools"
)

// TestSsms runs a sanity test of `sqlcmd open ssms`
func TestSsms(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("SSMS is only available on Windows")
	}

	tool := tools.NewTool("ssms")
	if !tool.IsInstalled() {
		t.Skip("SSMS is not installed")
	}

	cmdparser.TestSetup(t)
	config.AddEndpoint(sqlconfig.Endpoint{
		AssetDetails: nil,
		EndpointDetails: sqlconfig.EndpointDetails{
			Address: "localhost",
			Port:    1433,
		},
		Name: "endpoint",
	})
	config.AddContext(sqlconfig.Context{
		ContextDetails: sqlconfig.ContextDetails{
			Endpoint: "endpoint",
			User:     nil,
		},
		Name: "context",
	})
	config.SetCurrentContextName("context")

	cmdparser.TestCmd[*SSMS]()
}
