// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"fmt"
	"runtime"
	"strings"

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
		Short: "Open SQL Server Management Studio and connect to current context",
		Examples: []cmdparser.ExampleOptions{{
			Description: "Open SSMS and connect using the current context",
			Steps:       []string{"sqlcmd open ssms"}}},
		Run: c.run,
	}

	c.Cmd.DefineCommand(options)
}

// Launch SSMS and connect to the current context
func (c *Ssms) run() {
	endpoint, user := config.CurrentContext()

	// If the context has a local container, ensure it is running, otherwise bail out
	if endpoint.AssetDetails != nil && endpoint.AssetDetails.ContainerDetails != nil {
		c.ensureContainerIsRunning(endpoint)
	}

	// Launch SSMS with connection parameters
	c.launchSsms(endpoint.EndpointDetails.Address, endpoint.EndpointDetails.Port, user)
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

// launchSsms launches SQL Server Management Studio using the specified server and user credentials.
func (c *Ssms) launchSsms(host string, port int, user *sqlconfig.User) {
	output := c.Output()
	
	// Build server connection string
	serverArg := fmt.Sprintf("%s,%d", host, port)
	
	args := []string{
		"-S", serverArg,
		"-nosplash",
	}

	// Add authentication parameters
	if user != nil && user.AuthenticationType == "basic" {
		// SQL Server authentication
		// Escape double quotes in username (SQL Server allows " in login names)
		username := strings.Replace(user.BasicAuth.Username, `"`, `\"`, -1)
		args = append(args, "-U", username)
		// Note: -P parameter was removed in SSMS 18+ for security reasons
		// User will need to enter password in the login dialog
		output.Info(localizer.Sprintf("Note: You will need to enter the password in the SSMS login dialog"))
	} else {
		// Windows integrated authentication
		if runtime.GOOS == "windows" {
			args = append(args, "-E")
		}
	}

	tool := tools.NewTool("ssms")
	if !tool.IsInstalled() {
		output.Fatal(tool.HowToInstall())
	}

	c.displayPreLaunchInfo()

	_, err := tool.Run(args)
	c.CheckErr(err)
}
