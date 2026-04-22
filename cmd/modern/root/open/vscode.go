// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/test"
	"github.com/microsoft/go-sqlcmd/internal/tools"
	"github.com/tidwall/jsonc"
)

// testSettingsEnvVar overrides getVSCodeSettingsPath in tests so they
// never touch the real VS Code settings.json. Set via t.Setenv.
const testSettingsEnvVar = "SQLCMD_TEST_VSCODE_SETTINGS_PATH"

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

func (c *VSCode) run() {
	endpoint, user := config.CurrentContext()
	isLocalConnection := isLocalEndpoint(endpoint)

	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		c.ensureContainerIsRunning(asset.Id)
	}

	c.createConnectionProfile(endpoint, user, isLocalConnection)
	copyPasswordToClipboard(user, c.Output())
	c.launchVSCode(endpoint, user, isLocalConnection)
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

func (c *VSCode) launchVSCode(endpoint sqlconfig.Endpoint, user *sqlconfig.User, isLocalConnection bool) {
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
		if !c.isMssqlExtensionInstalled() {
			output.Warn(localizer.Sprintf("The MSSQL extension (ms-mssql.mssql) is not installed in VS Code"))
			output.Info(localizer.Sprintf("To install: sqlcmd open vscode --install-extension"))
		}
	}

	c.displayPreLaunchInfo()

	if test.IsRunningInTestExecutor() {
		return
	}

	// Build a vscode:// URI that triggers the mssql extension's protocol
	// handler, the same mechanism Fabric uses. The OS protocol handler routes
	// it to VS Code without opening a second window.
	connectURI := c.buildConnectURI(endpoint, user, isLocalConnection)
	if err := openURI(connectURI); err != nil {
		output.Warn(localizer.Sprintf("Could not open connection URI: %s", err.Error()))
		// Fall back to just opening VS Code
		_, err = tool.Run(nil)
		c.CheckErr(err)
	}
}

// buildConnectURI creates a vscode://ms-mssql.mssql/connect URI with query
// params matching the connection profile. The mssql extension's protocol
// handler parses these to find a matching profile or open the connect dialog.
func (c *VSCode) buildConnectURI(endpoint sqlconfig.Endpoint, user *sqlconfig.User, isLocalConnection bool) string {
	params := url.Values{}
	params.Set("server", fmt.Sprintf("%s,%d", endpoint.Address, endpoint.Port))
	params.Set("profileName", config.CurrentContextName())

	if isLocalConnection {
		params.Set("encrypt", "Optional")
		params.Set("trustServerCertificate", "true")
	} else {
		params.Set("encrypt", "Mandatory")
		params.Set("trustServerCertificate", "false")
	}

	if user != nil && user.AuthenticationType == "basic" && user.BasicAuth != nil {
		params.Set("user", user.BasicAuth.Username)
		params.Set("authenticationType", "SqlLogin")
	}

	return "vscode://ms-mssql.mssql/connect?" + params.Encode()
}

func (c *VSCode) createConnectionProfile(endpoint sqlconfig.Endpoint, user *sqlconfig.User, isLocalConnection bool) {
	output := c.Output()

	settingsPath := c.getVSCodeSettingsPath()

	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("Error"), err.Error()},
		}, localizer.Sprintf("Failed to create VS Code settings directory"))
	}

	raw, readErr := os.ReadFile(settingsPath)
	if readErr != nil && !os.IsNotExist(readErr) {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("Error"), readErr.Error()},
		}, localizer.Sprintf("Failed to read VS Code settings"))
	}

	settings := c.parseSettings(raw)
	profile := c.createProfile(endpoint, user, isLocalConnection)
	connections := c.getConnectionsArray(settings)
	connections = c.updateOrAddProfile(connections, profile)

	// Patch only mssql.connections, preserving comments
	patched, err := patchJSONCKey(raw, "mssql.connections", connections)
	if err != nil {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("Error"), err.Error()},
		}, localizer.Sprintf("Failed to update VS Code settings"))
	}

	c.writeSettingsRaw(settingsPath, patched)

	output.Info(localizer.Sprintf("Connection profile created in VS Code settings"))
}

func (c *VSCode) parseSettings(data []byte) map[string]interface{} {
	settings := make(map[string]interface{})
	if len(data) > 0 {
		clean := jsonc.ToJSON(data)
		if err := json.Unmarshal(clean, &settings); err != nil {
			output := c.Output()
			output.FatalWithHintExamples([][]string{
				{localizer.Sprintf("Error"), err.Error()},
			}, localizer.Sprintf("Failed to parse VS Code settings"))
		}
	}
	return settings
}

func (c *VSCode) writeSettingsRaw(path string, data []byte) {
	output := c.Output()

	// Preserve existing file permissions, or use 0600 for new files.
	mode := os.FileMode(0600)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode()
	}

	// Atomic write: temp file + rename, with direct-write fallback.
	dir := filepath.Dir(path)
	tmp, tmpErr := os.CreateTemp(dir, ".settings-*.tmp")
	if tmpErr == nil {
		tmpPath := tmp.Name()
		_ = tmp.Chmod(mode)
		_, writeErr := tmp.Write(data)
		closeErr := tmp.Close()
		if writeErr != nil || closeErr != nil {
			_ = os.Remove(tmpPath)
		} else if renameErr := os.Rename(tmpPath, path); renameErr != nil {
			_ = os.Remove(tmpPath)
		} else {
			return // atomic write succeeded
		}
	}

	// Fallback: direct write
	if err := os.WriteFile(path, data, mode); err != nil {
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
	contextName := config.CurrentContextName()

	encrypt := "Mandatory"
	trustServerCertificate := false

	// Local connections (containers, localhost) commonly use self-signed certs
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
		profile["authenticationType"] = "SqlLogin"
		profile["savePassword"] = true
	}

	return profile
}

func (c *VSCode) updateOrAddProfile(connections []interface{}, newProfile map[string]interface{}) []interface{} {
	profileName, ok := newProfile["profileName"].(string)
	if !ok {
		return append(connections, newProfile)
	}

	for i, conn := range connections {
		if connMap, ok := conn.(map[string]interface{}); ok {
			if name, ok := connMap["profileName"].(string); ok && name == profileName {
				connections[i] = newProfile
				return connections
			}
		}
	}

	return append(connections, newProfile)
}

func (c *VSCode) getVSCodeSettingsPath() string {
	if override := os.Getenv(testSettingsEnvVar); override != "" {
		return override
	}

	var stableDir string
	var insidersDir string

	homeDir := func() string {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return "."
	}

	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("APPDATA")
		if base == "" {
			base = filepath.Join(homeDir(), "AppData", "Roaming")
		}
		stableDir = filepath.Join(base, "Code", "User")
		insidersDir = filepath.Join(base, "Code - Insiders", "User")
	case "darwin":
		base := filepath.Join(homeDir(), "Library", "Application Support")
		stableDir = filepath.Join(base, "Code", "User")
		insidersDir = filepath.Join(base, "Code - Insiders", "User")
	default: // linux and others
		base := filepath.Join(homeDir(), ".config")
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

// isMssqlExtensionInstalled checks the VS Code extensions directory on disk
// instead of running Code.exe --list-extensions (which opens a window).
func (c *VSCode) isMssqlExtensionInstalled() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return true // assume installed if we can't check
	}

	for _, dir := range []string{".vscode-insiders", ".vscode"} {
		ext := filepath.Join(home, dir, "extensions")
		entries, err := os.ReadDir(ext)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasPrefix(strings.ToLower(e.Name()), "ms-mssql.mssql-") {
				return true
			}
		}
	}
	return false
}

func isLocalEndpoint(endpoint sqlconfig.Endpoint) bool {
	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		return true
	}

	addr := strings.ToLower(endpoint.Address)
	return addr == "localhost" || addr == "127.0.0.1" || addr == "::1" || addr == "host.docker.internal"
}
