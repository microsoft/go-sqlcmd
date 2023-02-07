// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/tools"
)

// Ads implements the `sqlcmd open ads` command. It opens
// Azure Data Studio and connects to the current context by using the
// credentials specified in the context.
func (c *Ads) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "ads",
		Short: "Open Azure Data Studio and connect to current context",
		Examples: []cmdparser.ExampleOptions{{
			Description: "Open ADS and connect using the current context",
			Steps:       []string{"sqlcmd open ads"}}},
		Run: c.run,
	}

	c.Cmd.DefineCommand(options)
}

// Launch ADS and connect to the current context. If the authentication type
// is basic, we need to securely store the password in an Operating System
// specific credential store, e.g. on Windows we use the Windows Credential
// Manager.
func (c *Ads) run() {
	endpoint, user := config.CurrentContext()

	// If the context has a local container, ensure it is running, otherwise bail out
	if endpoint.AssetDetails != nil && endpoint.AssetDetails.ContainerDetails != nil {
		c.ensureContainerIsRunning(endpoint)
	}

	hostname := endpoint.EndpointDetails.Address

	// If the hostname is localhost, ADS will not connect to the container
	// because it will try to connect to the host machine. To work around
	// BUG(stuartpa): I think this might be a bug in ADS?
	if hostname == "localhost" {
		hostname = "127.0.0.1"
	}

	// If basic auth is used, we need to persist the password in the OS in a way
	// that ADS can access it.  The method used is OS specific.
	if user != nil && user.AuthenticationType == "basic" {
		c.persistCredentialForAds(hostname, endpoint, user)
		c.launchAds(hostname, endpoint.EndpointDetails.Port, user.BasicAuth.Username)
	} else {
		c.launchAds(hostname, endpoint.EndpointDetails.Port, "")
	}
}

func (c *Ads) ensureContainerIsRunning(endpoint sqlconfig.Endpoint) {
	output := c.Output()
	controller := container.NewController()
	if !controller.ContainerRunning(endpoint.AssetDetails.ContainerDetails.Id) {
		output.FatalfWithHintExamples([][]string{
			{"To start the container", "sqlcmd start"},
		}, "Container is not running")
	}
}

// launchAds launches the Azure Data Studio using the specified server and username.
func (c *Ads) launchAds(localhost string, port int, username string) {
	output := c.Output()
	args := []string{
		"-r",
		fmt.Sprintf(
			"--server=%s", fmt.Sprintf(
				"%s,%d",
				localhost,
				port)),
	}

	if username != "" {
		args = append(args, fmt.Sprintf("--user=%s", username))
	}

	tool := tools.NewTool("ads")
	if !tool.IsInstalled() {
		output.Fatalf(tool.HowToInstall())
	} else {
		_, err := tool.Run(args)
		c.CheckErr(err)
	}
}
