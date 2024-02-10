// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/tools/tool"
	"strings"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/tools"
)

// Ads implements the `sqlcmd open ads` command. It opens
// Azure Data Studio and connects to the current context by using the
// credentials specified in the context.
func (c *Ssms) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "ssms",
		Short: "Open Sql Server Management Studio and connect to current context",
		Examples: []cmdparser.ExampleOptions{{
			Description: "Open SSMS and connect using the current context",
			Steps:       []string{"sqlcmd open ssms"}}},
		Run: c.run,
	}

	c.Cmd.DefineCommand(options)
}

// Launch ADS and connect to the current context. If the authentication type
// is basic, we need to securely store the password in an Operating System
// specific credential store, e.g. on Windows we use the Windows Credential
// Manager.
func (c *Ssms) run() {
	endpoint, user := config.CurrentContext()

	// If the context has a local container, ensure it is running, otherwise bail out
	if endpoint.AssetDetails != nil && endpoint.AssetDetails.ContainerDetails != nil {
		c.ensureContainerIsRunning(endpoint)
	}

	// If basic auth is used, we need to persist the password in the OS in a way
	// that ADS can access it.  The method used is OS specific.
	if user != nil && user.AuthenticationType == "basic" {
		c.PersistCredentialForAds(endpoint.EndpointDetails.Address, endpoint, user)
		c.launchAds(endpoint.EndpointDetails.Address, endpoint.EndpointDetails.Port, user.BasicAuth.Username)
	} else {
		c.launchAds(endpoint.EndpointDetails.Address, endpoint.EndpointDetails.Port, "")
	}
}

func (c *Ssms) ensureContainerIsRunning(endpoint sqlconfig.Endpoint) {
	output := c.Output()
	controller := container.NewController()
	if !controller.ContainerRunning(endpoint.AssetDetails.ContainerDetails.Id) {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("To start the container"), localizer.Sprintf("sqlcmd start")},
		}, localizer.Sprintf("Container is not running"))
	}
}

// launchAds launches the Azure Data Studio using the specified server and username.
func (c *Ssms) launchAds(host string, port int, username string) {
	output := c.Output()
	args := []string{
		"ssms ",
		"-S",
		fmt.Sprintf("%s,%d", host, port),
	}

	// If a username is specified, use that (basic auth), otherwise use integrated auth
	if username != "" {

		// Here's a fun SQL Server behavior  - it allows you to create database
		// and login names that include the " character. SSMS escapes those
		// with \" when invoking ADS on the command line, we do the same here
		args = append(args, "-U")
		args = append(args, fmt.Sprintf("%s", strings.Replace(username, `"`, `\"`, -1)))
	} else {
		args = append(args, "-E")
	}

	ssms := tools.NewTool("ssms")
	if !ssms.IsInstalled() {
		output.Fatalf(ssms.HowToInstall())
	}

	c.displayPreLaunchInfo()

	_, err := ssms.Run(args, tool.RunOptions{})
	c.CheckErr(err)
}
