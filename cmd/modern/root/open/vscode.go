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

func (c *VSCode) run() {
	if config.CurrentContextName() == "" {
		c.Output().FatalWithHintExamples([][]string{
			{localizer.Sprintf("To view available contexts"), "sqlcmd config get-contexts"},
		}, localizer.Sprintf("No current context"))
	}

	endpoint, user := config.CurrentContext()

	build := c.resolveBuild()
	isLocalConnection := isLocalEndpoint(endpoint)

	// Verify VS Code is installed before touching settings.json. Otherwise a
	// failed launch would leave behind a freshly written settings file (and a
	// plaintext local-context password) that the user never asked for.
	t := tools.NewTool("vscode")
	if vs, ok := t.(*tool.VSCode); ok {
		vs.SetBuild(build)
	}
	if !t.IsInstalled() {
		c.Output().Fatal(t.HowToInstall())
	}

	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		c.ensureContainerIsRunning(asset.Id)
	}

	c.createConnectionProfile(build, endpoint, user, isLocalConnection)

	// Launch VS Code and tell the mssql extension to connect to the profile
	// we just wrote. This focuses the SQL Server activity bar view instead of
	// landing on whatever was open last.
	c.launchVSCode(t, endpoint, user, isLocalConnection)
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

func (c *VSCode) launchVSCode(t tool.Tool, endpoint sqlconfig.Endpoint, user *sqlconfig.User, isLocalConnection bool) {
	// Don't pre-check or install the mssql extension ourselves. When VS Code
	// follows the vscode://ms-mssql.mssql/... URL and the extension isn't
	// installed, it prompts the user to install it. That UX is better than
	// our fire-and-forget `--install-extension` shell-out, which couldn't
	// report success or failure anyway.

	// For remote SQL auth, the password isn't written to settings.json, so the
	// mssql extension will prompt for it. Stage it on the clipboard so the user
	// can paste rather than retype.
	if !isLocalConnection {
		copyPasswordToClipboard(user, c.Output())
	}

	c.displayPreLaunchInfo()

	if test.IsRunningInTestExecutor() {
		return
	}

	_, err := t.Run([]string{"--open-url", mssqlConnectURI(endpoint, user)})
	c.CheckErr(err)
}

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
		}, localizer.Sprintf("Failed to encode VS Code settings"))
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

	settings, err := parseJSONCSettings(append([]byte(nil), data...))
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

		// Only persist the decrypted password for the local-container dev
		// flow. For remote servers, the user can save credentials through
		// the mssql extension's own prompt rather than have sqlcmd write
		// them into settings.json.
		if isLocalConnection {
			if _, _, password := config.GetCurrentContextInfo(); password != "" {
				profile["savePassword"] = true
				profile["password"] = password
			}
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
	path, err := vsCodeSettingsPath(build)
	if err != nil {
		hint := [][]string{
			{localizer.Sprintf("Set the HOME environment variable"), "export HOME=/your/home"},
		}
		if runtime.GOOS == "windows" {
			hint = [][]string{
				{localizer.Sprintf("Set the USERPROFILE environment variable"), `set USERPROFILE=C:\Users\you`},
			}
		}
		c.Output().FatalWithHintExamples(hint, localizer.Sprintf("Could not resolve home directory: %s", err.Error()))
	}
	return path
}

// vsCodeSettingsPath resolves the settings.json path for the given VS Code
// build without exiting on failure, so callers like the uninstall cleanup
// path can degrade gracefully when the home directory is unavailable.
func vsCodeSettingsPath(build string) (string, error) {
	if testSettingsPathOverride != "" {
		return testSettingsPathOverride, nil
	}

	appName := "Code"
	if build == "insiders" {
		appName = "Code - Insiders"
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if home == "" {
		return "", fmt.Errorf("empty home directory")
	}

	var configDir string
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("APPDATA")
		if base == "" {
			base = filepath.Join(home, "AppData", "Roaming")
		}
		configDir = filepath.Join(base, appName, "User")
	case "darwin":
		configDir = filepath.Join(home, "Library", "Application Support", appName, "User")
	default: // linux and others
		configDir = linuxVSCodeConfigDir(home, appName, build, vsCodeExePath(build), os.Getenv("XDG_CONFIG_HOME"))
	}

	return filepath.Join(configDir, "settings.json"), nil
}

// linuxVSCodeConfigDir resolves the VS Code User config directory on Linux.
// Snap installs are sandboxed so their settings live under
// $HOME/snap/<snapName>/current/.config/<appName>/User regardless of
// XDG_CONFIG_HOME; non-snap installs honor XDG_CONFIG_HOME and fall back to
// $HOME/.config.
func linuxVSCodeConfigDir(home, appName, build, exePath, xdgConfigHome string) string {
	if strings.HasPrefix(exePath, "/snap/") {
		snapName := "code"
		if build == "insiders" {
			snapName = "code-insiders"
		}
		return filepath.Join(home, "snap", snapName, "current", ".config", appName, "User")
	}
	base := xdgConfigHome
	if base == "" {
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, appName, "User")
}

// vsCodeExePath returns the resolved VS Code executable path for the given
// build, or "" if VS Code is not installed.
func vsCodeExePath(build string) string {
	t := tools.NewTool("vscode")
	vs, ok := t.(*tool.VSCode)
	if !ok {
		return ""
	}
	vs.SetBuild(build)
	if !t.IsInstalled() {
		return ""
	}
	return vs.ExePath()
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

// isLocalEndpoint reports whether the endpoint is a sqlcmd-managed local
// container. Only container-backed endpoints get the "local dev" treatment
// (trusting the self-signed cert and writing a plaintext password into VS
// Code settings); loopback hosts that happen to be port-forwards to remote
// servers must not opt into those tradeoffs.
func isLocalEndpoint(endpoint sqlconfig.Endpoint) bool {
	asset := endpoint.AssetDetails
	return asset != nil && asset.ContainerDetails != nil
}

// RemoveContextFromVSCodeSettings deletes any mssql connection profile named
// contextName from each known VS Code build's settings.json. It is best
// effort: missing files, unresolvable home dirs, and parse/write errors are
// swallowed so `sqlcmd delete` never fails on cleanup. Returns the list of
// settings paths that were actually modified, so callers can report what
// they cleaned up.
func RemoveContextFromVSCodeSettings(contextName string) []string {
	if contextName == "" {
		return nil
	}
	var cleaned []string
	seen := map[string]bool{}
	for _, build := range []string{"stable", "insiders"} {
		path, err := vsCodeSettingsPath(build)
		if err != nil || path == "" || seen[path] {
			continue
		}
		seen[path] = true
		removed, err := removeProfileFromVSCodeSettings(path, contextName)
		if err == nil && removed {
			cleaned = append(cleaned, path)
		}
	}
	return cleaned
}

// removeProfileFromVSCodeSettings rewrites settings.json with any
// mssql.connections entry whose profileName matches contextName stripped out.
// Returns (true, nil) when the file was modified, (false, nil) when no
// matching profile was present or the file does not exist.
func removeProfileFromVSCodeSettings(settingsPath, contextName string) (bool, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// parseJSONCSettings may mutate the input bytes via hujson.Standardize,
	// so peek on a copy and keep the pristine original for the JSONC-aware
	// rewrite below.
	settings, err := parseJSONCSettings(append([]byte(nil), data...))
	if err != nil {
		return false, err
	}

	existing, ok := settings["mssql.connections"].([]interface{})
	if !ok || len(existing) == 0 {
		return false, nil
	}

	filtered := make([]interface{}, 0, len(existing))
	for _, conn := range existing {
		if connMap, ok := conn.(map[string]interface{}); ok {
			if name, _ := connMap["profileName"].(string); name == contextName {
				continue
			}
		}
		filtered = append(filtered, conn)
	}
	if len(filtered) == len(existing) {
		return false, nil
	}

	out, err := applyJSONCSettingsUpdates(data, map[string]interface{}{
		"mssql.connections": filtered,
	})
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(settingsPath, out, 0600); err != nil {
		return false, err
	}
	return true, nil
}
