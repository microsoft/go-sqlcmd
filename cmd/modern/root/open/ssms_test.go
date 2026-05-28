// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"net/url"
	"runtime"
	"strings"
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

	// Skip if the ssms:// URL handler isn't registered (SSMS not installed).
	tool := tools.NewTool("ssms")
	if !tool.IsInstalled() {
		t.Skip("SSMS is not installed (ssms:// URL handler not registered)")
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

	cmdparser.TestCmd[*Ssms]()
}

// TestBuildSsmsURLNoUser covers the integrated-auth case: only the server
// parameter is set.
func TestBuildSsmsURLNoUser(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("buildSsmsURL is Windows-only")
	}
	got := buildSsmsURL("myserver.database.windows.net", 1433, nil, "")

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("invalid URL %q: %v", got, err)
	}
	if parsed.Scheme != "ssms" || parsed.Host != "connect" {
		t.Fatalf("expected ssms://connect, got scheme=%q host=%q", parsed.Scheme, parsed.Host)
	}

	q := parsed.Query()
	if q.Get("s") != "myserver.database.windows.net,1433" {
		t.Errorf("s param: got %q", q.Get("s"))
	}
	for _, k := range []string{"u", "a", "p"} {
		if q.Has(k) {
			t.Errorf("did not expect %q for integrated auth, got %q", k, q.Get(k))
		}
	}
}

// TestBuildSsmsURLBasicAuth covers SQL login with username but no password:
// the URL must include u and a=SqlLogin, and the username must be URL-encoded.
func TestBuildSsmsURLBasicAuth(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("buildSsmsURL is Windows-only")
	}
	cmdparser.TestSetup(t)
	addBasicContext(t, "localhost", 1433, "admin user", "")

	user := userPointerFromContext(t)
	got := buildSsmsURL("localhost", 1433, user, "")

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("invalid URL %q: %v", got, err)
	}
	q := parsed.Query()
	if q.Get("s") != "localhost,1433" {
		t.Errorf("s param: got %q", q.Get("s"))
	}
	if q.Get("u") != "admin user" {
		t.Errorf("u param: got %q", q.Get("u"))
	}
	if q.Get("a") != "SqlLogin" {
		t.Errorf("a param: got %q, want SqlLogin", q.Get("a"))
	}
	if q.Has("p") {
		t.Errorf("p param should be omitted when password is empty, got %q", q.Get("p"))
	}
	// Spaces must be percent-encoded in the raw URL so SSMS receives the original username.
	if !strings.Contains(got, "u=admin+user") && !strings.Contains(got, "u=admin%20user") {
		t.Errorf("expected username to be URL-encoded in %q", got)
	}
}

// TestBuildSsmsURLBasicAuthWithPassword verifies the password parameter is
// included and URL-encoded when the context carries one.
func TestBuildSsmsURLBasicAuthWithPassword(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("buildSsmsURL is Windows-only")
	}
	cmdparser.TestSetup(t)
	addBasicContext(t, "localhost", 1433, "sa", "P@ss w0rd&=?")

	user := userPointerFromContext(t)
	got := buildSsmsURL("localhost", 1433, user, "P@ss w0rd&=?")

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("invalid URL %q: %v", got, err)
	}
	q := parsed.Query()
	if q.Get("p") != "P@ss w0rd&=?" {
		t.Errorf("p param: got %q, want %q", q.Get("p"), "P@ss w0rd&=?")
	}
}

// addBasicContext sets up a SQL-login context for a buildSsmsURL test.
func addBasicContext(t *testing.T, address string, port int, username, password string) {
	t.Helper()
	config.AddEndpoint(sqlconfig.Endpoint{
		EndpointDetails: sqlconfig.EndpointDetails{Address: address, Port: port},
		Name:            "ssms-url-endpoint",
	})
	config.AddUser(sqlconfig.User{
		AuthenticationType: "basic",
		BasicAuth: &sqlconfig.BasicAuthDetails{
			Username:           username,
			PasswordEncryption: "",
			Password:           password,
		},
		Name: "ssms-url-user",
	})
	userName := "ssms-url-user"
	config.AddContext(sqlconfig.Context{
		ContextDetails: sqlconfig.ContextDetails{Endpoint: "ssms-url-endpoint", User: &userName},
		Name:           "ssms-url-context",
	})
	config.SetCurrentContextName("ssms-url-context")
}

func userPointerFromContext(t *testing.T) *sqlconfig.User {
	t.Helper()
	_, user := config.CurrentContext()
	if user == nil {
		t.Fatal("expected user in current context")
	}
	return user
}
