// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"fmt"
	"strconv"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
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

// Launch SSMS and connect to the current context
func (c *Ssms) run() {
	c.validateVersion()

	endpoint, user := config.CurrentContext()
	isLocalConnection := isLocalEndpoint(endpoint)

	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		c.ensureContainerIsRunning(asset.Id)
	}

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

	args := []string{
		"-S", fmt.Sprintf("%s,%d", host, port),
		"-nosplash",
	}

	// -C trusts the self-signed cert that local SQL Server containers ship with.
	if isLocalConnection {
		args = append(args, "-C")
	}

	if user != nil && user.AuthenticationType == "basic" && user.BasicAuth != nil {
		// SSMS removed -P in 18+; hand the password off via the clipboard.
		args = append(args, "-U", user.BasicAuth.Username)
	} else {
		// No SQL credentials in the context, so connect with Windows integrated auth.
		args = append(args, "-E")
	}

	t := tools.NewTool("ssms")
	if ssms, ok := t.(*tool.SSMS); ok {
		ssms.SetVersion(c.version)
	}
	if !t.IsInstalled() {
		output.Fatal(t.HowToInstall())
	}

	// Copy the password only after confirming SSMS is installed; otherwise a
	// fatal install message would leave the password sitting in the clipboard.
	if user != nil && user.AuthenticationType == "basic" && user.BasicAuth != nil {
		copyPasswordToClipboard(user, output)
	}

	c.displayPreLaunchInfo()

	if test.IsRunningInTestExecutor() {
		return
	}

	_, err := t.Run(args)
	c.CheckErr(err)
}
