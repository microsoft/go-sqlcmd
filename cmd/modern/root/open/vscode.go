// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/tools"
	"github.com/microsoft/go-sqlcmd/internal/tools/tool"
)

// VSCode implements the `sqlcmd open vscode` command. It opens
// Visual Studio Code and configures a connection profile for the
// current context using the MSSQL extension.
func (c *VSCode) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "vscode",
		Short: localizer.Sprintf("Open Visual Studio Code and configure connection for current context"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("Open VS Code and configure connection using the current context"),
				Steps:       []string{"sqlcmd open vscode"},
			},
			{
				Description: localizer.Sprintf("Open VS Code and install the MSSQL extension if needed"),
				Steps:       []string{"sqlcmd open vscode --install-extension"},
			},
		},
		Run: c.run,
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.installExtension,
		Name:  "install-extension",
		Usage: localizer.Sprintf("Install the MSSQL extension in VS Code if not already installed"),
	})
}

// Launch VS Code and configure connection profile for the current context.
// The connection profile will be added to VS Code's user settings to work
// with the MSSQL extension.
func (c *VSCode) run() {
	endpoint, user := config.CurrentContext()

	// Check if this is a local container connection
	isLocalConnection := isLocalEndpoint(endpoint)

	// If the context has a local container, ensure it is running, otherwise bail out
	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		c.ensureContainerIsRunning(asset.Id)
	}

	// Create or update connection profile in VS Code settings
	c.createConnectionProfile(endpoint, user, isLocalConnection)

	// Copy password to clipboard if using SQL authentication
	copyPasswordToClipboard(user, c.Output())

	// Launch VS Code
	c.launchVSCode()
}

func (c *VSCode) ensureContainerIsRunning(containerID string) {
	output := c.Output()
	controller := container.NewController()
	if !controller.ContainerRunning(containerID) {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("To start the container"), localizer.Sprintf("sqlcmd start")},
		}, localizer.Sprintf("Container is not running"))
	}
}

// launchVSCode launches Visual Studio Code
func (c *VSCode) launchVSCode() {
	output := c.Output()

	tool := tools.NewTool("vscode")
	if !tool.IsInstalled() {
		output.Fatal(tool.HowToInstall())
	}

	// Install the MSSQL extension if explicitly requested
	if c.installExtension {
		output.Info(localizer.Sprintf("Installing MSSQL extension..."))
		_, err := tool.Run([]string{"--install-extension", "ms-mssql.mssql", "--force"})
		if err != nil {
			output.Warn(localizer.Sprintf("Could not install MSSQL extension: %s", err.Error()))
		} else {
			output.Info(localizer.Sprintf("MSSQL extension installed successfully"))
		}
	} else {
		// Check if MSSQL extension is installed, warn if not
		if !c.isMssqlExtensionInstalled(tool) {
			output.FatalWithHintExamples([][]string{
				{localizer.Sprintf("To install the MSSQL extension"), "sqlcmd open vscode --install-extension"},
			}, localizer.Sprintf("The MSSQL extension (ms-mssql.mssql) is not installed in VS Code"))
		}
	}

	c.displayPreLaunchInfo()

	// Open VS Code
	_, err := tool.Run([]string{})
	c.CheckErr(err)
}

// createConnectionProfile creates or updates a connection profile in VS Code's user settings
func (c *VSCode) createConnectionProfile(endpoint sqlconfig.Endpoint, user *sqlconfig.User, isLocalConnection bool) {
	output := c.Output()

	settingsPath := c.getVSCodeSettingsPath()

	// Ensure the directory exists
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("Error"), err.Error()},
		}, localizer.Sprintf("Failed to create VS Code settings directory"))
	}

	// Read existing settings or create new
	settings := c.readSettings(settingsPath)

	// Create connection profile
	profile := c.createProfile(endpoint, user, isLocalConnection)

	// Add or update the connection profile
	connections := c.getConnectionsArray(settings)
	connections = c.updateOrAddProfile(connections, profile)
	settings["mssql.connections"] = connections

	// Write settings back
	c.writeSettings(settingsPath, settings)

	output.Info(localizer.Sprintf("Connection profile created in VS Code settings"))
}

func (c *VSCode) readSettings(path string) map[string]interface{} {
	settings := make(map[string]interface{})

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return settings
		}
		output := c.Output()
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("Error"), err.Error()},
		}, localizer.Sprintf("Failed to read VS Code settings"))
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			output := c.Output()
			output.FatalWithHintExamples([][]string{
				{localizer.Sprintf("Error"), err.Error()},
			}, localizer.Sprintf("Failed to parse VS Code settings"))
		}
	}

	return settings
}

func (c *VSCode) writeSettings(path string, settings map[string]interface{}) {
	output := c.Output()

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("Error"), err.Error()},
		}, localizer.Sprintf("Failed to encode VS Code settings"))
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("Error"), err.Error()},
		}, localizer.Sprintf("Failed to write VS Code settings"))
	}
}

func (c *VSCode) getConnectionsArray(settings map[string]interface{}) []interface{} {
	connections := []interface{}{}
	if existing, ok := settings["mssql.connections"]; ok {
		if arr, ok := existing.([]interface{}); ok {
			connections = arr
		}
	}
	return connections
}

func (c *VSCode) createProfile(endpoint sqlconfig.Endpoint, user *sqlconfig.User, isLocalConnection bool) map[string]interface{} {
	// Use context name as the profile name - this is the user's chosen identifier
	// and matches what they use with sqlcmd commands
	contextName := config.CurrentContextName()

	// Default to secure settings for production connections
	encrypt := "Mandatory"
	trustServerCertificate := false

	// Relax settings for local connections (containers, localhost) that commonly use
	// self-signed certificates. Users can still adjust these values in VS Code settings.
	if isLocalConnection {
		encrypt = "Optional"
		trustServerCertificate = true
	}

	profile := map[string]interface{}{
		"server":                 fmt.Sprintf("%s,%d", endpoint.Address, endpoint.Port),
		"profileName":            contextName,
		"encrypt":                encrypt,
		"trustServerCertificate": trustServerCertificate,
	}

	if user != nil && user.AuthenticationType == "basic" && user.BasicAuth != nil {
		profile["user"] = user.BasicAuth.Username
		// SQL authentication contexts use SqlLogin
		profile["authenticationType"] = "SqlLogin"
		profile["savePassword"] = true
	}

	return profile
}

func (c *VSCode) updateOrAddProfile(connections []interface{}, newProfile map[string]interface{}) []interface{} {
	profileName, ok := newProfile["profileName"].(string)
	if !ok {
		// If profileName is not a valid string, just append the profile
		return append(connections, newProfile)
	}

	// Check if profile with same name exists and update it
	for i, conn := range connections {
		if connMap, ok := conn.(map[string]interface{}); ok {
			if name, ok := connMap["profileName"].(string); ok && name == profileName {
				connections[i] = newProfile
				return connections
			}
		}
	}

	// Add new profile
	return append(connections, newProfile)
}

func (c *VSCode) getVSCodeSettingsPath() string {
	var stableDir string
	var insidersDir string

	getHomeDir := func() string {
		if home := os.Getenv("HOME"); home != "" {
			return home
		}
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return "."
	}

	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("APPDATA")
		if base == "" {
			// Fallback to deriving APPDATA from user home
			if home, err := os.UserHomeDir(); err == nil {
				base = filepath.Join(home, "AppData", "Roaming")
			} else {
				base = "."
			}
		}
		stableDir = filepath.Join(base, "Code", "User")
		insidersDir = filepath.Join(base, "Code - Insiders", "User")
	case "darwin":
		base := filepath.Join(getHomeDir(), "Library", "Application Support")
		stableDir = filepath.Join(base, "Code", "User")
		insidersDir = filepath.Join(base, "Code - Insiders", "User")
	default: // linux and others
		base := filepath.Join(getHomeDir(), ".config")
		stableDir = filepath.Join(base, "Code", "User")
		insidersDir = filepath.Join(base, "Code - Insiders", "User")
	}

	// Prefer VS Code Insiders settings if the directory exists, since the tool
	// searches for and launches Insiders first. Fall back to stable Code.
	configDir := stableDir
	if info, err := os.Stat(insidersDir); err == nil && info.IsDir() {
		configDir = insidersDir
	}

	return filepath.Join(configDir, "settings.json")
}

// isMssqlExtensionInstalled checks if the MSSQL extension is installed in VS Code
func (c *VSCode) isMssqlExtensionInstalled(t tool.Tool) bool {
	output, _, err := t.RunWithOutput([]string{"--list-extensions"})
	if err != nil {
		// If we can't list extensions, assume it's installed to avoid blocking the user,
		// but emit a warning so the user is aware that verification failed.
		c.Output().Warn(localizer.Sprintf("Could not verify MSSQL extension installation: %s", err.Error()))
		return true
	}

	// Check if the MSSQL extension is in the list (case-insensitive)
	extensions := strings.ToLower(output)
	return strings.Contains(extensions, "ms-mssql.mssql")
}

// isLocalEndpoint checks if the endpoint is a local connection (container, localhost, etc.)
// This is used to determine whether to use relaxed TLS settings.
func isLocalEndpoint(endpoint sqlconfig.Endpoint) bool {
	// Check if this is a container-based connection
	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		return true
	}

	// Check for common local addresses
	addr := strings.ToLower(endpoint.Address)
	return addr == "localhost" || addr == "127.0.0.1" || addr == "::1" || addr == "host.docker.internal"
}
