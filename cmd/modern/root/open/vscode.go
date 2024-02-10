// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/tools"
	"github.com/microsoft/go-sqlcmd/internal/tools/tool"
)

// Vscode implements the `sqlcmd open vscode` (or just `sqlcmd open code`)
// command. It opens VS Code and connects to the current context by using the
// credentials specified in the context.
func (c *Vscode) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:     "vscode",
		Aliases: []string{"code"},
		Short:   "Open Visual Studio Code and connect to current context",
		Examples: []cmdparser.ExampleOptions{{
			Description: "Open VS Code and connect using the current context",
			Steps:       []string{"sqlcmd open vscode"}}},
		Run: c.run,
	}

	c.Cmd.DefineCommand(options)
}

// Launch ADS and connect to the current context. If the authentication type
// is basic, we need to securely store the password in an Operating System
// specific credential store, e.g. on Windows we use the Windows Credential
// Manager.
func (c *Vscode) run() {
	endpoint, _ := config.CurrentContext()

	// If the context has a local container, ensure it is running, otherwise bail out
	if endpoint.AssetDetails != nil && endpoint.AssetDetails.ContainerDetails != nil {
		c.ensureContainerIsRunning(endpoint)
	}

	c.launchAds(endpoint.EndpointDetails.Address, endpoint.EndpointDetails.Port, "")
}

func (c *Vscode) ensureContainerIsRunning(endpoint sqlconfig.Endpoint) {
	output := c.Output()
	controller := container.NewController()
	if !controller.ContainerRunning(endpoint.AssetDetails.ContainerDetails.Id) {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("To start the container"), localizer.Sprintf("sqlcmd start")},
		}, localizer.Sprintf("Container is not running"))
	}
}

// launchAds launches VS Code
func (c *Vscode) launchAds(host string, port int, username string) {
	output := c.Output()
	args := []string{}

	vscode := tools.NewTool("vscode")
	if !vscode.IsInstalled() {
		output.Fatalf(vscode.HowToInstall())
	}

	c.displayPreLaunchInfo()

	_, err := vscode.Run(args, tool.RunOptions{})
	c.CheckErr(err)
}
