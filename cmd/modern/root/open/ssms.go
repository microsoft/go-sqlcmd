//go:build windows

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/test"
	"github.com/microsoft/go-sqlcmd/internal/tools"
	"github.com/microsoft/go-sqlcmd/internal/tools/tool"
)

// minSsmsVersion is the oldest SSMS major version this command supports. SSMS
// 21+ registers with the Visual Studio Installer and is discoverable via
// vswhere; older releases (legacy MSI) are out of support.
const minSsmsVersion = 21

// Ssms implements the `sqlcmd open ssms` command. It opens
// SQL Server Management Studio and connects to the current context using the
// credentials specified in the context.
func (c *Ssms) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "ssms",
		Short: localizer.Sprintf("Open SQL Server Management Studio and connect to current context"),
		Examples: []cmdparser.ExampleOptions{{
			Description: localizer.Sprintf("Open SSMS and connect using the current context"),
			Steps:       []string{"sqlcmd open ssms"}}},
		Run: c.run,
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.version,
		Name:   "version",
		Usage:  localizer.Sprintf("SSMS major version to launch (for example 21); defaults to the latest installed"),
	})
}

// Launch SSMS and connect to the current context
func (c *Ssms) run() {
	c.validateVersion()

	endpoint, user := config.CurrentContext()

	// Check if this is a local container connection
	isLocalConnection := isLocalEndpoint(endpoint)

	// If the context has a local container, ensure it is running, otherwise bail out
	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		c.ensureContainerIsRunning(asset.Id)
	}

	// Launch SSMS with connection parameters
	c.launchSsms(endpoint.Address, endpoint.Port, user, isLocalConnection)
}

// validateVersion rejects --version values below the supported SSMS floor.
func (c *Ssms) validateVersion() {
	if c.version == "" {
		return
	}
	major, err := strconv.Atoi(c.version)
	if err != nil || major < minSsmsVersion {
		c.Output().FatalWithHintExamples([][]string{
			{localizer.Sprintf("Open the latest SSMS"), "sqlcmd open ssms"},
		}, localizer.Sprintf("'sqlcmd open ssms' supports SSMS %d and later; '--version %s' is not supported", minSsmsVersion, c.version))
	}
}

func (c *Ssms) ensureContainerIsRunning(containerID string) {
	output := c.Output()
	controller := container.NewController()
	if !controller.ContainerRunning(containerID) {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("To start the container"), localizer.Sprintf("sqlcmd start")},
		}, localizer.Sprintf("Container is not running"))
	}
}

// launchSsms launches SQL Server Management Studio using the specified server and user credentials.
func (c *Ssms) launchSsms(host string, port int, user *sqlconfig.User, isLocalConnection bool) {
	output := c.Output()

	// Build server connection string
	serverArg := fmt.Sprintf("%s,%d", host, port)

	args := []string{
		"-S", serverArg,
		"-nosplash",
	}

	// Only add -C (trust server certificate) for local connections with self-signed certs
	if isLocalConnection {
		args = append(args, "-C")
	}

	// Use SQL authentication if configured (commonly used for SQL Server containers)
	if user != nil && user.AuthenticationType == "basic" && user.BasicAuth != nil {
		// Escape double quotes in username (SQL Server allows " in login names)
		username := strings.ReplaceAll(user.BasicAuth.Username, `"`, `\"`)
		args = append(args, "-U", username)
		// Note: -P parameter was removed in SSMS 18+ for security reasons
		// Copy password to clipboard so user can paste it in the login dialog
		copyPasswordToClipboard(user, output)
	}

	t := tools.NewTool("ssms")
	if ssms, ok := t.(*tool.SSMS); ok {
		ssms.SetVersion(c.version)
	}
	if !t.IsInstalled() {
		output.Fatal(t.HowToInstall())
	}

	c.displayPreLaunchInfo()

	if test.IsRunningInTestExecutor() {
		return
	}

	_, err := t.Run(args)
	c.CheckErr(err)
}
