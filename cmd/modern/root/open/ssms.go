//go:build windows

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"fmt"
	"net/url"
	"os/exec"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/tools"
)

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
}

// run launches SSMS via the ssms:// URL handler with connection parameters
// from the current context.
func (c *Ssms) run() {
	endpoint, user := config.CurrentContext()

	// If the context has a local container, ensure it is running, otherwise bail out
	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		c.ensureContainerIsRunning(asset.Id)
	}

	c.launchSsms(endpoint.Address, endpoint.Port, user)
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

// launchSsms hands off the connection to SSMS through the ssms:// URL handler.
// The handler resolves to Microsoft.VisualStudio.SSMSProtocolSelector.exe
// (SSMS 21+) or Ssms.exe (legacy MSI installs) and accepts short-form keys
// s, d, u, a, p observed in the SQL database in Fabric "Open in SSMS" link.
func (c *Ssms) launchSsms(host string, port int, user *sqlconfig.User) {
	output := c.Output()

	tool := tools.NewTool("ssms")
	if !tool.IsInstalled() {
		output.Fatal(tool.HowToInstall())
	}

	var password string
	if user != nil && user.AuthenticationType == "basic" && user.BasicAuth != nil {
		_, _, password = config.GetCurrentContextInfo()
	}
	ssmsURL := buildSsmsURL(host, port, user, password)

	output.Info(localizer.Sprintf("Launching SQL Server Management Studio..."))

	// cmd /c start "" "<url>" routes the URL through ShellExecute,
	// which uses the HKCR\ssms\shell\open\command registration.
	cmd := exec.Command("cmd", "/c", "start", "", ssmsURL)
	err := cmd.Start()
	c.CheckErr(err)
}

// buildSsmsURL constructs an ssms://connect URL for the supplied connection.
// The grammar follows the Fabric "Open in SSMS" link format:
//
//	ssms://connect?s=<server,port>&u=<user>&a=<auth>&p=<password>
//
// Database (d) and other parameters are omitted because the current sqlcmd
// context does not carry a database name.
func buildSsmsURL(host string, port int, user *sqlconfig.User, password string) string {
	q := url.Values{}
	q.Set("s", fmt.Sprintf("%s,%d", host, port))

	if user != nil && user.AuthenticationType == "basic" && user.BasicAuth != nil {
		q.Set("u", user.BasicAuth.Username)
		q.Set("a", "SqlLogin")
		if password != "" {
			q.Set("p", password)
		}
	}

	return "ssms://connect?" + q.Encode()
}
