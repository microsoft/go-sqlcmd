// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"strconv"

	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/telemetry"
)

// View implements the `sqlcmd config view` command
type View struct {
	cmdparser.Cmd

	raw bool
}

func (c *View) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "view",
		Short: localizer.Sprintf("Display merged sqlconfig settings or a specified sqlconfig file"),
		Examples: []cmdparser.ExampleOptions{
			{
				Description: localizer.Sprintf("Show sqlconfig settings, with REDACTED authentication data"),
				Steps:       []string{"sqlcmd config view"},
			},
			{
				Description: localizer.Sprintf("Show sqlconfig settings and raw authentication data"),
				Steps:       []string{"sqlcmd config view --raw"},
			},
		},
		Aliases: []string{"show"},
		Run:     c.run,
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		Name:  "raw",
		Bool:  &c.raw,
		Usage: localizer.Sprintf("Display raw byte data"),
	})
}

func (c *View) run() {
	output := c.Output()

	contents := config.RedactedConfig(c.raw)
	output.Struct(contents)
	c.LogTelemtry()
}

func (c *View) LogTelemtry() {
	eventName := "config-view"
	properties := map[string]string{}
	properties["Command"] = "Config"
	properties["SubCommand"] = "View"
	properties["Flag"] = strconv.FormatBool(c.raw)
	telemetry.TrackEvent(eventName, properties)
	telemetry.CloseTelemetry()
}
