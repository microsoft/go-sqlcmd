// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build windows

package open

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/tools"
)

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

func (c *Ssms) run() {
	endpoint, user := config.CurrentContext()
	isLocalConnection := isLocalEndpoint(endpoint)

	if asset := endpoint.AssetDetails; asset != nil && asset.ContainerDetails != nil {
		c.ensureContainerIsRunning(asset.Id)
	}

	c.launchSsms(endpoint.Address, endpoint.Port, user, isLocalConnection)
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

func (c *Ssms) launchSsms(host string, port int, user *sqlconfig.User, isLocalConnection bool) {
	output := c.Output()

	if user != nil && user.AuthenticationType == "basic" && user.BasicAuth != nil {
		copyPasswordToClipboard(user, output)
	}

	c.displayPreLaunchInfo()

	serverArg := fmt.Sprintf("%s,%d", host, port)

	args := []string{
		"-S", serverArg,
		"-nosplash",
	}

	if db := os.Getenv("SQLCMDDATABASE"); db != "" {
		args = append(args, "-d", db)
	}

	if user != nil && user.AuthenticationType == "basic" && user.BasicAuth != nil {
		username := strings.ReplaceAll(user.BasicAuth.Username, `"`, `\"`)
		args = append(args, "-U", username)
	}

	tool := tools.NewTool("ssms")
	if !tool.IsInstalled() {
		output.Fatal(tool.HowToInstall())
	}

	// -C (trust server certificate) for self-signed certs on local containers.
	// Only supported by SSMS 21+.
	if isLocalConnection && ssmsVersion(tool.ExePath()) >= 21 {
		args = append(args, "-C")
	}

	_, err := tool.Run(args)
	c.CheckErr(err)
}

var ssmsVersionRe = regexp.MustCompile(`Management Studio (\d+)`)

// ssmsVersion returns 0 if the version cannot be determined from the path.
func ssmsVersion(exePath string) int {
	m := ssmsVersionRe.FindStringSubmatch(exePath)
	if len(m) < 2 {
		return 0
	}
	v, _ := strconv.Atoi(m[1])
	return v
}
