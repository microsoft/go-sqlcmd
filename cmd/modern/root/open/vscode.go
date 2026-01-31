// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/tools"
)

// VSCode implements the `sqlcmd open vscode` command. It opens
// Visual Studio Code and configures a connection profile for the
// current context using the MSSQL extension.
func (c *VSCode) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "vscode",
		Short: "Open Visual Studio Code and configure connection for current context",
		Examples: []cmdparser.ExampleOptions{{
			Description: "Open VS Code and configure connection using the current context",
			Steps:       []string{"sqlcmd open vscode"}}},
		Run: c.run,
	}

	c.Cmd.DefineCommand(options)
}

// Launch VS Code and configure connection profile for the current context.
// The connection profile will be added to VS Code's user settings to work
// with the MSSQL extension.
func (c *VSCode) run() {
	endpoint, user := config.CurrentContext()

	// If the context has a local container, ensure it is running, otherwise bail out
	if endpoint.AssetDetails != nil && endpoint.AssetDetails.ContainerDetails != nil {
		c.ensureContainerIsRunning(endpoint)
	}

	// Create or update connection profile in VS Code settings
	c.createConnectionProfile(endpoint, user)

	// Launch VS Code
	c.launchVSCode()
}

func (c *VSCode) ensureContainerIsRunning(endpoint sqlconfig.Endpoint) {
	output := c.Output()
	controller := container.NewController()
	if !controller.ContainerRunning(endpoint.AssetDetails.ContainerDetails.Id) {
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

	c.displayPreLaunchInfo()

	// Just open VS Code, the connection profile is already configured
	_, err := tool.Run([]string{})
	c.CheckErr(err)
}

// createConnectionProfile creates or updates a connection profile in VS Code's user settings
func (c *VSCode) createConnectionProfile(endpoint sqlconfig.Endpoint, user *sqlconfig.User) {
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
	profile := c.createProfile(endpoint, user)

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

	if err := os.WriteFile(path, data, 0644); err != nil {
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

func (c *VSCode) createProfile(endpoint sqlconfig.Endpoint, user *sqlconfig.User) map[string]interface{} {
	profile := map[string]interface{}{
		"server":      fmt.Sprintf("%s,%d", endpoint.EndpointDetails.Address, endpoint.EndpointDetails.Port),
		"profileName": fmt.Sprintf("sqlcmd-%s", config.CurrentContextName()),
		// Set encrypt to "Optional" for compatibility with local development containers
		// which often use self-signed certificates. Users can modify this in VS Code settings
		// to "Mandatory" or "Strict" for production connections.
		"encrypt": "Optional",
	}

	if user != nil && user.AuthenticationType == "basic" {
		profile["authenticationType"] = "SqlLogin"
		profile["user"] = user.BasicAuth.Username
		// Note: We don't store the password in settings, VS Code will prompt for it
		// or user can configure it through the extension
	} else {
		if runtime.GOOS == "windows" {
			profile["authenticationType"] = "Integrated"
		} else {
			profile["authenticationType"] = "SqlLogin"
		}
	}

	return profile
}

func (c *VSCode) updateOrAddProfile(connections []interface{}, newProfile map[string]interface{}) []interface{} {
	profileName := newProfile["profileName"].(string)
	
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
	var configDir string
	
	switch runtime.GOOS {
	case "windows":
		configDir = filepath.Join(os.Getenv("APPDATA"), "Code", "User")
	case "darwin":
		configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Code", "User")
	default: // linux and others
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "Code", "User")
	}
	
	return filepath.Join(configDir, "settings.json")
}
