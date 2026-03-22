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

// TestAds runs a sanity test of `sqlcmd open ads`
func TestAds(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("ADS support only on Windows at this time")
	}

	tool := tools.NewTool("ads")
	if !tool.IsInstalled() {
		t.Skip("Azure Data Studio is not installed")
	}

	cmdparser.TestSetup(t)
	config.AddEndpoint(sqlconfig.Endpoint{
		AssetDetails:    nil,
		EndpointDetails: sqlconfig.EndpointDetails{},
		Name:            "endpoint",
	})
	config.AddContext(sqlconfig.Context{
		ContextDetails: sqlconfig.ContextDetails{
			Endpoint: "endpoint",
			User:     nil,
		},
		Name: "context",
	})
	config.SetCurrentContextName("context")

	cmdparser.TestCmd[*Ads]()
}
