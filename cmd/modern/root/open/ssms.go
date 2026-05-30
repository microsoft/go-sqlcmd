// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// Type Ssms implements the `sqlcmd open ssms` command. The struct and command
// surface live in an untagged file so that gotext extracts the localizable
// strings on every GOOS; the actual run() body is platform-specific.
type Ssms struct {
	cmdparser.Cmd

	// version pins the SSMS major version to launch (for example "21"). Empty
	// means the latest installed version. Only consumed on Windows; harmless
	// on other platforms because run() fatals before reading it.
	version string
}

// DefineCommand registers `sqlcmd open ssms` with its flags and examples.
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
