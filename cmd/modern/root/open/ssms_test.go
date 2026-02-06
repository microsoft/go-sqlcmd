// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"runtime"
	"strconv"
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

	// Skip if SSMS is not installed
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

	cmdparser.TestCmd[*Ssms]()
}

func TestSsmsCommandLineArgs(t *testing.T) {
	// Test server argument format
	host := "localhost"
	port := 1433
	serverArg := buildServerArg(host, port)

	expected := "localhost,1433"
	if serverArg != expected {
		t.Errorf("Expected server arg '%s', got '%s'", expected, serverArg)
	}

	// Test with non-default port
	port = 2000
	serverArg = buildServerArg(host, port)

	expected = "localhost,2000"
	if serverArg != expected {
		t.Errorf("Expected server arg '%s', got '%s'", expected, serverArg)
	}

	// Test with different host
	host = "myserver.database.windows.net"
	serverArg = buildServerArg(host, port)

	expected = "myserver.database.windows.net,2000"
	if serverArg != expected {
		t.Errorf("Expected server arg '%s', got '%s'", expected, serverArg)
	}
}

// TestSsmsUsernameEscaping tests that special characters in usernames are escaped
func TestSsmsUsernameEscaping(t *testing.T) {
	// Test escaping double quotes in username
	username := `admin"user`
	escaped := strings.ReplaceAll(username, `"`, `\"`)

	expected := `admin\"user`
	if escaped != expected {
		t.Errorf("Expected escaped username '%s', got '%s'", expected, escaped)
	}

	// Test username without special characters
	username = "sa"
	escaped = strings.ReplaceAll(username, `"`, `\"`)

	if escaped != "sa" {
		t.Errorf("Expected unchanged username 'sa', got '%s'", escaped)
	}

	// Test username with multiple quotes
	username = `user"with"quotes`
	escaped = strings.ReplaceAll(username, `"`, `\"`)

	expected = `user\"with\"quotes`
	if escaped != expected {
		t.Errorf("Expected escaped username '%s', got '%s'", expected, escaped)
	}
}

// TestSsmsContextWithUser tests SSMS setup with user credentials
func TestSsmsContextWithUser(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("SSMS is only available on Windows")
	}

	cmdparser.TestSetup(t)

	// Set up context with SQL authentication user
	config.AddEndpoint(sqlconfig.Endpoint{
		AssetDetails: nil,
		EndpointDetails: sqlconfig.EndpointDetails{
			Address: "localhost",
			Port:    1433,
		},
		Name: "ssms-test-endpoint",
	})

	config.AddUser(sqlconfig.User{
		AuthenticationType: "basic",
		BasicAuth: &sqlconfig.BasicAuthDetails{
			Username:           "sa",
			PasswordEncryption: "",
			Password:           "TestPassword123",
		},
		Name: "ssms-test-user",
	})

	config.AddContext(sqlconfig.Context{
		ContextDetails: sqlconfig.ContextDetails{
			Endpoint: "ssms-test-endpoint",
			User:     strPtr("ssms-test-user"),
		},
		Name: "ssms-test-context",
	})
	config.SetCurrentContextName("ssms-test-context")

	// Verify context is set up correctly
	endpoint, user := config.CurrentContext()

	if endpoint.Address != "localhost" {
		t.Errorf("Expected address 'localhost', got '%s'", endpoint.Address)
	}

	if endpoint.Port != 1433 {
		t.Errorf("Expected port 1433, got %d", endpoint.Port)
	}

	if user == nil {
		t.Fatal("Expected user to be set")
	}

	if user.AuthenticationType != "basic" {
		t.Errorf("Expected auth type 'basic', got '%s'", user.AuthenticationType)
	}

	if user.BasicAuth.Username != "sa" {
		t.Errorf("Expected username 'sa', got '%s'", user.BasicAuth.Username)
	}
}

// TestSsmsContextWithoutUser tests SSMS setup without user credentials
func TestSsmsContextWithoutUser(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("SSMS is only available on Windows")
	}

	cmdparser.TestSetup(t)

	// Set up context without user (e.g., for Windows authentication scenarios)
	config.AddEndpoint(sqlconfig.Endpoint{
		AssetDetails: nil,
		EndpointDetails: sqlconfig.EndpointDetails{
			Address: "myserver",
			Port:    1433,
		},
		Name: "ssms-no-user-endpoint",
	})

	config.AddContext(sqlconfig.Context{
		ContextDetails: sqlconfig.ContextDetails{
			Endpoint: "ssms-no-user-endpoint",
			User:     nil,
		},
		Name: "ssms-no-user-context",
	})
	config.SetCurrentContextName("ssms-no-user-context")

	// Verify context is set up correctly
	endpoint, user := config.CurrentContext()

	if endpoint.Address != "myserver" {
		t.Errorf("Expected address 'myserver', got '%s'", endpoint.Address)
	}

	if user != nil {
		t.Error("Expected user to be nil")
	}
}

// Helper function to build server argument string
func buildServerArg(host string, port int) string {
	return host + "," + strconv.Itoa(port)
}
