package open

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"runtime"
	"testing"
)

// TestOpen runs a sanity test of `sqlcmd open`
func TestAds(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Ads support only on Windows at this time")
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

	// TODO: Need to test this without launching the ADS UI itself
	// cmdparser.TestCmd[*Ads]()
}
