// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"github.com/microsoft/go-sqlcmd/internal/tools"
)

// TestVSCode runs a sanity test of `sqlcmd open vscode`
func TestVSCode(t *testing.T) {
	tool := tools.NewTool("vscode")
	if !tool.IsInstalled() {
		t.Skip("VS Code is not installed")
	}

	// Redirect settings writes to a temp directory so the test never
	// touches the real VS Code settings.json.
	testSettingsPathOverride = filepath.Join(t.TempDir(), "settings.json")
	t.Cleanup(func() { testSettingsPathOverride = "" })

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

	cmdparser.TestCmd[*VSCode]()
}

// TestVSCodeCreateProfile tests that createProfile generates correct profile structure
func TestVSCodeCreateProfile(t *testing.T) {
	cmdparser.TestSetup(t)

	// Set up a context with user credentials
	config.AddEndpoint(sqlconfig.Endpoint{
		AssetDetails: nil,
		EndpointDetails: sqlconfig.EndpointDetails{
			Address: "localhost",
			Port:    1433,
		},
		Name: "test-endpoint",
	})

	config.AddUser(sqlconfig.User{
		AuthenticationType: "basic",
		BasicAuth: &sqlconfig.BasicAuthDetails{
			Username:           "sa",
			PasswordEncryption: "none",
			Password:           secret.Encode("testpassword", "none"),
		},
		Name: "test-user",
	})

	config.AddContext(sqlconfig.Context{
		ContextDetails: sqlconfig.ContextDetails{
			Endpoint: "test-endpoint",
			User:     strPtr("test-user"),
		},
		Name: "my-database",
	})
	config.SetCurrentContextName("my-database")

	// Create a VSCode command instance and test profile creation
	vscode := &VSCode{}
	endpoint, user := config.CurrentContext()

	profile := vscode.createProfile(endpoint, user, true) // true for local connection

	// Verify profile structure
	if profile["server"] != "localhost,1433" {
		t.Errorf("Expected server 'localhost,1433', got '%v'", profile["server"])
	}

	if profile["profileName"] != "my-database" {
		t.Errorf("Expected profileName 'my-database', got '%v'", profile["profileName"])
	}

	if profile["authenticationType"] != "SqlLogin" {
		t.Errorf("Expected authenticationType 'SqlLogin', got '%v'", profile["authenticationType"])
	}

	if profile["user"] != "sa" {
		t.Errorf("Expected user 'sa', got '%v'", profile["user"])
	}

	if profile["encrypt"] != "Mandatory" {
		t.Errorf("Expected encrypt 'Mandatory', got '%v'", profile["encrypt"])
	}

	if profile["trustServerCertificate"] != true {
		t.Errorf("Expected trustServerCertificate true, got '%v'", profile["trustServerCertificate"])
	}

	if profile["savePassword"] != true {
		t.Errorf("Expected savePassword true, got '%v'", profile["savePassword"])
	}
}

// TestVSCodeCreateProfileRemoteDoesNotPersistPassword asserts that for non-local
// connections we hand the credentials off to the mssql extension's own prompt
// rather than writing the plaintext password into settings.json.
func TestVSCodeCreateProfileRemoteDoesNotPersistPassword(t *testing.T) {
	cmdparser.TestSetup(t)

	config.AddEndpoint(sqlconfig.Endpoint{
		EndpointDetails: sqlconfig.EndpointDetails{Address: "remote.example.com", Port: 1433},
		Name:            "remote-endpoint",
	})
	config.AddUser(sqlconfig.User{
		AuthenticationType: "basic",
		BasicAuth: &sqlconfig.BasicAuthDetails{
			Username:           "sa",
			PasswordEncryption: "none",
			Password:           secret.Encode("testpassword", "none"),
		},
		Name: "remote-user",
	})
	config.AddContext(sqlconfig.Context{
		ContextDetails: sqlconfig.ContextDetails{
			Endpoint: "remote-endpoint",
			User:     strPtr("remote-user"),
		},
		Name: "remote-context",
	})
	config.SetCurrentContextName("remote-context")

	endpoint, user := config.CurrentContext()
	profile := (&VSCode{}).createProfile(endpoint, user, false)

	if profile["user"] != "sa" {
		t.Errorf("Expected user 'sa', got %v", profile["user"])
	}
	if _, ok := profile["savePassword"]; ok {
		t.Error("Expected savePassword to be absent for remote connections")
	}
	if _, ok := profile["password"]; ok {
		t.Error("Expected password to be absent for remote connections")
	}
}

// TestVSCodeUpdateOrAddProfile tests profile update and add logic
func TestVSCodeUpdateOrAddProfile(t *testing.T) {
	cmdparser.TestSetup(t)

	vscode := &VSCode{}

	// Test adding a new profile to empty list
	connections := []interface{}{}
	newProfile := map[string]interface{}{
		"profileName": "test-profile",
		"server":      "localhost,1433",
	}

	result := vscode.updateOrAddProfile(connections, newProfile)
	if len(result) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(result))
	}

	// Test adding a second profile with different name
	secondProfile := map[string]interface{}{
		"profileName": "another-profile",
		"server":      "server2,1434",
	}

	result = vscode.updateOrAddProfile(result, secondProfile)
	if len(result) != 2 {
		t.Errorf("Expected 2 connections, got %d", len(result))
	}

	// Test updating existing profile (same name)
	updatedProfile := map[string]interface{}{
		"profileName": "test-profile",
		"server":      "localhost,2000",
		"user":        "newuser",
	}

	result = vscode.updateOrAddProfile(result, updatedProfile)
	if len(result) != 2 {
		t.Errorf("Expected 2 connections after update, got %d", len(result))
	}

	// Verify the profile was updated, not duplicated
	found := false
	for _, conn := range result {
		if connMap, ok := conn.(map[string]interface{}); ok {
			if connMap["profileName"] == "test-profile" {
				found = true
				if connMap["server"] != "localhost,2000" {
					t.Errorf("Expected updated server 'localhost,2000', got '%v'", connMap["server"])
				}
				if connMap["user"] != "newuser" {
					t.Errorf("Expected updated user 'newuser', got '%v'", connMap["user"])
				}
			}
		}
	}
	if !found {
		t.Error("Updated profile not found in connections")
	}
}

// TestVSCodeGetConnectionsArray tests extracting connections array from settings
func TestVSCodeGetConnectionsArray(t *testing.T) {
	cmdparser.TestSetup(t)

	vscode := &VSCode{}

	// Test with no connections key
	settings := map[string]interface{}{}
	connections := vscode.getConnectionsArray(settings)
	if len(connections) != 0 {
		t.Errorf("Expected empty array, got %d items", len(connections))
	}

	// Test with connections array
	settings["mssql.connections"] = []interface{}{
		map[string]interface{}{"profileName": "test1"},
		map[string]interface{}{"profileName": "test2"},
	}
	connections = vscode.getConnectionsArray(settings)
	if len(connections) != 2 {
		t.Errorf("Expected 2 connections, got %d", len(connections))
	}

	// Test with wrong type (should return empty array)
	settings["mssql.connections"] = "not an array"
	connections = vscode.getConnectionsArray(settings)
	if len(connections) != 0 {
		t.Errorf("Expected empty array for invalid type, got %d items", len(connections))
	}
}

// TestVSCodeGetSettingsPath tests that the settings path routes to the
// requested build's user directory.
func TestVSCodeGetSettingsPath(t *testing.T) {
	cmdparser.TestSetup(t)

	vscode := &VSCode{}

	stable := vscode.getVSCodeSettingsPath("stable")
	insiders := vscode.getVSCodeSettingsPath("insiders")

	for _, path := range []string{stable, insiders} {
		if filepath.Base(path) != "settings.json" {
			t.Errorf("Expected path to end with 'settings.json', got '%s'", filepath.Base(path))
		}
	}

	if !strings.Contains(insiders, "Code - Insiders") {
		t.Errorf("Expected insiders path to contain 'Code - Insiders', got '%s'", insiders)
	}
	if strings.Contains(stable, "Code - Insiders") {
		t.Errorf("Expected stable path to not contain 'Code - Insiders', got '%s'", stable)
	}

	switch runtime.GOOS {
	case "windows":
		want := filepath.Join("AppData", "Roaming", "Code", "User", "settings.json")
		if !strings.HasSuffix(stable, want) {
			t.Errorf("Expected Windows path to end with %q, got %q", want, stable)
		}
	case "darwin":
		want := filepath.Join("Library", "Application Support", "Code", "User", "settings.json")
		if !strings.HasSuffix(stable, want) {
			t.Errorf("Expected macOS path to end with %q, got %q", want, stable)
		}
	default:
		// Real linux path depends on env (XDG_CONFIG_HOME) and snap detection;
		// pure-function behavior is covered by TestLinuxVSCodeConfigDir.
		if !strings.Contains(stable, filepath.Join("Code", "User")) {
			t.Errorf("Expected Linux path to contain Code/User, got %q", stable)
		}
	}
}

func TestLinuxVSCodeConfigDir(t *testing.T) {
	home := filepath.FromSlash("/home/user")

	cases := []struct {
		name    string
		appName string
		build   string
		exe     string
		xdg     string
		want    string
	}{
		{
			name:    "stable defaults to ~/.config",
			appName: "Code",
			build:   "stable",
			want:    filepath.Join(home, ".config", "Code", "User"),
		},
		{
			name:    "insiders defaults to ~/.config/Code - Insiders",
			appName: "Code - Insiders",
			build:   "insiders",
			want:    filepath.Join(home, ".config", "Code - Insiders", "User"),
		},
		{
			name:    "XDG_CONFIG_HOME overrides ~/.config",
			appName: "Code",
			build:   "stable",
			xdg:     filepath.FromSlash("/custom/xdg"),
			want:    filepath.FromSlash("/custom/xdg/Code/User"),
		},
		{
			name:    "snap stable redirects under ~/snap/code",
			appName: "Code",
			build:   "stable",
			exe:     "/snap/bin/code",
			xdg:     filepath.FromSlash("/custom/xdg"), // ignored under snap confinement
			want:    filepath.Join(home, "snap", "code", "current", ".config", "Code", "User"),
		},
		{
			name:    "snap insiders redirects under ~/snap/code-insiders",
			appName: "Code - Insiders",
			build:   "insiders",
			exe:     "/snap/bin/code-insiders",
			want:    filepath.Join(home, "snap", "code-insiders", "current", ".config", "Code - Insiders", "User"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := linuxVSCodeConfigDir(home, tc.appName, tc.build, tc.exe, tc.xdg)
			if got != tc.want {
				t.Errorf("linuxVSCodeConfigDir = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestVSCodeProfileWithoutUser tests profile creation when no user is configured
func TestVSCodeProfileWithoutUser(t *testing.T) {
	cmdparser.TestSetup(t)

	config.AddEndpoint(sqlconfig.Endpoint{
		AssetDetails: nil,
		EndpointDetails: sqlconfig.EndpointDetails{
			Address: "myserver",
			Port:    1433,
		},
		Name: "no-user-endpoint",
	})

	config.AddContext(sqlconfig.Context{
		ContextDetails: sqlconfig.ContextDetails{
			Endpoint: "no-user-endpoint",
			User:     nil,
		},
		Name: "no-user-context",
	})
	config.SetCurrentContextName("no-user-context")

	vscode := &VSCode{}
	endpoint, user := config.CurrentContext()

	profile := vscode.createProfile(endpoint, user, false) // false for non-local connection

	// Verify profile doesn't have user field when no user is configured
	if _, hasUser := profile["user"]; hasUser {
		t.Error("Expected profile to not have 'user' field when no user configured")
	}

	// Verify other fields are still set correctly
	if profile["profileName"] != "no-user-context" {
		t.Errorf("Expected profileName 'no-user-context', got '%v'", profile["profileName"])
	}

	// Verify secure TLS settings for non-local connections
	if profile["encrypt"] != "Mandatory" {
		t.Errorf("Expected encrypt 'Mandatory' for non-local connection, got '%v'", profile["encrypt"])
	}

	if profile["trustServerCertificate"] != false {
		t.Errorf("Expected trustServerCertificate false for non-local connection, got '%v'", profile["trustServerCertificate"])
	}
}

// Helper to create string pointer
func strPtr(s string) *string {
	return &s
}
