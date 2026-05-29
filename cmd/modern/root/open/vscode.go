// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/test"
	"github.com/microsoft/go-sqlcmd/internal/tools"
	"github.com/microsoft/go-sqlcmd/internal/tools/tool"
)

// testSettingsPathOverride, when non-empty, overrides getVSCodeSettingsPath
// so tests never touch the real VS Code settings.json.
var testSettingsPathOverride string

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
				Description: localizer.Sprintf("Open a specific VS Code build"),
				Steps:       []string{"sqlcmd open vscode --build insiders"},
			},
		},
		Run: c.run,
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.build,
		Name:   "build",
		Usage:  localizer.Sprintf("VS Code build to open: 'stable' or 'insiders'; defaults to stable when both are installed"),
	})
}

// Launch VS Code and configure connection profile for the current context.
// The connection profile will be added to VS Code's user settings to work
// with the MSSQL extension.
func (c *VSCode) run() {
	endpoint, user := config.CurrentContext()

	build := c.resolveBuild()
	isLocalConnection := isLocalEndpoint(endpoint)

	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		c.ensureContainerIsRunning(asset.Id)
	}

	c.createConnectionProfile(build, endpoint, user, isLocalConnection)

	// Launch VS Code and tell the mssql extension to connect to the profile
	// we just wrote. This focuses the SQL Server activity bar view instead of
	// landing on whatever was open last.
	c.launchVSCode(build, endpoint, user)
}

// resolveBuild validates an explicit --build value and otherwise picks the
// build to configure and launch. An unset --build prefers stable, then
// insiders; if neither is installed it returns stable so the settings path is
// deterministic and launchVSCode reports how to install.
func (c *VSCode) resolveBuild() string {
	switch strings.ToLower(c.build) {
	case "":
		for _, b := range []string{"stable", "insiders"} {
			if vsCodeBuildInstalled(b) {
				return b
			}
		}
		return "stable"
	case "stable":
		return "stable"
	case "insiders":
		return "insiders"
	default:
		c.Output().FatalWithHintExamples([][]string{
			{localizer.Sprintf("Open the stable build"), "sqlcmd open vscode --build stable"},
			{localizer.Sprintf("Open the insiders build"), "sqlcmd open vscode --build insiders"},
		}, localizer.Sprintf("'--build %s' is not supported; use 'stable' or 'insiders'", c.build))
		return ""
	}
}

// vsCodeBuildInstalled reports whether the given VS Code build resolves to an
// installed executable.
func vsCodeBuildInstalled(build string) bool {
	t := tools.NewTool("vscode")
	if vs, ok := t.(*tool.VSCode); ok {
		vs.SetBuild(build)
	}
	return t.IsInstalled()
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

func (c *VSCode) launchVSCode(build string, endpoint sqlconfig.Endpoint, user *sqlconfig.User) {
	output := c.Output()

	t := tools.NewTool("vscode")
	if vs, ok := t.(*tool.VSCode); ok {
		vs.SetBuild(build)
	}
	if !t.IsInstalled() {
		output.Fatal(t.HowToInstall())
	}

	// Don't pre-check or install the mssql extension ourselves. When VS Code
	// follows the vscode://ms-mssql.mssql/... URL and the extension isn't
	// installed, it prompts the user to install it. That UX is better than
	// our fire-and-forget `--install-extension` shell-out, which couldn't
	// report success or failure anyway.

	c.displayPreLaunchInfo()

	if test.IsRunningInTestExecutor() {
		return
	}

	_, err := t.Run([]string{"--open-url", mssqlConnectURI(endpoint, user)})
	c.CheckErr(err)
}

// createConnectionProfile creates or updates a connection profile in VS Code's user settings
func (c *VSCode) createConnectionProfile(build string, endpoint sqlconfig.Endpoint, user *sqlconfig.User, isLocalConnection bool) {
	output := c.Output()

	settingsPath := c.getVSCodeSettingsPath(build)

	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("Error"), err.Error()},
		}, localizer.Sprintf("Failed to create VS Code settings directory"))
	}

	original, settings := c.readSettings(settingsPath)
	profile := c.createProfile(endpoint, user, isLocalConnection)

	connections := c.getConnectionsArray(settings)
	connections = c.updateOrAddProfile(connections, profile)

	// Patch only the two keys we own so hand-edited user settings round-trip.
	updates := map[string]interface{}{
		"mssql.connections":      connections,
		"mssql.connectionGroups": ensureRootGroup(settings["mssql.connectionGroups"]),
	}
	out, err := applyJSONCSettingsUpdates(original, updates)
	if err != nil {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("Error"), err.Error()},
		}, localizer.Sprintf("Failed to update VS Code settings"))
	}
	c.writeSettings(settingsPath, out)

	output.Info(localizer.Sprintf("Connection profile created in VS Code settings"))
}

// readSettings reads settings.json, returning both the original bytes (for AST
// preservation on write) and the parsed map (for reading existing values).
// A missing file is treated as empty.
func (c *VSCode) readSettings(path string) ([]byte, map[string]interface{}) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, make(map[string]interface{})
		}
		c.Output().FatalWithHintExamples([][]string{
			{localizer.Sprintf("Error"), err.Error()},
		}, localizer.Sprintf("Failed to read VS Code settings"))
	}

	settings, err := parseJSONCSettings(data)
	if err != nil {
		c.Output().FatalWithHintExamples([][]string{
			{localizer.Sprintf("Error"), err.Error()},
		}, localizer.Sprintf("Failed to parse VS Code settings"))
	}
	return data, settings
}

func (c *VSCode) writeSettings(path string, data []byte) {
	output := c.Output()

	// Write to a sibling temp file and rename for atomicity; fall back to a
	// direct write if another process holds the file.
	dir := filepath.Dir(path)
	tmp, tmpErr := os.CreateTemp(dir, ".settings-*.tmp")
	if tmpErr == nil {
		tmpPath := tmp.Name()
		_, writeErr := tmp.Write(data)
		closeErr := tmp.Close()
		if writeErr != nil || closeErr != nil {
			_ = os.Remove(tmpPath)
		} else if renameErr := os.Rename(tmpPath, path); renameErr != nil {
			_ = os.Remove(tmpPath)
		} else {
			return
		}
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
	contextName := config.CurrentContextName()

	// trustServerCertificate=true accepts the self-signed certs that local
	// SQL Server containers ship with; encrypt stays Mandatory either way.
	encrypt := "Mandatory"
	trustServerCertificate := isLocalConnection

	profile := map[string]interface{}{
		"applicationName":        "vscode-mssql",
		"commandTimeout":         30,
		"connectRetryCount":      1,
		"connectRetryInterval":   10,
		"connectTimeout":         30,
		"database":               "master",
		"encrypt":                encrypt,
		"groupId":                rootGroupID,
		"id":                     uuid.NewString(),
		"port":                   endpoint.Port,
		"profileName":            contextName,
		"server":                 fmt.Sprintf("%s,%d", endpoint.Address, endpoint.Port),
		"trustServerCertificate": trustServerCertificate,
	}

	// If the endpoint is backed by a local container, surface the container
	// name so the mssql extension can show docker actions in its connection
	// tree.
	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		if name := container.NewController().ContainerName(asset.Id); name != "" {
			profile["containerName"] = name
		}
	}

	if user != nil && user.AuthenticationType == "basic" && user.BasicAuth != nil {
		profile["user"] = user.BasicAuth.Username
		profile["authenticationType"] = "SqlLogin"
		profile["savePassword"] = true

		// Include the decrypted password so the mssql extension can
		// auto-connect without prompting. The extension reads it from the
		// profile on first use and migrates it to the OS credential store,
		// removing it from settings.json.
		if _, _, password := config.GetCurrentContextInfo(); password != "" {
			profile["password"] = password
		}
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
				// Preserve the user's existing group assignment and the
				// extension-assigned id so credentials stay linked.
				if existingGroup, ok := connMap["groupId"].(string); ok && existingGroup != "" {
					newProfile["groupId"] = existingGroup
				}
				if existingID, ok := connMap["id"].(string); ok && existingID != "" {
					newProfile["id"] = existingID
				}
				connections[i] = newProfile
				return connections
			}
		}
	}

	return append(connections, newProfile)
}

// rootGroupID is the stable id of the default connection group the mssql
// extension creates for ungrouped profiles.
const rootGroupID = "ROOT"

// ensureRootGroup returns a connectionGroups array that contains a ROOT entry,
// preserving any other groups the user already has.
func ensureRootGroup(existing interface{}) []interface{} {
	groups, _ := existing.([]interface{})
	for _, g := range groups {
		if gm, ok := g.(map[string]interface{}); ok {
			if id, _ := gm["id"].(string); id == rootGroupID {
				return groups
			}
		}
	}
	return append(groups, map[string]interface{}{
		"id":   rootGroupID,
		"name": rootGroupID,
	})
}

func (c *VSCode) getVSCodeSettingsPath(build string) string {
	if testSettingsPathOverride != "" {
		return testSettingsPathOverride
	}

	stableName := "Code"
	insidersName := "Code - Insiders"
	appName := stableName
	if build == "insiders" {
		appName = insidersName
	}

	getHomeDir := func() string {
		if home := os.Getenv("HOME"); home != "" {
			return home
		}
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return "."
	}

	var configDir string
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("APPDATA")
		if base == "" {
			if home, err := os.UserHomeDir(); err == nil {
				base = filepath.Join(home, "AppData", "Roaming")
			} else {
				base = "."
			}
		}
		configDir = filepath.Join(base, appName, "User")
	case "darwin":
		base := filepath.Join(getHomeDir(), "Library", "Application Support")
		configDir = filepath.Join(base, appName, "User")
	default: // linux and others
		base := filepath.Join(getHomeDir(), ".config")
		configDir = filepath.Join(base, appName, "User")
	}

	return filepath.Join(configDir, "settings.json")
}

// mssqlConnectURI builds a vscode:// URI that the mssql extension's protocol
// handler uses to find the matching saved profile, open an Object Explorer
// session, and focus the SQL Server view.
func mssqlConnectURI(endpoint sqlconfig.Endpoint, user *sqlconfig.User) string {
	q := url.Values{}
	q.Set("profileName", config.CurrentContextName())
	q.Set("server", fmt.Sprintf("%s,%d", endpoint.Address, endpoint.Port))
	q.Set("database", "master")
	if user != nil && user.AuthenticationType == "basic" && user.BasicAuth != nil {
		q.Set("user", user.BasicAuth.Username)
		q.Set("authenticationType", "SqlLogin")
	}
	return "vscode://ms-mssql.mssql/connect?" + q.Encode()
}

func isLocalEndpoint(endpoint sqlconfig.Endpoint) bool {
	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		return true
	}
	addr := strings.ToLower(endpoint.Address)
	return addr == "localhost" || addr == "127.0.0.1" || addr == "::1" || addr == "host.docker.internal"
}
